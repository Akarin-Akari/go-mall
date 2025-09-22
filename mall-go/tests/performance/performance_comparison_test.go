package performance

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	"mall-go/internal/model"
	"mall-go/pkg/cache"
	"mall-go/pkg/logger"

	"github.com/stretchr/testify/assert"
)

// PerformanceComparisonSuite 性能对比测试套件
type PerformanceComparisonSuite struct {
	*PerformanceTestSuite
	cacheManager  cache.CacheManager
	keyManager    *cache.CacheKeyManager
	monitoringMgr *cache.CacheMonitoringManager
}

// ComparisonResult 性能对比结果
type ComparisonResult struct {
	TestName          string                `json:"test_name"`
	Timestamp         time.Time             `json:"timestamp"`
	WithoutCache      *PerformanceMetrics   `json:"without_cache"`
	WithCache         *PerformanceMetrics   `json:"with_cache"`
	Improvement       *ImprovementMetrics   `json:"improvement"`
	CacheMetrics      *CacheSpecificMetrics `json:"cache_metrics"`
	TestConfiguration *TestConfiguration    `json:"test_configuration"`
}

// PerformanceMetrics 性能指标
type PerformanceMetrics struct {
	AverageResponseTime time.Duration `json:"average_response_time"`
	P95ResponseTime     time.Duration `json:"p95_response_time"`
	P99ResponseTime     time.Duration `json:"p99_response_time"`
	QPS                 float64       `json:"qps"`
	ErrorRate           float64       `json:"error_rate"`
	TotalRequests       int           `json:"total_requests"`
	SuccessRequests     int           `json:"success_requests"`
	DatabaseQueries     int           `json:"database_queries"`
}

// ImprovementMetrics 改进指标
type ImprovementMetrics struct {
	ResponseTimeImprovement float64 `json:"response_time_improvement"` // 百分比
	QPSImprovement          float64 `json:"qps_improvement"`           // 百分比
	DatabaseQueryReduction  float64 `json:"database_query_reduction"`  // 百分比
	ErrorRateReduction      float64 `json:"error_rate_reduction"`      // 百分比
}

// CacheSpecificMetrics 缓存特定指标
type CacheSpecificMetrics struct {
	HitRate         float64       `json:"hit_rate"`
	MissRate        float64       `json:"miss_rate"`
	AvgCacheTime    time.Duration `json:"avg_cache_time"`
	CacheOperations int           `json:"cache_operations"`
	WarmupTime      time.Duration `json:"warmup_time"`
	MemoryUsageMB   float64       `json:"memory_usage_mb"`
}

// TestConfiguration 测试配置
type TestConfiguration struct {
	ConcurrentUsers int           `json:"concurrent_users"`
	TestDuration    time.Duration `json:"test_duration"`
	TotalRequests   int           `json:"total_requests"`
	DataSetSize     int           `json:"data_set_size"`
	CacheEnabled    bool          `json:"cache_enabled"`
	TestScenario    string        `json:"test_scenario"`
}

// SetupPerformanceComparison 设置性能对比测试环境
func SetupPerformanceComparison(t *testing.T) *PerformanceComparisonSuite {
	baseSuite := SetupPerformanceTest(t)

	// 初始化日志
	logger.Init()

	// 初始化Redis客户端（可选）
	var cacheManager cache.CacheManager
	var keyManager *cache.CacheKeyManager
	var monitoringMgr *cache.CacheMonitoringManager

	// 尝试连接Redis，如果失败则使用内存缓存
	redisConfig := cache.RedisConfig{
		Host:         "localhost",
		Port:         6379,
		Password:     "",
		DB:           1,
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
	if err == nil {
		redisClient.FlushDB()
		cacheManager = cache.NewRedisCacheManager(redisClient)

		// 初始化键管理器
		cache.InitKeyManager("perf_comp")
		keyManager = cache.GetKeyManager()

		// 创建监控管理器
		monitoringConfig := cache.DefaultCacheMonitoringConfig()
		monitoringConfig.CollectInterval = 1 * time.Second
		monitoringMgr = cache.NewCacheMonitoringManager(monitoringConfig, cacheManager, keyManager, nil, nil, nil, nil)
		monitoringMgr.Start()
	} else {
		t.Logf("Redis不可用，将使用模拟缓存进行对比测试")
		// 这里可以使用内存缓存实现
	}

	return &PerformanceComparisonSuite{
		PerformanceTestSuite: baseSuite,
		cacheManager:         cacheManager,
		keyManager:           keyManager,
		monitoringMgr:        monitoringMgr,
	}
}

// CleanupPerformanceComparison 清理性能对比测试环境
func (suite *PerformanceComparisonSuite) CleanupPerformanceComparison() {
	if suite.monitoringMgr != nil {
		suite.monitoringMgr.Stop()
	}
	suite.CleanupPerformanceTest()
}

// TestProductQueryPerformanceComparison 测试商品查询性能对比
func TestProductQueryPerformanceComparison(t *testing.T) {
	suite := SetupPerformanceComparison(t)
	if suite == nil {
		return
	}
	defer suite.CleanupPerformanceComparison()

	// 创建测试数据
	suite.CreateTestData(t)

	t.Run("商品查询性能对比测试", func(t *testing.T) {
		concurrency := 100
		totalRequests := 2000

		// 测试配置
		config := &TestConfiguration{
			ConcurrentUsers: concurrency,
			TotalRequests:   totalRequests,
			DataSetSize:     1000,
			TestScenario:    "商品查询",
		}

		// 1. 测试无缓存性能
		t.Logf("🔍 测试无缓存性能...")
		withoutCacheMetrics := suite.testWithoutCache(t, concurrency, totalRequests)

		// 2. 预热缓存
		if suite.cacheManager != nil {
			t.Logf("🔥 预热缓存...")
			suite.warmupProductCache(t)
		}

		// 3. 测试有缓存性能
		t.Logf("⚡ 测试有缓存性能...")
		withCacheMetrics, cacheMetrics := suite.testWithCache(t, concurrency, totalRequests)

		// 4. 计算改进指标
		improvement := suite.calculateImprovement(withoutCacheMetrics, withCacheMetrics)

		// 5. 生成对比结果
		result := &ComparisonResult{
			TestName:          "商品查询性能对比",
			Timestamp:         time.Now(),
			WithoutCache:      withoutCacheMetrics,
			WithCache:         withCacheMetrics,
			Improvement:       improvement,
			CacheMetrics:      cacheMetrics,
			TestConfiguration: config,
		}

		// 6. 验证性能改进
		assert.GreaterOrEqual(t, improvement.ResponseTimeImprovement, 70.0, "响应时间改进应≥70%")
		assert.GreaterOrEqual(t, improvement.QPSImprovement, 200.0, "QPS改进应≥200%")
		assert.GreaterOrEqual(t, improvement.DatabaseQueryReduction, 80.0, "数据库查询减少应≥80%")

		if suite.cacheManager != nil {
			assert.GreaterOrEqual(t, cacheMetrics.HitRate, 85.0, "缓存命中率应≥85%")
		}

		// 7. 保存测试报告
		suite.saveComparisonReport(t, result)

		// 8. 输出结果
		suite.printComparisonResults(t, result)

		t.Logf("✅ 商品查询性能对比测试通过")
	})
}

// testWithoutCache 测试无缓存性能
func (suite *PerformanceComparisonSuite) testWithoutCache(t *testing.T, concurrency, totalRequests int) *PerformanceMetrics {
	var dbQueries int64

	requestFunc := func() *RequestResult {
		start := time.Now()

		// 直接从数据库查询
		productID := uint(rand.Intn(1000) + 1)
		var product model.Product
		err := suite.db.Where("id = ?", productID).First(&product).Error

		duration := time.Since(start)
		dbQueries++

		return &RequestResult{
			Success:  err == nil,
			Duration: duration,
			Error:    err,
		}
	}

	result := suite.RunConcurrentTest(t, "无缓存查询", concurrency, totalRequests, requestFunc)

	return &PerformanceMetrics{
		AverageResponseTime: result.AverageTime,
		P95ResponseTime:     result.P95ResponseTime,
		P99ResponseTime:     result.P99ResponseTime,
		QPS:                 result.RequestsPerSec,
		ErrorRate:           result.ErrorRate,
		TotalRequests:       result.TotalRequests,
		SuccessRequests:     result.SuccessRequests,
		DatabaseQueries:     int(dbQueries),
	}
}

// testWithCache 测试有缓存性能
func (suite *PerformanceComparisonSuite) testWithCache(t *testing.T, concurrency, totalRequests int) (*PerformanceMetrics, *CacheSpecificMetrics) {
	if suite.cacheManager == nil {
		// 如果没有缓存，返回模拟数据
		return suite.testWithoutCache(t, concurrency, totalRequests), &CacheSpecificMetrics{
			HitRate:         0,
			MissRate:        100,
			AvgCacheTime:    0,
			CacheOperations: 0,
		}
	}

	var dbQueries int64
	var cacheHits int64
	var cacheMisses int64
	var cacheOperations int64

	requestFunc := func() *RequestResult {
		start := time.Now()

		// 先尝试从缓存获取
		productID := uint(rand.Intn(1000) + 1)
		key := suite.keyManager.GenerateProductKey(productID)

		cachedData, err := suite.cacheManager.Get(key)
		cacheOperations++

		if err == nil && cachedData != nil {
			// 缓存命中
			cacheHits++
			duration := time.Since(start)

			// 记录监控数据
			if suite.monitoringMgr != nil {
				suite.monitoringMgr.RecordResponseTime(duration)
				suite.monitoringMgr.RecordHotKey(key, true)
			}

			return &RequestResult{
				Success:  true,
				Duration: duration,
				Error:    nil,
			}
		} else {
			// 缓存未命中，从数据库查询
			cacheMisses++
			var product model.Product
			dbErr := suite.db.Where("id = ?", productID).First(&product).Error
			dbQueries++

			if dbErr == nil {
				// 缓存数据
				suite.cacheManager.Set(key, product, 1*time.Hour)
			}

			duration := time.Since(start)

			// 记录监控数据
			if suite.monitoringMgr != nil {
				suite.monitoringMgr.RecordResponseTime(duration)
				suite.monitoringMgr.RecordHotKey(key, false)
			}

			return &RequestResult{
				Success:  dbErr == nil,
				Duration: duration,
				Error:    dbErr,
			}
		}
	}

	result := suite.RunConcurrentTest(t, "有缓存查询", concurrency, totalRequests, requestFunc)

	// 计算缓存指标
	totalCacheOps := cacheHits + cacheMisses
	hitRate := float64(cacheHits) / float64(totalCacheOps) * 100
	missRate := float64(cacheMisses) / float64(totalCacheOps) * 100

	performanceMetrics := &PerformanceMetrics{
		AverageResponseTime: result.AverageTime,
		P95ResponseTime:     result.P95ResponseTime,
		P99ResponseTime:     result.P99ResponseTime,
		QPS:                 result.RequestsPerSec,
		ErrorRate:           result.ErrorRate,
		TotalRequests:       result.TotalRequests,
		SuccessRequests:     result.SuccessRequests,
		DatabaseQueries:     int(dbQueries),
	}

	cacheMetrics := &CacheSpecificMetrics{
		HitRate:         hitRate,
		MissRate:        missRate,
		AvgCacheTime:    result.AverageTime, // 简化处理
		CacheOperations: int(cacheOperations),
		MemoryUsageMB:   0, // 需要实际测量
	}

	return performanceMetrics, cacheMetrics
}

// warmupProductCache 预热商品缓存
func (suite *PerformanceComparisonSuite) warmupProductCache(t *testing.T) {
	if suite.cacheManager == nil {
		return
	}

	start := time.Now()

	// 预热前500个商品
	for i := 1; i <= 500; i++ {
		var product model.Product
		err := suite.db.Where("id = ?", i).First(&product).Error
		if err == nil {
			key := suite.keyManager.GenerateProductKey(product.ID)
			suite.cacheManager.Set(key, product, 1*time.Hour)
		}
	}

	warmupDuration := time.Since(start)
	t.Logf("   缓存预热完成，耗时: %v", warmupDuration)
}

// calculateImprovement 计算性能改进
func (suite *PerformanceComparisonSuite) calculateImprovement(without, with *PerformanceMetrics) *ImprovementMetrics {
	responseTimeImprovement := (1 - float64(with.AverageResponseTime)/float64(without.AverageResponseTime)) * 100
	qpsImprovement := (with.QPS/without.QPS - 1) * 100
	dbQueryReduction := (1 - float64(with.DatabaseQueries)/float64(without.DatabaseQueries)) * 100
	errorRateReduction := (without.ErrorRate - with.ErrorRate) / without.ErrorRate * 100

	return &ImprovementMetrics{
		ResponseTimeImprovement: responseTimeImprovement,
		QPSImprovement:          qpsImprovement,
		DatabaseQueryReduction:  dbQueryReduction,
		ErrorRateReduction:      errorRateReduction,
	}
}

// saveComparisonReport 保存对比报告
func (suite *PerformanceComparisonSuite) saveComparisonReport(t *testing.T, result *ComparisonResult) {
	// 创建报告目录
	reportDir := "test-reports"
	if err := os.MkdirAll(reportDir, 0755); err != nil {
		t.Logf("创建报告目录失败: %v", err)
		return
	}

	// 生成报告文件名
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("performance_comparison_%s.json", timestamp)
	filepath := filepath.Join(reportDir, filename)

	// 保存JSON报告
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		t.Logf("序列化报告失败: %v", err)
		return
	}

	if err := os.WriteFile(filepath, data, 0644); err != nil {
		t.Logf("保存报告失败: %v", err)
		return
	}

	t.Logf("📊 性能对比报告已保存: %s", filepath)
}

// printComparisonResults 打印对比结果
func (suite *PerformanceComparisonSuite) printComparisonResults(t *testing.T, result *ComparisonResult) {
	t.Logf("\n📊 性能对比测试结果:")
	t.Logf("=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=")

	t.Logf("🔍 无缓存性能:")
	t.Logf("   - 平均响应时间: %v", result.WithoutCache.AverageResponseTime)
	t.Logf("   - P95响应时间: %v", result.WithoutCache.P95ResponseTime)
	t.Logf("   - QPS: %.2f", result.WithoutCache.QPS)
	t.Logf("   - 错误率: %.2f%%", result.WithoutCache.ErrorRate)
	t.Logf("   - 数据库查询: %d", result.WithoutCache.DatabaseQueries)

	t.Logf("\n⚡ 有缓存性能:")
	t.Logf("   - 平均响应时间: %v", result.WithCache.AverageResponseTime)
	t.Logf("   - P95响应时间: %v", result.WithCache.P95ResponseTime)
	t.Logf("   - QPS: %.2f", result.WithCache.QPS)
	t.Logf("   - 错误率: %.2f%%", result.WithCache.ErrorRate)
	t.Logf("   - 数据库查询: %d", result.WithCache.DatabaseQueries)

	if result.CacheMetrics != nil {
		t.Logf("\n🎯 缓存指标:")
		t.Logf("   - 缓存命中率: %.2f%%", result.CacheMetrics.HitRate)
		t.Logf("   - 缓存未命中率: %.2f%%", result.CacheMetrics.MissRate)
		t.Logf("   - 缓存操作数: %d", result.CacheMetrics.CacheOperations)
	}

	t.Logf("\n📈 性能改进:")
	t.Logf("   - 响应时间改进: %.2f%%", result.Improvement.ResponseTimeImprovement)
	t.Logf("   - QPS改进: %.2f%%", result.Improvement.QPSImprovement)
	t.Logf("   - 数据库查询减少: %.2f%%", result.Improvement.DatabaseQueryReduction)
	t.Logf("   - 错误率减少: %.2f%%", result.Improvement.ErrorRateReduction)
}
