package cache

import (
	"fmt"
	"mall-go/internal/model"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// 创建测试用的商品数据
func createTestProductForPrice(id uint, price float64) *model.Product {
	return &model.Product{
		ID:          id,
		Name:        "测试商品",
		Price:       decimal.NewFromFloat(price),
		OriginPrice: decimal.NewFromFloat(price * 1.5),
		CostPrice:   decimal.NewFromFloat(price * 0.6),
		Version:     1,
		Status:      "active",

		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func setupPriceCacheService() (*PriceCacheService, *MockCacheManager) {
	mockCache := new(MockCacheManager)
	keyManager := NewCacheKeyManager("test")
	service := NewPriceCacheService(mockCache, keyManager)
	return service, mockCache
}

func TestPriceCacheService_GetPrice(t *testing.T) {
	service, mockCache := setupPriceCacheService()

	// 测试缓存命中
	t.Run("缓存命中", func(t *testing.T) {
		productID := uint(123)
		expectedKey := "test:price:123"

		// 模拟缓存数据
		cacheData := `{"product_id":123,"price":"99.99","origin_price":"149.99","price_status":"normal","cached_at":"2025-01-10T10:00:00Z"}`

		mockCache.On("Get", expectedKey).Return(cacheData, nil)

		result, err := service.GetPrice(productID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, uint(123), result.ProductID)
		assert.Equal(t, "99.99", result.Price)
		assert.Equal(t, "149.99", result.OriginPrice)
		assert.Equal(t, "normal", result.PriceStatus)

		mockCache.AssertExpectations(t)
	})

	// 测试缓存未命中
	t.Run("缓存未命中", func(t *testing.T) {
		productID := uint(456)
		expectedKey := "test:price:456"

		mockCache.On("Get", expectedKey).Return(nil, nil)

		result, err := service.GetPrice(productID)

		assert.NoError(t, err)
		assert.Nil(t, result)

		mockCache.AssertExpectations(t)
	})
}

func TestPriceCacheService_SetPrice(t *testing.T) {
	service, mockCache := setupPriceCacheService()

	product := createTestProductForPrice(123, 99.99)
	expectedKey := "test:price:123"

	mockCache.On("Set", expectedKey, mock.AnythingOfType("string"), 15*time.Minute).Return(nil)

	err := service.SetPrice(product)

	assert.NoError(t, err)
	mockCache.AssertExpectations(t)
}

func TestPriceCacheService_UpdatePrice(t *testing.T) {
	service, mockCache := setupPriceCacheService()

	productID := uint(123)
	expectedKey := "test:price:123"

	// 模拟当前价格数据
	currentData := `{"product_id":123,"price":"99.99","origin_price":"149.99","price_status":"normal","version":1,"cached_at":"2025-01-10T10:00:00Z"}`

	request := &PriceUpdateRequest{
		ProductID: productID,
		Price:     decimal.NewFromFloat(89.99),
		Reason:    "price_adjustment",
	}

	mockCache.On("Get", expectedKey).Return(currentData, nil)
	mockCache.On("Set", expectedKey, mock.AnythingOfType("string"), 15*time.Minute).Return(nil)
	mockCache.On("LPush", "mall:price_history:123", mock.MatchedBy(func(args []interface{}) bool {
		return len(args) == 1
	})).Return(nil)
	mockCache.On("LLen", "mall:price_history:123").Return(int64(5), nil)

	err := service.UpdatePrice(request)

	assert.NoError(t, err)
	mockCache.AssertExpectations(t)
}

func TestPriceCacheService_SetPromotionPrice(t *testing.T) {
	service, mockCache := setupPriceCacheService()

	productID := uint(123)
	expectedKey := "test:price:123"

	// 模拟当前价格数据
	currentData := `{"product_id":123,"price":"99.99","origin_price":"149.99","price_status":"normal","version":1,"cached_at":"2025-01-10T10:00:00Z"}`

	request := &PromotionPriceRequest{
		ProductID:      productID,
		PromotionPrice: decimal.NewFromFloat(79.99),
		StartTime:      time.Now(),
		EndTime:        time.Now().Add(24 * time.Hour),
		PromotionType:  "discount",
		PromotionValue: decimal.NewFromFloat(20),
	}

	mockCache.On("Get", expectedKey).Return(currentData, nil)
	mockCache.On("Set", expectedKey, mock.AnythingOfType("string"), 15*time.Minute).Return(nil)
	mockCache.On("LPush", "mall:price_history:123", mock.MatchedBy(func(args []interface{}) bool {
		return len(args) == 1
	})).Return(nil)
	mockCache.On("LLen", "mall:price_history:123").Return(int64(5), nil)

	err := service.SetPromotionPrice(request)

	assert.NoError(t, err)
	mockCache.AssertExpectations(t)
}

func TestPriceCacheService_GetEffectivePrice(t *testing.T) {
	service, mockCache := setupPriceCacheService()

	// 测试普通价格
	t.Run("普通价格", func(t *testing.T) {
		productID := uint(123)
		expectedKey := "test:price:123"

		// 模拟普通价格数据
		cacheData := `{"product_id":123,"price":"99.99","price_status":"normal","is_promotion":false,"cached_at":"2025-01-10T10:00:00Z"}`

		mockCache.On("Get", expectedKey).Return(cacheData, nil)

		price, err := service.GetEffectivePrice(productID, "normal")

		assert.NoError(t, err)
		assert.NotNil(t, price)
		assert.Equal(t, "99.99", price.String())

		mockCache.AssertExpectations(t)
	})

	// 测试促销价格
	t.Run("促销价格", func(t *testing.T) {
		productID := uint(456)
		expectedKey := "test:price:456"

		// 模拟促销价格数据
		now := time.Now()
		startTime := now.Add(-1 * time.Hour)
		endTime := now.Add(1 * time.Hour)

		cacheData := fmt.Sprintf(`{"product_id":456,"price":"99.99","promotion_price":"79.99","price_status":"promotion","is_promotion":true,"promotion_start_time":"%s","promotion_end_time":"%s","cached_at":"2025-01-10T10:00:00Z"}`,
			startTime.Format(time.RFC3339), endTime.Format(time.RFC3339))

		mockCache.On("Get", expectedKey).Return(cacheData, nil)

		price, err := service.GetEffectivePrice(productID, "normal")

		assert.NoError(t, err)
		assert.NotNil(t, price)
		assert.Equal(t, "79.99", price.String())

		mockCache.AssertExpectations(t)
	})
}

func TestPriceCacheService_GetPrices(t *testing.T) {
	service, mockCache := setupPriceCacheService()

	productIDs := []uint{123, 456, 789}
	expectedKeys := []string{"test:price:123", "test:price:456", "test:price:789"}

	// 模拟批量获取结果
	cacheValues := []interface{}{
		`{"product_id":123,"price":"99.99","price_status":"normal"}`,
		nil, // 缓存未命中
		`{"product_id":789,"price":"199.99","price_status":"normal"}`,
	}

	mockCache.On("MGet", expectedKeys).Return(cacheValues, nil)

	result, err := service.GetPrices(productIDs)

	assert.NoError(t, err)
	assert.Len(t, result, 2) // 只有2个命中
	assert.Contains(t, result, uint(123))
	assert.Contains(t, result, uint(789))
	assert.NotContains(t, result, uint(456)) // 未命中的不包含

	mockCache.AssertExpectations(t)
}

func TestPriceCacheService_GetPriceHistory(t *testing.T) {
	service, mockCache := setupPriceCacheService()

	productID := uint(123)
	historyKey := "mall:price_history:123"

	// 模拟价格历史数据
	historyData := []interface{}{
		`{"product_id":123,"old_price":"99.99","new_price":"89.99","change_type":"decrease","reason":"price_adjustment","change_time":"2025-01-10T10:00:00Z"}`,
		`{"product_id":123,"old_price":"89.99","new_price":"79.99","change_type":"promotion_start","reason":"促销活动开始: discount","change_time":"2025-01-10T11:00:00Z"}`,
	}

	mockCache.On("LRange", historyKey, int64(0), int64(9)).Return(historyData, nil)

	history, err := service.GetPriceHistory(productID, 10)

	assert.NoError(t, err)
	assert.Len(t, history, 2)
	assert.Equal(t, "decrease", history[0].ChangeType)
	assert.Equal(t, "promotion_start", history[1].ChangeType)

	mockCache.AssertExpectations(t)
}

func TestPriceCacheService_PromotionProducts(t *testing.T) {
	service, mockCache := setupPriceCacheService()

	// 测试获取促销商品
	t.Run("获取促销商品", func(t *testing.T) {
		key := "mall:promotion_products"

		mockCache.On("SMembers", key).Return([]interface{}{"123", "456"}, nil)

		productIDs, err := service.GetPromotionProducts()

		assert.NoError(t, err)
		assert.Len(t, productIDs, 2)
		assert.Contains(t, productIDs, uint(123))
		assert.Contains(t, productIDs, uint(456))

		mockCache.AssertExpectations(t)
	})

	// 测试添加促销商品
	t.Run("添加促销商品", func(t *testing.T) {
		key := "mall:promotion_products"
		productID := uint(789)

		mockCache.On("SAdd", key, []interface{}{"789"}).Return(nil)

		err := service.AddPromotionProduct(productID)

		assert.NoError(t, err)
		mockCache.AssertExpectations(t)
	})

	// 测试移除促销商品
	t.Run("移除促销商品", func(t *testing.T) {
		key := "mall:promotion_products"
		productID := uint(789)

		mockCache.On("SRem", key, []interface{}{"789"}).Return(nil)

		err := service.RemovePromotionProduct(productID)

		assert.NoError(t, err)
		mockCache.AssertExpectations(t)
	})
}

func TestPriceCacheService_PriceManagement(t *testing.T) {
	service, mockCache := setupPriceCacheService()

	// 测试检查价格存在
	t.Run("检查价格存在", func(t *testing.T) {
		productID := uint(123)
		expectedKey := "test:price:123"

		mockCache.On("Exists", expectedKey).Return(true)

		exists := service.ExistsPrice(productID)

		assert.True(t, exists)
		mockCache.AssertExpectations(t)
	})

	// 测试获取TTL
	t.Run("获取TTL", func(t *testing.T) {
		productID := uint(123)
		expectedKey := "test:price:123"
		expectedTTL := 10 * time.Minute

		mockCache.On("TTL", expectedKey).Return(expectedTTL, nil)

		ttl, err := service.GetPriceTTL(productID)

		assert.NoError(t, err)
		assert.Equal(t, expectedTTL, ttl)
		mockCache.AssertExpectations(t)
	})

	// 测试刷新TTL
	t.Run("刷新TTL", func(t *testing.T) {
		productID := uint(123)
		expectedKey := "test:price:123"

		mockCache.On("Expire", expectedKey, 15*time.Minute).Return(nil)

		err := service.RefreshPriceTTL(productID)

		assert.NoError(t, err)
		mockCache.AssertExpectations(t)
	})
}

func TestConvertToPriceCacheData(t *testing.T) {
	product := createTestProductForPrice(123, 99.99)

	cacheData := ConvertToPriceCacheData(product)

	assert.Equal(t, product.ID, cacheData.ProductID)
	assert.Equal(t, product.Price.String(), cacheData.Price)
	assert.Equal(t, product.OriginPrice.String(), cacheData.OriginPrice)
	assert.Equal(t, product.CostPrice.String(), cacheData.CostPrice)
	assert.Equal(t, "normal", cacheData.PriceStatus)
	assert.False(t, cacheData.IsPromotion)
	assert.Equal(t, product.Version, cacheData.Version)
	assert.NotZero(t, cacheData.CachedAt)
}

func TestPriceCacheService_GetPriceStats(t *testing.T) {
	service, mockCache := setupPriceCacheService()

	// 模拟缓存指标
	metrics := &CacheMetrics{
		TotalOps:    1000,
		HitCount:    800,
		MissCount:   200,
		HitRate:     0.8,
		ErrorCount:  5,
		LastUpdated: time.Now(),
	}

	// 模拟连接池统计
	connStats := &redis.PoolStats{
		TotalConns: 10,
		IdleConns:  5,
		Hits:       100,
		Misses:     20,
	}

	// 模拟促销商品
	mockCache.On("GetMetrics").Return(metrics)
	mockCache.On("GetConnectionStats").Return(connStats)
	mockCache.On("SMembers", "mall:promotion_products").Return([]interface{}{"123", "456"}, nil)

	stats := service.GetPriceStats()

	assert.NotNil(t, stats)
	assert.Equal(t, int64(1000), stats["total_ops"])
	assert.Equal(t, int64(800), stats["hit_count"])
	assert.Equal(t, float64(0.8), stats["hit_rate"])
	assert.Equal(t, uint32(10), stats["total_conns"])
	assert.Equal(t, uint32(5), stats["idle_conns"])
	assert.Equal(t, 2, stats["promotion_count"])

	mockCache.AssertExpectations(t)
}
