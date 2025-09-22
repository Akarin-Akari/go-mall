package cache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"mall-go/pkg/optimistic"
)

// 创建测试用的缓存防护管理器
func createTestProtectionManager() (*CacheProtectionManager, *SharedMockCacheManager, *SharedMockOptimisticLockService) {
	mockCache := &SharedMockCacheManager{}
	mockOptimistic := &SharedMockOptimisticLockService{}

	// 初始化键管理器
	InitKeyManager("test")
	keyManager := GetKeyManager()

	// 创建一致性管理器
	consistencyConfig := DefaultCacheConsistencyConfig()
	consistencyMgr := NewCacheConsistencyManager(consistencyConfig, mockCache, keyManager, mockOptimistic)

	// 创建预热管理器
	warmupConfig := DefaultCacheWarmupConfig()
	warmupMgr := NewCacheWarmupManager(warmupConfig, mockCache, keyManager, consistencyMgr, mockOptimistic)

	// 创建防护配置
	config := DefaultCacheProtectionConfig()
	config.Strategies = []ProtectionStrategy{
		ProtectionNullCache,
		ProtectionRandomTTL,
	} // 移除需要Redis的策略用于测试

	cpm := NewCacheProtectionManager(config, mockCache, keyManager, consistencyMgr, warmupMgr, mockOptimistic)

	return cpm, mockCache, mockOptimistic
}

// TestCacheProtectionManagerCreation 测试缓存防护管理器创建
func TestCacheProtectionManagerCreation(t *testing.T) {
	cpm, _, _ := createTestProtectionManager()

	assert.NotNil(t, cpm)
	assert.NotNil(t, cpm.config)
	assert.NotNil(t, cpm.cacheManager)
	assert.NotNil(t, cpm.keyManager)
	assert.NotNil(t, cpm.consistencyMgr)
	assert.NotNil(t, cpm.warmupMgr)
	assert.NotNil(t, cpm.metrics)
	assert.False(t, cpm.IsRunning())
}

// TestCacheProtectionManagerStartStop 测试启动和停止
func TestCacheProtectionManagerStartStop(t *testing.T) {
	cpm, _, _ := createTestProtectionManager()

	// 测试启动
	err := cpm.Start()
	assert.NoError(t, err)
	assert.True(t, cpm.IsRunning())

	// 测试重复启动
	err = cpm.Start()
	assert.Error(t, err)

	// 等待一小段时间让后台任务启动
	time.Sleep(50 * time.Millisecond)

	// 测试停止
	err = cpm.Stop()
	assert.NoError(t, err)
	assert.False(t, cpm.IsRunning())

	// 测试重复停止
	err = cpm.Stop()
	assert.NoError(t, err)
}

// TestCacheProtectionManagerConfig 测试配置管理
func TestCacheProtectionManagerConfig(t *testing.T) {
	cpm, _, _ := createTestProtectionManager()

	config := cpm.GetConfig()
	assert.NotNil(t, config)
	assert.True(t, config.Enabled)
	assert.Equal(t, ProtectionLevelStandard, config.Level)
	assert.Contains(t, config.Strategies, ProtectionNullCache)
	assert.Contains(t, config.Strategies, ProtectionRandomTTL)
}

// TestCacheProtectionManagerNullCache 测试空值缓存
func TestCacheProtectionManagerNullCache(t *testing.T) {
	cpm, _, _ := createTestProtectionManager()

	testKey := "test:null:key"

	// 初始状态不应该有空值缓存
	assert.False(t, cpm.isNullCached(testKey))

	// 缓存空值
	cpm.cacheNullValue(testKey)

	// 现在应该有空值缓存
	assert.True(t, cpm.isNullCached(testKey))

	// 测试过期
	cpm.config.NullCacheConfig.TTL = 1 * time.Millisecond
	time.Sleep(2 * time.Millisecond)
	assert.False(t, cpm.isNullCached(testKey))
}

// TestCacheProtectionManagerTTLCalculation 测试TTL计算
func TestCacheProtectionManagerTTLCalculation(t *testing.T) {
	cpm, _, _ := createTestProtectionManager()

	testKey := "test:ttl:key"

	// 测试随机TTL策略
	ttl1 := cpm.calculateTTL(testKey)
	ttl2 := cpm.calculateTTL(testKey)

	// 同一个键应该得到相同的TTL（基于哈希）
	assert.Equal(t, ttl1, ttl2)

	// TTL应该在配置的范围内
	config := cpm.config.RandomTTLConfig
	assert.True(t, ttl1 >= config.MinTTL)
	assert.True(t, ttl1 <= config.MaxTTL)
}

// TestCacheProtectionManagerMetrics 测试指标管理
func TestCacheProtectionManagerMetrics(t *testing.T) {
	cpm, _, _ := createTestProtectionManager()

	// 获取初始指标
	metrics := cpm.GetMetrics()
	assert.NotNil(t, metrics)
	assert.Equal(t, int64(0), metrics.TotalRequests)
	assert.Equal(t, int64(0), metrics.ProtectedRequests)

	// 增加指标
	cpm.incrementMetric("total_requests")
	cpm.incrementMetric("protected_requests")
	cpm.incrementMetric("penetration_blocked")

	// 检查指标更新
	metrics = cpm.GetMetrics()
	assert.Equal(t, int64(1), metrics.TotalRequests)
	assert.Equal(t, int64(1), metrics.ProtectedRequests)
	assert.Equal(t, int64(1), metrics.PenetrationBlocked)
	assert.Equal(t, float64(100), metrics.ProtectionRate)

	// 测试指标重置
	cpm.ResetMetrics()
	metrics = cpm.GetMetrics()
	assert.Equal(t, int64(0), metrics.TotalRequests)
	assert.Equal(t, int64(0), metrics.ProtectedRequests)
}

// TestCacheProtectionManagerResponseTimeMetrics 测试响应时间指标
func TestCacheProtectionManagerResponseTimeMetrics(t *testing.T) {
	cpm, _, _ := createTestProtectionManager()

	// 更新响应时间指标
	cpm.updateMetrics(100 * time.Millisecond)
	cpm.updateMetrics(200 * time.Millisecond)
	cpm.updateMetrics(50 * time.Millisecond)

	metrics := cpm.GetMetrics()
	assert.Equal(t, 50*time.Millisecond, metrics.MinResponseTime)
	assert.Equal(t, 200*time.Millisecond, metrics.MaxResponseTime)
	assert.True(t, metrics.AverageResponseTime > 0)
}

// TestCacheProtectionManagerProtectedGet 测试受保护的缓存获取
func TestCacheProtectionManagerProtectedGet(t *testing.T) {
	cpm, mockCache, _ := createTestProtectionManager()

	testKey := "test:protected:key"
	testValue := "test_value"

	// 设置mock期望
	mockCache.On("Get", testKey).Return(nil, nil).Once() // 缓存未命中
	mockCache.On("Set", testKey, testValue, mock.AnythingOfType("time.Duration")).Return(nil).Once()

	// 定义数据加载器
	loader := func() (interface{}, error) {
		return testValue, nil
	}

	// 测试受保护的获取
	result, err := cpm.ProtectedGet(testKey, loader)
	assert.NoError(t, err)
	assert.Equal(t, testValue, result)

	// 验证mock调用
	mockCache.AssertExpectations(t)
}

// TestCacheProtectionManagerProtectedGetWithNullCache 测试空值缓存防护
func TestCacheProtectionManagerProtectedGetWithNullCache(t *testing.T) {
	cpm, mockCache, _ := createTestProtectionManager()

	testKey := "test:null:protected:key"

	// 先缓存空值
	cpm.cacheNullValue(testKey)

	// 设置mock期望
	mockCache.On("Get", testKey).Return(nil, nil).Once() // 缓存未命中

	// 定义数据加载器
	loader := func() (interface{}, error) {
		return "should_not_be_called", nil
	}

	// 测试受保护的获取（应该被空值缓存阻止）
	result, err := cpm.ProtectedGet(testKey, loader)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "数据不存在（空值缓存）")

	// 验证mock调用
	mockCache.AssertExpectations(t)
}

// TestCacheProtectionManagerHashString 测试字符串哈希
func TestCacheProtectionManagerHashString(t *testing.T) {
	cpm, _, _ := createTestProtectionManager()

	testString := "test_string"

	// 测试哈希一致性
	hash1 := cpm.hashString(testString)
	hash2 := cpm.hashString(testString)
	assert.Equal(t, hash1, hash2)

	// 测试不同字符串产生不同哈希
	hash3 := cpm.hashString("different_string")
	assert.NotEqual(t, hash1, hash3)
}

// TestCacheProtectionManagerStrategyEnabled 测试策略启用检查
func TestCacheProtectionManagerStrategyEnabled(t *testing.T) {
	cpm, _, _ := createTestProtectionManager()

	// 测试启用的策略
	assert.True(t, cpm.isStrategyEnabled(ProtectionNullCache))
	assert.True(t, cpm.isStrategyEnabled(ProtectionRandomTTL))

	// 测试未启用的策略
	assert.False(t, cpm.isStrategyEnabled(ProtectionBloomFilter))
	assert.False(t, cpm.isStrategyEnabled(ProtectionDistributedLock))
}

// TestCacheProtectionManagerCleanupNullCache 测试空值缓存清理
func TestCacheProtectionManagerCleanupNullCache(t *testing.T) {
	cpm, _, _ := createTestProtectionManager()

	// 设置短TTL用于测试
	cpm.config.NullCacheConfig.TTL = 1 * time.Millisecond

	testKey := "test:cleanup:key"

	// 缓存空值
	cpm.cacheNullValue(testKey)
	assert.True(t, cpm.isNullCached(testKey))

	// 等待过期
	time.Sleep(2 * time.Millisecond)

	// 执行清理
	cpm.cleanupNullCache()

	// 验证已清理
	assert.False(t, cpm.isNullCached(testKey))
}

// TestCacheProtectionManagerGenerateLockValue 测试锁值生成
func TestCacheProtectionManagerGenerateLockValue(t *testing.T) {
	cpm, _, _ := createTestProtectionManager()

	// 生成多个锁值
	value1 := cpm.generateLockValue()
	value2 := cpm.generateLockValue()

	// 验证锁值不为空且不相同
	assert.NotEmpty(t, value1)
	assert.NotEmpty(t, value2)
	assert.NotEqual(t, value1, value2)

	// 验证锁值长度（16字节的十六进制表示应该是32个字符）
	assert.Equal(t, 32, len(value1))
	assert.Equal(t, 32, len(value2))
}
