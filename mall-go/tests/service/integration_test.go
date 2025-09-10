package service

import (
	"fmt"
	"testing"

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

// TestServiceIntegration 跨服务集成测试
func TestServiceIntegration(t *testing.T) {
	// 初始化配置
	config.GlobalConfig = config.Config{
		JWT: config.JWTConfig{
			Secret: "test-secret-key-for-integration-testing",
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
	userService := user.NewUserService(db)
	productService := product.NewProductService(db)
	cartService := cart.NewCartService(db)

	t.Run("用户-商品集成测试", func(t *testing.T) {
		// 1. 用户注册
		registerReq := &model.RegisterRequest{
			Username: "integrationuser1",
			Email:    "integration1@example.com",
			Password: "password123",
			Phone:    "13800138001",
		}

		userResp, err := userService.Register(registerReq)
		assert.NoError(t, err, "用户注册应该成功")
		assert.NotNil(t, userResp, "用户响应不应为空")
		assert.Equal(t, registerReq.Username, userResp.Username, "用户名应该匹配")

		// 2. 创建商家用户
		merchant := &model.User{
			Username: "integrationmerchant1",
			Email:    "integrationmerchant1@example.com",
			Password: "hashedpassword",
			Phone:    "13800138002",
			Role:     "merchant",
			Status:   "active",
		}
		err = db.Create(merchant).Error
		assert.NoError(t, err, "创建商家用户失败")

		// 3. 创建分类
		category := &model.Category{
			Name:        "集成测试分类1",
			Description: "integration-test-category-1",
			Status:      "active",
		}
		err = db.Create(category).Error
		assert.NoError(t, err, "创建分类失败")

		// 4. 创建商品
		price, _ := decimal.NewFromString("199.99")
		createProductReq := &model.CreateProductRequest{
			Name:        "集成测试商品1",
			Description: "用于集成测试的商品1",
			CategoryID:  category.ID,
			Price:       price,
			Stock:       100,
		}

		productResp, err := productService.CreateProduct(merchant.ID, createProductReq)
		assert.NoError(t, err, "创建商品应该成功")
		assert.NotNil(t, productResp, "商品响应不应为空")
		assert.Equal(t, createProductReq.Name, productResp.Name, "商品名称应该匹配")

		// 5. 用户浏览商品
		getProductResp, err := productService.GetProduct(productResp.ID)
		assert.NoError(t, err, "获取商品应该成功")
		assert.NotNil(t, getProductResp, "商品详情不应为空")
		assert.Equal(t, productResp.ID, getProductResp.ID, "商品ID应该匹配")

		t.Logf("✅ 用户-商品集成测试通过 - 用户: %s, 商品: %s",
			userResp.Username, productResp.Name)
	})

	t.Run("购物车-商品集成测试", func(t *testing.T) {
		// 1. 创建测试用户
		user := &model.User{
			Username: "integrationuser2",
			Email:    "integration2@example.com",
			Password: "hashedpassword",
			Phone:    "13800138003",
			Status:   "active",
		}
		err := db.Create(user).Error
		assert.NoError(t, err, "创建测试用户失败")

		// 2. 创建商家用户
		merchant := &model.User{
			Username: "integrationmerchant2",
			Email:    "integrationmerchant2@example.com",
			Password: "hashedpassword",
			Phone:    "13800138004",
			Role:     "merchant",
			Status:   "active",
		}
		err = db.Create(merchant).Error
		assert.NoError(t, err, "创建商家用户失败")

		// 3. 创建分类
		category := &model.Category{
			Name:        "集成测试分类2",
			Description: "integration-test-category-2",
			Status:      "active",
		}
		err = db.Create(category).Error
		assert.NoError(t, err, "创建分类失败")

		// 4. 创建商品
		price, _ := decimal.NewFromString("299.99")
		createProductReq := &model.CreateProductRequest{
			Name:        "集成测试商品2",
			Description: "用于集成测试的商品2",
			CategoryID:  category.ID,
			Price:       price,
			Stock:       50,
		}

		productResp, err := productService.CreateProduct(merchant.ID, createProductReq)
		assert.NoError(t, err, "创建商品应该成功")

		// 5. 添加商品到购物车
		addToCartReq := &model.AddToCartRequest{
			ProductID: productResp.ID,
			Quantity:  3,
		}

		cartItem, err := cartService.AddToCart(user.ID, "", addToCartReq)
		assert.NoError(t, err, "添加商品到购物车应该成功")
		assert.NotNil(t, cartItem, "购物车项不应为空")
		assert.Equal(t, productResp.ID, cartItem.ProductID, "商品ID应该匹配")
		assert.Equal(t, 3, cartItem.Quantity, "数量应该匹配")

		// 6. 获取购物车列表
		cartResponse, err := cartService.GetCart(user.ID, "", false)
		assert.NoError(t, err, "获取购物车应该成功")
		assert.NotNil(t, cartResponse, "购物车响应不应为空")
		assert.NotNil(t, cartResponse.Cart, "购物车不应为空")
		assert.GreaterOrEqual(t, len(cartResponse.Cart.Items), 1, "购物车应该至少有1个商品")

		// 7. 验证购物车中的商品信息
		found := false
		for _, item := range cartResponse.Cart.Items {
			if item.ProductID == productResp.ID {
				assert.Equal(t, 3, item.Quantity, "购物车中商品数量应该匹配")
				found = true
				break
			}
		}
		assert.True(t, found, "应该在购物车中找到添加的商品")

		t.Logf("✅ 购物车-商品集成测试通过 - 商品: %s, 购物车商品数: %d",
			productResp.Name, len(cartResponse.Cart.Items))
	})

	t.Run("完整购物流程集成测试", func(t *testing.T) {
		// 1. 用户注册
		registerReq := &model.RegisterRequest{
			Username: "integrationuser3",
			Email:    "integration3@example.com",
			Password: "password123",
			Phone:    "13800138005",
		}

		userResp, err := userService.Register(registerReq)
		assert.NoError(t, err, "用户注册应该成功")

		// 2. 创建商家用户
		merchant := &model.User{
			Username: "integrationmerchant3",
			Email:    "integrationmerchant3@example.com",
			Password: "hashedpassword",
			Phone:    "13800138006",
			Role:     "merchant",
			Status:   "active",
		}
		err = db.Create(merchant).Error
		assert.NoError(t, err, "创建商家用户失败")

		// 3. 创建分类
		category := &model.Category{
			Name:        "集成测试分类3",
			Description: "integration-test-category-3",
			Status:      "active",
		}
		err = db.Create(category).Error
		assert.NoError(t, err, "创建分类失败")

		// 4. 创建多个商品
		products := make([]*model.Product, 0)
		for i := 1; i <= 3; i++ {
			price, _ := decimal.NewFromString("99.99")
			createProductReq := &model.CreateProductRequest{
				Name:        fmt.Sprintf("集成测试商品%d", i),
				Description: fmt.Sprintf("用于集成测试的商品%d", i),
				CategoryID:  category.ID,
				Price:       price.Mul(decimal.NewFromInt(int64(i))),
				Stock:       20 * i,
			}

			productResp, err := productService.CreateProduct(merchant.ID, createProductReq)
			assert.NoError(t, err, "创建商品应该成功")
			products = append(products, productResp)
		}

		// 5. 用户浏览并添加多个商品到购物车
		totalItems := 0
		for i, product := range products {
			// 浏览商品
			getProductResp, err := productService.GetProduct(product.ID)
			assert.NoError(t, err, "获取商品应该成功")
			assert.Equal(t, product.ID, getProductResp.ID, "商品ID应该匹配")

			// 添加到购物车
			addToCartReq := &model.AddToCartRequest{
				ProductID: product.ID,
				Quantity:  i + 1, // 数量递增
			}

			cartItem, err := cartService.AddToCart(userResp.ID, "", addToCartReq)
			assert.NoError(t, err, "添加商品到购物车应该成功")
			assert.Equal(t, product.ID, cartItem.ProductID, "商品ID应该匹配")
			assert.Equal(t, i+1, cartItem.Quantity, "数量应该匹配")

			totalItems += i + 1
		}

		// 6. 验证购物车状态
		cartResponse, err := cartService.GetCart(userResp.ID, "", false)
		assert.NoError(t, err, "获取购物车应该成功")
		assert.Equal(t, len(products), len(cartResponse.Cart.Items), "购物车商品种类数应该匹配")

		// 7. 验证购物车总数量
		actualTotalQty := 0
		for _, item := range cartResponse.Cart.Items {
			actualTotalQty += item.Quantity
		}
		assert.Equal(t, totalItems, actualTotalQty, "购物车总数量应该匹配")

		t.Logf("✅ 完整购物流程集成测试通过 - 用户: %s, 商品种类: %d, 总数量: %d",
			userResp.Username, len(products), totalItems)
	})

	t.Run("业务规则集成验证", func(t *testing.T) {
		// 1. 创建测试用户
		user := &model.User{
			Username: "integrationuser4",
			Email:    "integration4@example.com",
			Password: "hashedpassword",
			Phone:    "13800138007",
			Status:   "active",
		}
		err := db.Create(user).Error
		assert.NoError(t, err, "创建测试用户失败")

		// 2. 创建商家用户
		merchant := &model.User{
			Username: "integrationmerchant4",
			Email:    "integrationmerchant4@example.com",
			Password: "hashedpassword",
			Phone:    "13800138008",
			Role:     "merchant",
			Status:   "active",
		}
		err = db.Create(merchant).Error
		assert.NoError(t, err, "创建商家用户失败")

		// 3. 创建分类
		category := &model.Category{
			Name:        "集成测试分类4",
			Description: "integration-test-category-4",
			Status:      "active",
		}
		err = db.Create(category).Error
		assert.NoError(t, err, "创建分类失败")

		// 4. 创建库存有限的商品
		price, _ := decimal.NewFromString("499.99")
		createProductReq := &model.CreateProductRequest{
			Name:        "集成测试商品4",
			Description: "用于集成测试的商品4",
			CategoryID:  category.ID,
			Price:       price,
			Stock:       5, // 低库存
		}

		productResp, err := productService.CreateProduct(merchant.ID, createProductReq)
		assert.NoError(t, err, "创建商品应该成功")

		// 5. 验证库存不足的业务规则
		addToCartReq := &model.AddToCartRequest{
			ProductID: productResp.ID,
			Quantity:  10, // 超过库存
		}

		cartItem, err := cartService.AddToCart(user.ID, "", addToCartReq)
		assert.Error(t, err, "库存不足时添加应该失败")
		assert.Contains(t, err.Error(), "库存不足", "错误信息应该包含库存不足")
		assert.Nil(t, cartItem, "失败时不应返回购物车项")

		// 6. 验证正常库存的业务规则
		normalReq := &model.AddToCartRequest{
			ProductID: productResp.ID,
			Quantity:  3, // 正常数量
		}

		normalCartItem, err := cartService.AddToCart(user.ID, "", normalReq)
		assert.NoError(t, err, "正常库存时添加应该成功")
		assert.NotNil(t, normalCartItem, "成功时应该返回购物车项")
		assert.Equal(t, 3, normalCartItem.Quantity, "数量应该匹配")

		t.Logf("✅ 业务规则集成验证测试通过 - 库存限制正常工作")
	})
}
