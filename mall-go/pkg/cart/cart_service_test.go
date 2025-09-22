package cart

import (
	"testing"

	"mall-go/internal/model"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// CartServiceTestSuite 购物车服务测试套件
type CartServiceTestSuite struct {
	suite.Suite
	db          *gorm.DB
	cartService *CartService
}

// SetupSuite 设置测试套件
func (suite *CartServiceTestSuite) SetupSuite() {
	// 创建内存数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	suite.Require().NoError(err)
	suite.db = db

	// 自动迁移
	err = db.AutoMigrate(
		&model.Cart{},
		&model.CartItem{},
		&model.Product{},
		&model.ProductSKU{},
		&model.Category{},
		&model.Brand{},
		&model.User{},
	)
	suite.Require().NoError(err)

	// 创建测试数据
	suite.createTestData()

	// 创建购物车服务
	suite.cartService = NewCartService(db)
}

// createTestData 创建测试数据
func (suite *CartServiceTestSuite) createTestData() {
	// 创建测试分类
	category := &model.Category{
		Name:   "测试分类",
		Status: model.CategoryStatusActive,
	}
	suite.db.Create(category)

	// 创建测试品牌
	brand := &model.Brand{
		Name:   "测试品牌",
		Status: model.BrandStatusActive,
	}
	suite.db.Create(brand)

	// 创建测试用户
	user := &model.User{
		Username: "testuser",
		Email:    "test@example.com",
		Role:     model.RoleUser,
		Status:   model.StatusActive,
	}
	suite.db.Create(user)

	// 创建测试商品
	product := &model.Product{
		Name:       "测试商品",
		CategoryID: 1,
		BrandID:    1,
		MerchantID: 1,
		Price:      decimal.NewFromFloat(99.99),
		Stock:      100,
		Status:     model.ProductStatusActive,
	}
	suite.db.Create(product)

	// 创建测试SKU
	sku := &model.ProductSKU{
		ProductID: 1,
		SKUCode:   "TEST-SKU-001",
		Name:      "测试SKU",
		Price:     decimal.NewFromFloat(89.99),
		Stock:     50,
		Status:    model.SKUStatusActive,
	}
	suite.db.Create(sku)
}

// TestCartService_AddToCart 测试添加商品到购物车
func (suite *CartServiceTestSuite) TestCartService_AddToCart() {
	req := &model.AddToCartRequest{
		ProductID: 1,
		Quantity:  2,
	}

	// 测试用户购物车
	cartItem, err := suite.cartService.AddToCart(1, "", req)
	suite.NoError(err)
	suite.NotNil(cartItem)
	suite.Equal(uint(1), cartItem.ProductID)
	suite.Equal(2, cartItem.Quantity)
	suite.True(cartItem.Selected)
	suite.Equal(model.CartItemStatusNormal, cartItem.Status)

	// 验证购物车已创建
	var cart model.Cart
	err = suite.db.Where("user_id = ?", 1).First(&cart).Error
	suite.NoError(err)
	suite.Equal(uint(1), cart.UserID)
	suite.Equal(model.CartStatusActive, cart.Status)

	// 测试游客购物车
	cartItem2, err := suite.cartService.AddToCart(0, "guest-session-123", req)
	suite.NoError(err)
	suite.NotNil(cartItem2)
	suite.Equal(uint(1), cartItem2.ProductID)

	// 验证游客购物车已创建
	var guestCart model.Cart
	err = suite.db.Where("session_id = ?", "guest-session-123").First(&guestCart).Error
	suite.NoError(err)
	suite.Equal(uint(0), guestCart.UserID)
	suite.Equal("guest-session-123", guestCart.SessionID)
}

// TestCartService_AddToCartWithSKU 测试添加SKU商品到购物车
func (suite *CartServiceTestSuite) TestCartService_AddToCartWithSKU() {
	req := &model.AddToCartRequest{
		ProductID: 1,
		SKUID:     1,
		Quantity:  1,
	}

	cartItem, err := suite.cartService.AddToCart(1, "", req)
	suite.NoError(err)
	suite.NotNil(cartItem)
	suite.Equal(uint(1), cartItem.ProductID)
	suite.Equal(uint(1), cartItem.SKUID)
	suite.Equal(1, cartItem.Quantity)
	suite.True(cartItem.Price.Equal(decimal.NewFromFloat(89.99))) // SKU价格
}

// TestCartService_AddToCartDuplicate 测试添加重复商品
func (suite *CartServiceTestSuite) TestCartService_AddToCartDuplicate() {
	req := &model.AddToCartRequest{
		ProductID: 1,
		Quantity:  2,
	}

	// 第一次添加
	cartItem1, err := suite.cartService.AddToCart(1, "", req)
	suite.NoError(err)
	suite.Equal(2, cartItem1.Quantity)

	// 第二次添加相同商品
	cartItem2, err := suite.cartService.AddToCart(1, "", req)
	suite.NoError(err)
	suite.Equal(4, cartItem2.Quantity)      // 数量应该累加
	suite.Equal(cartItem1.ID, cartItem2.ID) // 应该是同一个商品项
}

// TestCartService_AddToCartInsufficientStock 测试库存不足
func (suite *CartServiceTestSuite) TestCartService_AddToCartInsufficientStock() {
	req := &model.AddToCartRequest{
		ProductID: 1,
		Quantity:  200, // 超过库存
	}

	cartItem, err := suite.cartService.AddToCart(1, "", req)
	suite.Error(err)
	suite.Nil(cartItem)
	suite.Contains(err.Error(), "库存不足")
}

// TestCartService_UpdateCartItem 测试更新购物车商品
func (suite *CartServiceTestSuite) TestCartService_UpdateCartItem() {
	// 先添加商品
	addReq := &model.AddToCartRequest{
		ProductID: 1,
		Quantity:  2,
	}
	cartItem, err := suite.cartService.AddToCart(1, "", addReq)
	suite.NoError(err)

	// 更新商品
	updateReq := &model.UpdateCartItemRequest{
		Quantity: 5,
		Selected: false,
	}

	updatedItem, err := suite.cartService.UpdateCartItem(1, "", cartItem.ID, updateReq)
	suite.NoError(err)
	suite.NotNil(updatedItem)
	suite.Equal(5, updatedItem.Quantity)
	suite.False(updatedItem.Selected)
}

// TestCartService_RemoveFromCart 测试从购物车移除商品
func (suite *CartServiceTestSuite) TestCartService_RemoveFromCart() {
	// 先添加商品
	addReq := &model.AddToCartRequest{
		ProductID: 1,
		Quantity:  2,
	}
	cartItem, err := suite.cartService.AddToCart(1, "", addReq)
	suite.NoError(err)

	// 移除商品
	err = suite.cartService.RemoveFromCart(1, "", cartItem.ID)
	suite.NoError(err)

	// 验证商品已被删除
	var deletedItem model.CartItem
	err = suite.db.First(&deletedItem, cartItem.ID).Error
	suite.Error(err) // 应该找不到记录
}

// TestCartService_ClearCart 测试清空购物车
func (suite *CartServiceTestSuite) TestCartService_ClearCart() {
	// 先添加多个商品
	for i := 0; i < 3; i++ {
		addReq := &model.AddToCartRequest{
			ProductID: 1,
			Quantity:  1,
		}
		suite.cartService.AddToCart(1, "", addReq)
	}

	// 清空购物车
	err := suite.cartService.ClearCart(1, "")
	suite.NoError(err)

	// 验证购物车已清空
	var count int64
	suite.db.Model(&model.CartItem{}).Where("cart_id IN (SELECT id FROM carts WHERE user_id = ?)", 1).Count(&count)
	suite.Equal(int64(0), count)
}

// TestCartService_GetCart 测试获取购物车
func (suite *CartServiceTestSuite) TestCartService_GetCart() {
	// 先添加商品
	addReq := &model.AddToCartRequest{
		ProductID: 1,
		Quantity:  2,
	}
	suite.cartService.AddToCart(1, "", addReq)

	// 获取购物车
	cartResponse, err := suite.cartService.GetCart(1, "", false)
	suite.NoError(err)
	suite.NotNil(cartResponse)
	suite.NotNil(cartResponse.Cart)
	suite.NotNil(cartResponse.Summary)
	suite.Len(cartResponse.Cart.Items, 1)
	suite.Equal(1, cartResponse.Summary.ItemCount)
	suite.Equal(2, cartResponse.Summary.TotalQty)
}

// TestCartService_GetEmptyCart 测试获取空购物车
func (suite *CartServiceTestSuite) TestCartService_GetEmptyCart() {
	cartResponse, err := suite.cartService.GetCart(999, "", false)
	suite.NoError(err)
	suite.NotNil(cartResponse)
	suite.NotNil(cartResponse.Cart)
	suite.NotNil(cartResponse.Summary)
	suite.Equal(0, cartResponse.Summary.ItemCount)
	suite.Equal(0, cartResponse.Summary.TotalQty)
}

// TestCartService_BatchUpdateCart 测试批量更新购物车
func (suite *CartServiceTestSuite) TestCartService_BatchUpdateCart() {
	// 先添加多个商品
	var itemIDs []uint
	for i := 0; i < 3; i++ {
		addReq := &model.AddToCartRequest{
			ProductID: 1,
			Quantity:  1,
		}
		cartItem, err := suite.cartService.AddToCart(1, "", addReq)
		suite.NoError(err)
		itemIDs = append(itemIDs, cartItem.ID)
	}

	// 批量更新
	batchReq := &model.BatchUpdateCartRequest{
		Items: []struct {
			ID       uint `json:"id" binding:"required"`
			Quantity int  `json:"quantity" binding:"min=1"`
			Selected bool `json:"selected"`
		}{
			{ID: itemIDs[0], Quantity: 5, Selected: true},
			{ID: itemIDs[1], Quantity: 3, Selected: false},
			{ID: itemIDs[2], Quantity: 2, Selected: true},
		},
	}

	err := suite.cartService.BatchUpdateCart(1, "", batchReq)
	suite.NoError(err)

	// 验证更新结果
	var items []model.CartItem
	suite.db.Where("id IN ?", itemIDs).Find(&items)
	suite.Len(items, 3)

	for _, item := range items {
		switch item.ID {
		case itemIDs[0]:
			suite.Equal(5, item.Quantity)
			suite.True(item.Selected)
		case itemIDs[1]:
			suite.Equal(3, item.Quantity)
			suite.False(item.Selected)
		case itemIDs[2]:
			suite.Equal(2, item.Quantity)
			suite.True(item.Selected)
		}
	}
}

// TestCartService_SelectAllItems 测试全选/取消全选
func (suite *CartServiceTestSuite) TestCartService_SelectAllItems() {
	// 先添加多个商品
	for i := 0; i < 3; i++ {
		addReq := &model.AddToCartRequest{
			ProductID: 1,
			Quantity:  1,
		}
		suite.cartService.AddToCart(1, "", addReq)
	}

	// 取消全选
	err := suite.cartService.SelectAllItems(1, "", false)
	suite.NoError(err)

	// 验证所有商品都未选中
	var selectedCount int64
	suite.db.Model(&model.CartItem{}).
		Where("cart_id IN (SELECT id FROM carts WHERE user_id = ?) AND selected = ?", 1, true).
		Count(&selectedCount)
	suite.Equal(int64(0), selectedCount)

	// 全选
	err = suite.cartService.SelectAllItems(1, "", true)
	suite.NoError(err)

	// 验证所有商品都已选中
	suite.db.Model(&model.CartItem{}).
		Where("cart_id IN (SELECT id FROM carts WHERE user_id = ?) AND selected = ?", 1, true).
		Count(&selectedCount)
	suite.Equal(int64(3), selectedCount)
}

// TestCartService_GetCartItemCount 测试获取购物车商品数量
func (suite *CartServiceTestSuite) TestCartService_GetCartItemCount() {
	// 先添加商品
	for i := 0; i < 5; i++ {
		addReq := &model.AddToCartRequest{
			ProductID: 1,
			Quantity:  1,
		}
		suite.cartService.AddToCart(1, "", addReq)
	}

	// 获取商品数量
	count, err := suite.cartService.GetCartItemCount(1, "")
	suite.NoError(err)
	suite.Equal(5, count)

	// 测试空购物车
	emptyCount, err := suite.cartService.GetCartItemCount(999, "")
	suite.NoError(err)
	suite.Equal(0, emptyCount)
}

// TestCartService_InvalidProduct 测试无效商品
func (suite *CartServiceTestSuite) TestCartService_InvalidProduct() {
	req := &model.AddToCartRequest{
		ProductID: 999, // 不存在的商品
		Quantity:  1,
	}

	cartItem, err := suite.cartService.AddToCart(1, "", req)
	suite.Error(err)
	suite.Nil(cartItem)
	suite.Contains(err.Error(), "商品不存在")
}

// TestCartService_InvalidSKU 测试无效SKU
func (suite *CartServiceTestSuite) TestCartService_InvalidSKU() {
	req := &model.AddToCartRequest{
		ProductID: 1,
		SKUID:     999, // 不存在的SKU
		Quantity:  1,
	}

	cartItem, err := suite.cartService.AddToCart(1, "", req)
	suite.Error(err)
	suite.Nil(cartItem)
	suite.Contains(err.Error(), "商品规格不存在")
}

// 运行购物车服务测试套件
func TestCartServiceSuite(t *testing.T) {
	suite.Run(t, new(CartServiceTestSuite))
}

// TestNewCartService 测试创建购物车服务
func TestNewCartService(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	service := NewCartService(db)
	assert.NotNil(t, service)
	assert.Equal(t, db, service.db)
}
