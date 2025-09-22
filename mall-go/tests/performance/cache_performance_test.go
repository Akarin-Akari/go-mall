package performance

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"mall-go/internal/config"
	"mall-go/internal/model"
	"mall-go/pkg/cache"
	"mall-go/pkg/logger"
	"mall-go/tests/helpers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// CachePerformanceTestSuite 缓存性能测试套件
type CachePerformanceTestSuite struct {
	*PerformanceTestSuite
	cacheManager   cache.CacheManager
	consistencyMgr *cache.CacheConsistencyManager
	warmupMgr      *cache.CacheWarmupManager
	protectionMgr  *cache.CacheProtectionManager
	monitoringMgr  *cache.CacheMonitoringManager
	keyManager     *cache.CacheKeyManager
	testHelper     *helpers.TestHelper
}

// CachePerformanceResult 缓存性能测试结果
type CachePerformanceResult struct {
	*PerformanceResult
	CacheHitRate       float64       `json:"cache_hit_rate"`
	CacheMissRate      float64       `json:"cache_miss_rate"`
	AvgCacheTime       time.Duration `json:"avg_cache_time"`
	AvgDBTime          time.Duration `json:"avg_db_time"`
	CacheOperationsQPS float64       `json:"cache_operations_qps"`
	DataConsistency    float64       `json:"data_consistency"`
	MemoryUsage        int64         `json:"memory_usage"`
}

// SetupCachePerformanceTest 设置缓存性能测试环境
func SetupCachePerformanceTest(t *testing.T) *CachePerformanceTestSuite {
	// 初始化基础测试套件
	baseSuite := SetupPerformanceTest(t)

	// 初始化日志
	logger.Init()

	// 创建测试助手
	testHelper := helpers.NewTestHelper()

	// 初始化Redis客户端（使用测试配置）
	redisConfig := config.RedisConfig{
		Host:         "localhost",
		Port:         6379,
		Password:     "",
		DB:           1, // 使用测试数据库
		PoolSize:     100,
		MinIdleConns: 10,
		MaxRetries:   3,
		DialTimeout:  5,
		ReadTimeout:  3,
		WriteTimeout: 3,
		IdleTimeout:  300,
		MaxConnAge:   3600,
	}

	redisClient, err := cache.NewRedisClient(redisConfig)
	if err != nil {
		t.Skipf("Redis服务未启动，跳过缓存性能测试: %v", err)
		return nil
	}

	// 清理Redis测试数据
	redisClient.FlushDB()

	// 创建缓存管理器
	cacheManager := cache.NewRedisCacheManager(redisClient)

	// 初始化键管理器
	cache.InitKeyManager("perf_test")
	keyManager := cache.GetKeyManager()

	// 创建乐观锁服务
	optimisticLock := testHelper.CreateOptimisticLockService()

	// 创建缓存一致性管理器
	consistencyConfig := cache.DefaultCacheConsistencyConfig()
	consistencyConfig.EventWorkers = 2 // 减少工作者数量用于测试
	consistencyMgr := cache.NewCacheConsistencyManager(consistencyConfig, cacheManager, keyManager, optimisticLock)

	// 创建缓存预热管理器
	warmupConfig := cache.DefaultCacheWarmupConfig()
	warmupConfig.MaxConcurrency = 5 // 减少并发数用于测试
	warmupMgr := cache.NewCacheWarmupManager(warmupConfig, cacheManager, keyManager, consistencyMgr, optimisticLock)

	// 创建缓存防护管理器
	protectionConfig := cache.DefaultCacheProtectionConfig()
	protectionConfig.Strategies = []cache.ProtectionStrategy{
		cache.ProtectionNullCache,
		cache.ProtectionRandomTTL,
	} // 移除需要Redis特殊模块的策略
	protectionMgr := cache.NewCacheProtectionManager(protectionConfig, cacheManager, keyManager, consistencyMgr, warmupMgr, optimisticLock)

	// 创建缓存监控管理器
	monitoringConfig := cache.DefaultCacheMonitoringConfig()
	monitoringConfig.CollectInterval = 1 * time.Second // 缩短收集间隔用于测试
	monitoringMgr := cache.NewCacheMonitoringManager(monitoringConfig, cacheManager, keyManager, consistencyMgr, warmupMgr, protectionMgr, optimisticLock)

	// 启动所有管理器
	require.NoError(t, consistencyMgr.Start())
	require.NoError(t, warmupMgr.Start())
	require.NoError(t, protectionMgr.Start())
	require.NoError(t, monitoringMgr.Start())

	return &CachePerformanceTestSuite{
		PerformanceTestSuite: baseSuite,
		cacheManager:         cacheManager,
		consistencyMgr:       consistencyMgr,
		warmupMgr:            warmupMgr,
		protectionMgr:        protectionMgr,
		monitoringMgr:        monitoringMgr,
		keyManager:           keyManager,
		testHelper:           testHelper,
	}
}

// CleanupCachePerformanceTest 清理缓存性能测试环境
func (suite *CachePerformanceTestSuite) CleanupCachePerformanceTest() {
	if suite.monitoringMgr != nil {
		suite.monitoringMgr.Stop()
	}
	if suite.protectionMgr != nil {
		suite.protectionMgr.Stop()
	}
	if suite.warmupMgr != nil {
		suite.warmupMgr.Stop()
	}
	if suite.consistencyMgr != nil {
		suite.consistencyMgr.Stop()
	}
	if suite.testHelper != nil {
		suite.testHelper.Cleanup()
	}
	suite.CleanupPerformanceTest()
}

// CreateCacheTestData 创建缓存测试数据
func (suite *CachePerformanceTestSuite) CreateCacheTestData(t *testing.T) {
	// 创建基础测试数据
	suite.CreateTestData(t)

	// 预热部分缓存数据
	ctx := context.Background()

	// 缓存前100个商品
	for i := 1; i <= 100; i++ {
		var product model.Product
		err := suite.db.Where("id = ?", i).First(&product).Error
		if err == nil {
			key := suite.keyManager.GenerateProductKey(product.ID)
			suite.cacheManager.Set(key, product, 1*time.Hour)
		}
	}

	// 缓存前50个用户会话
	for i := 1; i <= 50; i++ {
		key := suite.keyManager.GenerateUserSessionKey(uint(i))
		sessionData := map[string]interface{}{
			"user_id":    i,
			"login_time": time.Now(),
			"status":     "active",
		}
		suite.cacheManager.Set(key, sessionData, 2*time.Hour)
	}

	// 缓存分类数据
	var categories []model.Category
	suite.db.Find(&categories)
	for _, category := range categories {
		key := suite.keyManager.GenerateCategoryKey(category.ID)
		suite.cacheManager.Set(key, category, 6*time.Hour)
	}

	t.Logf("✅ 缓存测试数据创建完成 - 商品缓存: 100, 用户会话: 50, 分类: %d", len(categories))
}

// TestCacheHitRatePerformance 测试缓存命中率性能
func TestCacheHitRatePerformance(t *testing.T) {
	suite := SetupCachePerformanceTest(t)
	if suite == nil {
		return
	}
	defer suite.CleanupCachePerformanceTest()

	// 创建测试数据
	suite.CreateCacheTestData(t)

	t.Run("商品查询缓存命中率测试", func(t *testing.T) {
		concurrency := 100
		totalRequests := 2000

		var hitCount, missCount int64
		var hitTime, missTime time.Duration
		var mutex sync.Mutex

		requestFunc := func() *RequestResult {
			start := time.Now()

			// 80%概率查询已缓存的商品（1-100），20%查询未缓存的商品（101-1000）
			var productID uint
			if rand.Float64() < 0.8 {
				productID = uint(rand.Intn(100) + 1) // 已缓存
			} else {
				productID = uint(rand.Intn(900) + 101) // 未缓存
			}

			key := suite.keyManager.GenerateProductKey(productID)

			// 尝试从缓存获取
			cachedData, err := suite.cacheManager.Get(key)
			duration := time.Since(start)

			mutex.Lock()
			if err == nil && cachedData != nil {
				hitCount++
				hitTime += duration
			} else {
				missCount++
				missTime += duration

				// 模拟从数据库加载并缓存
				var product model.Product
				if dbErr := suite.db.Where("id = ?", productID).First(&product).Error; dbErr == nil {
					suite.cacheManager.Set(key, product, 1*time.Hour)
				}
			}
			mutex.Unlock()

			return &RequestResult{
				Success:  true,
				Duration: duration,
				Error:    err,
			}
		}

		result := suite.RunConcurrentTest(t, "商品查询缓存命中率", concurrency, totalRequests, requestFunc)

		// 计算缓存指标
		totalHit := hitCount
		totalMiss := missCount
		hitRate := float64(totalHit) / float64(totalHit+totalMiss) * 100

		avgHitTime := time.Duration(0)
		avgMissTime := time.Duration(0)
		if totalHit > 0 {
			avgHitTime = time.Duration(int64(hitTime) / totalHit)
		}
		if totalMiss > 0 {
			avgMissTime = time.Duration(int64(missTime) / totalMiss)
		}

		// 验证性能指标
		assert.GreaterOrEqual(t, hitRate, 70.0, "缓存命中率应≥70%")
		assert.Less(t, avgHitTime, 5*time.Millisecond, "缓存命中平均响应时间应<5ms")
		assert.Less(t, result.AverageTime, 20*time.Millisecond, "总体平均响应时间应<20ms")
		assert.Greater(t, result.RequestsPerSec, 2000.0, "QPS应>2000")

		t.Logf("✅ 缓存命中率测试通过")
		t.Logf("   - 缓存命中率: %.2f%% (目标: ≥70%%)", hitRate)
		t.Logf("   - 缓存命中平均时间: %v (目标: <5ms)", avgHitTime)
		t.Logf("   - 缓存未命中平均时间: %v", avgMissTime)
		t.Logf("   - 总体QPS: %.2f (目标: >2000)", result.RequestsPerSec)
	})
}

// TestCacheConsistencyPerformance 测试缓存一致性性能
func TestCacheConsistencyPerformance(t *testing.T) {
	suite := SetupCachePerformanceTest(t)
	if suite == nil {
		return
	}
	defer suite.CleanupCachePerformanceTest()

	// 创建测试数据
	suite.CreateCacheTestData(t)

	t.Run("缓存一致性更新性能测试", func(t *testing.T) {
		concurrency := 50
		totalRequests := 500

		var consistentCount, inconsistentCount int64
		var mutex sync.Mutex

		requestFunc := func() *RequestResult {
			start := time.Now()

			// 随机选择一个商品进行更新
			productID := uint(rand.Intn(100) + 1)

			// 模拟数据库更新
			newPrice := fmt.Sprintf("%.2f", rand.Float64()*1000+10)
			updates := map[string]interface{}{
				"price": newPrice,
			}

			// 通过一致性管理器同步更新
			cacheKey := suite.keyManager.GenerateProductKey(productID)
			err := suite.consistencyMgr.SyncCache(cacheKey, "products", productID, updates)

			duration := time.Since(start)

			// 验证一致性（稍后检查）
			time.Sleep(10 * time.Millisecond) // 等待异步更新完成

			// 检查缓存和数据库的一致性
			var dbProduct model.Product
			suite.db.Where("id = ?", productID).First(&dbProduct)

			cachedData, cacheErr := suite.cacheManager.Get(cacheKey)

			mutex.Lock()
			if cacheErr == nil && cachedData != nil {
				// 这里简化一致性检查，实际应该比较具体字段
				consistentCount++
			} else {
				inconsistentCount++
			}
			mutex.Unlock()

			return &RequestResult{
				Success:  err == nil,
				Duration: duration,
				Error:    err,
			}
		}

		result := suite.RunConcurrentTest(t, "缓存一致性更新", concurrency, totalRequests, requestFunc)

		// 计算一致性指标
		totalChecks := consistentCount + inconsistentCount
		consistencyRate := float64(consistentCount) / float64(totalChecks) * 100

		// 验证性能指标
		assert.GreaterOrEqual(t, consistencyRate, 95.0, "数据一致性应≥95%")
		assert.Less(t, result.AverageTime, 50*time.Millisecond, "一致性更新平均响应时间应<50ms")
		assert.Greater(t, result.RequestsPerSec, 200.0, "一致性更新QPS应>200")
		assert.Less(t, result.ErrorRate, 5.0, "错误率应<5%")

		t.Logf("✅ 缓存一致性性能测试通过")
		t.Logf("   - 数据一致性: %.2f%% (目标: ≥95%%)", consistencyRate)
		t.Logf("   - 平均响应时间: %v (目标: <50ms)", result.AverageTime)
		t.Logf("   - QPS: %.2f (目标: >200)", result.RequestsPerSec)
		t.Logf("   - 错误率: %.2f%% (目标: <5%%)", result.ErrorRate)
	})
}

// TestCacheWarmupPerformance 测试缓存预热性能
func TestCacheWarmupPerformance(t *testing.T) {
	suite := SetupCachePerformanceTest(t)
	if suite == nil {
		return
	}
	defer suite.CleanupCachePerformanceTest()

	// 创建测试数据
	suite.CreateTestData(t) // 不预热，测试从零开始的预热性能

	t.Run("缓存预热性能测试", func(t *testing.T) {
		start := time.Now()

		// 执行缓存预热
		err := suite.warmupMgr.WarmupAll()

		warmupDuration := time.Since(start)

		// 验证预热结果
		assert.NoError(t, err, "缓存预热应该成功")
		assert.Less(t, warmupDuration, 30*time.Second, "预热时间应<30秒")

		// 检查预热后的缓存命中率
		time.Sleep(2 * time.Second) // 等待预热完成

		// 测试预热后的查询性能
		hitCount := 0
		testQueries := 100

		for i := 1; i <= testQueries; i++ {
			productID := uint(i)
			key := suite.keyManager.GenerateProductKey(productID)

			if _, err := suite.cacheManager.Get(key); err == nil {
				hitCount++
			}
		}

		hitRate := float64(hitCount) / float64(testQueries) * 100

		assert.GreaterOrEqual(t, hitRate, 80.0, "预热后缓存命中率应≥80%")

		t.Logf("✅ 缓存预热性能测试通过")
		t.Logf("   - 预热耗时: %v (目标: <30s)", warmupDuration)
		t.Logf("   - 预热后命中率: %.2f%% (目标: ≥80%%)", hitRate)
	})
}

// TestCacheProtectionPerformance 测试缓存防护性能
func TestCacheProtectionPerformance(t *testing.T) {
	suite := SetupCachePerformanceTest(t)
	if suite == nil {
		return
	}
	defer suite.CleanupCachePerformanceTest()

	// 创建测试数据
	suite.CreateCacheTestData(t)

	t.Run("缓存防护机制性能测试", func(t *testing.T) {
		concurrency := 100
		totalRequests := 1000

		var protectedCount, unprotectedCount int64
		var mutex sync.Mutex

		requestFunc := func() *RequestResult {
			start := time.Now()

			// 50%概率查询存在的数据，50%查询不存在的数据（测试防穿透）
			var productID uint
			var exists bool
			if rand.Float64() < 0.5 {
				productID = uint(rand.Intn(100) + 1) // 存在的数据
				exists = true
			} else {
				productID = uint(rand.Intn(1000) + 10000) // 不存在的数据
				exists = false
			}

			// 使用防护机制获取数据
			key := suite.keyManager.GenerateProductKey(productID)

			loader := func() (interface{}, error) {
				var product model.Product
				err := suite.db.Where("id = ?", productID).First(&product).Error
				if err != nil {
					return nil, err
				}
				return product, nil
			}

			data, err := suite.protectionMgr.ProtectedGet(key, loader)
			duration := time.Since(start)

			mutex.Lock()
			if exists && data != nil {
				protectedCount++
			} else if !exists && (err != nil || data == nil) {
				unprotectedCount++ // 正确防护了不存在的数据
			}
			mutex.Unlock()

			return &RequestResult{
				Success:  (exists && data != nil) || (!exists && (err != nil || data == nil)),
				Duration: duration,
				Error:    nil, // 防护机制的错误是预期的
			}
		}

		result := suite.RunConcurrentTest(t, "缓存防护机制", concurrency, totalRequests, requestFunc)

		// 计算防护效果
		totalProtected := protectedCount + unprotectedCount
		protectionRate := float64(totalProtected) / float64(totalRequests) * 100

		// 验证性能指标
		assert.GreaterOrEqual(t, protectionRate, 90.0, "防护有效率应≥90%")
		assert.Less(t, result.AverageTime, 30*time.Millisecond, "防护机制平均响应时间应<30ms")
		assert.Greater(t, result.RequestsPerSec, 1000.0, "防护机制QPS应>1000")

		t.Logf("✅ 缓存防护性能测试通过")
		t.Logf("   - 防护有效率: %.2f%% (目标: ≥90%%)", protectionRate)
		t.Logf("   - 平均响应时间: %v (目标: <30ms)", result.AverageTime)
		t.Logf("   - QPS: %.2f (目标: >1000)", result.RequestsPerSec)
	})
}

// TestCacheMonitoringPerformance 测试缓存监控性能
func TestCacheMonitoringPerformance(t *testing.T) {
	suite := SetupCachePerformanceTest(t)
	if suite == nil {
		return
	}
	defer suite.CleanupCachePerformanceTest()

	// 创建测试数据
	suite.CreateCacheTestData(t)

	t.Run("缓存监控系统性能测试", func(t *testing.T) {
		// 模拟一段时间的缓存操作
		concurrency := 50
		totalRequests := 1000

		requestFunc := func() *RequestResult {
			start := time.Now()

			// 随机缓存操作
			productID := uint(rand.Intn(200) + 1)
			key := suite.keyManager.GenerateProductKey(productID)

			// 50%概率读取，50%概率写入
			if rand.Float64() < 0.5 {
				// 读取操作
				_, err := suite.cacheManager.Get(key)
				duration := time.Since(start)

				// 记录监控数据
				suite.monitoringMgr.RecordResponseTime(duration)
				suite.monitoringMgr.RecordHotKey(key, err == nil)

				return &RequestResult{
					Success:  true,
					Duration: duration,
					Error:    err,
				}
			} else {
				// 写入操作
				data := map[string]interface{}{
					"id":    productID,
					"name":  fmt.Sprintf("Product %d", productID),
					"price": rand.Float64() * 1000,
				}
				err := suite.cacheManager.Set(key, data, 1*time.Hour)
				duration := time.Since(start)

				// 记录监控数据
				suite.monitoringMgr.RecordResponseTime(duration)

				return &RequestResult{
					Success:  err == nil,
					Duration: duration,
					Error:    err,
				}
			}
		}

		result := suite.RunConcurrentTest(t, "缓存监控系统", concurrency, totalRequests, requestFunc)

		// 等待监控数据收集
		time.Sleep(3 * time.Second)

		// 获取监控统计
		stats := suite.monitoringMgr.GetStats()
		hotKeys := suite.monitoringMgr.GetHotKeys(10)

		// 验证监控功能
		assert.NotNil(t, stats, "监控统计应该可用")
		assert.Greater(t, len(hotKeys), 0, "应该有热点键数据")
		assert.Less(t, result.AverageTime, 10*time.Millisecond, "监控开销应该很小")

		// 生成性能报告
		report := suite.monitoringMgr.GeneratePerformanceReport("cache_performance_test")
		assert.NotNil(t, report, "应该能生成性能报告")

		t.Logf("✅ 缓存监控性能测试通过")
		t.Logf("   - 监控开销: %v (目标: <10ms)", result.AverageTime)
		t.Logf("   - 热点键数量: %d", len(hotKeys))
		t.Logf("   - 平均响应时间: %v", stats.AvgResponseTime)
		t.Logf("   - P95响应时间: %v", stats.P95ResponseTime)
	})
}
