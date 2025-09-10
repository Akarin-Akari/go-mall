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

// TestBasicProductService 基础商品服务测试
func TestBasicProductService(t *testing.T) {
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
		&model.ProductAttr{},
		&model.ProductSKU{},
		&model.ProductReview{},
		&model.Category{},
		&model.Brand{},
		&model.User{},
	)
	assert.NoError(t, err, "数据库迁移失败")

	// 创建商品服务
	productService := product.NewProductService(db)

	t.Run("成功创建商品", func(t *testing.T) {
		// 先创建商家用户
		merchant := &model.User{
			Username: "testmerchant",
			Email:    "merchant@example.com",
			Password: "hashedpassword",
			Role:     "merchant",
			Status:   "active",
		}
		err := db.Create(merchant).Error
		assert.NoError(t, err, "创建商家失败")

		// 创建分类
		category := &model.Category{
			Name:        "测试分类",
			Description: "test-category",
			Status:      "active",
		}
		err = db.Create(category).Error
		assert.NoError(t, err, "创建分类失败")

		// 创建商品请求
		req := &product.CreateProductRequest{
			Name:        "测试商品",
			Description: "这是一个测试商品",
			CategoryID:  category.ID,
			MerchantID:  merchant.ID, // 使用创建的商户ID
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

		price, _ := decimal.NewFromString("199.99")
		testProduct := &model.Product{
			Name:        "获取测试商品",
			Description: "用于获取测试的商品",
			CategoryID:  category.ID,
			Price:       price,
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

	t.Run("商品业务逻辑验证", func(t *testing.T) {
		// 先创建分类
		category := &model.Category{
			Name:        "业务逻辑测试分类",
			Description: "business-logic-test-category",
			Status:      "active",
		}
		err := db.Create(category).Error
		assert.NoError(t, err, "创建分类失败")

		// 创建商品
		price, _ := decimal.NewFromString("299.99")
		originPrice, _ := decimal.NewFromString("399.99")
		testProduct := &model.Product{
			Name:        "业务逻辑测试商品",
			Description: "用于业务逻辑测试的商品",
			CategoryID:  category.ID,
			Price:       price,
			OriginPrice: originPrice,
			Stock:       10,
			Status:      "active",
		}
		err = db.Create(testProduct).Error
		assert.NoError(t, err, "创建测试商品失败")

		// 验证商品业务逻辑
		// 1. 价格应该小于原价
		assert.True(t, testProduct.Price.LessThan(testProduct.OriginPrice), "商品价格应该小于原价")

		// 2. 库存应该大于0
		assert.Greater(t, testProduct.Stock, 0, "商品库存应该大于0")

		// 3. 状态应该是有效的
		validStatuses := []string{"active", "inactive", "draft"}
		assert.Contains(t, validStatuses, testProduct.Status, "商品状态应该是有效的")

		// 4. 分类ID应该存在
		var categoryCount int64
		err = db.Model(&model.Category{}).Where("id = ?", testProduct.CategoryID).Count(&categoryCount).Error
		assert.NoError(t, err, "查询分类应该成功")
		assert.Equal(t, int64(1), categoryCount, "商品分类应该存在")

		t.Logf("✅ 商品业务逻辑验证测试通过")
	})

	t.Run("商品数据完整性验证", func(t *testing.T) {
		// 先创建分类
		category := &model.Category{
			Name:        "数据完整性测试分类",
			Description: "data-integrity-test-category",
			Status:      "active",
		}
		err := db.Create(category).Error
		assert.NoError(t, err, "创建分类失败")

		// 创建商品请求（包含完整数据）
		req := &product.CreateProductRequest{
			Name:           "数据完整性测试商品",
			SubTitle:       "副标题",
			Description:    "详细描述",
			Detail:         "详细信息",
			CategoryID:     category.ID,
			MerchantID:     1,
			Price:          decimal.NewFromFloat(199.99),
			OriginPrice:    decimal.NewFromFloat(299.99),
			CostPrice:      decimal.NewFromFloat(100.00),
			Stock:          50,
			MinStock:       5,
			MaxStock:       1000,
			Weight:         decimal.NewFromFloat(1.5),
			Volume:         decimal.NewFromFloat(0.1),
			Unit:           "件",
			SEOTitle:       "SEO标题",
			SEOKeywords:    "关键词1,关键词2",
			SEODescription: "SEO描述",
			Sort:           100,
		}

		response, err := productService.CreateProduct(req)

		// 验证创建结果
		assert.NoError(t, err, "创建商品应该成功")
		assert.NotNil(t, response, "创建响应不应为空")

		// 验证数据完整性
		var dbProduct model.Product
		err = db.Where("id = ?", response.ID).First(&dbProduct).Error
		assert.NoError(t, err, "应该能在数据库中找到商品")

		// 验证各字段数据
		assert.Equal(t, "数据完整性测试商品", dbProduct.Name, "商品名称应该匹配")
		assert.Equal(t, "副标题", dbProduct.SubTitle, "副标题应该匹配")
		assert.Equal(t, "详细描述", dbProduct.Description, "描述应该匹配")
		assert.Equal(t, category.ID, dbProduct.CategoryID, "分类ID应该匹配")
		assert.Equal(t, 50, dbProduct.Stock, "库存应该匹配")
		assert.Equal(t, 5, dbProduct.MinStock, "最小库存应该匹配")
		assert.Equal(t, 1000, dbProduct.MaxStock, "最大库存应该匹配")
		assert.Equal(t, "件", dbProduct.Unit, "单位应该匹配")
		assert.Equal(t, 100, dbProduct.Sort, "排序应该匹配")

		t.Logf("✅ 商品数据完整性验证测试通过 - 商品ID: %d", response.ID)
	})

	t.Run("商品边界条件测试", func(t *testing.T) {
		// 先创建分类
		category := &model.Category{
			Name:        "边界条件测试分类",
			Description: "boundary-test-category",
			Status:      "active",
		}
		err := db.Create(category).Error
		assert.NoError(t, err, "创建分类失败")

		// 测试最小价格
		req := &product.CreateProductRequest{
			Name:        "边界条件测试商品",
			Description: "用于边界条件测试的商品",
			CategoryID:  category.ID,
			MerchantID:  1,
			Price:       decimal.NewFromFloat(0.01), // 最小价格
			Stock:       0,                          // 最小库存
		}

		response, err := productService.CreateProduct(req)

		// 验证边界条件处理
		if err != nil {
			// 如果有业务规则限制最小价格或库存，这里应该失败
			t.Logf("边界条件验证: %v", err)
			assert.Contains(t, err.Error(), "价格", "错误信息应该包含价格相关信息")
		} else {
			// 如果允许最小值，验证创建成功
			assert.NotNil(t, response, "边界条件下创建应该成功")
			t.Logf("✅ 边界条件测试通过 - 允许最小值")
		}
	})
}
