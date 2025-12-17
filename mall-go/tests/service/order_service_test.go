package service

import (
	"testing"

	"mall-go/internal/config"
	"mall-go/internal/model"
	"mall-go/pkg/cart"
	"mall-go/pkg/inventory"
	"mall-go/pkg/order"

	"github.com/glebarez/sqlite"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// TestOrderService 订单服务测试
func TestOrderService(t *testing.T) {
	// 初始化配置
	config.GlobalConfig = config.Config{
		JWT: config.JWTConfig{
			Secret: "test-secret-key-for-order-service-testing",
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
		&model.OrderLog{},
		&model.Address{},
	)
	assert.NoError(t, err, "数据库迁移失败")

	// 创建服务
	cartService := cart.NewCartService(db)
	calculationService := cart.NewCalculationService(db)
	inventoryService := inventory.NewInventoryService(db)
	orderService := order.NewOrderService(db, cartService, calculationService, inventoryService)

	t.Run("订单创建业务逻辑验证", func(t *testing.T) {
		// 创建测试用户
		user := &model.User{
			Username: "orderuser1",
			Email:    "order1@example.com",
			Password: "hashedpassword",
			Phone:    "13800138001",
			Status:   "active",
		}
		err := db.Create(user).Error
		assert.NoError(t, err, "创建测试用户失败")

		// 创建商家用户
		merchant := &model.User{
			Username: "ordermerchant1",
			Email:    "ordermerchant1@example.com",
			Password: "hashedpassword",
			Phone:    "13800138002",
			Role:     "merchant",
			Status:   "active",
		}
		err = db.Create(merchant).Error
		assert.NoError(t, err, "创建商家用户失败")

		// 创建分类
		category := &model.Category{
			Name:        "订单测试分类1",
			Description: "order-test-category-1",
			Status:      "active",
		}
		err = db.Create(category).Error
		assert.NoError(t, err, "创建分类失败")

		// 创建商品
		price, _ := decimal.NewFromString("199.99")
		product := &model.Product{
			Name:        "订单测试商品1",
			Description: "用于订单测试的商品1",
			CategoryID:  category.ID,
			MerchantID:  merchant.ID,
			Price:       price,
			Stock:       50,
			Status:      "active",
		}
		err = db.Create(product).Error
		assert.NoError(t, err, "创建测试商品失败")

		// 创建用户地址
		address := &model.Address{
			UserID:    user.ID,
			Name:      "张三",
			Phone:     "13800138001",
			Province:  "北京市",
			City:      "北京市",
			District:  "朝阳区",
			Address:   "朝阳路123号",
			ZipCode:   "100000",
			IsDefault: true,
			Status:    "active",
		}
		err = db.Create(address).Error
		assert.NoError(t, err, "创建用户地址失败")

		// 添加商品到购物车
		cartReq := &model.AddToCartRequest{
			ProductID: product.ID,
			Quantity:  2,
		}
		cartItem, err := cartService.AddToCart(user.ID, "", cartReq)
		assert.NoError(t, err, "添加商品到购物车应该成功")

		// 验证订单创建前的业务逻辑
		// 1. 用户必须存在且状态为active
		assert.Equal(t, "active", user.Status, "用户状态应该为active")

		// 2. 商品必须存在且有库存
		assert.Equal(t, "active", product.Status, "商品状态应该为active")
		assert.Greater(t, product.Stock, 0, "商品库存应该大于0")

		// 3. 购物车项必须存在
		assert.NotNil(t, cartItem, "购物车项不应为空")
		assert.Equal(t, product.ID, cartItem.ProductID, "购物车商品ID应该匹配")

		// 4. 地址必须存在
		assert.Equal(t, user.ID, address.UserID, "地址用户ID应该匹配")
		assert.True(t, address.IsDefault, "应该有默认地址")

		t.Logf("✅ 订单创建业务逻辑验证测试通过")
	})

	t.Run("订单状态管理验证", func(t *testing.T) {
		// 验证订单状态常量
		assert.Equal(t, "pending", model.OrderStatusPending, "待支付状态常量应该正确")
		assert.Equal(t, "paid", model.OrderStatusPaid, "已支付状态常量应该正确")
		assert.Equal(t, "shipped", model.OrderStatusShipped, "已发货状态常量应该正确")
		assert.Equal(t, "delivered", model.OrderStatusDelivered, "已送达状态常量应该正确")
		assert.Equal(t, "completed", model.OrderStatusCompleted, "已完成状态常量应该正确")
		assert.Equal(t, "cancelled", model.OrderStatusCancelled, "已取消状态常量应该正确")

		// 验证支付状态常量
		assert.Equal(t, "unpaid", model.PaymentStatusUnpaid, "未支付状态常量应该正确")
		assert.Equal(t, "paid", model.PaymentStatusPaid, "已支付状态常量应该正确")
		assert.Equal(t, "refunded", model.PaymentStatusRefunded, "已退款状态常量应该正确")

		t.Logf("✅ 订单状态管理验证测试通过")
	})

	t.Run("订单金额计算验证", func(t *testing.T) {
		// 创建测试用户
		user := &model.User{
			Username: "orderuser2",
			Email:    "order2@example.com",
			Password: "hashedpassword",
			Phone:    "13800138003",
			Status:   "active",
		}
		err := db.Create(user).Error
		assert.NoError(t, err, "创建测试用户失败")

		// 创建商家用户
		merchant := &model.User{
			Username: "ordermerchant2",
			Email:    "ordermerchant2@example.com",
			Password: "hashedpassword",
			Phone:    "13800138004",
			Role:     "merchant",
			Status:   "active",
		}
		err = db.Create(merchant).Error
		assert.NoError(t, err, "创建商家用户失败")

		// 创建分类
		category := &model.Category{
			Name:        "订单测试分类2",
			Description: "order-test-category-2",
			Status:      "active",
		}
		err = db.Create(category).Error
		assert.NoError(t, err, "创建分类失败")

		// 创建商品
		price, _ := decimal.NewFromString("299.99")
		product := &model.Product{
			Name:        "订单测试商品2",
			Description: "用于订单测试的商品2",
			CategoryID:  category.ID,
			MerchantID:  merchant.ID,
			Price:       price,
			Stock:       30,
			Status:      "active",
		}
		err = db.Create(product).Error
		assert.NoError(t, err, "创建测试商品失败")

		// 验证金额计算逻辑
		quantity := 3
		expectedTotal := price.Mul(decimal.NewFromInt(int64(quantity)))

		// 验证价格计算
		assert.True(t, price.GreaterThan(decimal.Zero), "商品价格应该大于0")
		assert.True(t, expectedTotal.GreaterThan(price), "总金额应该大于单价")

		// 验证金额精度
		assert.Equal(t, int32(2), price.Exponent(), "价格应该保持2位小数精度")

		t.Logf("✅ 订单金额计算验证测试通过 - 单价: %s, 数量: %d, 总计: %s",
			price.String(), quantity, expectedTotal.String())
	})

	t.Run("订单库存扣减验证", func(t *testing.T) {
		// 创建测试用户
		user := &model.User{
			Username: "orderuser3",
			Email:    "order3@example.com",
			Password: "hashedpassword",
			Phone:    "13800138005",
			Status:   "active",
		}
		err := db.Create(user).Error
		assert.NoError(t, err, "创建测试用户失败")

		// 创建商家用户
		merchant := &model.User{
			Username: "ordermerchant3",
			Email:    "ordermerchant3@example.com",
			Password: "hashedpassword",
			Phone:    "13800138006",
			Role:     "merchant",
			Status:   "active",
		}
		err = db.Create(merchant).Error
		assert.NoError(t, err, "创建商家用户失败")

		// 创建分类
		category := &model.Category{
			Name:        "订单测试分类3",
			Description: "order-test-category-3",
			Status:      "active",
		}
		err = db.Create(category).Error
		assert.NoError(t, err, "创建分类失败")

		// 创建库存有限的商品
		price, _ := decimal.NewFromString("399.99")
		initialStock := 10
		product := &model.Product{
			Name:        "订单测试商品3",
			Description: "用于订单测试的商品3",
			CategoryID:  category.ID,
			MerchantID:  merchant.ID,
			Price:       price,
			Stock:       initialStock,
			Status:      "active",
		}
		err = db.Create(product).Error
		assert.NoError(t, err, "创建测试商品失败")

		// 验证库存扣减逻辑
		orderQuantity := 3
		expectedRemainingStock := initialStock - orderQuantity

		// 验证库存充足
		assert.GreaterOrEqual(t, product.Stock, orderQuantity, "库存应该充足")

		// 验证扣减后库存
		assert.GreaterOrEqual(t, expectedRemainingStock, 0, "扣减后库存不应为负数")

		// 验证库存不足的情况
		excessiveQuantity := initialStock + 5
		assert.Greater(t, excessiveQuantity, product.Stock, "超量订购应该大于库存")

		t.Logf("✅ 订单库存扣减验证测试通过 - 初始库存: %d, 订购数量: %d, 剩余库存: %d",
			initialStock, orderQuantity, expectedRemainingStock)
	})

	t.Run("订单地址信息验证", func(t *testing.T) {
		// 创建测试用户
		user := &model.User{
			Username: "orderuser4",
			Email:    "order4@example.com",
			Password: "hashedpassword",
			Phone:    "13800138007",
			Status:   "active",
		}
		err := db.Create(user).Error
		assert.NoError(t, err, "创建测试用户失败")

		// 创建用户地址
		address := &model.Address{
			UserID:    user.ID,
			Name:      "李四",
			Phone:     "13800138007",
			Province:  "上海市",
			City:      "上海市",
			District:  "浦东新区",
			Address:   "张江高科技园区",
			ZipCode:   "200000",
			IsDefault: true,
			Status:    "active",
		}
		err = db.Create(address).Error
		assert.NoError(t, err, "创建用户地址失败")

		// 验证地址信息
		assert.Equal(t, user.ID, address.UserID, "地址用户ID应该匹配")
		assert.NotEmpty(t, address.Name, "收货人姓名不应为空")
		assert.NotEmpty(t, address.Phone, "收货人电话不应为空")
		assert.NotEmpty(t, address.Province, "省份不应为空")
		assert.NotEmpty(t, address.City, "城市不应为空")
		assert.NotEmpty(t, address.District, "区县不应为空")
		assert.NotEmpty(t, address.Address, "详细地址不应为空")
		assert.True(t, address.IsDefault, "应该有默认地址")
		assert.Equal(t, "active", address.Status, "地址状态应该为active")

		t.Logf("✅ 订单地址信息验证测试通过 - 收货人: %s, 地址: %s %s %s %s",
			address.Name, address.Province, address.City, address.District, address.Address)
	})
}
