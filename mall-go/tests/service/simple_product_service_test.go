package service

import (
	"testing"

	"mall-go/internal/config"
	"mall-go/internal/model"
	"mall-go/pkg/product"
	testConfig "mall-go/tests/config"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

// TestSimpleProductService 简单的商品服务测试
func TestSimpleProductService(t *testing.T) {
	// 初始化配置
	config.GlobalConfig = config.Config{
		JWT: config.JWTConfig{
			Secret: "test-secret-key-for-service-layer-testing",
			Expire: "24h",
		},
	}

	// 初始化测试数据库
	db := testConfig.SetupTestDB()
	defer testConfig.CleanupTestDB(db)

	// 自动迁移
	err := db.AutoMigrate(
		&model.Product{},
		&model.ProductImage{},
		&model.Category{},
		&model.User{},
	)
	assert.NoError(t, err, "数据库迁移失败")

	// 创建商品服务
	productService := product.NewProductService(db)

	t.Run("成功创建商品", func(t *testing.T) {
		// 先创建分类
		category := &model.Category{
			Name:        "测试分类",
			Description: "test-category",
			Status:      "active",
		}
		err := db.Create(category).Error
		assert.NoError(t, err, "创建分类失败")

		// 创建商品请求
		req := &product.CreateProductRequest{
			Name:        "测试商品",
			Description: "这是一个测试商品",
			CategoryID:  category.ID,
			MerchantID:  1, // 添加必需的商户ID
			Price:       decimal.NewFromFloat(99.99),
			Stock:       100,
		}

		response, err := productService.CreateProduct(req)

		// 验证创建结果
		if err != nil {
			t.Logf("创建商品错误: %v", err)
			t.FailNow()
		}

		assert.NotNil(t, response, "创建响应不应为空")
		assert.NotZero(t, response.ID, "商品ID应该不为0")
		assert.Equal(t, "测试商品", response.Name, "商品名称应该匹配")

		// 验证数据库中的商品数据
		var dbProduct model.Product
		err = db.Where("id = ?", response.ID).First(&dbProduct).Error
		assert.NoError(t, err, "应该能在数据库中找到商品")
		assert.Equal(t, "测试商品", dbProduct.Name, "数据库中商品名称应该匹配")

		t.Logf("✅ 商品创建测试通过 - 商品ID: %d", response.ID)
	})

	t.Run("获取商品详情", func(t *testing.T) {
		// 先创建分类和商品
		category := &model.Category{
			Name:        "测试分类2",
			Description: "test-category-2",
			Status:      "active",
		}
		err := db.Create(category).Error
		assert.NoError(t, err, "创建分类失败")

		testProduct := &model.Product{
			Name:        "获取测试商品",
			Description: "用于获取测试的商品",
			CategoryID:  category.ID,
			Price:       decimal.NewFromString("199.99"),
			Stock:       50,
			Status:      "active",
		}
		err = db.Create(testProduct).Error
		assert.NoError(t, err, "创建测试商品失败")

		// 获取商品详情
		response, err := productService.GetProduct(testProduct.ID)

		// 验证获取结果
		assert.NoError(t, err, "获取商品详情应该成功")
		assert.NotNil(t, response, "获取响应不应为空")
		assert.Equal(t, testProduct.ID, response.ID, "商品ID应该匹配")
		assert.Equal(t, "获取测试商品", response.Name, "商品名称应该匹配")

		t.Logf("✅ 商品获取测试通过 - 商品ID: %d", response.ID)
	})

	t.Run("商品不存在", func(t *testing.T) {
		// 尝试获取不存在的商品
		response, err := productService.GetProduct(99999)

		// 验证错误处理
		assert.Error(t, err, "获取不存在商品应该失败")
		assert.Nil(t, response, "获取失败时不应返回结果")

		t.Logf("✅ 商品不存在测试通过")
	})

	t.Run("商品列表查询", func(t *testing.T) {
		// 先创建分类
		category := &model.Category{
			Name:        "列表测试分类",
			Description: "list-test-category",
			Status:      "active",
		}
		err := db.Create(category).Error
		assert.NoError(t, err, "创建分类失败")

		// 创建多个测试商品
		products := []*model.Product{
			{
				Name:        "列表测试商品1",
				Description: "用于列表测试的商品1",
				CategoryID:  category.ID,
				Price:       decimal.NewFromString("100.00"),
				Stock:       10,
				Status:      "active",
			},
			{
				Name:        "列表测试商品2",
				Description: "用于列表测试的商品2",
				CategoryID:  category.ID,
				Price:       decimal.NewFromString("200.00"),
				Stock:       20,
				Status:      "active",
			},
		}

		for _, p := range products {
			err = db.Create(p).Error
			assert.NoError(t, err, "创建测试商品失败")
		}

		// 查询商品列表
		req := &product.ListProductRequest{
			Page:     1,
			PageSize: 10,
			Status:   "active",
		}

		response, err := productService.ListProducts(req)

		// 验证查询结果
		assert.NoError(t, err, "查询商品列表应该成功")
		assert.NotNil(t, response, "查询响应不应为空")
		assert.GreaterOrEqual(t, len(response.Products), 2, "应该至少有2个商品")

		t.Logf("✅ 商品列表查询测试通过 - 商品数量: %d", len(response.Products))
	})

	t.Run("商品搜索功能", func(t *testing.T) {
		// 先创建分类
		category := &model.Category{
			Name:        "搜索测试分类",
			Description: "search-test-category",
			Status:      "active",
		}
		err := db.Create(category).Error
		assert.NoError(t, err, "创建分类失败")

		// 创建测试商品
		testProduct := &model.Product{
			Name:        "搜索测试商品iPhone",
			Description: "用于搜索测试的iPhone商品",
			CategoryID:  category.ID,
			Price:       decimal.NewFromString("5999.00"),
			Stock:       5,
			Status:      "active",
		}
		err = db.Create(testProduct).Error
		assert.NoError(t, err, "创建测试商品失败")

		// 搜索商品
		req := &product.SearchProductRequest{
			Keyword:  "iPhone",
			Page:     1,
			PageSize: 10,
		}

		response, err := productService.SearchProducts(req)

		// 验证搜索结果
		assert.NoError(t, err, "搜索商品应该成功")
		assert.NotNil(t, response, "搜索响应不应为空")
		assert.GreaterOrEqual(t, len(response.Products), 1, "应该至少找到1个商品")

		// 验证搜索结果包含关键词
		found := false
		for _, p := range response.Products {
			if p.Name == "搜索测试商品iPhone" {
				found = true
				break
			}
		}
		assert.True(t, found, "搜索结果应该包含目标商品")

		t.Logf("✅ 商品搜索功能测试通过 - 找到商品数量: %d", len(response.Products))
	})

	t.Run("商品分类筛选", func(t *testing.T) {
		// 先创建分类
		category := &model.Category{
			Name:        "筛选测试分类",
			Description: "filter-test-category",
			Status:      "active",
		}
		err := db.Create(category).Error
		assert.NoError(t, err, "创建分类失败")

		// 创建该分类下的商品
		testProduct := &model.Product{
			Name:        "筛选测试商品",
			Description: "用于筛选测试的商品",
			CategoryID:  category.ID,
			Price:       decimal.NewFromString("299.00"),
			Stock:       15,
			Status:      "active",
		}
		err = db.Create(testProduct).Error
		assert.NoError(t, err, "创建测试商品失败")

		// 按分类筛选商品
		req := &product.ListProductRequest{
			CategoryID: category.ID,
			Page:       1,
			PageSize:   10,
		}

		response, err := productService.ListProducts(req)

		// 验证筛选结果
		assert.NoError(t, err, "按分类筛选商品应该成功")
		assert.NotNil(t, response, "筛选响应不应为空")
		assert.GreaterOrEqual(t, len(response.Products), 1, "应该至少找到1个商品")

		// 验证所有商品都属于指定分类
		for _, p := range response.Products {
			assert.Equal(t, category.ID, p.CategoryID, "商品应该属于指定分类")
		}

		t.Logf("✅ 商品分类筛选测试通过 - 分类ID: %d, 商品数量: %d", category.ID, len(response.Products))
	})
}
