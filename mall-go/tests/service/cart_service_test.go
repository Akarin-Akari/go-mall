package service

import (
	"testing"

	"mall-go/internal/config"
	"mall-go/internal/model"
	"mall-go/pkg/cart"
	testConfig "mall-go/tests/config"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

// TestCartService 购物车服务测试
func TestCartService(t *testing.T) {
	// 初始化配置
	config.GlobalConfig = config.Config{
		JWT: config.JWTConfig{
			Secret: "test-secret-key-for-cart-service-testing",
			Expire: "24h",
		},
	}

	// 初始化测试数据库
	db := testConfig.SetupTestDB()
	defer testConfig.CleanupTestDB(db)

	// 自动迁移
	err := db.AutoMigrate(
		&model.User{},
		&model.Product{},
		&model.ProductImage{},
		&model.ProductSKU{},
		&model.Category{},
		&model.Brand{},
		&model.Cart{},
		&model.CartItem{},
	)
	assert.NoError(t, err, "数据库迁移失败")

	// 创建购物车服务
	cartService := cart.NewCartService(db)

	// 创建测试数据
	setupTestData := func() (*model.User, *model.Product, *model.Category) {
		// 创建测试用户
		user := &model.User{
			Username: "cartuser",
			Email:    "cart@example.com",
			Password: "hashedpassword",
			Status:   "active",
		}
		err := db.Create(user).Error
		assert.NoError(t, err, "创建测试用户失败")

		// 创建商家用户
		merchant := &model.User{
			Username: "cartmerchant",
			Email:    "cartmerchant@example.com",
			Password: "hashedpassword",
			Role:     "merchant",
			Status:   "active",
		}
		err = db.Create(merchant).Error
		assert.NoError(t, err, "创建商家用户失败")

		// 创建分类
		category := &model.Category{
			Name:        "购物车测试分类",
			Description: "cart-test-category",
			Status:      "active",
		}
		err = db.Create(category).Error
		assert.NoError(t, err, "创建分类失败")

		// 创建商品
		price, _ := decimal.NewFromString("99.99")
		product := &model.Product{
			Name:        "购物车测试商品",
			Description: "用于购物车测试的商品",
			CategoryID:  category.ID,
			MerchantID:  merchant.ID,
			Price:       price,
			Stock:       100,
			Status:      "active",
		}
		err = db.Create(product).Error
		assert.NoError(t, err, "创建测试商品失败")

		return user, product, category
	}

	t.Run("成功添加商品到购物车", func(t *testing.T) {
		user, product, _ := setupTestData()

		// 添加商品到购物车
		req := &model.AddToCartRequest{
			ProductID: product.ID,
			Quantity:  2,
		}

		cartItem, err := cartService.AddToCart(user.ID, "", req)

		// 验证添加结果
		assert.NoError(t, err, "添加商品到购物车应该成功")
		assert.NotNil(t, cartItem, "购物车项不应为空")
		assert.Equal(t, product.ID, cartItem.ProductID, "商品ID应该匹配")
		assert.Equal(t, 2, cartItem.Quantity, "数量应该匹配")
		// UserID在CartItem中不存在，通过Cart关联验证

		// 验证数据库中的购物车数据
		var dbCartItem model.CartItem
		err = db.Where("user_id = ? AND product_id = ?", user.ID, product.ID).First(&dbCartItem).Error
		assert.NoError(t, err, "应该能在数据库中找到购物车项")
		assert.Equal(t, 2, dbCartItem.Quantity, "数据库中数量应该匹配")

		t.Logf("✅ 添加商品到购物车测试通过 - 商品ID: %d, 数量: %d", cartItem.ProductID, cartItem.Quantity)
	})

	t.Run("重复添加商品到购物车", func(t *testing.T) {
		user, product, _ := setupTestData()

		// 第一次添加
		req1 := &model.AddToCartRequest{
			ProductID: product.ID,
			Quantity:  2,
		}
		_, err := cartService.AddToCart(user.ID, "", req1)
		assert.NoError(t, err, "第一次添加应该成功")

		// 第二次添加相同商品
		req2 := &model.AddToCartRequest{
			ProductID: product.ID,
			Quantity:  3,
		}
		cartItem, err := cartService.AddToCart(user.ID, "", req2)

		// 验证数量累加
		assert.NoError(t, err, "第二次添加应该成功")
		assert.NotNil(t, cartItem, "购物车项不应为空")
		assert.Equal(t, 5, cartItem.Quantity, "数量应该累加为5")

		t.Logf("✅ 重复添加商品测试通过 - 累加数量: %d", cartItem.Quantity)
	})

	t.Run("添加不存在的商品", func(t *testing.T) {
		user, _, _ := setupTestData()

		// 尝试添加不存在的商品
		req := &model.AddToCartRequest{
			ProductID: 99999, // 不存在的商品ID
			Quantity:  1,
		}

		cartItem, err := cartService.AddToCart(user.ID, "", req)

		// 验证错误处理
		assert.Error(t, err, "添加不存在商品应该失败")
		assert.Contains(t, err.Error(), "商品不存在", "错误信息应该包含商品不存在")
		assert.Nil(t, cartItem, "失败时不应返回购物车项")

		t.Logf("✅ 添加不存在商品测试通过")
	})

	t.Run("库存不足时添加商品", func(t *testing.T) {
		user, product, _ := setupTestData()

		// 尝试添加超过库存的数量
		req := &model.AddToCartRequest{
			ProductID: product.ID,
			Quantity:  150, // 超过库存100
		}

		cartItem, err := cartService.AddToCart(user.ID, "", req)

		// 验证库存检查
		assert.Error(t, err, "库存不足时添加应该失败")
		assert.Contains(t, err.Error(), "库存不足", "错误信息应该包含库存不足")
		assert.Nil(t, cartItem, "失败时不应返回购物车项")

		t.Logf("✅ 库存不足测试通过")
	})

	t.Run("获取购物车列表", func(t *testing.T) {
		user, product, _ := setupTestData()

		// 先添加商品到购物车
		req := &model.AddToCartRequest{
			ProductID: product.ID,
			Quantity:  3,
		}
		_, err := cartService.AddToCart(user.ID, "", req)
		assert.NoError(t, err, "添加商品应该成功")

		// 获取购物车列表
		cartResponse, err := cartService.GetCart(user.ID, "", false)

		// 验证获取结果
		assert.NoError(t, err, "获取购物车应该成功")
		assert.NotNil(t, cartResponse, "购物车响应不应为空")
		assert.GreaterOrEqual(t, len(cartResponse.Cart.Items), 1, "购物车应该至少有1个商品")

		// 验证购物车项数据
		found := false
		for _, item := range cartResponse.Cart.Items {
			if item.ProductID == product.ID {
				assert.Equal(t, 3, item.Quantity, "商品数量应该匹配")
				assert.NotNil(t, item.Product, "商品信息不应为空")
				assert.Equal(t, "购物车测试商品", item.Product.Name, "商品名称应该匹配")
				found = true
				break
			}
		}
		assert.True(t, found, "应该找到添加的商品")

		t.Logf("✅ 获取购物车列表测试通过 - 商品数量: %d", len(cartResponse.Cart.Items))
	})

	t.Run("更新购物车商品数量", func(t *testing.T) {
		user, product, _ := setupTestData()

		// 先添加商品到购物车
		req := &model.AddToCartRequest{
			ProductID: product.ID,
			Quantity:  2,
		}
		cartItem, err := cartService.AddToCart(user.ID, "", req)
		assert.NoError(t, err, "添加商品应该成功")

		// 更新商品数量
		updateReq := &model.UpdateCartItemRequest{
			Quantity: 5,
			Selected: true,
		}
		updatedItem, err := cartService.UpdateCartItem(user.ID, "", cartItem.ID, updateReq)

		// 验证更新结果
		assert.NoError(t, err, "更新购物车商品数量应该成功")
		assert.NotNil(t, updatedItem, "更新后的购物车项不应为空")
		assert.Equal(t, 5, updatedItem.Quantity, "数量应该更新为5")

		// 验证数据库中的数据
		var dbCartItem model.CartItem
		err = db.Where("id = ?", cartItem.ID).First(&dbCartItem).Error
		assert.NoError(t, err, "应该能在数据库中找到购物车项")
		assert.Equal(t, 5, dbCartItem.Quantity, "数据库中数量应该已更新")

		t.Logf("✅ 更新购物车商品数量测试通过 - 新数量: %d", updatedItem.Quantity)
	})

	t.Run("删除购物车商品", func(t *testing.T) {
		user, product, _ := setupTestData()

		// 先添加商品到购物车
		req := &model.AddToCartRequest{
			ProductID: product.ID,
			Quantity:  2,
		}
		cartItem, err := cartService.AddToCart(user.ID, "", req)
		assert.NoError(t, err, "添加商品应该成功")

		// 删除购物车商品
		err = cartService.RemoveFromCart(user.ID, "", cartItem.ID)

		// 验证删除结果
		assert.NoError(t, err, "删除购物车商品应该成功")

		// 验证数据库中已删除
		var dbCartItem model.CartItem
		err = db.Where("id = ?", cartItem.ID).First(&dbCartItem).Error
		assert.Error(t, err, "删除后应该在数据库中找不到购物车项")

		t.Logf("✅ 删除购物车商品测试通过")
	})

	t.Run("清空购物车", func(t *testing.T) {
		user, product, _ := setupTestData()

		// 先添加多个商品到购物车
		for i := 0; i < 3; i++ {
			req := &model.AddToCartRequest{
				ProductID: product.ID,
				Quantity:  1,
			}
			_, err := cartService.AddToCart(user.ID, "", req)
			assert.NoError(t, err, "添加商品应该成功")
		}

		// 清空购物车
		err := cartService.ClearCart(user.ID, "")

		// 验证清空结果
		assert.NoError(t, err, "清空购物车应该成功")

		// 验证购物车已空
		cartResponse, err := cartService.GetCart(user.ID, "", false)
		assert.NoError(t, err, "获取购物车应该成功")
		assert.Equal(t, 0, len(cartResponse.Cart.Items), "购物车应该为空")

		t.Logf("✅ 清空购物车测试通过")
	})
}
