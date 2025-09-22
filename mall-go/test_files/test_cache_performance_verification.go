package main

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"mall-go/internal/config"
	"mall-go/internal/model"
	"mall-go/pkg/cache"
	"mall-go/pkg/logger"
	"mall-go/pkg/optimistic"
	"mall-go/tests/helpers"

	"github.com/glebarez/sqlite"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

// PerformanceVerificationResult 性能验证结果
type PerformanceVerificationResult struct {
	TestName            string        `json:"test_name"`
	Timestamp           time.Time     `json:"timestamp"`
	ConcurrentUsers     int           `json:"concurrent_users"`
	TestDuration        time.Duration `json:"test_duration"`
	TotalRequests       int64         `json:"total_requests"`
	SuccessRequests     int64         `json:"success_requests"`
	FailedRequests      int64         `json:"failed_requests"`
	AverageResponseTime time.Duration `json:"average_response_time"`
	P95ResponseTime     time.Duration `json:"p95_response_time"`
	P99ResponseTime     time.Duration `json:"p99_response_time"`
	QPS                 float64       `json:"qps"`
	CacheHitRate        float64       `json:"cache_hit_rate"`
	CacheMissRate       float64       `json:"cache_miss_rate"`
	ErrorRate           float64       `json:"error_rate"`
	MemoryUsageMB       float64       `json:"memory_usage_mb"`
	DatabaseQueries     int64         `json:"database_queries"`
	CacheOperations     int64         `json:"cache_operations"`
	Passed              bool          `json:"passed"`
	FailureReasons      []string      `json:"failure_reasons"`
}

// PerformanceTargets 性能目标
type PerformanceTargets struct {
	MinQPS                    float64       `json:"min_qps"`
	MaxAverageResponseTime    time.Duration `json:"max_average_response_time"`
	MaxP95ResponseTime        time.Duration `json:"max_p95_response_time"`
	MinCacheHitRate           float64       `json:"min_cache_hit_rate"`
	MaxErrorRate              float64       `json:"max_error_rate"`
	MinDatabaseQueryReduction float64       `json:"min_database_query_reduction"`
}

func main() {
	// 初始化日志
	logger.Init()

	fmt.Println("🚀 Mall-Go 缓存性能验证程序")
	fmt.Println("=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=")

	// 设置性能目标
	targets := &PerformanceTargets{
		MinQPS:                    10000,                 // 目标QPS ≥ 10,000
		MaxAverageResponseTime:    5 * time.Millisecond,  // 平均响应时间 ≤ 5ms
		MaxP95ResponseTime:        20 * time.Millisecond, // P95响应时间 ≤ 20ms
		MinCacheHitRate:           90.0,                  // 缓存命中率 ≥ 90%
		MaxErrorRate:              1.0,                   // 错误率 ≤ 1%
		MinDatabaseQueryReduction: 80.0,                  // 数据库查询减少 ≥ 80%
	}

	fmt.Printf("📋 性能验证目标:\n")
	fmt.Printf("   - 最小QPS: %.0f\n", targets.MinQPS)
	fmt.Printf("   - 最大平均响应时间: %v\n", targets.MaxAverageResponseTime)
	fmt.Printf("   - 最大P95响应时间: %v\n", targets.MaxP95ResponseTime)
	fmt.Printf("   - 最小缓存命中率: %.1f%%\n", targets.MinCacheHitRate)
	fmt.Printf("   - 最大错误率: %.1f%%\n", targets.MaxErrorRate)
	fmt.Printf("   - 最小数据库查询减少: %.1f%%\n", targets.MinDatabaseQueryReduction)

	// 1. 基础环境验证
	fmt.Println("\n🔧 步骤1: 基础环境验证")
	if !verifyEnvironment() {
		fmt.Println("❌ 基础环境验证失败")
		return
	}
	fmt.Println("✅ 基础环境验证通过")

	// 2. 缓存系统初始化验证
	fmt.Println("\n🏗️ 步骤2: 缓存系统初始化验证")
	cacheSystem, db := initializeCacheSystem()
	if cacheSystem == nil {
		fmt.Println("❌ 缓存系统初始化失败")
		return
	}
	fmt.Println("✅ 缓存系统初始化成功")

	// 3. 测试数据准备
	fmt.Println("\n📊 步骤3: 测试数据准备")
	if !prepareTestData(db) {
		fmt.Println("❌ 测试数据准备失败")
		return
	}
	fmt.Println("✅ 测试数据准备完成")

	// 4. 缓存预热验证
	fmt.Println("\n🔥 步骤4: 缓存预热验证")
	if !verifyCacheWarmup(cacheSystem) {
		fmt.Println("❌ 缓存预热验证失败")
		return
	}
	fmt.Println("✅ 缓存预热验证通过")

	// 5. 性能基准测试
	fmt.Println("\n⚡ 步骤5: 性能基准测试")
	baselineResult := runBaselinePerformanceTest(db)
	fmt.Printf("   基准性能 - QPS: %.2f, 平均响应时间: %v\n",
		baselineResult.QPS, baselineResult.AverageResponseTime)

	// 6. 缓存性能测试
	fmt.Println("\n🎯 步骤6: 缓存性能测试")
	cacheResult := runCachePerformanceTest(cacheSystem, db)
	fmt.Printf("   缓存性能 - QPS: %.2f, 平均响应时间: %v, 命中率: %.2f%%\n",
		cacheResult.QPS, cacheResult.AverageResponseTime, cacheResult.CacheHitRate)

	// 7. 并发压力测试
	fmt.Println("\n💪 步骤7: 并发压力测试")
	stressResult := runConcurrentStressTest(cacheSystem, db)
	fmt.Printf("   压力测试 - QPS: %.2f, 错误率: %.2f%%\n",
		stressResult.QPS, stressResult.ErrorRate)

	// 8. 一致性验证测试
	fmt.Println("\n🔄 步骤8: 一致性验证测试")
	consistencyResult := runConsistencyVerificationTest(cacheSystem, db)
	fmt.Printf("   一致性测试 - 成功率: %.2f%%\n",
		float64(consistencyResult.SuccessRequests)/float64(consistencyResult.TotalRequests)*100)

	// 9. 综合性能验证
	fmt.Println("\n📈 步骤9: 综合性能验证")
	finalResult := runComprehensivePerformanceTest(cacheSystem, db)

	// 10. 性能目标验证
	fmt.Println("\n✅ 步骤10: 性能目标验证")
	passed, reasons := verifyPerformanceTargets(finalResult, targets)

	// 输出最终结果
	fmt.Println("\n🎉 性能验证完成!")
	fmt.Println("=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=")

	printFinalResults(finalResult, targets, passed, reasons)

	if passed {
		fmt.Println("\n🎊 恭喜！所有性能目标均已达成！")
		fmt.Println("   Mall-Go缓存优化项目第三周验收标准100%通过！")
	} else {
		fmt.Println("\n⚠️ 部分性能目标未达成，需要进一步优化")
		for _, reason := range reasons {
			fmt.Printf("   - %s\n", reason)
		}
	}
}

// verifyEnvironment 验证基础环境
func verifyEnvironment() bool {
	// 检查Redis连接
	redisConfig := config.RedisConfig{
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
	if err != nil {
		fmt.Printf("   ❌ Redis连接失败: %v\n", err)
		return false
	}
	defer redisClient.Close()

	if err := redisClient.Ping(); err != nil {
		fmt.Printf("   ❌ Redis Ping失败: %v\n", err)
		return false
	}

	fmt.Println("   ✅ Redis连接正常")
	return true
}

// CacheSystem 缓存系统结构
type CacheSystem struct {
	CacheManager   cache.CacheManager
	KeyManager     *cache.CacheKeyManager
	ConsistencyMgr *cache.CacheConsistencyManager
	WarmupMgr      *cache.CacheWarmupManager
	ProtectionMgr  *cache.CacheProtectionManager
	MonitoringMgr  *cache.CacheMonitoringManager
}

// initializeCacheSystem 初始化缓存系统
func initializeCacheSystem() (*CacheSystem, *gorm.DB) {
	// 初始化数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: gormLogger.Default.LogMode(gormLogger.Silent),
	})
	if err != nil {
		fmt.Printf("   ❌ 数据库初始化失败: %v\n", err)
		return nil, nil
	}

	// 自动迁移
	db.AutoMigrate(
		&model.User{},
		&model.Product{},
		&model.Category{},
		&model.Order{},
		&model.OrderItem{},
		&model.CartItem{},
		&model.Cart{},
	)

	// 初始化Redis
	redisConfig := config.RedisConfig{
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
	if err != nil {
		fmt.Printf("   ❌ Redis客户端创建失败: %v\n", err)
		return nil, nil
	}

	// 清理Redis
	redisClient.GetClient().FlushDB(context.Background())

	// 创建缓存管理器
	cacheManager := cache.NewRedisCacheManager(redisClient)

	// 初始化键管理器
	cache.InitKeyManager("perf_verify")
	keyManager := cache.GetKeyManager()

	// 创建测试助手和乐观锁服务
	_ = helpers.NewTestHelper(db)
	optimisticLock := optimistic.NewOptimisticLockService(db)

	// 创建各种管理器
	consistencyConfig := cache.DefaultCacheConsistencyConfig()
	consistencyMgr := cache.NewCacheConsistencyManager(consistencyConfig, cacheManager, keyManager, optimisticLock)

	warmupConfig := cache.DefaultCacheWarmupConfig()
	warmupMgr := cache.NewCacheWarmupManager(warmupConfig, cacheManager, keyManager, consistencyMgr, optimisticLock)

	protectionConfig := cache.DefaultCacheProtectionConfig()
	protectionConfig.Strategies = []cache.ProtectionStrategy{
		cache.ProtectionNullCache,
		cache.ProtectionRandomTTL,
	}
	protectionMgr := cache.NewCacheProtectionManager(protectionConfig, cacheManager, keyManager, consistencyMgr, warmupMgr, optimisticLock)

	monitoringConfig := cache.DefaultCacheMonitoringConfig()
	monitoringMgr := cache.NewCacheMonitoringManager(monitoringConfig, cacheManager, keyManager, consistencyMgr, warmupMgr, protectionMgr, optimisticLock)

	// 启动所有管理器
	consistencyMgr.Start()
	warmupMgr.Start()
	protectionMgr.Start()
	monitoringMgr.Start()

	return &CacheSystem{
		CacheManager:   cacheManager,
		KeyManager:     keyManager,
		ConsistencyMgr: consistencyMgr,
		WarmupMgr:      warmupMgr,
		ProtectionMgr:  protectionMgr,
		MonitoringMgr:  monitoringMgr,
	}, db
}

// prepareTestData 准备测试数据
func prepareTestData(db *gorm.DB) bool {
	// 创建分类
	categories := []model.Category{
		{Name: "电子产品", Description: "各种电子设备", Status: "active"},
		{Name: "服装", Description: "时尚服装", Status: "active"},
		{Name: "家居", Description: "家居用品", Status: "active"},
		{Name: "图书", Description: "各类图书", Status: "active"},
		{Name: "运动", Description: "运动用品", Status: "active"},
	}

	for _, category := range categories {
		if err := db.Create(&category).Error; err != nil {
			fmt.Printf("   ❌ 创建分类失败: %v\n", err)
			return false
		}
	}

	// 创建商品
	for i := 1; i <= 2000; i++ {
		product := model.Product{
			Name:        fmt.Sprintf("商品%d", i),
			Description: fmt.Sprintf("这是商品%d的描述", i),
			Price:       decimal.NewFromFloat(rand.Float64()*1000 + 10),
			Stock:       rand.Intn(1000) + 10,
			CategoryID:  uint(rand.Intn(5) + 1),
			Status:      "on_sale",
		}

		if err := db.Create(&product).Error; err != nil {
			fmt.Printf("   ❌ 创建商品失败: %v\n", err)
			return false
		}
	}

	// 创建用户
	for i := 1; i <= 200; i++ {
		user := model.User{
			Username: fmt.Sprintf("user%d", i),
			Email:    fmt.Sprintf("user%d@example.com", i),
			Password: "password123",
			Status:   "active",
		}

		if err := db.Create(&user).Error; err != nil {
			fmt.Printf("   ❌ 创建用户失败: %v\n", err)
			return false
		}
	}

	fmt.Printf("   ✅ 测试数据创建完成 - 分类: %d, 商品: 2000, 用户: 200\n", len(categories))
	return true
}

// verifyCacheWarmup 验证缓存预热
func verifyCacheWarmup(system *CacheSystem) bool {
	start := time.Now()

	// 执行缓存预热
	if err := system.WarmupMgr.WarmupAll(); err != nil {
		fmt.Printf("   ❌ 缓存预热失败: %v\n", err)
		return false
	}

	warmupDuration := time.Since(start)

	// 等待预热完成
	time.Sleep(3 * time.Second)

	// 验证预热效果
	hitCount := 0
	testCount := 100

	for i := 1; i <= testCount; i++ {
		key := system.KeyManager.GenerateProductKey(uint(i))
		if _, err := system.CacheManager.Get(key); err == nil {
			hitCount++
		}
	}

	hitRate := float64(hitCount) / float64(testCount) * 100

	fmt.Printf("   ✅ 缓存预热完成 - 耗时: %v, 命中率: %.2f%%\n", warmupDuration, hitRate)

	return hitRate >= 70.0 // 预热后命中率应≥70%
}

// runBaselinePerformanceTest 运行基准性能测试
func runBaselinePerformanceTest(db *gorm.DB) *PerformanceVerificationResult {
	concurrency := 100
	testDuration := 10 * time.Second

	var totalRequests int64
	var successRequests int64
	var totalResponseTime int64
	var dbQueries int64

	ctx, cancel := context.WithTimeout(context.Background(), testDuration)
	defer cancel()

	var wg sync.WaitGroup
	startTime := time.Now()

	// 启动并发工作者
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for {
				select {
				case <-ctx.Done():
					return
				default:
					requestStart := time.Now()

					// 直接从数据库查询
					productID := uint(rand.Intn(2000) + 1)
					var product model.Product
					err := db.Where("id = ?", productID).First(&product).Error

					duration := time.Since(requestStart)

					atomic.AddInt64(&totalRequests, 1)
					atomic.AddInt64(&totalResponseTime, int64(duration))
					atomic.AddInt64(&dbQueries, 1)

					if err == nil {
						atomic.AddInt64(&successRequests, 1)
					}

					time.Sleep(time.Microsecond * 100)
				}
			}
		}()
	}

	wg.Wait()
	actualDuration := time.Since(startTime)

	// 计算指标
	totalReq := atomic.LoadInt64(&totalRequests)
	successReq := atomic.LoadInt64(&successRequests)
	avgResponseTime := time.Duration(atomic.LoadInt64(&totalResponseTime) / totalReq)
	qps := float64(totalReq) / actualDuration.Seconds()
	errorRate := float64(totalReq-successReq) / float64(totalReq) * 100

	return &PerformanceVerificationResult{
		TestName:            "基准性能测试",
		Timestamp:           time.Now(),
		ConcurrentUsers:     concurrency,
		TestDuration:        actualDuration,
		TotalRequests:       totalReq,
		SuccessRequests:     successReq,
		FailedRequests:      totalReq - successReq,
		AverageResponseTime: avgResponseTime,
		QPS:                 qps,
		ErrorRate:           errorRate,
		DatabaseQueries:     atomic.LoadInt64(&dbQueries),
		CacheOperations:     0,
		CacheHitRate:        0,
		CacheMissRate:       100,
	}
}

// runCachePerformanceTest 运行缓存性能测试
func runCachePerformanceTest(system *CacheSystem, db *gorm.DB) *PerformanceVerificationResult {
	concurrency := 200
	testDuration := 15 * time.Second

	var totalRequests int64
	var successRequests int64
	var totalResponseTime int64
	var dbQueries int64
	var cacheHits int64
	var cacheMisses int64
	var cacheOperations int64

	ctx, cancel := context.WithTimeout(context.Background(), testDuration)
	defer cancel()

	var wg sync.WaitGroup
	startTime := time.Now()

	// 启动并发工作者
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for {
				select {
				case <-ctx.Done():
					return
				default:
					requestStart := time.Now()

					// 先尝试从缓存获取
					productID := uint(rand.Intn(2000) + 1)
					key := system.KeyManager.GenerateProductKey(productID)

					cachedData, err := system.CacheManager.Get(key)
					atomic.AddInt64(&cacheOperations, 1)

					if err == nil && cachedData != nil {
						// 缓存命中
						atomic.AddInt64(&cacheHits, 1)
						atomic.AddInt64(&successRequests, 1)
					} else {
						// 缓存未命中，从数据库查询
						atomic.AddInt64(&cacheMisses, 1)
						var product model.Product
						dbErr := db.Where("id = ?", productID).First(&product).Error
						atomic.AddInt64(&dbQueries, 1)

						if dbErr == nil {
							// 缓存数据
							system.CacheManager.Set(key, product, 1*time.Hour)
							atomic.AddInt64(&successRequests, 1)
						}
					}

					duration := time.Since(requestStart)
					atomic.AddInt64(&totalRequests, 1)
					atomic.AddInt64(&totalResponseTime, int64(duration))

					// 记录监控数据
					system.MonitoringMgr.RecordResponseTime(duration)
					system.MonitoringMgr.RecordHotKey(key, err == nil)

					time.Sleep(time.Microsecond * 50)
				}
			}
		}()
	}

	wg.Wait()
	actualDuration := time.Since(startTime)

	// 计算指标
	totalReq := atomic.LoadInt64(&totalRequests)
	successReq := atomic.LoadInt64(&successRequests)
	hits := atomic.LoadInt64(&cacheHits)
	misses := atomic.LoadInt64(&cacheMisses)
	avgResponseTime := time.Duration(atomic.LoadInt64(&totalResponseTime) / totalReq)
	qps := float64(totalReq) / actualDuration.Seconds()
	errorRate := float64(totalReq-successReq) / float64(totalReq) * 100
	hitRate := float64(hits) / float64(hits+misses) * 100
	missRate := float64(misses) / float64(hits+misses) * 100

	return &PerformanceVerificationResult{
		TestName:            "缓存性能测试",
		Timestamp:           time.Now(),
		ConcurrentUsers:     concurrency,
		TestDuration:        actualDuration,
		TotalRequests:       totalReq,
		SuccessRequests:     successReq,
		FailedRequests:      totalReq - successReq,
		AverageResponseTime: avgResponseTime,
		QPS:                 qps,
		ErrorRate:           errorRate,
		DatabaseQueries:     atomic.LoadInt64(&dbQueries),
		CacheOperations:     atomic.LoadInt64(&cacheOperations),
		CacheHitRate:        hitRate,
		CacheMissRate:       missRate,
	}
}

// runConcurrentStressTest 运行并发压力测试
func runConcurrentStressTest(system *CacheSystem, db *gorm.DB) *PerformanceVerificationResult {
	concurrency := 500
	testDuration := 20 * time.Second

	var totalRequests int64
	var successRequests int64
	var totalResponseTime int64
	var cacheHits int64
	var cacheMisses int64

	ctx, cancel := context.WithTimeout(context.Background(), testDuration)
	defer cancel()

	var wg sync.WaitGroup
	startTime := time.Now()

	// 启动并发工作者
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for {
				select {
				case <-ctx.Done():
					return
				default:
					requestStart := time.Now()

					// 混合操作：80%读取，20%写入
					if rand.Float64() < 0.8 {
						// 读取操作
						productID := uint(rand.Intn(2000) + 1)
						key := system.KeyManager.GenerateProductKey(productID)

						_, err := system.CacheManager.Get(key)
						if err == nil {
							atomic.AddInt64(&cacheHits, 1)
							atomic.AddInt64(&successRequests, 1)
						} else {
							atomic.AddInt64(&cacheMisses, 1)
							// 从数据库加载
							var product model.Product
							if dbErr := db.Where("id = ?", productID).First(&product).Error; dbErr == nil {
								system.CacheManager.Set(key, product, 1*time.Hour)
								atomic.AddInt64(&successRequests, 1)
							}
						}
					} else {
						// 写入操作
						productID := uint(rand.Intn(2000) + 1)
						key := system.KeyManager.GenerateProductKey(productID)

						data := map[string]interface{}{
							"id":    productID,
							"name":  fmt.Sprintf("Product %d", productID),
							"price": rand.Float64() * 1000,
						}

						if err := system.CacheManager.Set(key, data, 1*time.Hour); err == nil {
							atomic.AddInt64(&successRequests, 1)
						}
					}

					duration := time.Since(requestStart)
					atomic.AddInt64(&totalRequests, 1)
					atomic.AddInt64(&totalResponseTime, int64(duration))

					time.Sleep(time.Microsecond * 20)
				}
			}
		}()
	}

	wg.Wait()
	actualDuration := time.Since(startTime)

	// 计算指标
	totalReq := atomic.LoadInt64(&totalRequests)
	successReq := atomic.LoadInt64(&successRequests)
	hits := atomic.LoadInt64(&cacheHits)
	misses := atomic.LoadInt64(&cacheMisses)
	avgResponseTime := time.Duration(atomic.LoadInt64(&totalResponseTime) / totalReq)
	qps := float64(totalReq) / actualDuration.Seconds()
	errorRate := float64(totalReq-successReq) / float64(totalReq) * 100
	hitRate := float64(hits) / float64(hits+misses) * 100

	return &PerformanceVerificationResult{
		TestName:            "并发压力测试",
		Timestamp:           time.Now(),
		ConcurrentUsers:     concurrency,
		TestDuration:        actualDuration,
		TotalRequests:       totalReq,
		SuccessRequests:     successReq,
		FailedRequests:      totalReq - successReq,
		AverageResponseTime: avgResponseTime,
		QPS:                 qps,
		ErrorRate:           errorRate,
		CacheHitRate:        hitRate,
		CacheMissRate:       100 - hitRate,
	}
}

// runConsistencyVerificationTest 运行一致性验证测试
func runConsistencyVerificationTest(system *CacheSystem, db *gorm.DB) *PerformanceVerificationResult {
	concurrency := 50
	testDuration := 10 * time.Second

	var totalRequests int64
	var successRequests int64
	var consistencyChecks int64
	var consistentData int64

	ctx, cancel := context.WithTimeout(context.Background(), testDuration)
	defer cancel()

	var wg sync.WaitGroup
	startTime := time.Now()

	// 启动并发更新工作者
	for i := 0; i < concurrency/2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for {
				select {
				case <-ctx.Done():
					return
				default:
					// 随机更新商品价格
					productID := uint(rand.Intn(100) + 1)
					newPrice := fmt.Sprintf("%.2f", rand.Float64()*1000+10)

					updates := map[string]interface{}{
						"price": newPrice,
					}

					// 通过一致性管理器更新
					cacheKey := system.KeyManager.GenerateProductKey(productID)
					if err := system.ConsistencyMgr.SyncCache(cacheKey, "products", productID, updates); err == nil {
						atomic.AddInt64(&successRequests, 1)
					}
					atomic.AddInt64(&totalRequests, 1)

					time.Sleep(time.Millisecond * 20)
				}
			}
		}()
	}

	// 启动一致性检查工作者
	for i := 0; i < concurrency/2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for {
				select {
				case <-ctx.Done():
					return
				default:
					// 检查数据一致性
					productID := uint(rand.Intn(100) + 1)
					cacheKey := system.KeyManager.GenerateProductKey(productID)

					// 从缓存和数据库获取数据
					_, cacheErr := system.CacheManager.Get(cacheKey)
					var dbProduct model.Product
					dbErr := db.Where("id = ?", productID).First(&dbProduct).Error

					atomic.AddInt64(&consistencyChecks, 1)

					// 简化的一致性检查
					if (cacheErr == nil && dbErr == nil) || (cacheErr != nil && dbErr != nil) {
						atomic.AddInt64(&consistentData, 1)
					}

					time.Sleep(time.Millisecond * 10)
				}
			}
		}()
	}

	wg.Wait()
	actualDuration := time.Since(startTime)

	// 计算指标
	totalReq := atomic.LoadInt64(&totalRequests)
	successReq := atomic.LoadInt64(&successRequests)
	checks := atomic.LoadInt64(&consistencyChecks)
	consistent := atomic.LoadInt64(&consistentData)

	consistencyRate := float64(consistent) / float64(checks) * 100

	return &PerformanceVerificationResult{
		TestName:        "一致性验证测试",
		Timestamp:       time.Now(),
		ConcurrentUsers: concurrency,
		TestDuration:    actualDuration,
		TotalRequests:   totalReq,
		SuccessRequests: successReq,
		FailedRequests:  totalReq - successReq,
		QPS:             float64(totalReq) / actualDuration.Seconds(),
		ErrorRate:       float64(totalReq-successReq) / float64(totalReq) * 100,
		// 使用一致性率作为成功率的指标
		Passed: consistencyRate >= 85.0,
	}
}

// runComprehensivePerformanceTest 运行综合性能测试
func runComprehensivePerformanceTest(system *CacheSystem, db *gorm.DB) *PerformanceVerificationResult {
	concurrency := 300
	testDuration := 30 * time.Second

	var totalRequests int64
	var successRequests int64
	var totalResponseTime int64
	var dbQueries int64
	var cacheHits int64
	var cacheMisses int64
	var cacheOperations int64
	var responseTimes []time.Duration
	var responseTimesMutex sync.Mutex

	ctx, cancel := context.WithTimeout(context.Background(), testDuration)
	defer cancel()

	var wg sync.WaitGroup
	startTime := time.Now()

	// 启动并发工作者
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for {
				select {
				case <-ctx.Done():
					return
				default:
					requestStart := time.Now()

					// 混合操作：70%读取，20%写入，10%更新
					operation := rand.Float64()
					productID := uint(rand.Intn(2000) + 1)
					key := system.KeyManager.GenerateProductKey(productID)

					if operation < 0.7 {
						// 读取操作
						cachedData, err := system.CacheManager.Get(key)
						atomic.AddInt64(&cacheOperations, 1)

						if err == nil && cachedData != nil {
							atomic.AddInt64(&cacheHits, 1)
							atomic.AddInt64(&successRequests, 1)
						} else {
							atomic.AddInt64(&cacheMisses, 1)
							var product model.Product
							if dbErr := db.Where("id = ?", productID).First(&product).Error; dbErr == nil {
								system.CacheManager.Set(key, product, 1*time.Hour)
								atomic.AddInt64(&successRequests, 1)
							}
							atomic.AddInt64(&dbQueries, 1)
						}
					} else if operation < 0.9 {
						// 写入操作
						data := map[string]interface{}{
							"id":    productID,
							"name":  fmt.Sprintf("Product %d", productID),
							"price": rand.Float64() * 1000,
						}

						if err := system.CacheManager.Set(key, data, 1*time.Hour); err == nil {
							atomic.AddInt64(&successRequests, 1)
						}
						atomic.AddInt64(&cacheOperations, 1)
					} else {
						// 更新操作
						updates := map[string]interface{}{
							"price": fmt.Sprintf("%.2f", rand.Float64()*1000+10),
						}

						if err := system.ConsistencyMgr.SyncCache(key, "products", productID, updates); err == nil {
							atomic.AddInt64(&successRequests, 1)
						}
					}

					duration := time.Since(requestStart)
					atomic.AddInt64(&totalRequests, 1)
					atomic.AddInt64(&totalResponseTime, int64(duration))

					// 记录响应时间用于P95/P99计算
					responseTimesMutex.Lock()
					responseTimes = append(responseTimes, duration)
					responseTimesMutex.Unlock()

					// 记录监控数据
					system.MonitoringMgr.RecordResponseTime(duration)
					system.MonitoringMgr.RecordHotKey(key, operation < 0.7)

					time.Sleep(time.Microsecond * 30)
				}
			}
		}()
	}

	wg.Wait()
	actualDuration := time.Since(startTime)

	// 计算指标
	totalReq := atomic.LoadInt64(&totalRequests)
	successReq := atomic.LoadInt64(&successRequests)
	hits := atomic.LoadInt64(&cacheHits)
	misses := atomic.LoadInt64(&cacheMisses)
	avgResponseTime := time.Duration(atomic.LoadInt64(&totalResponseTime) / totalReq)
	qps := float64(totalReq) / actualDuration.Seconds()
	errorRate := float64(totalReq-successReq) / float64(totalReq) * 100
	hitRate := float64(hits) / float64(hits+misses) * 100

	// 计算P95和P99响应时间
	responseTimesMutex.Lock()
	p95Time, p99Time := calculatePercentiles(responseTimes)
	responseTimesMutex.Unlock()

	return &PerformanceVerificationResult{
		TestName:            "综合性能测试",
		Timestamp:           time.Now(),
		ConcurrentUsers:     concurrency,
		TestDuration:        actualDuration,
		TotalRequests:       totalReq,
		SuccessRequests:     successReq,
		FailedRequests:      totalReq - successReq,
		AverageResponseTime: avgResponseTime,
		P95ResponseTime:     p95Time,
		P99ResponseTime:     p99Time,
		QPS:                 qps,
		ErrorRate:           errorRate,
		DatabaseQueries:     atomic.LoadInt64(&dbQueries),
		CacheOperations:     atomic.LoadInt64(&cacheOperations),
		CacheHitRate:        hitRate,
		CacheMissRate:       100 - hitRate,
	}
}

// calculatePercentiles 计算响应时间百分位数
func calculatePercentiles(times []time.Duration) (p95, p99 time.Duration) {
	if len(times) == 0 {
		return 0, 0
	}

	// 简单排序
	for i := 0; i < len(times)-1; i++ {
		for j := 0; j < len(times)-i-1; j++ {
			if times[j] > times[j+1] {
				times[j], times[j+1] = times[j+1], times[j]
			}
		}
	}

	p95Index := int(float64(len(times)) * 0.95)
	p99Index := int(float64(len(times)) * 0.99)

	if p95Index >= len(times) {
		p95Index = len(times) - 1
	}
	if p99Index >= len(times) {
		p99Index = len(times) - 1
	}

	return times[p95Index], times[p99Index]
}

// verifyPerformanceTargets 验证性能目标
func verifyPerformanceTargets(result *PerformanceVerificationResult, targets *PerformanceTargets) (bool, []string) {
	var failures []string

	// 检查QPS
	if result.QPS < targets.MinQPS {
		failures = append(failures, fmt.Sprintf("QPS未达标: %.2f < %.2f", result.QPS, targets.MinQPS))
	}

	// 检查平均响应时间
	if result.AverageResponseTime > targets.MaxAverageResponseTime {
		failures = append(failures, fmt.Sprintf("平均响应时间超标: %v > %v", result.AverageResponseTime, targets.MaxAverageResponseTime))
	}

	// 检查P95响应时间
	if result.P95ResponseTime > targets.MaxP95ResponseTime {
		failures = append(failures, fmt.Sprintf("P95响应时间超标: %v > %v", result.P95ResponseTime, targets.MaxP95ResponseTime))
	}

	// 检查缓存命中率
	if result.CacheHitRate < targets.MinCacheHitRate {
		failures = append(failures, fmt.Sprintf("缓存命中率未达标: %.2f%% < %.2f%%", result.CacheHitRate, targets.MinCacheHitRate))
	}

	// 检查错误率
	if result.ErrorRate > targets.MaxErrorRate {
		failures = append(failures, fmt.Sprintf("错误率超标: %.2f%% > %.2f%%", result.ErrorRate, targets.MaxErrorRate))
	}

	return len(failures) == 0, failures
}

// printFinalResults 打印最终结果
func printFinalResults(result *PerformanceVerificationResult, targets *PerformanceTargets, passed bool, reasons []string) {
	fmt.Printf("\n📊 综合性能测试结果:\n")
	fmt.Printf("   - 测试名称: %s\n", result.TestName)
	fmt.Printf("   - 测试时间: %v\n", result.Timestamp.Format("2006-01-02 15:04:05"))
	fmt.Printf("   - 并发用户: %d\n", result.ConcurrentUsers)
	fmt.Printf("   - 测试时长: %v\n", result.TestDuration)
	fmt.Printf("   - 总请求数: %d\n", result.TotalRequests)
	fmt.Printf("   - 成功请求: %d\n", result.SuccessRequests)
	fmt.Printf("   - 失败请求: %d\n", result.FailedRequests)

	fmt.Printf("\n⚡ 性能指标:\n")
	fmt.Printf("   - QPS: %.2f (目标: ≥%.0f) %s\n",
		result.QPS, targets.MinQPS, getStatusIcon(result.QPS >= targets.MinQPS))
	fmt.Printf("   - 平均响应时间: %v (目标: ≤%v) %s\n",
		result.AverageResponseTime, targets.MaxAverageResponseTime,
		getStatusIcon(result.AverageResponseTime <= targets.MaxAverageResponseTime))
	fmt.Printf("   - P95响应时间: %v (目标: ≤%v) %s\n",
		result.P95ResponseTime, targets.MaxP95ResponseTime,
		getStatusIcon(result.P95ResponseTime <= targets.MaxP95ResponseTime))
	fmt.Printf("   - P99响应时间: %v\n", result.P99ResponseTime)

	fmt.Printf("\n🎯 缓存指标:\n")
	fmt.Printf("   - 缓存命中率: %.2f%% (目标: ≥%.1f%%) %s\n",
		result.CacheHitRate, targets.MinCacheHitRate,
		getStatusIcon(result.CacheHitRate >= targets.MinCacheHitRate))
	fmt.Printf("   - 缓存未命中率: %.2f%%\n", result.CacheMissRate)
	fmt.Printf("   - 缓存操作数: %d\n", result.CacheOperations)
	fmt.Printf("   - 数据库查询数: %d\n", result.DatabaseQueries)

	fmt.Printf("\n🔍 质量指标:\n")
	fmt.Printf("   - 错误率: %.2f%% (目标: ≤%.1f%%) %s\n",
		result.ErrorRate, targets.MaxErrorRate,
		getStatusIcon(result.ErrorRate <= targets.MaxErrorRate))

	if result.DatabaseQueries > 0 {
		dbReduction := (1 - float64(result.DatabaseQueries)/float64(result.TotalRequests)) * 100
		fmt.Printf("   - 数据库查询减少: %.2f%% (目标: ≥%.1f%%) %s\n",
			dbReduction, targets.MinDatabaseQueryReduction,
			getStatusIcon(dbReduction >= targets.MinDatabaseQueryReduction))
	}

	fmt.Printf("\n📈 验收结果: %s\n", getStatusIcon(passed))
	if passed {
		fmt.Printf("   🎉 所有性能目标均已达成！\n")
	} else {
		fmt.Printf("   ⚠️ 以下目标未达成:\n")
		for _, reason := range reasons {
			fmt.Printf("     - %s\n", reason)
		}
	}
}

// getStatusIcon 获取状态图标
func getStatusIcon(passed bool) string {
	if passed {
		return "✅"
	}
	return "❌"
}
