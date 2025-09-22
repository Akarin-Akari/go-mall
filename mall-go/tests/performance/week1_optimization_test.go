package performance

import (
	"fmt"
	"log"
	"sync"
	"testing"
	"time"

	"mall-go/internal/model"
	"mall-go/pkg/product"

	"github.com/shopspring/decimal"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Week1OptimizationTest 第一周数据模型优化测试
type Week1OptimizationTest struct {
	db             *gorm.DB
	productService *product.ProductService
	syncService    *product.SyncService
}

// NewWeek1OptimizationTest 创建第一周优化测试
func NewWeek1OptimizationTest() *Week1OptimizationTest {
	// 创建内存数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatal("连接数据库失败:", err)
	}

	// 自动迁移
	db.AutoMigrate(
		&model.User{},
		&model.Category{},
		&model.Brand{},
		&model.Product{},
		&model.ProductImage{},
		&model.ProductAttr{},
		&model.ProductSKU{},
		&model.ProductReview{},
	)

	return &Week1OptimizationTest{
		db:             db,
		productService: product.NewProductService(db),
		syncService:    product.NewSyncService(db),
	}
}

// setupTestData 创建测试数据
func (w *Week1OptimizationTest) setupTestData() error {
	// 创建测试分类
	categories := []model.Category{
		{Name: "电子产品", Description: "各类电子设备", Status: "active"},
		{Name: "服装鞋帽", Description: "时尚服饰", Status: "active"},
		{Name: "家居用品", Description: "家庭生活用品", Status: "active"},
		{Name: "运动户外", Description: "运动健身用品", Status: "active"},
		{Name: "美妆护肤", Description: "美容护肤产品", Status: "active"},
	}

	for _, category := range categories {
		if err := w.db.Create(&category).Error; err != nil {
			return err
		}
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
		if err := w.db.Create(&brand).Error; err != nil {
			return err
		}
	}

	// 创建测试商家
	merchants := []model.User{
		{Username: "merchant1", Email: "merchant1@example.com", Role: "merchant", Status: "active"},
		{Username: "merchant2", Email: "merchant2@example.com", Role: "merchant", Status: "active"},
		{Username: "merchant3", Email: "merchant3@example.com", Role: "merchant", Status: "active"},
	}

	for _, merchant := range merchants {
		if err := w.db.Create(&merchant).Error; err != nil {
			return err
		}
	}

	// 创建测试商品（1000个）
	for i := 1; i <= 1000; i++ {
		price, _ := decimal.NewFromString(fmt.Sprintf("%.2f", float64(i)*10.99))
		product := model.Product{
			Name:         fmt.Sprintf("测试商品 %d", i),
			Description:  fmt.Sprintf("这是第 %d 个测试商品", i),
			CategoryID:   uint((i % 5) + 1),
			BrandID:      uint((i % 5) + 1),
			MerchantID:   uint((i % 3) + 1),
			Price:        price,
			Stock:        100,
			Status:       "active",
			CategoryName: categories[(i % 5)].Name,
			BrandName:    brands[(i % 5)].Name,
			MerchantName: merchants[(i % 3)].Username,
		}

		if err := w.db.Create(&product).Error; err != nil {
			return err
		}

		// 为每个商品创建图片
		image := model.ProductImage{
			ProductID: product.ID,
			URL:       fmt.Sprintf("https://example.com/product%d.jpg", i),
			IsMain:    true,
			Sort:      1,
		}
		w.db.Create(&image)
	}

	return nil
}

// TestProductQueryPerformance 测试商品查询性能
func TestProductQueryPerformance(t *testing.T) {
	test := NewWeek1OptimizationTest()

	// 设置测试数据
	if err := test.setupTestData(); err != nil {
		t.Fatalf("设置测试数据失败: %v", err)
	}

	t.Run("单个商品查询性能测试", func(t *testing.T) {
		test.testSingleProductQuery(t)
	})

	t.Run("商品列表查询性能测试", func(t *testing.T) {
		test.testProductListQuery(t)
	})

	t.Run("商品搜索性能测试", func(t *testing.T) {
		test.testProductSearch(t)
	})

	t.Run("并发查询性能测试", func(t *testing.T) {
		test.testConcurrentQuery(t)
	})

	t.Run("数据同步性能测试", func(t *testing.T) {
		test.testDataSync(t)
	})
}

// testSingleProductQuery 测试单个商品查询
func (w *Week1OptimizationTest) testSingleProductQuery(t *testing.T) {
	fmt.Println("\n🔍 单个商品查询性能测试")

	// 测试优化后的查询
	start := time.Now()
	successCount := 0
	totalQueries := 100

	for i := 1; i <= totalQueries; i++ {
		product, err := w.productService.GetProduct(uint(i))
		if err == nil && product != nil {
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

	// 验收标准：成功率 > 95%，平均响应时间 < 50ms
	if successRate < 95.0 {
		t.Errorf("查询成功率 %.2f%% 低于预期 95%%", successRate)
	}

	if avgResponseTime > 50*time.Millisecond {
		t.Errorf("平均响应时间 %v 超过预期 50ms", avgResponseTime)
	}
}

// testProductListQuery 测试商品列表查询
func (w *Week1OptimizationTest) testProductListQuery(t *testing.T) {
	fmt.Println("\n📋 商品列表查询性能测试")

	req := &product.ProductListRequest{
		Page:     1,
		PageSize: 20,
		Status:   "active",
	}

	start := time.Now()
	products, total, err := w.productService.GetProductList(req)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("商品列表查询失败: %v", err)
	}

	fmt.Printf("✅ 查询结果: %d 个商品，总计 %d 个\n", len(products), total)
	fmt.Printf("✅ 响应时间: %v\n", duration)

	// 验收标准：响应时间 < 100ms
	if duration > 100*time.Millisecond {
		t.Errorf("列表查询响应时间 %v 超过预期 100ms", duration)
	}

	// 验证冗余字段是否正确填充
	for _, product := range products {
		if product.CategoryName == "" || product.BrandName == "" || product.MerchantName == "" {
			t.Errorf("商品 %d 的冗余字段未正确填充", product.ID)
		}
	}
}

// testProductSearch 测试商品搜索
func (w *Week1OptimizationTest) testProductSearch(t *testing.T) {
	fmt.Println("\n🔍 商品搜索性能测试")

	req := &product.ProductListRequest{
		Page:     1,
		PageSize: 20,
		Keyword:  "Apple",
		Status:   "active",
	}

	start := time.Now()
	products, total, err := w.productService.GetProductList(req)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("商品搜索失败: %v", err)
	}

	fmt.Printf("✅ 搜索结果: %d 个商品，总计 %d 个\n", len(products), total)
	fmt.Printf("✅ 响应时间: %v\n", duration)

	// 验收标准：响应时间 < 100ms
	if duration > 100*time.Millisecond {
		t.Errorf("搜索响应时间 %v 超过预期 100ms", duration)
	}
}

// testConcurrentQuery 测试并发查询
func (w *Week1OptimizationTest) testConcurrentQuery(t *testing.T) {
	fmt.Println("\n⚡ 并发查询性能测试")

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
				_, err := w.productService.GetProduct(productID)
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

	// 验收标准：成功率 > 95%，QPS > 1000
	if successRate < 95.0 {
		t.Errorf("并发查询成功率 %.2f%% 低于预期 95%%", successRate)
	}

	if qps < 1000 {
		t.Errorf("QPS %.2f 低于预期 1000", qps)
	}
}

// testDataSync 测试数据同步
func (w *Week1OptimizationTest) testDataSync(t *testing.T) {
	fmt.Println("\n🔄 数据同步性能测试")

	start := time.Now()
	err := w.syncService.SyncProductRedundantFields()
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("数据同步失败: %v", err)
	}

	fmt.Printf("✅ 数据同步完成，耗时: %v\n", duration)

	// 验证数据一致性
	validation, err := w.syncService.ValidateRedundantData()
	if err != nil {
		t.Fatalf("验证数据一致性失败: %v", err)
	}

	totalMismatch := validation["category_mismatch"] + validation["brand_mismatch"] + validation["merchant_mismatch"]
	fmt.Printf("✅ 数据一致性验证: %d 条不一致记录\n", totalMismatch)

	if totalMismatch > 0 {
		t.Errorf("发现 %d 条数据不一致", totalMismatch)
	}
}
