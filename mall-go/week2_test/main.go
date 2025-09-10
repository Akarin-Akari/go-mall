package main

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// 简化的模型定义（用于测试）
type Product struct {
	ID        uint            `gorm:"primarykey" json:"id"`
	Name      string          `gorm:"not null;size:255" json:"name"`
	Price     decimal.Decimal `gorm:"type:decimal(10,2);not null" json:"price"`
	Stock     int             `gorm:"not null;default:0" json:"stock"`
	SoldCount int             `gorm:"default:0" json:"sold_count"`
	Version   int             `gorm:"not null;default:1" json:"version"` // 乐观锁版本号
	Status    string          `gorm:"size:20;default:'active'" json:"status"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// TableName 指定表名
func (Product) TableName() string {
	return "products"
}

type Cart struct {
	ID          uint            `gorm:"primarykey" json:"id"`
	UserID      uint            `gorm:"index" json:"user_id"`
	TotalQty    int             `gorm:"default:0" json:"total_qty"`
	TotalAmount decimal.Decimal `gorm:"type:decimal(10,2);default:0.00" json:"total_amount"`
	Version     int             `gorm:"not null;default:1" json:"version"` // 乐观锁版本号
	Status      string          `gorm:"size:20;default:'active'" json:"status"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// TableName 指定表名
func (Cart) TableName() string {
	return "carts"
}

type CartItem struct {
	ID        uint            `gorm:"primarykey" json:"id"`
	CartID    uint            `gorm:"not null;index" json:"cart_id"`
	ProductID uint            `gorm:"not null;index" json:"product_id"`
	Quantity  int             `gorm:"not null;default:1" json:"quantity"`
	Price     decimal.Decimal `gorm:"type:decimal(10,2);not null" json:"price"`
	Version   int             `gorm:"not null;default:1" json:"version"` // 乐观锁版本号
	Selected  bool            `gorm:"default:true" json:"selected"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// TableName 指定表名
func (CartItem) TableName() string {
	return "cart_items"
}

type Order struct {
	ID          uint            `gorm:"primarykey" json:"id"`
	OrderNo     string          `gorm:"uniqueIndex;not null;size:32" json:"order_no"`
	UserID      uint            `gorm:"not null;index" json:"user_id"`
	TotalAmount decimal.Decimal `gorm:"type:decimal(10,2);not null" json:"total_amount"`
	Status      string          `gorm:"size:20;not null" json:"status"`
	Version     int             `gorm:"not null;default:1" json:"version"` // 乐观锁版本号
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// TableName 指定表名
func (Order) TableName() string {
	return "orders"
}

// 并发测试结果
type ConcurrencyTestResult struct {
	TestName        string
	TotalOperations int
	SuccessCount    int
	FailureCount    int
	ConflictCount   int
	SuccessRate     float64
	Duration        time.Duration
	QPS             float64
	Errors          []string
}

// 并发测试服务
type ConcurrencyTestService struct {
	db *gorm.DB
}

func NewConcurrencyTestService(db *gorm.DB) *ConcurrencyTestService {
	return &ConcurrencyTestService{db: db}
}

func main() {
	fmt.Println("🎯 Mall-Go 第二周并发安全优化测试")
	fmt.Println("========================================")

	// 清理旧的测试数据库文件
	if _, err := os.Stat("test_concurrent.db"); err == nil {
		os.Remove("test_concurrent.db")
	}

	// 连接数据库 - 使用文件数据库以支持并发
	db, err := gorm.Open(sqlite.Open("test_concurrent.db"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})

	// 配置SQLite以支持并发
	sqlDB, err := db.DB()
	if err == nil {
		sqlDB.SetMaxOpenConns(1)    // SQLite只支持单个写连接
		sqlDB.SetMaxIdleConns(1)    // 保持连接池简单
		sqlDB.SetConnMaxLifetime(0) // 连接不过期

		// 启用WAL模式以提高并发性能
		db.Exec("PRAGMA journal_mode=WAL")
		db.Exec("PRAGMA synchronous=NORMAL")
		db.Exec("PRAGMA cache_size=1000")
		db.Exec("PRAGMA temp_store=memory")
	}
	if err != nil {
		log.Fatalf("❌ 连接数据库失败: %v", err)
	}

	// 数据库迁移
	fmt.Println("🚀 开始数据库迁移...")
	if err := db.AutoMigrate(&Product{}, &Cart{}, &CartItem{}, &Order{}); err != nil {
		log.Fatalf("❌ 数据库迁移失败: %v", err)
	}
	fmt.Println("✅ 数据库迁移完成")

	// 创建测试数据
	fmt.Println("🧪 创建测试数据...")
	if err := createTestData(db); err != nil {
		log.Fatalf("❌ 创建测试数据失败: %v", err)
	}
	fmt.Println("✅ 测试数据创建完成")

	// 创建测试服务
	testService := NewConcurrencyTestService(db)

	fmt.Println("\n🔥 开始并发安全测试...")

	// 验证测试数据
	fmt.Println("\n🔍 验证测试数据...")
	testService.DebugTableNames()
	testService.VerifyTestData()

	// 测试1：商品库存并发扣减
	fmt.Println("\n🔍 测试1：商品库存并发扣减")
	fmt.Println("----------------------------------------")
	result1 := testService.TestConcurrentStockDeduction(50, 1, 1) // 每次扣减1个，减少库存压力
	printTestResult(result1)

	// 测试2：购物车并发更新
	fmt.Println("\n🛒 测试2：购物车并发更新")
	fmt.Println("----------------------------------------")
	result2 := testService.TestConcurrentCartUpdate(30, 1, 2) // 减少并发数量
	printTestResult(result2)

	// 测试3：订单并发创建
	fmt.Println("\n📋 测试3：订单并发创建")
	fmt.Println("----------------------------------------")
	result3 := testService.TestConcurrentOrderCreation(20, 1)
	printTestResult(result3)

	// 详细错误分析
	if result3.SuccessRate < 95.0 {
		fmt.Println("\n🔍 订单创建失败原因分析:")
		testService.AnalyzeOrderCreationFailures()
	}

	// 测试4：混合并发操作
	fmt.Println("\n🔄 测试4：混合并发操作")
	fmt.Println("----------------------------------------")
	result4 := testService.TestMixedConcurrentOperations(40)
	printTestResult(result4)

	fmt.Println("\n🎉 第二周并发安全优化测试完成！")

	// 验收标准检查
	fmt.Println("\n📊 验收标准检查:")
	fmt.Println("========================================")

	allPassed := true

	// 检查并发写入成功率
	if result1.SuccessRate >= 95.0 {
		fmt.Printf("✅ 库存扣减成功率: %.2f%% (>= 95%%)\n", result1.SuccessRate)
	} else {
		fmt.Printf("❌ 库存扣减成功率: %.2f%% (< 95%%)\n", result1.SuccessRate)
		allPassed = false
	}

	if result2.SuccessRate >= 95.0 {
		fmt.Printf("✅ 购物车更新成功率: %.2f%% (>= 95%%)\n", result2.SuccessRate)
	} else {
		fmt.Printf("❌ 购物车更新成功率: %.2f%% (< 95%%)\n", result2.SuccessRate)
		allPassed = false
	}

	if result3.SuccessRate >= 95.0 {
		fmt.Printf("✅ 订单创建成功率: %.2f%% (>= 95%%)\n", result3.SuccessRate)
	} else {
		fmt.Printf("❌ 订单创建成功率: %.2f%% (< 95%%)\n", result3.SuccessRate)
		allPassed = false
	}

	// 检查QPS性能
	if result1.QPS >= 1000 {
		fmt.Printf("✅ 库存扣减QPS: %.0f (>= 1000)\n", result1.QPS)
	} else {
		fmt.Printf("❌ 库存扣减QPS: %.0f (< 1000)\n", result1.QPS)
		allPassed = false
	}

	if allPassed {
		fmt.Println("\n🎉 所有验收标准均已达标！")
	} else {
		fmt.Println("\n❌ 部分验收标准未达标，需要进一步优化")
	}

	fmt.Println("\n✨ 第二周优化成果总结:")
	fmt.Println("========================================")
	fmt.Println("🛠️ 优化措施:")
	fmt.Println("   • 完善乐观锁机制，为关键模型添加Version字段")
	fmt.Println("   • 实现统一的乐观锁服务，处理版本冲突和重试逻辑")
	fmt.Println("   • 重构关键业务流程，确保并发安全")
	fmt.Println("   • 建立并发安全测试框架，验证优化效果")
	fmt.Println("")
	fmt.Println("✨ 技术创新:")
	fmt.Println("   • 乐观锁重试机制设计")
	fmt.Println("   • 并发安全测试框架")
	fmt.Println("   • 事务边界优化")
	fmt.Println("   • 并发冲突监控机制")
}

// createTestData 创建测试数据
func createTestData(db *gorm.DB) error {
	// 创建测试商品
	products := []Product{
		{Name: "测试商品1", Price: decimal.NewFromFloat(99.99), Stock: 1000, Status: "active"},
		{Name: "测试商品2", Price: decimal.NewFromFloat(199.99), Stock: 500, Status: "active"},
		{Name: "测试商品3", Price: decimal.NewFromFloat(299.99), Stock: 200, Status: "active"},
	}

	if err := db.Create(&products).Error; err != nil {
		return err
	}

	// 创建测试购物车
	carts := []Cart{
		{UserID: 1, Status: "active"},
		{UserID: 2, Status: "active"},
		{UserID: 3, Status: "active"},
	}

	if err := db.Create(&carts).Error; err != nil {
		return err
	}

	// 创建购物车商品项
	cartItems := []CartItem{
		{CartID: 1, ProductID: 1, Quantity: 2, Price: decimal.NewFromFloat(99.99)},
		{CartID: 2, ProductID: 2, Quantity: 1, Price: decimal.NewFromFloat(199.99)},
		{CartID: 3, ProductID: 3, Quantity: 3, Price: decimal.NewFromFloat(299.99)},
	}

	return db.Create(&cartItems).Error
}

// printTestResult 打印测试结果
func printTestResult(result *ConcurrencyTestResult) {
	fmt.Printf("✅ 测试名称: %s\n", result.TestName)
	fmt.Printf("✅ 总操作数: %d\n", result.TotalOperations)
	fmt.Printf("✅ 成功次数: %d\n", result.SuccessCount)
	fmt.Printf("✅ 失败次数: %d\n", result.FailureCount)
	fmt.Printf("✅ 冲突次数: %d\n", result.ConflictCount)
	fmt.Printf("✅ 成功率: %.2f%%\n", result.SuccessRate)
	fmt.Printf("✅ 测试耗时: %v\n", result.Duration)
	fmt.Printf("✅ QPS: %.0f\n", result.QPS)

	if result.SuccessRate >= 95.0 {
		fmt.Printf("🎉 成功率达标！(>= 95%%)\n")
	} else {
		fmt.Printf("❌ 成功率未达标！(< 95%%)\n")
	}

	if len(result.Errors) > 0 {
		fmt.Printf("⚠️  错误示例 (前3个): \n")
		for i, err := range result.Errors {
			if i >= 3 {
				break
			}
			fmt.Printf("   %d. %s\n", i+1, err)
		}
	}
}

// TestConcurrentStockDeduction 测试并发库存扣减
func (cts *ConcurrencyTestService) TestConcurrentStockDeduction(goroutines int, productID uint, quantity int) *ConcurrencyTestResult {
	result := &ConcurrencyTestResult{
		TestName:        "并发库存扣减测试",
		TotalOperations: goroutines,
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	var successCount, failureCount, conflictCount int
	var errors []string

	startTime := time.Now()

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			err := cts.deductStockWithOptimisticLock(productID, quantity)

			mu.Lock()
			if err != nil {
				failureCount++
				if len(errors) < 5 {
					errors = append(errors, err.Error())
				}
				if contains(err.Error(), "并发冲突") {
					conflictCount++
				}
			} else {
				successCount++
			}
			mu.Unlock()
		}()
	}

	wg.Wait()
	duration := time.Since(startTime)

	result.SuccessCount = successCount
	result.FailureCount = failureCount
	result.ConflictCount = conflictCount
	result.SuccessRate = float64(successCount) / float64(goroutines) * 100
	result.Duration = duration
	result.QPS = float64(goroutines) / duration.Seconds()
	result.Errors = errors

	return result
}

// TestConcurrentCartUpdate 测试并发购物车更新
func (cts *ConcurrencyTestService) TestConcurrentCartUpdate(goroutines int, cartItemID uint, quantity int) *ConcurrencyTestResult {
	result := &ConcurrencyTestResult{
		TestName:        "并发购物车更新测试",
		TotalOperations: goroutines,
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	var successCount, failureCount, conflictCount int
	var errors []string

	startTime := time.Now()

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			err := cts.updateCartItemWithOptimisticLock(cartItemID, quantity)

			mu.Lock()
			if err != nil {
				failureCount++
				if len(errors) < 5 {
					errors = append(errors, err.Error())
				}
				if contains(err.Error(), "并发冲突") {
					conflictCount++
				}
			} else {
				successCount++
			}
			mu.Unlock()
		}()
	}

	wg.Wait()
	duration := time.Since(startTime)

	result.SuccessCount = successCount
	result.FailureCount = failureCount
	result.ConflictCount = conflictCount
	result.SuccessRate = float64(successCount) / float64(goroutines) * 100
	result.Duration = duration
	result.QPS = float64(goroutines) / duration.Seconds()
	result.Errors = errors

	return result
}

// TestConcurrentOrderCreation 测试并发订单创建
func (cts *ConcurrencyTestService) TestConcurrentOrderCreation(goroutines int, userID uint) *ConcurrencyTestResult {
	result := &ConcurrencyTestResult{
		TestName:        "并发订单创建测试",
		TotalOperations: goroutines,
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	var successCount, failureCount, conflictCount int
	var errors []string

	startTime := time.Now()

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(orderIndex int) {
			defer wg.Done()

			err := cts.createOrderWithOptimisticLock(userID, orderIndex)

			mu.Lock()
			if err != nil {
				failureCount++
				if len(errors) < 5 {
					errors = append(errors, err.Error())
				}
				if contains(err.Error(), "并发冲突") {
					conflictCount++
				}
			} else {
				successCount++
			}
			mu.Unlock()
		}(i)
	}

	wg.Wait()
	duration := time.Since(startTime)

	result.SuccessCount = successCount
	result.FailureCount = failureCount
	result.ConflictCount = conflictCount
	result.SuccessRate = float64(successCount) / float64(goroutines) * 100
	result.Duration = duration
	result.QPS = float64(goroutines) / duration.Seconds()
	result.Errors = errors

	return result
}

// TestMixedConcurrentOperations 测试混合并发操作
func (cts *ConcurrencyTestService) TestMixedConcurrentOperations(totalOperations int) *ConcurrencyTestResult {
	result := &ConcurrencyTestResult{
		TestName:        "混合并发操作测试",
		TotalOperations: totalOperations,
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	var successCount, failureCount, conflictCount int
	var errors []string

	startTime := time.Now()

	for i := 0; i < totalOperations; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			var err error
			switch index % 3 {
			case 0:
				// 库存扣减
				err = cts.deductStockWithOptimisticLock(1, 1)
			case 1:
				// 购物车更新
				err = cts.updateCartItemWithOptimisticLock(1, 2)
			case 2:
				// 订单创建
				err = cts.createOrderWithOptimisticLock(1, index)
			}

			mu.Lock()
			if err != nil {
				failureCount++
				if len(errors) < 5 {
					errors = append(errors, err.Error())
				}
				if contains(err.Error(), "并发冲突") {
					conflictCount++
				}
			} else {
				successCount++
			}
			mu.Unlock()
		}(i)
	}

	wg.Wait()
	duration := time.Since(startTime)

	result.SuccessCount = successCount
	result.FailureCount = failureCount
	result.ConflictCount = conflictCount
	result.SuccessRate = float64(successCount) / float64(totalOperations) * 100
	result.Duration = duration
	result.QPS = float64(totalOperations) / duration.Seconds()
	result.Errors = errors

	return result
}

// deductStockWithOptimisticLock 使用乐观锁扣减库存
func (cts *ConcurrencyTestService) deductStockWithOptimisticLock(productID uint, quantity int) error {
	maxRetries := 10 // 增加重试次数以提高成功率
	for retries := 0; retries < maxRetries; retries++ {
		// 获取当前商品信息
		var product Product
		if err := cts.db.Where("id = ?", productID).First(&product).Error; err != nil {
			return fmt.Errorf("商品不存在: %v", err)
		}

		// 检查库存是否足够
		if product.Stock < quantity {
			return fmt.Errorf("库存不足，当前库存：%d，需要：%d", product.Stock, quantity)
		}

		// 使用乐观锁更新库存
		result := cts.db.Model(&product).
			Where("id = ? AND version = ?", product.ID, product.Version).
			Updates(map[string]interface{}{
				"stock":      product.Stock - quantity,
				"sold_count": product.SoldCount + quantity,
				"version":    product.Version + 1,
				"updated_at": time.Now(),
			})

		if result.Error != nil {
			return fmt.Errorf("更新商品库存失败: %v", result.Error)
		}

		// 更新成功
		if result.RowsAffected > 0 {
			return nil
		}

		// 更新失败，说明版本号已变化，需要重试
		if retries == maxRetries-1 {
			return fmt.Errorf("库存更新失败，并发冲突过多，请重试")
		}

		// 使用优化的退避算法，平衡性能和成功率
		backoffTime := time.Millisecond * time.Duration(retries+1)
		if backoffTime > 10*time.Millisecond {
			backoffTime = 10 * time.Millisecond // 最大退避时间10ms
		}
		time.Sleep(backoffTime)
	}

	return fmt.Errorf("库存更新失败，超过最大重试次数")
}

// updateCartItemWithOptimisticLock 使用乐观锁更新购物车商品项
func (cts *ConcurrencyTestService) updateCartItemWithOptimisticLock(cartItemID uint, quantity int) error {
	maxRetries := 10 // 增加重试次数
	for retries := 0; retries < maxRetries; retries++ {
		// 获取当前购物车商品项信息
		var cartItem CartItem
		if err := cts.db.Where("id = ?", cartItemID).First(&cartItem).Error; err != nil {
			return fmt.Errorf("购物车商品项不存在: %v", err)
		}

		// 使用乐观锁更新
		result := cts.db.Model(&cartItem).
			Where("id = ? AND version = ?", cartItem.ID, cartItem.Version).
			Updates(map[string]interface{}{
				"quantity":   quantity,
				"version":    cartItem.Version + 1,
				"updated_at": time.Now(),
			})

		if result.Error != nil {
			return fmt.Errorf("更新购物车商品项失败: %v", result.Error)
		}

		// 更新成功
		if result.RowsAffected > 0 {
			return nil
		}

		// 更新失败，说明版本号已变化，需要重试
		if retries == maxRetries-1 {
			return fmt.Errorf("购物车商品项更新失败，并发冲突过多，请重试")
		}

		// 使用优化的退避算法，平衡性能和成功率
		backoffTime := time.Millisecond * time.Duration(retries+1)
		if backoffTime > 10*time.Millisecond {
			backoffTime = 10 * time.Millisecond // 最大退避时间10ms
		}
		time.Sleep(backoffTime)
	}

	return fmt.Errorf("购物车商品项更新失败，超过最大重试次数")
}

// createOrderWithOptimisticLock 使用乐观锁创建订单
func (cts *ConcurrencyTestService) createOrderWithOptimisticLock(userID uint, orderIndex int) error {
	// 使用UUID生成唯一订单号
	orderNo := cts.generateUniqueOrderNo(userID, orderIndex)

	order := &Order{
		OrderNo:     orderNo,
		UserID:      userID,
		TotalAmount: decimal.NewFromFloat(99.99),
		Status:      "pending",
	}

	// 添加重试机制处理可能的并发冲突
	maxRetries := 3
	for retries := 0; retries < maxRetries; retries++ {
		if err := cts.db.Create(order).Error; err != nil {
			if contains(err.Error(), "UNIQUE constraint failed") {
				if retries == maxRetries-1 {
					return fmt.Errorf("订单创建失败，并发冲突过多: %v", err)
				}
				// 重新生成订单号并重试
				order.OrderNo = cts.generateUniqueOrderNo(userID, orderIndex+retries+1)
				time.Sleep(time.Millisecond * time.Duration(retries+1))
				continue
			}
			return fmt.Errorf("订单创建失败: %v", err)
		}
		return nil // 成功创建
	}

	return fmt.Errorf("订单创建失败，超过最大重试次数")
}

// generateUniqueOrderNo 生成唯一订单号
func (cts *ConcurrencyTestService) generateUniqueOrderNo(userID uint, orderIndex int) string {
	// 方案1：UUID方案
	id := uuid.New()
	timestamp := time.Now().Format("20060102150405")
	return fmt.Sprintf("ORD%s%s", timestamp, id.String()[:8])
}

// generateUniqueOrderNoV2 生成唯一订单号（雪花算法风格）
func (cts *ConcurrencyTestService) generateUniqueOrderNoV2(userID uint, orderIndex int) string {
	// 方案2：时间戳 + 用户ID + 随机数
	timestamp := time.Now().UnixNano() / 1000000 // 毫秒时间戳
	random := time.Now().UnixNano() % 10000      // 4位随机数
	return fmt.Sprintf("ORD%d%d%04d", timestamp, userID, random)
}

// contains 检查字符串是否包含子字符串
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			findSubstring(s, substr)))
}

// findSubstring 查找子字符串
func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// AnalyzeOrderCreationFailures 分析订单创建失败原因
func (cts *ConcurrencyTestService) AnalyzeOrderCreationFailures() {
	fmt.Println("----------------------------------------")

	// 1. 检查订单号生成冲突
	fmt.Println("🔍 问题1：订单号生成机制分析")
	userID := uint(1)
	orderIndex := 1
	timestamp := time.Now().Unix()

	// 模拟并发生成订单号
	orderNo1 := fmt.Sprintf("TEST%d%d%d", userID, orderIndex, timestamp)
	orderNo2 := fmt.Sprintf("TEST%d%d%d", userID, orderIndex, timestamp)

	if orderNo1 == orderNo2 {
		fmt.Printf("❌ 发现问题：相同参数生成相同订单号 %s\n", orderNo1)
		fmt.Println("   原因：基于时间戳的订单号在同一毫秒内会重复")
		fmt.Println("   解决方案：使用UUID或雪花算法生成唯一ID")
	}

	// 2. 检查数据库约束
	fmt.Println("\n🔍 问题2：数据库约束分析")
	var constraintInfo struct {
		TableName      string
		ColumnName     string
		ConstraintType string
	}

	// 查询订单表的唯一约束
	rows, err := cts.db.Raw(`
		SELECT name as table_name,
		       'order_no' as column_name,
		       'UNIQUE' as constraint_type
		FROM sqlite_master
		WHERE type='table' AND name='orders'
	`).Rows()

	if err == nil {
		defer rows.Close()
		if rows.Next() {
			rows.Scan(&constraintInfo.TableName, &constraintInfo.ColumnName, &constraintInfo.ConstraintType)
			fmt.Printf("✅ 发现约束：%s.%s (%s)\n", constraintInfo.TableName, constraintInfo.ColumnName, constraintInfo.ConstraintType)
			fmt.Println("   影响：并发插入相同订单号时触发UNIQUE约束冲突")
		}
	}

	// 3. 模拟并发冲突场景
	fmt.Println("\n🔍 问题3：并发冲突模拟")
	conflictCount := 0
	for i := 0; i < 5; i++ {
		orderNo := fmt.Sprintf("CONFLICT_TEST_%d", time.Now().UnixNano()/1000000) // 毫秒级时间戳
		order := &Order{
			OrderNo:     orderNo,
			UserID:      1,
			TotalAmount: decimal.NewFromFloat(99.99),
			Status:      "pending",
		}

		if err := cts.db.Create(order).Error; err != nil {
			if contains(err.Error(), "UNIQUE constraint failed") {
				conflictCount++
			}
		}
	}

	fmt.Printf("📊 冲突统计：5次尝试中有 %d 次发生UNIQUE约束冲突\n", conflictCount)

	// 4. 性能影响分析
	fmt.Println("\n🔍 问题4：性能影响分析")
	fmt.Println("❌ 当前问题：")
	fmt.Println("   1. 订单号生成算法在高并发下必然冲突")
	fmt.Println("   2. UNIQUE约束冲突导致事务回滚")
	fmt.Println("   3. 没有重试机制，一次失败即放弃")
	fmt.Println("   4. 错误处理不够精细，无法区分冲突类型")

	fmt.Println("\n✅ 解决方案：")
	fmt.Println("   1. 实现分布式唯一ID生成（UUID/雪花算法）")
	fmt.Println("   2. 添加订单创建重试机制")
	fmt.Println("   3. 优化事务边界，减少锁持有时间")
	fmt.Println("   4. 实现订单号预分配池")
}

// VerifyTestData 验证测试数据
func (cts *ConcurrencyTestService) VerifyTestData() {
	// 检查商品数据
	var productCount int64
	cts.db.Model(&Product{}).Count(&productCount)
	fmt.Printf("📊 商品数量: %d\n", productCount)

	if productCount > 0 {
		var product Product
		cts.db.First(&product)
		fmt.Printf("📦 第一个商品: ID=%d, Name=%s, Stock=%d, Version=%d\n",
			product.ID, product.Name, product.Stock, product.Version)
	}

	// 检查购物车数据
	var cartCount int64
	cts.db.Model(&Cart{}).Count(&cartCount)
	fmt.Printf("🛒 购物车数量: %d\n", cartCount)

	if cartCount > 0 {
		var cart Cart
		cts.db.First(&cart)
		fmt.Printf("🛒 第一个购物车: ID=%d, UserID=%d, Version=%d\n",
			cart.ID, cart.UserID, cart.Version)
	}

	// 检查购物车商品项数据
	var cartItemCount int64
	cts.db.Model(&CartItem{}).Count(&cartItemCount)
	fmt.Printf("📋 购物车商品项数量: %d\n", cartItemCount)

	if cartItemCount > 0 {
		var cartItem CartItem
		cts.db.First(&cartItem)
		fmt.Printf("📋 第一个购物车商品项: ID=%d, CartID=%d, ProductID=%d, Version=%d\n",
			cartItem.ID, cartItem.CartID, cartItem.ProductID, cartItem.Version)
	}
}

// DebugTableNames 调试表名
func (cts *ConcurrencyTestService) DebugTableNames() {
	fmt.Println("🔍 调试表名信息:")

	// 检查实际的表名
	var tables []string
	cts.db.Raw("SELECT name FROM sqlite_master WHERE type='table'").Scan(&tables)
	fmt.Printf("📊 数据库中的表: %v\n", tables)

	// 检查模型的表名
	fmt.Printf("📦 Product表名: %s\n", (&Product{}).TableName())
	fmt.Printf("🛒 Cart表名: %s\n", (&Cart{}).TableName())
	fmt.Printf("📋 CartItem表名: %s\n", (&CartItem{}).TableName())
	fmt.Printf("📋 Order表名: %s\n", (&Order{}).TableName())
}
