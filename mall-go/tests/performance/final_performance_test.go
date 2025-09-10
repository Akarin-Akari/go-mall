package performance

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"mall-go/internal/config"
	"mall-go/internal/model"

	"github.com/glebarez/sqlite"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// TestFinalPerformance 最终性能测试
func TestFinalPerformance(t *testing.T) {
	// 初始化配置
	config.GlobalConfig = config.Config{
		JWT: config.JWTConfig{
			Secret: "test-secret-key-for-final-performance-testing",
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

	// 创建测试数据
	createFinalTestData(t, db)

	t.Run("数据库查询性能测试", func(t *testing.T) {
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

				var product model.Product
				productID := uint(requestID%1000 + 1)
				err := db.Where("id = ?", productID).First(&product).Error

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
		successCount, averageTime, qps := calculateFinalMetrics(results, totalTime, totalRequests)

		t.Logf("📊 数据库查询性能测试结果:")
		t.Logf("   总请求数: %d", totalRequests)
		t.Logf("   成功请求: %d", successCount)
		t.Logf("   失败请求: %d", totalRequests-successCount)
		t.Logf("   平均响应时间: %v", averageTime)
		t.Logf("   QPS: %.2f", qps)
		t.Logf("   成功率: %.2f%%", float64(successCount)/float64(totalRequests)*100)

		// 验证性能指标
		assert.Greater(t, successCount, totalRequests*8/10, "成功率应大于80%")
		assert.Less(t, averageTime, 10*time.Millisecond, "平均响应时间应小于10ms")
		assert.Greater(t, qps, 1000.0, "QPS应大于1000")

		t.Logf("✅ 数据库查询性能测试通过 - 平均响应时间: %v, QPS: %.2f", averageTime, qps)
	})

	t.Run("并发写入性能测试", func(t *testing.T) {
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

				// 创建新用户
				user := &model.User{
					Username: fmt.Sprintf("finaluser%d", requestID+30000),
					Email:    fmt.Sprintf("finaluser%d@example.com", requestID+30000),
					Password: "hashedpassword",
					Phone:    fmt.Sprintf("1380030%04d", requestID%10000),
					Status:   "active",
				}

				err := db.Create(user).Error
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
		successCount, averageTime, qps := calculateFinalMetrics(results, totalTime, totalRequests)

		t.Logf("📊 并发写入性能测试结果:")
		t.Logf("   总请求数: %d", totalRequests)
		t.Logf("   成功请求: %d", successCount)
		t.Logf("   失败请求: %d", totalRequests-successCount)
		t.Logf("   平均响应时间: %v", averageTime)
		t.Logf("   QPS: %.2f", qps)
		t.Logf("   成功率: %.2f%%", float64(successCount)/float64(totalRequests)*100)

		// 验证性能指标
		assert.Greater(t, successCount, totalRequests*8/10, "成功率应大于80%")
		assert.Less(t, averageTime, 50*time.Millisecond, "平均响应时间应小于50ms")
		assert.Greater(t, qps, 200.0, "QPS应大于200")

		t.Logf("✅ 并发写入性能测试通过 - 平均响应时间: %v, QPS: %.2f", averageTime, qps)
	})

	t.Run("复杂查询性能测试", func(t *testing.T) {
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

				var products []model.Product
				categoryID := uint(requestID%20 + 1)
				err := db.Where("category_id = ? AND status = ?", categoryID, "active").
					Limit(20).Find(&products).Error

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
		successCount, averageTime, qps := calculateFinalMetrics(results, totalTime, totalRequests)

		t.Logf("📊 复杂查询性能测试结果:")
		t.Logf("   总请求数: %d", totalRequests)
		t.Logf("   成功请求: %d", successCount)
		t.Logf("   失败请求: %d", totalRequests-successCount)
		t.Logf("   平均响应时间: %v", averageTime)
		t.Logf("   QPS: %.2f", qps)
		t.Logf("   成功率: %.2f%%", float64(successCount)/float64(totalRequests)*100)

		// 验证性能指标
		assert.Greater(t, successCount, totalRequests*8/10, "成功率应大于80%")
		assert.Less(t, averageTime, 30*time.Millisecond, "平均响应时间应小于30ms")
		assert.Greater(t, qps, 300.0, "QPS应大于300")

		t.Logf("✅ 复杂查询性能测试通过 - 平均响应时间: %v, QPS: %.2f", averageTime, qps)
	})

	t.Run("混合操作性能测试", func(t *testing.T) {
		concurrency := 80
		totalRequests := 800

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

				// 随机选择操作类型
				operation := requestID % 4
				var err error

				switch operation {
				case 0: // 查询商品
					var product model.Product
					productID := uint(requestID%1000 + 1)
					err = db.Where("id = ?", productID).First(&product).Error

				case 1: // 查询用户
					var user model.User
					userID := uint(requestID%100 + 1)
					err = db.Where("id = ?", userID).First(&user).Error

				case 2: // 创建购物车项
					cartItem := &model.CartItem{
						CartID:    uint(requestID%100 + 1),
						ProductID: uint(requestID%1000 + 1),
						Quantity:  1,
					}
					err = db.Create(cartItem).Error

				case 3: // 查询分类商品
					var products []model.Product
					categoryID := uint(requestID%20 + 1)
					err = db.Where("category_id = ?", categoryID).Limit(10).Find(&products).Error
				}

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
		successCount, averageTime, qps := calculateFinalMetrics(results, totalTime, totalRequests)

		t.Logf("📊 混合操作性能测试结果:")
		t.Logf("   总请求数: %d", totalRequests)
		t.Logf("   成功请求: %d", successCount)
		t.Logf("   失败请求: %d", totalRequests-successCount)
		t.Logf("   平均响应时间: %v", averageTime)
		t.Logf("   QPS: %.2f", qps)
		t.Logf("   成功率: %.2f%%", float64(successCount)/float64(totalRequests)*100)

		// 验证性能指标
		assert.Greater(t, successCount, totalRequests*7/10, "成功率应大于70%")
		assert.Less(t, averageTime, 20*time.Millisecond, "平均响应时间应小于20ms")
		assert.Greater(t, qps, 500.0, "QPS应大于500")

		t.Logf("✅ 混合操作性能测试通过 - 平均响应时间: %v, QPS: %.2f", averageTime, qps)
	})
}

// calculateFinalMetrics 计算最终性能指标
func calculateFinalMetrics(results chan time.Duration, totalTime time.Duration, totalRequests int) (int, time.Duration, float64) {
	var responseTimes []time.Duration
	successCount := 0

	for duration := range results {
		if duration > 0 {
			responseTimes = append(responseTimes, duration)
			successCount++
		}
	}

	if len(responseTimes) == 0 {
		return 0, 0, 0
	}

	// 计算平均时间
	var totalResponseTime time.Duration
	for _, duration := range responseTimes {
		totalResponseTime += duration
	}
	averageTime := totalResponseTime / time.Duration(len(responseTimes))

	// 计算QPS
	qps := float64(successCount) / totalTime.Seconds()

	return successCount, averageTime, qps
}

// createFinalTestData 创建最终测试数据
func createFinalTestData(t *testing.T, db *gorm.DB) {
	// 创建测试用户
	for i := 1; i <= 100; i++ {
		user := &model.User{
			Username: fmt.Sprintf("finaluser%d", i),
			Email:    fmt.Sprintf("finaluser%d@example.com", i),
			Password: "hashedpassword",
			Phone:    fmt.Sprintf("1380031%04d", i),
			Status:   "active",
		}
		err := db.Create(user).Error
		assert.NoError(t, err, "创建测试用户失败")
	}

	// 创建分类
	for i := 1; i <= 20; i++ {
		category := &model.Category{
			Name:        fmt.Sprintf("最终测试分类%d", i),
			Description: fmt.Sprintf("final-test-category-%d", i),
			Status:      "active",
		}
		err := db.Create(category).Error
		assert.NoError(t, err, "创建分类失败")
	}

	// 创建商品
	for i := 1; i <= 1000; i++ {
		price, _ := decimal.NewFromString(fmt.Sprintf("%.2f", float64(i)*29.99))
		product := &model.Product{
			Name:        fmt.Sprintf("最终测试商品%d", i),
			Description: fmt.Sprintf("用于最终测试的商品%d", i),
			CategoryID:  uint((i-1)%20 + 1), // 分配到不同分类
			Price:       price,
			Stock:       500,
			Status:      "active",
		}
		err := db.Create(product).Error
		assert.NoError(t, err, "创建测试商品失败")
	}

	t.Logf("✅ 最终测试数据创建完成 - 用户: 100, 分类: 20, 商品: 1000")
}
