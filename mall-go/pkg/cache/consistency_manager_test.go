package cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"mall-go/pkg/optimistic"
)

// 创建测试用的缓存一致性管理器
func createTestConsistencyManager() (*CacheConsistencyManager, *SharedMockCacheManager, *SharedMockOptimisticLockService) {
	mockCache := &SharedMockCacheManager{}
	mockOptimistic := &SharedMockOptimisticLockService{}

	// 初始化键管理器
	InitKeyManager("test")
	keyManager := GetKeyManager()

	config := DefaultCacheConsistencyConfig()
	config.EventWorkers = 1                       // 减少工作者数量用于测试
	config.CheckInterval = 100 * time.Millisecond // 缩短检查间隔

	ccm := NewCacheConsistencyManager(config, mockCache, keyManager, mockOptimistic)

	return ccm, mockCache, mockOptimistic
}

// TestCacheConsistencyManager_Creation 测试缓存一致性管理器创建
func TestCacheConsistencyManager_Creation(t *testing.T) {
	ccm, _, _ := createTestConsistencyManager()

	assert.NotNil(t, ccm)
	assert.NotNil(t, ccm.config)
	assert.NotNil(t, ccm.cacheManager)
	assert.NotNil(t, ccm.keyManager)
	assert.NotNil(t, ccm.optimisticLock)
	assert.NotNil(t, ccm.stats)
	assert.False(t, ccm.IsRunning())
}

// TestCacheConsistencyManager_StartStop 测试启动和停止
func TestCacheConsistencyManager_StartStop(t *testing.T) {
	ccm, _, _ := createTestConsistencyManager()

	// 测试启动
	err := ccm.Start()
	assert.NoError(t, err)
	assert.True(t, ccm.IsRunning())

	// 测试重复启动
	err = ccm.Start()
	assert.Error(t, err)

	// 等待一小段时间让工作者启动
	time.Sleep(50 * time.Millisecond)

	// 测试停止
	err = ccm.Stop()
	assert.NoError(t, err)
	assert.False(t, ccm.IsRunning())

	// 测试重复停止
	err = ccm.Stop()
	assert.NoError(t, err)
}

// TestCacheConsistencyManager_PublishEvent 测试发布事件
func TestCacheConsistencyManager_PublishEvent(t *testing.T) {
	ccm, mockCache, _ := createTestConsistencyManager()

	// 启动管理器
	err := ccm.Start()
	assert.NoError(t, err)
	defer ccm.Stop()

	// 创建测试事件
	event := &CacheUpdateEvent{
		Type:      "update",
		TableName: "products",
		RecordID:  1,
		Data: map[string]interface{}{
			"id":      1,
			"name":    "Test Product",
			"version": 2,
		},
		CacheKeys: []string{"test:product:1"},
	}

	// 设置mock期望
	mockCache.On("Set", mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("time.Duration")).Return(nil)

	// 发布事件
	err = ccm.PublishEvent(event)
	assert.NoError(t, err)

	// 等待事件处理
	time.Sleep(100 * time.Millisecond)

	// 验证统计信息
	stats := ccm.GetStats()
	assert.Equal(t, int64(1), stats.TotalEvents)
}

// TestCacheConsistencyManager_CheckConsistency 测试一致性检查
func TestCacheConsistencyManager_CheckConsistency(t *testing.T) {
	ccm, mockCache, _ := createTestConsistencyManager()

	t.Run("缓存数据一致", func(t *testing.T) {
		// 设置mock返回
		cacheData := `{"id":1,"name":"Test Product","version":1}`
		mockCache.On("Get", "test:product:1").Return(cacheData, nil)

		// 执行一致性检查
		result, err := ccm.CheckConsistency("test:product:1", "products", 1)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.IsConsistent)
		assert.Equal(t, 1, result.CacheVersion)
		assert.Equal(t, 1, result.DatabaseVersion)

		mockCache.AssertExpectations(t)
	})

	t.Run("缓存未命中", func(t *testing.T) {
		mockCache.On("Get", "test:product:2").Return(nil, nil)

		result, err := ccm.CheckConsistency("test:product:2", "products", 2)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.IsConsistent) // 缓存和数据库都为空，认为一致

		mockCache.AssertExpectations(t)
	})
}

// TestCacheConsistencyManager_SyncStrategies 测试同步策略
func TestCacheConsistencyManager_SyncStrategies(t *testing.T) {
	ccm, mockCache, mockOptimistic := createTestConsistencyManager()

	testData := map[string]interface{}{
		"id":      1,
		"name":    "Test Product",
		"version": 2,
	}

	t.Run("写穿透策略", func(t *testing.T) {
		ccm.config.Strategy = WriteThrough

		// 设置mock期望
		mockOptimistic.On("UpdateWithOptimisticLock", "products", uint(1), mock.AnythingOfType("map[string]interface {}"), mock.Anything).Return(&optimistic.UpdateResult{
			Success:      true,
			RowsAffected: 1,
		})
		mockCache.On("Set", "test:product:1", mock.AnythingOfType("string"), mock.AnythingOfType("time.Duration")).Return(nil)

		err := ccm.SyncCache("test:product:1", "products", 1, testData)
		assert.NoError(t, err)

		mockOptimistic.AssertExpectations(t)
		mockCache.AssertExpectations(t)
	})

	t.Run("写回策略", func(t *testing.T) {
		ccm.config.Strategy = WriteBehind

		// 设置mock期望
		mockCache.On("Set", "test:product:1", mock.AnythingOfType("string"), mock.AnythingOfType("time.Duration")).Return(nil)

		err := ccm.SyncCache("test:product:1", "products", 1, testData)
		assert.NoError(t, err)

		mockCache.AssertExpectations(t)
	})

	t.Run("缓存旁路策略", func(t *testing.T) {
		ccm.config.Strategy = CacheAside

		// 设置mock期望
		mockCache.On("Delete", "test:product:1").Return(nil)

		err := ccm.SyncCache("test:product:1", "products", 1, testData)
		assert.NoError(t, err)

		mockCache.AssertExpectations(t)
	})

	t.Run("提前刷新策略", func(t *testing.T) {
		ccm.config.Strategy = RefreshAhead

		// 设置mock期望
		mockCache.On("Set", "test:product:1", mock.AnythingOfType("string"), mock.AnythingOfType("time.Duration")).Return(nil)

		err := ccm.SyncCache("test:product:1", "products", 1, testData)
		assert.NoError(t, err)

		mockCache.AssertExpectations(t)
	})
}

// TestCacheConsistencyManager_InvalidateCache 测试缓存失效
func TestCacheConsistencyManager_InvalidateCache(t *testing.T) {
	ccm, mockCache, _ := createTestConsistencyManager()

	t.Run("批量失效缓存", func(t *testing.T) {
		cacheKeys := []string{"test:product:1", "test:product:2", "test:product:3"}

		// 设置mock期望
		mockCache.On("MDelete", cacheKeys).Return(nil)

		err := ccm.InvalidateCache(cacheKeys)
		assert.NoError(t, err)

		mockCache.AssertExpectations(t)
	})

	t.Run("空键列表", func(t *testing.T) {
		err := ccm.InvalidateCache([]string{})
		assert.NoError(t, err)
	})
}

// TestCacheConsistencyManager_Stats 测试统计功能
func TestCacheConsistencyManager_Stats(t *testing.T) {
	ccm, mockCache, _ := createTestConsistencyManager()

	// 获取初始统计
	stats := ccm.GetStats()
	assert.NotNil(t, stats)
	assert.Equal(t, int64(0), stats.TotalChecks)
	assert.Equal(t, int64(0), stats.TotalSyncs)

	// 执行一些操作来更新统计
	mockCache.On("Get", "test:product:1").Return(`{"id":1,"version":1}`, nil)

	_, err := ccm.CheckConsistency("test:product:1", "products", 1)
	assert.NoError(t, err)

	// 检查统计更新
	stats = ccm.GetStats()
	assert.Equal(t, int64(1), stats.TotalChecks)
	assert.Equal(t, int64(1), stats.ConsistentCount)
	assert.Equal(t, float64(100), stats.ConsistencyRate)

	// 重置统计
	ccm.ResetStats()
	stats = ccm.GetStats()
	assert.Equal(t, int64(0), stats.TotalChecks)
	assert.Equal(t, int64(0), stats.ConsistentCount)

	mockCache.AssertExpectations(t)
}
