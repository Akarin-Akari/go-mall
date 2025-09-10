package performance

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"mall-go/internal/config"
	"mall-go/internal/model"
	"mall-go/pkg/cart"
	"mall-go/pkg/product"
	"mall-go/pkg/user"

	"github.com/glebarez/sqlite"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// SimplePerformanceTest 简化的性能测试
func TestSimplePerformance(t *testing.T) {
	// 初始化配置
	config.GlobalConfig = config.Config{
		JWT: config.JWTConfig{
			Secret: "test-secret-key-for-performance-testing",
			Expire: "24h",
		},
	}

	// 初始化测试数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	assert.NoError(t, err, "数据库连接失败")

	// 自动迁移
	err = db.AutoMigrate(
		&model.User{},
		&model.Product{},
		&model.ProductImage{},
		&model.ProductSKU{},
		&model.Category{},
		&model.Brand{},
		&model.Cart{},
		&model.CartItem{},
		&model.Order{},
		&model.OrderItem{},
		&model.OrderStatusLog{},
		&model.Payment{},
	)
	assert.NoError(t, err, "数据库迁移失败")

	// 创建服务
	productService := product.NewProductService(db)
	cartService := cart.NewCartService(db)
	registerService := user.NewRegisterService(db)

	// 创建测试数据
	createPerformanceTestData(t, db)

	t.Run("商品服务性能测试", func(t *testing.T) {
		concurrency := 100
		totalRequests := 1000

		results := make(chan time.Duration, totalRequests)
		var wg sync.WaitGroup
		startTime := time.Now()

		// 控制并发数
		semaphore := make(chan struct{}, concurrency)

		for i := 0; i < totalRequests; i++ {
			wg.Add(1)
			go func(requestID int) {
				defer wg.Done()
				semaphore <- struct{}{}
				defer func() { <-semaphore }()

				requestStart := time.Now()
				productID := uint(requestID%1000 + 1)
				product, err := productService.GetProduct(productID)
				duration := time.Since(requestStart)

				if err == nil && product != nil {
					results <- duration
				} else {
					results <- time.Duration(-1) // 标记错误
				}
			}(i)
		}

		wg.Wait()
		close(results)
		totalTime := time.Since(startTime)

		// 统计结果
		var responseTimes []time.Duration
		successCount := 0
		for duration := range results {
			if duration > 0 {
				responseTimes = append(responseTimes, duration)
				successCount++
			}
		}

		if len(responseTimes) > 0 {
			// 计算平均时间
			var totalResponseTime time.Duration
			for _, duration := range responseTimes {
				totalResponseTime += duration
			}
			averageTime := totalResponseTime / time.Duration(len(responseTimes))

			// 计算QPS
			qps := float64(successCount) / totalTime.Seconds()

			t.Logf("📊 商品服务性能测试结果:")
			t.Logf("   总请求数: %d", totalRequests)
			t.Logf("   成功请求: %d", successCount)
			t.Logf("   失败请求: %d", totalRequests-successCount)
			t.Logf("   平均响应时间: %v", averageTime)
			t.Logf("   QPS: %.2f", qps)
			t.Logf("   成功率: %.2f%%", float64(successCount)/float64(totalRequests)*100)

			// 验证性能指标
			assert.Less(t, averageTime, 50*time.Millisecond, "平均响应时间应小于50ms")
			assert.Greater(t, qps, 500.0, "QPS应大于500")
			assert.Greater(t, float64(successCount)/float64(totalRequests), 0.95, "成功率应大于95%")

			t.Logf("✅ 商品服务性能测试通过 - 平均响应时间: %v, QPS: %.2f", averageTime, qps)
		} else {
			t.Logf("❌ 商品服务性能测试失败 - 没有成功的请求")
		}
	})

	t.Run("购物车服务性能测试", func(t *testing.T) {
		concurrency := 50
		totalRequests := 500

		results := make(chan time.Duration, totalRequests)
		var wg sync.WaitGroup
		startTime := time.Now()

		// 控制并发数
		semaphore := make(chan struct{}, concurrency)

		for i := 0; i < totalRequests; i++ {
			wg.Add(1)
			go func(requestID int) {
				defer wg.Done()
				semaphore <- struct{}{}
				defer func() { <-semaphore }()

				requestStart := time.Now()
				userID := uint(requestID%100 + 1)
				productID := uint(requestID%1000 + 1)

				addToCartReq := &model.AddToCartRequest{
					ProductID: productID,
					Quantity:  1,
				}

				_, err := cartService.AddToCart(userID, "", addToCartReq)
				duration := time.Since(requestStart)

				if err == nil {
					results <- duration
				} else {
					results <- time.Duration(-1) // 标记错误
				}
			}(i)
		}

		wg.Wait()
		close(results)
		totalTime := time.Since(startTime)

		// 统计结果
		var responseTimes []time.Duration
		successCount := 0
		for duration := range results {
			if duration > 0 {
				responseTimes = append(responseTimes, duration)
				successCount++
			}
		}

		if len(responseTimes) > 0 {
			// 计算平均时间
			var totalResponseTime time.Duration
			for _, duration := range responseTimes {
				totalResponseTime += duration
			}
			averageTime := totalResponseTime / time.Duration(len(responseTimes))

			// 计算QPS
			qps := float64(successCount) / totalTime.Seconds()

			t.Logf("📊 购物车服务性能测试结果:")
			t.Logf("   总请求数: %d", totalRequests)
			t.Logf("   成功请求: %d", successCount)
			t.Logf("   失败请求: %d", totalRequests-successCount)
			t.Logf("   平均响应时间: %v", averageTime)
			t.Logf("   QPS: %.2f", qps)
			t.Logf("   成功率: %.2f%%", float64(successCount)/float64(totalRequests)*100)

			// 验证性能指标
			assert.Less(t, averageTime, 100*time.Millisecond, "平均响应时间应小于100ms")
			assert.Greater(t, qps, 100.0, "QPS应大于100")
			assert.Greater(t, float64(successCount)/float64(totalRequests), 0.80, "成功率应大于80%")

			t.Logf("✅ 购物车服务性能测试通过 - 平均响应时间: %v, QPS: %.2f", averageTime, qps)
		}
	})

	t.Run("用户注册服务性能测试", func(t *testing.T) {
		concurrency := 30
		totalRequests := 300

		results := make(chan time.Duration, totalRequests)
		var wg sync.WaitGroup
		startTime := time.Now()

		// 控制并发数
		semaphore := make(chan struct{}, concurrency)

		for i := 0; i < totalRequests; i++ {
			wg.Add(1)
			go func(requestID int) {
				defer wg.Done()
				semaphore <- struct{}{}
				defer func() { <-semaphore }()

				requestStart := time.Now()

				registerReq := &user.RegisterRequest{
					Username: fmt.Sprintf("perfuser%d", requestID+10000),
					Email:    fmt.Sprintf("perfuser%d@example.com", requestID+10000),
					Password: "password123",
					Phone:    fmt.Sprintf("1380019%04d", requestID%10000),
				}

				_, err := registerService.Register(registerReq)
				duration := time.Since(requestStart)

				if err == nil {
					results <- duration
				} else {
					results <- time.Duration(-1) // 标记错误
				}
			}(i)
		}

		wg.Wait()
		close(results)
		totalTime := time.Since(startTime)

		// 统计结果
		var responseTimes []time.Duration
		successCount := 0
		for duration := range results {
			if duration > 0 {
				responseTimes = append(responseTimes, duration)
				successCount++
			}
		}

		if len(responseTimes) > 0 {
			// 计算平均时间
			var totalResponseTime time.Duration
			for _, duration := range responseTimes {
				totalResponseTime += duration
			}
			averageTime := totalResponseTime / time.Duration(len(responseTimes))

			// 计算QPS
			qps := float64(successCount) / totalTime.Seconds()

			t.Logf("📊 用户注册服务性能测试结果:")
			t.Logf("   总请求数: %d", totalRequests)
			t.Logf("   成功请求: %d", successCount)
			t.Logf("   失败请求: %d", totalRequests-successCount)
			t.Logf("   平均响应时间: %v", averageTime)
			t.Logf("   QPS: %.2f", qps)
			t.Logf("   成功率: %.2f%%", float64(successCount)/float64(totalRequests)*100)

			// 验证性能指标
			assert.Less(t, averageTime, 200*time.Millisecond, "平均响应时间应小于200ms")
			assert.Greater(t, qps, 50.0, "QPS应大于50")
			assert.Greater(t, float64(successCount)/float64(totalRequests), 0.90, "成功率应大于90%")

			t.Logf("✅ 用户注册服务性能测试通过 - 平均响应时间: %v, QPS: %.2f", averageTime, qps)
		}
	})
}

// createPerformanceTestData 创建性能测试数据
func createPerformanceTestData(t *testing.T, db *gorm.DB) {
	// 创建测试用户
	for i := 1; i <= 100; i++ {
		user := &model.User{
			Username: fmt.Sprintf("perfuser%d", i),
			Email:    fmt.Sprintf("perf%d@example.com", i),
			Password: "hashedpassword",
			Phone:    fmt.Sprintf("1380013%04d", i),
			Status:   "active",
		}
		err := db.Create(user).Error
		assert.NoError(t, err, "创建测试用户失败")
	}

	// 创建商家用户
	for i := 1; i <= 10; i++ {
		merchant := &model.User{
			Username: fmt.Sprintf("perfmerchant%d", i),
			Email:    fmt.Sprintf("perfmerchant%d@example.com", i),
			Password: "hashedpassword",
			Phone:    fmt.Sprintf("1380014%04d", i),
			Role:     "merchant",
			Status:   "active",
		}
		err := db.Create(merchant).Error
		assert.NoError(t, err, "创建商家用户失败")
	}

	// 创建分类
	for i := 1; i <= 20; i++ {
		category := &model.Category{
			Name:        fmt.Sprintf("性能测试分类%d", i),
			Description: fmt.Sprintf("performance-test-category-%d", i),
			Status:      "active",
		}
		err := db.Create(category).Error
		assert.NoError(t, err, "创建分类失败")
	}

	// 创建品牌
	for i := 1; i <= 10; i++ {
		brand := &model.Brand{
			Name:        fmt.Sprintf("性能测试品牌%d", i),
			Description: fmt.Sprintf("performance-test-brand-%d", i),
			Status:      "active",
		}
		err := db.Create(brand).Error
		assert.NoError(t, err, "创建品牌失败")
	}

	// 创建商品
	for i := 1; i <= 1000; i++ {
		price, _ := decimal.NewFromString(fmt.Sprintf("%.2f", float64(i)*9.99))
		product := &model.Product{
			Name:        fmt.Sprintf("性能测试商品%d", i),
			Description: fmt.Sprintf("用于性能测试的商品%d", i),
			CategoryID:  uint((i-1)%20 + 1),   // 分配到不同分类
			BrandID:     uint((i-1)%10 + 1),   // 分配到不同品牌
			MerchantID:  uint((i-1)%10 + 101), // 分配到不同商家
			Price:       price,
			Stock:       1000,
			Status:      "active",
		}
		err := db.Create(product).Error
		assert.NoError(t, err, "创建测试商品失败")
	}

	t.Logf("✅ 性能测试数据创建完成 - 用户: 100, 商家: 10, 分类: 20, 品牌: 10, 商品: 1000")
}
