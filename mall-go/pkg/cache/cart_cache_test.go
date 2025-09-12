package cache

import (
	"mall-go/internal/model"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// 创建测试用的购物车数据
func createTestCart(id uint, userID uint, sessionID string) *model.Cart {
	return &model.Cart{
		ID:          id,
		UserID:      userID,
		SessionID:   sessionID,
		Status:      model.CartStatusActive,
		ItemCount:   2,
		TotalQty:    3,
		TotalAmount: decimal.NewFromFloat(299.98),
		Version:     1,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Items: []model.CartItem{
			{
				ID:           1,
				CartID:       id,
				ProductID:    101,
				SKUID:        0,
				Quantity:     1,
				Price:        decimal.NewFromFloat(199.99),
				ProductName:  "测试商品1",
				ProductImage: "https://example.com/product1.jpg",
				SKUName:      "",
				SKUImage:     "",
				SKUAttrs:     "",
				Selected:     true,
				Status:       model.CartItemStatusNormal,
				Version:      1,
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			},
			{
				ID:           2,
				CartID:       id,
				ProductID:    102,
				SKUID:        201,
				Quantity:     2,
				Price:        decimal.NewFromFloat(49.99),
				ProductName:  "测试商品2",
				ProductImage: "https://example.com/product2.jpg",
				SKUName:      "红色-L",
				SKUImage:     "https://example.com/sku2.jpg",
				SKUAttrs:     `{"color":"红色","size":"L"}`,
				Selected:     true,
				Status:       model.CartItemStatusNormal,
				Version:      1,
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			},
		},
	}
}

// 创建测试用的购物车汇总数据
func createTestCartSummary() *model.CartSummary {
	return &model.CartSummary{
		ItemCount:      2,
		TotalQty:       3,
		SelectedCount:  2,
		SelectedQty:    3,
		TotalAmount:    decimal.NewFromFloat(299.98),
		SelectedAmount: decimal.NewFromFloat(299.98),
		DiscountAmount: decimal.NewFromFloat(0),
		ShippingFee:    decimal.NewFromFloat(10.00),
		FinalAmount:    decimal.NewFromFloat(309.98),
		InvalidItems:   []model.CartItem{},
	}
}

func setupCartCacheService() (*CartCacheService, *MockCacheManager) {
	mockCache := new(MockCacheManager)
	keyManager := NewCacheKeyManager("test")
	service := NewCartCacheService(mockCache, keyManager)
	return service, mockCache
}

func TestCartCacheService_GetUserCart(t *testing.T) {
	service, mockCache := setupCartCacheService()

	// 测试缓存命中
	t.Run("用户购物车缓存命中", func(t *testing.T) {
		userID := uint(1)

		// 模拟缓存数据
		cacheData := `{"cart_id":1,"user_id":1,"session_id":"","status":"active","item_count":2,"total_qty":3,"total_amount":"299.98","items":[{"id":1,"cart_id":1,"product_id":101,"sku_id":0,"quantity":1,"price":"199.99","product_name":"测试商品1","product_image":"https://example.com/product1.jpg","sku_name":"","sku_image":"","sku_attrs":"","selected":true,"status":"normal","cached_at":"2025-01-10T10:00:00Z","updated_at":"2025-01-10T09:00:00Z","version":1}],"cached_at":"2025-01-10T10:00:00Z","updated_at":"2025-01-10T09:00:00Z","version":1}`

		mockCache.On("Get", mock.AnythingOfType("string")).Return(cacheData, nil)

		result, err := service.GetUserCart(userID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, uint(1), result.CartID)
		assert.Equal(t, uint(1), result.UserID)
		assert.Equal(t, "active", result.Status)
		assert.Equal(t, 2, result.ItemCount)
		assert.Equal(t, 3, result.TotalQty)
		assert.Equal(t, "299.98", result.TotalAmount)
		assert.Len(t, result.Items, 1)

		mockCache.AssertExpectations(t)
	})

	// 测试缓存未命中
	t.Run("用户购物车缓存未命中", func(t *testing.T) {
		// 创建新的mock实例避免冲突
		newMockCache := new(MockCacheManager)
		newKeyManager := NewCacheKeyManager("test")
		newService := NewCartCacheService(newMockCache, newKeyManager)

		userID := uint(2)

		newMockCache.On("Get", mock.AnythingOfType("string")).Return(nil, nil)

		result, err := newService.GetUserCart(userID)

		assert.NoError(t, err)
		assert.Nil(t, result)

		newMockCache.AssertExpectations(t)
	})
}

func TestCartCacheService_SetUserCart(t *testing.T) {
	service, mockCache := setupCartCacheService()

	cart := createTestCart(1, 1, "")

	mockCache.On("Set", mock.AnythingOfType("string"), mock.AnythingOfType("string"), 24*time.Hour).Return(nil)

	err := service.SetUserCart(cart)

	assert.NoError(t, err)
	mockCache.AssertExpectations(t)
}

func TestCartCacheService_GetGuestCart(t *testing.T) {
	service, mockCache := setupCartCacheService()

	// 测试游客购物车缓存命中
	t.Run("游客购物车缓存命中", func(t *testing.T) {
		sessionID := "guest_session_123"

		// 模拟缓存数据
		cacheData := `{"cart_id":2,"user_id":0,"session_id":"guest_session_123","status":"active","item_count":1,"total_qty":1,"total_amount":"199.99","items":[{"id":3,"cart_id":2,"product_id":103,"sku_id":0,"quantity":1,"price":"199.99","product_name":"游客商品","product_image":"https://example.com/guest.jpg","sku_name":"","sku_image":"","sku_attrs":"","selected":true,"status":"normal","cached_at":"2025-01-10T10:00:00Z","updated_at":"2025-01-10T09:00:00Z","version":1}],"cached_at":"2025-01-10T10:00:00Z","updated_at":"2025-01-10T09:00:00Z","version":1}`

		mockCache.On("Get", mock.AnythingOfType("string")).Return(cacheData, nil)

		result, err := service.GetGuestCart(sessionID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, uint(2), result.CartID)
		assert.Equal(t, uint(0), result.UserID)
		assert.Equal(t, sessionID, result.SessionID)
		assert.Equal(t, "active", result.Status)

		mockCache.AssertExpectations(t)
	})
}

func TestCartCacheService_SetGuestCart(t *testing.T) {
	service, mockCache := setupCartCacheService()

	cart := createTestCart(2, 0, "guest_session_123")

	mockCache.On("Set", mock.AnythingOfType("string"), mock.AnythingOfType("string"), 24*time.Hour).Return(nil)

	err := service.SetGuestCart(cart)

	assert.NoError(t, err)
	mockCache.AssertExpectations(t)
}

func TestCartCacheService_GetCartSummary(t *testing.T) {
	service, mockCache := setupCartCacheService()

	cartID := uint(1)

	// 模拟缓存数据
	cacheData := `{"item_count":2,"total_qty":3,"selected_count":2,"selected_qty":3,"total_amount":"299.98","selected_amount":"299.98","discount_amount":"0","shipping_fee":"10.00","final_amount":"309.98","invalid_items":[],"cached_at":"2025-01-10T10:00:00Z"}`

	mockCache.On("Get", mock.AnythingOfType("string")).Return(cacheData, nil)

	result, err := service.GetCartSummary(cartID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 2, result.ItemCount)
	assert.Equal(t, 3, result.TotalQty)
	assert.Equal(t, 2, result.SelectedCount)
	assert.Equal(t, 3, result.SelectedQty)
	assert.Equal(t, "299.98", result.TotalAmount)
	assert.Equal(t, "309.98", result.FinalAmount)

	mockCache.AssertExpectations(t)
}

func TestCartCacheService_SetCartSummary(t *testing.T) {
	service, mockCache := setupCartCacheService()

	cartID := uint(1)
	summary := createTestCartSummary()

	mockCache.On("Set", mock.AnythingOfType("string"), mock.AnythingOfType("string"), 24*time.Hour).Return(nil)

	err := service.SetCartSummary(cartID, summary)

	assert.NoError(t, err)
	mockCache.AssertExpectations(t)
}

func TestCartCacheService_CartItemOperations(t *testing.T) {
	service, mockCache := setupCartCacheService()

	// 测试设置购物车商品项
	t.Run("设置购物车商品项", func(t *testing.T) {
		item := &model.CartItem{
			ID:           1,
			CartID:       1,
			ProductID:    101,
			Quantity:     1,
			Price:        decimal.NewFromFloat(199.99),
			ProductName:  "测试商品",
			ProductImage: "https://example.com/product.jpg",
			Selected:     true,
			Status:       model.CartItemStatusNormal,
			Version:      1,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		mockCache.On("Set", mock.AnythingOfType("string"), mock.AnythingOfType("string"), 24*time.Hour).Return(nil)

		err := service.SetCartItem(item)

		assert.NoError(t, err)
		mockCache.AssertExpectations(t)
	})

	// 测试获取购物车商品项
	t.Run("获取购物车商品项", func(t *testing.T) {
		// 创建新的mock实例避免冲突
		newMockCache := new(MockCacheManager)
		newKeyManager := NewCacheKeyManager("test")
		newService := NewCartCacheService(newMockCache, newKeyManager)

		cartID := uint(1)
		itemID := uint(1)

		// 模拟缓存数据
		cacheData := `{"id":1,"cart_id":1,"product_id":101,"sku_id":0,"quantity":1,"price":"199.99","product_name":"测试商品","product_image":"https://example.com/product.jpg","sku_name":"","sku_image":"","sku_attrs":"","selected":true,"status":"normal","cached_at":"2025-01-10T10:00:00Z","updated_at":"2025-01-10T09:00:00Z","version":1}`

		newMockCache.On("Get", mock.AnythingOfType("string")).Return(cacheData, nil)

		result, err := newService.GetCartItem(cartID, itemID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, uint(1), result.ID)
		assert.Equal(t, uint(1), result.CartID)
		assert.Equal(t, uint(101), result.ProductID)
		assert.Equal(t, 1, result.Quantity)
		assert.Equal(t, "199.99", result.Price)
		assert.True(t, result.Selected)

		newMockCache.AssertExpectations(t)
	})

	// 测试删除购物车商品项
	t.Run("删除购物车商品项", func(t *testing.T) {
		// 创建新的mock实例避免冲突
		newMockCache := new(MockCacheManager)
		newKeyManager := NewCacheKeyManager("test")
		newService := NewCartCacheService(newMockCache, newKeyManager)

		cartID := uint(1)
		itemID := uint(1)

		newMockCache.On("Delete", mock.AnythingOfType("string")).Return(nil)

		err := newService.DeleteCartItem(cartID, itemID)

		assert.NoError(t, err)
		newMockCache.AssertExpectations(t)
	})
}

func TestCartCacheService_BatchOperations(t *testing.T) {
	service, mockCache := setupCartCacheService()

	// 测试批量删除购物车商品项
	t.Run("批量删除购物车商品项", func(t *testing.T) {
		cartID := uint(1)
		itemIDs := []uint{1, 2, 3}

		mockCache.On("MDelete", mock.AnythingOfType("[]string")).Return(nil)

		err := service.BatchDeleteCartItems(cartID, itemIDs)

		assert.NoError(t, err)
		mockCache.AssertExpectations(t)
	})
}

func TestCartCacheService_DeleteOperations(t *testing.T) {
	service, mockCache := setupCartCacheService()

	// 测试删除用户购物车
	t.Run("删除用户购物车", func(t *testing.T) {
		userID := uint(1)

		mockCache.On("Delete", mock.AnythingOfType("string")).Return(nil)

		err := service.DeleteUserCart(userID)

		assert.NoError(t, err)
		mockCache.AssertExpectations(t)
	})

	// 测试删除游客购物车
	t.Run("删除游客购物车", func(t *testing.T) {
		// 创建新的mock实例避免冲突
		newMockCache := new(MockCacheManager)
		newKeyManager := NewCacheKeyManager("test")
		newService := NewCartCacheService(newMockCache, newKeyManager)

		sessionID := "guest_session_123"

		newMockCache.On("Delete", mock.AnythingOfType("string")).Return(nil)

		err := newService.DeleteGuestCart(sessionID)

		assert.NoError(t, err)
		newMockCache.AssertExpectations(t)
	})
}

func TestCartCacheService_ExistsOperations(t *testing.T) {
	service, mockCache := setupCartCacheService()

	// 测试检查用户购物车存在
	t.Run("检查用户购物车存在", func(t *testing.T) {
		userID := uint(1)

		mockCache.On("Exists", mock.AnythingOfType("string")).Return(true)

		exists := service.ExistsUserCart(userID)

		assert.True(t, exists)
		mockCache.AssertExpectations(t)
	})

	// 测试检查游客购物车存在
	t.Run("检查游客购物车存在", func(t *testing.T) {
		// 创建新的mock实例避免冲突
		newMockCache := new(MockCacheManager)
		newKeyManager := NewCacheKeyManager("test")
		newService := NewCartCacheService(newMockCache, newKeyManager)

		sessionID := "guest_session_123"

		newMockCache.On("Exists", mock.AnythingOfType("string")).Return(true)

		exists := newService.ExistsGuestCart(sessionID)

		assert.True(t, exists)
		newMockCache.AssertExpectations(t)
	})
}

func TestCartCacheService_TTLOperations(t *testing.T) {
	service, mockCache := setupCartCacheService()

	// 测试获取TTL
	t.Run("获取用户购物车TTL", func(t *testing.T) {
		userID := uint(1)
		expectedTTL := 24 * time.Hour

		mockCache.On("TTL", mock.AnythingOfType("string")).Return(expectedTTL, nil)

		ttl, err := service.GetUserCartTTL(userID)

		assert.NoError(t, err)
		assert.Equal(t, expectedTTL, ttl)
		mockCache.AssertExpectations(t)
	})

	// 测试刷新TTL
	t.Run("刷新用户购物车TTL", func(t *testing.T) {
		// 创建新的mock实例避免冲突
		newMockCache := new(MockCacheManager)
		newKeyManager := NewCacheKeyManager("test")
		newService := NewCartCacheService(newMockCache, newKeyManager)

		userID := uint(1)

		newMockCache.On("Expire", mock.AnythingOfType("string"), 24*time.Hour).Return(nil)

		err := newService.RefreshUserCartTTL(userID)

		assert.NoError(t, err)
		newMockCache.AssertExpectations(t)
	})
}
