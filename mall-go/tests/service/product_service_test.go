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
		assert.Equal(t, "99.99", response.Price, "商品价格应该匹配")
		assert.Equal(t, 100, response.Stock, "商品库存应该匹配")

		// 验证数据库中的商品数据
		var dbProduct model.Product
		err = db.Where("id = ?", response.ID).First(&dbProduct).Error
		assert.NoError(t, err, "应该能在数据库中找到商品")
		assert.Equal(t, "测试商品", dbProduct.Name, "数据库中商品名称应该匹配")
		assert.Equal(t, "99.99", dbProduct.Price, "数据库中商品价格应该匹配")

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
			Price:       "199.99",
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
		assert.Equal(t, "199.99", response.Price, "商品价格应该匹配")
		assert.Equal(t, 50, response.Stock, "商品库存应该匹配")

		t.Logf("✅ 商品获取测试通过 - 商品ID: %d", response.ID)
	})

	t.Run("商品不存在", func(t *testing.T) {
		// 尝试获取不存在的商品
		response, err := productService.GetProduct(99999)

		// 验证错误处理
		assert.Error(t, err, "获取不存在商品应该失败")
		assert.Contains(t, err.Error(), "商品不存在", "错误信息应该包含商品不存在")
		assert.Nil(t, response, "获取失败时不应返回结果")

		t.Logf("✅ 商品不存在测试通过")
	})

	t.Run("更新商品库存", func(t *testing.T) {
		// 先创建分类和商品
		category := &model.Category{
			Name:        "库存测试分类",
			Description: "stock-test-category",
			Status:      "active",
		}
		err := db.Create(category).Error
		assert.NoError(t, err, "创建分类失败")

		testProduct := &model.Product{
			Name:        "库存测试商品",
			Description: "用于库存测试的商品",
			CategoryID:  category.ID,
			Price:       "299.99",
			Stock:       100,
			Status:      "active",
		}
		err = db.Create(testProduct).Error
		assert.NoError(t, err, "创建测试商品失败")

		// 更新库存
		newStock := 80
		err = productService.UpdateStock(testProduct.ID, newStock)

		// 验证更新结果
		assert.NoError(t, err, "更新库存应该成功")

		// 验证数据库中的库存已更新
		var updatedProduct model.Product
		err = db.Where("id = ?", testProduct.ID).First(&updatedProduct).Error
		assert.NoError(t, err, "应该能在数据库中找到商品")
		assert.Equal(t, newStock, updatedProduct.Stock, "库存应该已更新")

		t.Logf("✅ 库存更新测试通过 - 新库存: %d", updatedProduct.Stock)
	})

	t.Run("库存不足检查", func(t *testing.T) {
		// 先创建分类和商品
		category := &model.Category{
			Name:        "库存不足测试分类",
			Description: "insufficient-stock-test-category",
			Status:      "active",
		}
		err := db.Create(category).Error
		assert.NoError(t, err, "创建分类失败")

		testProduct := &model.Product{
			Name:        "库存不足测试商品",
			Description: "用于库存不足测试的商品",
			CategoryID:  category.ID,
			Price:       "399.99",
			Stock:       5, // 低库存
			Status:      "active",
		}
		err = db.Create(testProduct).Error
		assert.NoError(t, err, "创建测试商品失败")

		// 检查库存是否充足
		isAvailable := productService.CheckStock(testProduct.ID, 10) // 需要10个，但只有5个
		assert.False(t, isAvailable, "库存不足时应该返回false")

		isAvailable = productService.CheckStock(testProduct.ID, 3) // 需要3个，有5个
		assert.True(t, isAvailable, "库存充足时应该返回true")

		t.Logf("✅ 库存不足检查测试通过")
	})

	t.Run("商品状态管理", func(t *testing.T) {
		// 先创建分类和商品
		category := &model.Category{
			Name:        "状态测试分类",
			Description: "status-test-category",
			Status:      "active",
		}
		err := db.Create(category).Error
		assert.NoError(t, err, "创建分类失败")

		testProduct := &model.Product{
			Name:        "状态测试商品",
			Description: "用于状态测试的商品",
			CategoryID:  category.ID,
			Price:       "499.99",
			Stock:       20,
			Status:      "active",
		}
		err = db.Create(testProduct).Error
		assert.NoError(t, err, "创建测试商品失败")

		// 更新商品状态为下架
		err = productService.UpdateStatus(testProduct.ID, "inactive")
		assert.NoError(t, err, "更新商品状态应该成功")

		// 验证数据库中的状态已更新
		var updatedProduct model.Product
		err = db.Where("id = ?", testProduct.ID).First(&updatedProduct).Error
		assert.NoError(t, err, "应该能在数据库中找到商品")
		assert.Equal(t, "inactive", updatedProduct.Status, "商品状态应该已更新")

		t.Logf("✅ 商品状态管理测试通过")
	})

	t.Run("商品价格计算", func(t *testing.T) {
		// 先创建分类和商品
		category := &model.Category{
			Name:        "价格测试分类",
			Description: "price-test-category",
			Status:      "active",
		}
		err := db.Create(category).Error
		assert.NoError(t, err, "创建分类失败")

		testProduct := &model.Product{
			Name:        "价格测试商品",
			Description: "用于价格测试的商品",
			CategoryID:  category.ID,
			Price:       "100.00",
			OriginPrice: "120.00",
			Stock:       30,
			Status:      "active",
		}
		err = db.Create(testProduct).Error
		assert.NoError(t, err, "创建测试商品失败")

		// 计算折扣
		discount := productService.CalculateDiscount(testProduct.ID)
		expectedDiscount := (120.00 - 100.00) / 120.00 * 100 // 约16.67%

		// 允许小的浮点数误差
		assert.InDelta(t, expectedDiscount, discount, 0.01, "折扣计算应该正确")

		t.Logf("✅ 商品价格计算测试通过 - 折扣: %.2f%%", discount)
	})
}
