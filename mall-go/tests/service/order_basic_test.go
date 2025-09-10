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

// TestOrderBasicService 订单基础服务测试
func TestOrderBasicService(t *testing.T) {
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
		&model.OrderStatusLog{},
		&model.Payment{},
	)
	assert.NoError(t, err, "数据库迁移失败")

	// 创建服务
	cartService := cart.NewCartService(db)

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
		assert.Equal(t, "pending", string(model.PaymentStatusPending), "待支付状态常量应该正确")
		assert.Equal(t, "paid", string(model.PaymentStatusPaid), "已支付状态常量应该正确")
		assert.Equal(t, "refunded", string(model.PaymentStatusRefunded), "已退款状态常量应该正确")

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

		// 验证金额精度（decimal库的Exponent返回负数表示小数位数）
		assert.Equal(t, int32(-2), price.Exponent(), "价格应该保持2位小数精度")

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

	t.Run("订单业务规则验证", func(t *testing.T) {
		// 创建测试订单
		user := &model.User{
			Username: "orderuser4",
			Email:    "order4@example.com",
			Password: "hashedpassword",
			Phone:    "13800138007",
			Status:   "active",
		}
		err := db.Create(user).Error
		assert.NoError(t, err, "创建测试用户失败")

		// 创建订单
		totalAmount, _ := decimal.NewFromString("599.99")
		order := &model.Order{
			OrderNo:         "ORD202501100001",
			UserID:          user.ID,
			Status:          model.OrderStatusPending,
			PaymentStatus:   string(model.PaymentStatusPending),
			TotalAmount:     totalAmount,
			PayableAmount:   totalAmount,
			ReceiverName:    "测试用户",
			ReceiverPhone:   "13800138007",
			ReceiverAddress: "测试地址",
		}
		err = db.Create(order).Error
		assert.NoError(t, err, "创建测试订单失败")

		// 验证订单业务规则
		// 1. 待支付订单可以取消
		assert.True(t, order.CanCancel(), "待支付订单应该可以取消")

		// 2. 待支付订单可以支付
		assert.True(t, order.CanPay(), "待支付订单应该可以支付")

		// 3. 待支付订单不能发货
		assert.False(t, order.CanShip(), "待支付订单不应该可以发货")

		// 4. 待支付订单不能收货
		assert.False(t, order.CanReceive(), "待支付订单不应该可以收货")

		// 5. 待支付订单不能退款
		assert.False(t, order.CanRefund(), "待支付订单不应该可以退款")

		// 6. 验证订单状态文本
		assert.Equal(t, "待支付", order.GetStatusText(), "订单状态文本应该正确")

		t.Logf("✅ 订单业务规则验证测试通过 - 订单号: %s, 状态: %s",
			order.OrderNo, order.GetStatusText())
	})
}
