package main

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"mall-go/internal/model"
	"mall-go/pkg/product"

	"github.com/shopspring/decimal"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Week1PerformanceTest 第一周性能测试
func main() {
	fmt.Println("🎯 Mall-Go 第一周数据模型优化性能测试")
	fmt.Println(strings.Repeat("=", 60))

	// 创建测试环境
	db, err := gorm.Open(sqlite.Open("week1_test.db"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatal("连接数据库失败:", err)
	}

	// 自动迁移
	fmt.Println("🚀 开始数据库迁移...")
	err = db.AutoMigrate(
		&model.User{},
		&model.Category{},
		&model.Brand{},
		&model.Product{},
		&model.ProductImage{},
		&model.ProductAttr{},
		&model.ProductSKU{},
		&model.ProductReview{},
	)
	if err != nil {
		log.Fatal("数据库迁移失败:", err)
	}
	fmt.Println("✅ 数据库迁移完成")

	// 创建服务
	productService := product.NewProductService(db)
	syncService := product.NewSyncService(db)

	// 创建测试数据
	fmt.Println("🧪 开始创建测试数据...")
	if err := setupTestData(db); err != nil {
		log.Fatal("创建测试数据失败:", err)
	}
	fmt.Println("✅ 测试数据创建完成")

	// 同步冗余数据
	fmt.Println("🔄 开始同步冗余数据...")
	if err := syncService.SyncProductRedundantFields(); err != nil {
		log.Fatal("同步冗余数据失败:", err)
	}
	fmt.Println("✅ 冗余数据同步完成")

	// 执行性能测试
	fmt.Println("\n🔥 开始性能测试...")

	// 测试1：单个商品查询性能
	testSingleProductQuery(productService)

	// 测试2：商品列表查询性能
	testProductListQuery(productService)

	// 测试3：商品搜索性能
	testProductSearch(productService)

	// 测试4：并发查询性能
	testConcurrentQuery(productService)

	// 测试5：数据一致性验证
	testDataConsistency(syncService)

	fmt.Println("\n🎉 第一周性能优化测试完成！")
	printSummary()
}

// setupTestData 创建测试数据
func setupTestData(db *gorm.DB) error {
	// 创建测试分类
	categories := []model.Category{
		{Name: "电子产品", Description: "各类电子设备", Status: "active"},
		{Name: "服装鞋帽", Description: "时尚服饰", Status: "active"},
		{Name: "家居用品", Description: "家庭生活用品", Status: "active"},
		{Name: "运动户外", Description: "运动健身用品", Status: "active"},
		{Name: "美妆护肤", Description: "美容护肤产品", Status: "active"},
	}

	for _, category := range categories {
		db.Create(&category)
	}

	// 创建测试品牌
	brands := []model.Brand{
		{Name: "Apple", Description: "苹果公司", Status: "active"},
		{Name: "Nike", Description: "耐克运动品牌", Status: "active"},
		{Name: "Samsung", Description: "三星电子", Status: "active"},
		{Name: "Adidas", Description: "阿迪达斯", Status: "active"},
		{Name: "Huawei", Description: "华为技术", Status: "active"},
	}

	for _, brand := range brands {
		db.Create(&brand)
	}

	// 创建测试商家
	merchants := []model.User{
		{Username: "merchant1", Email: "merchant1@example.com", Role: "merchant", Status: "active"},
		{Username: "merchant2", Email: "merchant2@example.com", Role: "merchant", Status: "active"},
		{Username: "merchant3", Email: "merchant3@example.com", Role: "merchant", Status: "active"},
	}

	for _, merchant := range merchants {
		db.Create(&merchant)
	}

	// 创建测试商品（1000个）
	fmt.Println("📦 创建1000个测试商品...")
	for i := 1; i <= 1000; i++ {
		price, _ := decimal.NewFromString(fmt.Sprintf("%.2f", float64(i)*10.99))
		product := model.Product{
			Name:        fmt.Sprintf("测试商品 %d", i),
			Description: fmt.Sprintf("这是第 %d 个测试商品", i),
			CategoryID:  uint((i % 5) + 1),
			BrandID:     uint((i % 5) + 1),
			MerchantID:  uint((i % 3) + 1),
			Price:       price,
			Stock:       100,
			Status:      "active",
		}

		db.Create(&product)

		// 为每个商品创建图片
		image := model.ProductImage{
			ProductID: product.ID,
			URL:       fmt.Sprintf("https://example.com/product%d.jpg", i),
			IsMain:    true,
			Sort:      1,
		}
		db.Create(&image)

		if i%100 == 0 {
			fmt.Printf("已创建 %d 个商品...\n", i)
		}
	}

	return nil
}

// testSingleProductQuery 测试单个商品查询
func testSingleProductQuery(productService *product.ProductService) {
	fmt.Println("\n🔍 测试1：单个商品查询性能")
	fmt.Println(strings.Repeat("-", 40))

	totalQueries := 100
	successCount := 0

	start := time.Now()
	for i := 1; i <= totalQueries; i++ {
		_, err := productService.GetProduct(uint(i))
		if err == nil {
			successCount++
		}
	}
	duration := time.Since(start)

	successRate := float64(successCount) / float64(totalQueries) * 100
	avgResponseTime := duration / time.Duration(totalQueries)

	fmt.Printf("✅ 查询结果: %d/%d 成功\n", successCount, totalQueries)
	fmt.Printf("✅ 成功率: %.2f%%\n", successRate)
	fmt.Printf("✅ 平均响应时间: %v\n", avgResponseTime)
	fmt.Printf("✅ 总耗时: %v\n", duration)

	// 验收标准检查
	if successRate >= 95.0 {
		fmt.Printf("🎉 成功率达标！(>= 95%%)\n")
	} else {
		fmt.Printf("❌ 成功率未达标！(< 95%%)\n")
	}

	if avgResponseTime <= 50*time.Millisecond {
		fmt.Printf("🎉 响应时间达标！(<= 50ms)\n")
	} else {
		fmt.Printf("❌ 响应时间未达标！(> 50ms)\n")
	}
}

// testProductListQuery 测试商品列表查询
func testProductListQuery(productService *product.ProductService) {
	fmt.Println("\n📋 测试2：商品列表查询性能")
	fmt.Println(strings.Repeat("-", 40))

	req := &product.ProductListRequest{
		Page:     1,
		PageSize: 20,
		Status:   "active",
	}

	start := time.Now()
	products, total, err := productService.GetProductList(req)
	duration := time.Since(start)

	if err != nil {
		fmt.Printf("❌ 查询失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 查询结果: %d 个商品，总计 %d 个\n", len(products), total)
	fmt.Printf("✅ 响应时间: %v\n", duration)

	// 验收标准检查
	if duration <= 100*time.Millisecond {
		fmt.Printf("🎉 响应时间达标！(<= 100ms)\n")
	} else {
		fmt.Printf("❌ 响应时间未达标！(> 100ms)\n")
	}

	// 验证冗余字段
	redundantFieldsOK := true
	for _, product := range products {
		if product.CategoryName == "" || product.BrandName == "" || product.MerchantName == "" {
			redundantFieldsOK = false
			break
		}
	}

	if redundantFieldsOK {
		fmt.Printf("🎉 冗余字段填充正确！\n")
	} else {
		fmt.Printf("❌ 冗余字段填充有误！\n")
	}
}

// testProductSearch 测试商品搜索
func testProductSearch(productService *product.ProductService) {
	fmt.Println("\n🔍 测试3：商品搜索性能")
	fmt.Println(strings.Repeat("-", 40))

	req := &product.ProductListRequest{
		Page:     1,
		PageSize: 20,
		Keyword:  "Apple",
		Status:   "active",
	}

	start := time.Now()
	products, total, err := productService.GetProductList(req)
	duration := time.Since(start)

	if err != nil {
		fmt.Printf("❌ 搜索失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 搜索结果: %d 个商品，总计 %d 个\n", len(products), total)
	fmt.Printf("✅ 响应时间: %v\n", duration)

	// 验收标准检查
	if duration <= 100*time.Millisecond {
		fmt.Printf("🎉 搜索响应时间达标！(<= 100ms)\n")
	} else {
		fmt.Printf("❌ 搜索响应时间未达标！(> 100ms)\n")
	}
}

// testConcurrentQuery 测试并发查询
func testConcurrentQuery(productService *product.ProductService) {
	fmt.Println("\n⚡ 测试4：并发查询性能")
	fmt.Println(strings.Repeat("-", 40))

	concurrency := 50
	queriesPerGoroutine := 20
	totalQueries := concurrency * queriesPerGoroutine

	var wg sync.WaitGroup
	var mu sync.Mutex
	successCount := 0

	start := time.Now()

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			localSuccess := 0

			for j := 1; j <= queriesPerGoroutine; j++ {
				productID := uint((goroutineID*queriesPerGoroutine+j)%1000 + 1)
				_, err := productService.GetProduct(productID)
				if err == nil {
					localSuccess++
				}
			}

			mu.Lock()
			successCount += localSuccess
			mu.Unlock()
		}(i)
	}

	wg.Wait()
	duration := time.Since(start)

	successRate := float64(successCount) / float64(totalQueries) * 100
	qps := float64(totalQueries) / duration.Seconds()

	fmt.Printf("✅ 并发查询结果: %d/%d 成功\n", successCount, totalQueries)
	fmt.Printf("✅ 成功率: %.2f%%\n", successRate)
	fmt.Printf("✅ QPS: %.2f\n", qps)
	fmt.Printf("✅ 总耗时: %v\n", duration)

	// 验收标准检查
	if successRate >= 95.0 {
		fmt.Printf("🎉 并发成功率达标！(>= 95%%)\n")
	} else {
		fmt.Printf("❌ 并发成功率未达标！(< 95%%)\n")
	}

	if qps >= 1000 {
		fmt.Printf("🎉 QPS达标！(>= 1000)\n")
	} else {
		fmt.Printf("❌ QPS未达标！(< 1000)\n")
	}
}

// testDataConsistency 测试数据一致性
func testDataConsistency(syncService *product.SyncService) {
	fmt.Println("\n🔄 测试5：数据一致性验证")
	fmt.Println(strings.Repeat("-", 40))

	validation, err := syncService.ValidateRedundantData()
	if err != nil {
		fmt.Printf("❌ 验证失败: %v\n", err)
		return
	}

	totalMismatch := validation["category_mismatch"] + validation["brand_mismatch"] + validation["merchant_mismatch"]
	fmt.Printf("✅ 数据一致性验证完成\n")
	fmt.Printf("✅ 分类不一致: %d 条\n", validation["category_mismatch"])
	fmt.Printf("✅ 品牌不一致: %d 条\n", validation["brand_mismatch"])
	fmt.Printf("✅ 商家不一致: %d 条\n", validation["merchant_mismatch"])
	fmt.Printf("✅ 总计不一致: %d 条\n", totalMismatch)

	if totalMismatch == 0 {
		fmt.Printf("🎉 数据一致性验证通过！\n")
	} else {
		fmt.Printf("❌ 发现数据不一致！\n")
	}
}

// printSummary 打印测试总结
func printSummary() {
	fmt.Println("\n📊 第一周优化成果总结")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("🎯 优化目标:")
	fmt.Println("   • 商品查询成功率: 1.2% → 95%+")
	fmt.Println("   • 查询平均响应时间: >1000ms → <50ms")
	fmt.Println("   • 复杂查询响应时间: >2000ms → <100ms")
	fmt.Println("   • 并发查询QPS: <200 → >1000")
	fmt.Println("")
	fmt.Println("🛠️ 优化措施:")
	fmt.Println("   • 添加冗余字段(CategoryName, BrandName, MerchantName)")
	fmt.Println("   • 重构查询逻辑，减少复杂JOIN操作")
	fmt.Println("   • 创建性能索引，优化查询路径")
	fmt.Println("   • 实现分步查询策略，按需加载关联数据")
	fmt.Println("   • 建立数据同步机制，确保冗余字段一致性")
	fmt.Println("")
	fmt.Println("✨ 技术创新:")
	fmt.Println("   • 数据冗余策略设计")
	fmt.Println("   • 自动数据同步服务")
	fmt.Println("   • 性能测试框架")
	fmt.Println("   • 数据一致性验证机制")
}
