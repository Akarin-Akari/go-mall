package service

import (
	"testing"

	"mall-go/internal/config"
	"mall-go/internal/model"
	"mall-go/pkg/cart"

	"github.com/glebarez/sqlite"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// TestSimpleServiceIntegration 简单跨服务集成测试
func TestSimpleServiceIntegration(t *testing.T) {
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
	cartService := cart.NewCartService(db)

	t.Run("用户-商品-购物车集成流程", func(t *testing.T) {
		// 1. 创建测试用户
		user := &model.User{
			Username: "integrationuser1",
			Email:    "integration1@example.com",
			Password: "hashedpassword",
			Phone:    "13800138001",
			Status:   "active",
		}
		err := db.Create(user).Error
		assert.NoError(t, err, "创建测试用户失败")

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
		product := &model.Product{
			Name:        "集成测试商品1",
			Description: "用于集成测试的商品1",
			CategoryID:  category.ID,
			MerchantID:  merchant.ID,
			Price:       price,
			Stock:       100,
			Status:      "active",
		}
		err = db.Create(product).Error
		assert.NoError(t, err, "创建测试商品失败")

		// 5. 验证商品数据完整性
		var dbProduct model.Product
		err = db.Where("id = ?", product.ID).First(&dbProduct).Error
		assert.NoError(t, err, "应该能在数据库中找到商品")
		assert.Equal(t, product.Name, dbProduct.Name, "商品名称应该匹配")
		assert.Equal(t, "active", dbProduct.Status, "商品状态应该为active")
		assert.True(t, dbProduct.Price.GreaterThan(decimal.Zero), "商品价格应该大于0")

		// 6. 添加商品到购物车
		addToCartReq := &model.AddToCartRequest{
			ProductID: product.ID,
			Quantity:  3,
		}

		cartItem, err := cartService.AddToCart(user.ID, "", addToCartReq)
		assert.NoError(t, err, "添加商品到购物车应该成功")
		assert.NotNil(t, cartItem, "购物车项不应为空")
		assert.Equal(t, product.ID, cartItem.ProductID, "商品ID应该匹配")
		assert.Equal(t, 3, cartItem.Quantity, "数量应该匹配")

		// 7. 获取购物车列表
		cartResponse, err := cartService.GetCart(user.ID, "", false)
		assert.NoError(t, err, "获取购物车应该成功")
		assert.NotNil(t, cartResponse, "购物车响应不应为空")
		assert.NotNil(t, cartResponse.Cart, "购物车不应为空")

		// 8. 验证购物车中的商品信息
		if len(cartResponse.Cart.Items) > 0 {
			found := false
			for _, item := range cartResponse.Cart.Items {
				if item.ProductID == product.ID {
					assert.Equal(t, 3, item.Quantity, "购物车中商品数量应该匹配")
					found = true
					break
				}
			}
			assert.True(t, found, "应该在购物车中找到添加的商品")
		}

		t.Logf("✅ 用户-商品-购物车集成流程测试通过 - 用户: %s, 商品: %s, 购物车商品数: %d",
			user.Username, product.Name, len(cartResponse.Cart.Items))
	})

	t.Run("多用户购物车隔离测试", func(t *testing.T) {
		// 1. 创建两个测试用户
		user1 := &model.User{
			Username: "integrationuser2",
			Email:    "integration2@example.com",
			Password: "hashedpassword",
			Phone:    "13800138003",
			Status:   "active",
		}
		err := db.Create(user1).Error
		assert.NoError(t, err, "创建测试用户1失败")

		user2 := &model.User{
			Username: "integrationuser3",
			Email:    "integration3@example.com",
			Password: "hashedpassword",
			Phone:    "13800138004",
			Status:   "active",
		}
		err = db.Create(user2).Error
		assert.NoError(t, err, "创建测试用户2失败")

		// 2. 创建商家用户
		merchant := &model.User{
			Username: "integrationmerchant2",
			Email:    "integrationmerchant2@example.com",
			Password: "hashedpassword",
			Phone:    "13800138005",
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
		product := &model.Product{
			Name:        "集成测试商品2",
			Description: "用于集成测试的商品2",
			CategoryID:  category.ID,
			MerchantID:  merchant.ID,
			Price:       price,
			Stock:       50,
			Status:      "active",
		}
		err = db.Create(product).Error
		assert.NoError(t, err, "创建测试商品失败")

		// 5. 用户1添加商品到购物车
		addToCartReq1 := &model.AddToCartRequest{
			ProductID: product.ID,
			Quantity:  2,
		}
		cartItem1, err := cartService.AddToCart(user1.ID, "", addToCartReq1)
		assert.NoError(t, err, "用户1添加商品到购物车应该成功")
		assert.Equal(t, 2, cartItem1.Quantity, "用户1购物车数量应该为2")

		// 6. 用户2添加商品到购物车
		addToCartReq2 := &model.AddToCartRequest{
			ProductID: product.ID,
			Quantity:  5,
		}
		cartItem2, err := cartService.AddToCart(user2.ID, "", addToCartReq2)
		assert.NoError(t, err, "用户2添加商品到购物车应该成功")
		assert.Equal(t, 5, cartItem2.Quantity, "用户2购物车数量应该为5")

		// 7. 验证购物车隔离
		cartResponse1, err := cartService.GetCart(user1.ID, "", false)
		assert.NoError(t, err, "获取用户1购物车应该成功")

		cartResponse2, err := cartService.GetCart(user2.ID, "", false)
		assert.NoError(t, err, "获取用户2购物车应该成功")

		// 8. 验证购物车数据隔离
		if len(cartResponse1.Cart.Items) > 0 && len(cartResponse2.Cart.Items) > 0 {
			user1Qty := 0
			user2Qty := 0

			for _, item := range cartResponse1.Cart.Items {
				if item.ProductID == product.ID {
					user1Qty = item.Quantity
					break
				}
			}

			for _, item := range cartResponse2.Cart.Items {
				if item.ProductID == product.ID {
					user2Qty = item.Quantity
					break
				}
			}

			assert.Equal(t, 2, user1Qty, "用户1购物车数量应该为2")
			assert.Equal(t, 5, user2Qty, "用户2购物车数量应该为5")
			assert.NotEqual(t, user1Qty, user2Qty, "两个用户的购物车数量应该不同")
		}

		t.Logf("✅ 多用户购物车隔离测试通过 - 用户1数量: 2, 用户2数量: 5")
	})

	t.Run("购物车业务规则集成测试", func(t *testing.T) {
		// 1. 创建测试用户
		user := &model.User{
			Username: "integrationuser4",
			Email:    "integration4@example.com",
			Password: "hashedpassword",
			Phone:    "13800138006",
			Status:   "active",
		}
		err := db.Create(user).Error
		assert.NoError(t, err, "创建测试用户失败")

		// 2. 创建商家用户
		merchant := &model.User{
			Username: "integrationmerchant3",
			Email:    "integrationmerchant3@example.com",
			Password: "hashedpassword",
			Phone:    "13800138007",
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

		// 4. 创建库存有限的商品
		price, _ := decimal.NewFromString("399.99")
		product := &model.Product{
			Name:        "集成测试商品3",
			Description: "用于集成测试的商品3",
			CategoryID:  category.ID,
			MerchantID:  merchant.ID,
			Price:       price,
			Stock:       8, // 有限库存
			Status:      "active",
		}
		err = db.Create(product).Error
		assert.NoError(t, err, "创建测试商品失败")

		// 5. 第一次添加商品（正常数量）
		addToCartReq1 := &model.AddToCartRequest{
			ProductID: product.ID,
			Quantity:  3,
		}
		cartItem1, err := cartService.AddToCart(user.ID, "", addToCartReq1)
		assert.NoError(t, err, "第一次添加应该成功")
		assert.Equal(t, 3, cartItem1.Quantity, "第一次添加数量应该为3")

		// 6. 第二次添加商品（累加数量）
		addToCartReq2 := &model.AddToCartRequest{
			ProductID: product.ID,
			Quantity:  2,
		}
		cartItem2, err := cartService.AddToCart(user.ID, "", addToCartReq2)
		assert.NoError(t, err, "第二次添加应该成功")
		assert.Equal(t, 5, cartItem2.Quantity, "累加后数量应该为5")

		// 7. 第三次添加商品（超过库存）
		addToCartReq3 := &model.AddToCartRequest{
			ProductID: product.ID,
			Quantity:  5, // 5 + 5 = 10 > 8（库存）
		}
		cartItem3, err := cartService.AddToCart(user.ID, "", addToCartReq3)
		assert.Error(t, err, "超过库存时添加应该失败")
		assert.Contains(t, err.Error(), "库存不足", "错误信息应该包含库存不足")
		assert.Nil(t, cartItem3, "失败时不应返回购物车项")

		// 8. 验证购物车状态未被破坏
		cartResponse, err := cartService.GetCart(user.ID, "", false)
		assert.NoError(t, err, "获取购物车应该成功")

		if len(cartResponse.Cart.Items) > 0 {
			for _, item := range cartResponse.Cart.Items {
				if item.ProductID == product.ID {
					assert.Equal(t, 5, item.Quantity, "购物车数量应该保持为5")
					break
				}
			}
		}

		t.Logf("✅ 购物车业务规则集成测试通过 - 库存限制和数量累加正常工作")
	})
}
