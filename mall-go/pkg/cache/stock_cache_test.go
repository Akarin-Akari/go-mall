package cache

import (
	"mall-go/internal/model"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// 创建测试用的商品数据
func createTestProductForStock(id uint, stock int, minStock int) *model.Product {
	return &model.Product{
		ID:       id,
		Name:     "测试商品",
		Stock:    stock,
		MinStock: minStock,
		MaxStock: 1000,
		Version:  1,
		Status:   "active",
		Price:    decimal.NewFromFloat(99.99),

		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func setupStockCacheService() (*StockCacheService, *MockCacheManager) {
	mockCache := new(MockCacheManager)
	keyManager := NewCacheKeyManager("test")
	service := NewStockCacheService(mockCache, keyManager)
	return service, mockCache
}

func TestStockCacheService_GetStock(t *testing.T) {
	service, mockCache := setupStockCacheService()

	// 测试缓存命中
	t.Run("缓存命中", func(t *testing.T) {
		productID := uint(123)
		expectedKey := "test:stock:123"

		// 模拟缓存数据
		cacheData := `{"product_id":123,"stock":100,"min_stock":10,"version":1,"status":"active","cached_at":"2025-01-10T10:00:00Z"}`

		mockCache.On("Get", expectedKey).Return(cacheData, nil)

		result, err := service.GetStock(productID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, uint(123), result.ProductID)
		assert.Equal(t, 100, result.Stock)
		assert.Equal(t, 1, result.Version)

		mockCache.AssertExpectations(t)
	})

	// 测试缓存未命中
	t.Run("缓存未命中", func(t *testing.T) {
		productID := uint(456)
		expectedKey := "test:stock:456"

		mockCache.On("Get", expectedKey).Return(nil, nil)

		result, err := service.GetStock(productID)

		assert.NoError(t, err)
		assert.Nil(t, result)

		mockCache.AssertExpectations(t)
	})
}

func TestStockCacheService_SetStock(t *testing.T) {
	service, mockCache := setupStockCacheService()

	product := createTestProductForStock(123, 100, 10)
	expectedKey := "test:stock:123"

	mockCache.On("Set", expectedKey, mock.AnythingOfType("string"), 10*time.Minute).Return(nil)

	err := service.SetStock(product)

	assert.NoError(t, err)
	mockCache.AssertExpectations(t)
}

func TestStockCacheService_UpdateStock(t *testing.T) {
	service, mockCache := setupStockCacheService()

	productID := uint(123)
	expectedKey := "test:stock:123"

	// 模拟当前库存数据
	currentData := `{"product_id":123,"stock":100,"min_stock":10,"version":1,"status":"active","cached_at":"2025-01-10T10:00:00Z"}`

	mockCache.On("Get", expectedKey).Return(currentData, nil)
	mockCache.On("Set", expectedKey, mock.AnythingOfType("string"), 10*time.Minute).Return(nil)

	err := service.UpdateStock(productID, 80, 2)

	assert.NoError(t, err)
	mockCache.AssertExpectations(t)
}

func TestStockCacheService_DeductStockWithOptimisticLock(t *testing.T) {
	service, mockCache := setupStockCacheService()

	// 测试成功扣减
	t.Run("成功扣减库存", func(t *testing.T) {
		productID := uint(123)
		expectedKey := "test:stock:123"

		// 模拟当前库存数据
		currentData := `{"product_id":123,"stock":100,"min_stock":10,"version":1,"status":"active","cached_at":"2025-01-10T10:00:00Z"}`

		request := &StockDeductionRequest{
			ProductID: productID,
			Quantity:  10,
			Reason:    "order",
		}

		mockCache.On("Get", expectedKey).Return(currentData, nil)
		mockCache.On("HGet", expectedKey, "version").Return("1", nil)
		mockCache.On("HMSet", expectedKey, mock.AnythingOfType("map[string]interface {}")).Return(nil)
		mockCache.On("Expire", expectedKey, 10*time.Minute).Return(nil)
		// 低库存预警可能触发，也可能不触发，取决于扣减后的库存
		mockCache.On("LPush", "mall:low_stock_alerts", mock.AnythingOfType("string")).Return(nil).Maybe()

		result, err := service.DeductStockWithOptimisticLock(request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.Success)
		assert.Equal(t, 100, result.OldStock)
		assert.Equal(t, 90, result.NewStock)
		assert.Equal(t, 1, result.OldVersion)
		assert.Equal(t, 2, result.NewVersion)

		mockCache.AssertExpectations(t)
	})

	// 测试库存不足
	t.Run("库存不足", func(t *testing.T) {
		productID := uint(456)
		expectedKey := "test:stock:456"

		// 模拟库存不足的数据
		currentData := `{"product_id":456,"stock":5,"min_stock":10,"version":1,"status":"active","cached_at":"2025-01-10T10:00:00Z"}`

		request := &StockDeductionRequest{
			ProductID: productID,
			Quantity:  10,
			Reason:    "order",
		}

		mockCache.On("Get", expectedKey).Return(currentData, nil)

		result, err := service.DeductStockWithOptimisticLock(request)

		assert.Error(t, err)
		assert.NotNil(t, result)
		assert.False(t, result.Success)
		assert.Contains(t, result.Error, "库存不足")

		mockCache.AssertExpectations(t)
	})
}

func TestStockCacheService_BatchDeductStock(t *testing.T) {
	service, mockCache := setupStockCacheService()

	requests := []*StockDeductionRequest{
		{ProductID: 123, Quantity: 5, Reason: "order"},
		{ProductID: 456, Quantity: 3, Reason: "order"},
	}

	// 模拟第一个商品的扣减
	mockCache.On("Get", "test:stock:123").Return(
		`{"product_id":123,"stock":100,"min_stock":10,"version":1,"status":"active","cached_at":"2025-01-10T10:00:00Z"}`, nil)
	mockCache.On("HGet", "test:stock:123", "version").Return("1", nil)
	mockCache.On("HMSet", "test:stock:123", mock.AnythingOfType("map[string]interface {}")).Return(nil)
	mockCache.On("Expire", "test:stock:123", 10*time.Minute).Return(nil)

	// 模拟第二个商品的扣减
	mockCache.On("Get", "test:stock:456").Return(
		`{"product_id":456,"stock":50,"min_stock":5,"version":1,"status":"active","cached_at":"2025-01-10T10:00:00Z"}`, nil)
	mockCache.On("HGet", "test:stock:456", "version").Return("1", nil)
	mockCache.On("HMSet", "test:stock:456", mock.AnythingOfType("map[string]interface {}")).Return(nil)
	mockCache.On("Expire", "test:stock:456", 10*time.Minute).Return(nil)

	results, err := service.BatchDeductStock(requests)

	assert.NoError(t, err)
	assert.Len(t, results, 2)
	assert.True(t, results[0].Success)
	assert.True(t, results[1].Success)

	mockCache.AssertExpectations(t)
}

func TestStockCacheService_GetStocks(t *testing.T) {
	service, mockCache := setupStockCacheService()

	productIDs := []uint{123, 456, 789}
	expectedKeys := []string{"test:stock:123", "test:stock:456", "test:stock:789"}

	// 模拟批量获取结果
	cacheValues := []interface{}{
		`{"product_id":123,"stock":100,"version":1}`,
		nil, // 缓存未命中
		`{"product_id":789,"stock":200,"version":1}`,
	}

	mockCache.On("MGet", expectedKeys).Return(cacheValues, nil)

	result, err := service.GetStocks(productIDs)

	assert.NoError(t, err)
	assert.Len(t, result, 2) // 只有2个命中
	assert.Contains(t, result, uint(123))
	assert.Contains(t, result, uint(789))
	assert.NotContains(t, result, uint(456)) // 未命中的不包含

	mockCache.AssertExpectations(t)
}

func TestStockCacheService_LowStockAlerts(t *testing.T) {
	service, mockCache := setupStockCacheService()

	// 测试获取低库存预警
	t.Run("获取低库存预警", func(t *testing.T) {
		alertKey := "mall:low_stock_alerts"
		alertData := `{"product_id":123,"current_stock":5,"min_stock":10,"alert_time":"2025-01-10T10:00:00Z"}`

		mockCache.On("LRange", alertKey, int64(0), int64(9)).Return([]interface{}{alertData}, nil)

		alerts, err := service.GetLowStockAlerts(10)

		assert.NoError(t, err)
		assert.Len(t, alerts, 1)
		assert.Equal(t, uint(123), alerts[0].ProductID)
		assert.Equal(t, 5, alerts[0].CurrentStock)
		assert.Equal(t, 10, alerts[0].MinStock)

		mockCache.AssertExpectations(t)
	})

	// 测试清空低库存预警
	t.Run("清空低库存预警", func(t *testing.T) {
		alertKey := "mall:low_stock_alerts"

		mockCache.On("Delete", alertKey).Return(nil)

		err := service.ClearLowStockAlerts()

		assert.NoError(t, err)
		mockCache.AssertExpectations(t)
	})
}

func TestStockCacheService_OutOfStockProducts(t *testing.T) {
	service, mockCache := setupStockCacheService()

	// 测试获取缺货商品
	t.Run("获取缺货商品", func(t *testing.T) {
		key := "mall:out_of_stock_products"

		mockCache.On("SMembers", key).Return([]interface{}{"123", "456"}, nil)

		productIDs, err := service.GetOutOfStockProducts()

		assert.NoError(t, err)
		assert.Len(t, productIDs, 2)
		assert.Contains(t, productIDs, uint(123))
		assert.Contains(t, productIDs, uint(456))

		mockCache.AssertExpectations(t)
	})

	// 测试添加缺货商品
	t.Run("添加缺货商品", func(t *testing.T) {
		key := "mall:out_of_stock_products"
		productID := uint(789)

		mockCache.On("SAdd", key, []interface{}{"789"}).Return(nil)

		err := service.AddOutOfStockProduct(productID)

		assert.NoError(t, err)
		mockCache.AssertExpectations(t)
	})

	// 测试移除缺货商品
	t.Run("移除缺货商品", func(t *testing.T) {
		key := "mall:out_of_stock_products"
		productID := uint(789)

		mockCache.On("SRem", key, []interface{}{"789"}).Return(nil)

		err := service.RemoveOutOfStockProduct(productID)

		assert.NoError(t, err)
		mockCache.AssertExpectations(t)
	})
}

func TestStockCacheService_StockManagement(t *testing.T) {
	service, mockCache := setupStockCacheService()

	// 测试检查库存存在
	t.Run("检查库存存在", func(t *testing.T) {
		productID := uint(123)
		expectedKey := "test:stock:123"

		mockCache.On("Exists", expectedKey).Return(true)

		exists := service.ExistsStock(productID)

		assert.True(t, exists)
		mockCache.AssertExpectations(t)
	})

	// 测试获取TTL
	t.Run("获取TTL", func(t *testing.T) {
		productID := uint(123)
		expectedKey := "test:stock:123"
		expectedTTL := 5 * time.Minute

		mockCache.On("TTL", expectedKey).Return(expectedTTL, nil)

		ttl, err := service.GetStockTTL(productID)

		assert.NoError(t, err)
		assert.Equal(t, expectedTTL, ttl)
		mockCache.AssertExpectations(t)
	})

	// 测试刷新TTL
	t.Run("刷新TTL", func(t *testing.T) {
		productID := uint(123)
		expectedKey := "test:stock:123"

		mockCache.On("Expire", expectedKey, 10*time.Minute).Return(nil)

		err := service.RefreshStockTTL(productID)

		assert.NoError(t, err)
		mockCache.AssertExpectations(t)
	})
}

func TestConvertToStockCacheData(t *testing.T) {
	product := createTestProductForStock(123, 100, 10)

	cacheData := ConvertToStockCacheData(product)

	assert.Equal(t, product.ID, cacheData.ProductID)
	assert.Equal(t, product.Stock, cacheData.Stock)
	assert.Equal(t, product.MinStock, cacheData.MinStock)
	assert.Equal(t, product.Version, cacheData.Version)
	assert.Equal(t, "active", cacheData.Status)
	assert.False(t, cacheData.IsLowStock) // 100 > 10，不是低库存
	assert.NotZero(t, cacheData.CachedAt)
}

func TestStockCacheService_GetStockStats(t *testing.T) {
	service, mockCache := setupStockCacheService()

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

	// 模拟低库存预警
	mockCache.On("GetMetrics").Return(metrics)
	mockCache.On("GetConnectionStats").Return(connStats)
	mockCache.On("LRange", "mall:low_stock_alerts", int64(0), int64(99)).Return([]interface{}{}, nil)
	mockCache.On("SMembers", "mall:out_of_stock_products").Return([]interface{}{}, nil)

	stats := service.GetStockStats()

	assert.NotNil(t, stats)
	assert.Equal(t, int64(1000), stats["total_ops"])
	assert.Equal(t, int64(800), stats["hit_count"])
	assert.Equal(t, float64(0.8), stats["hit_rate"])
	assert.Equal(t, uint32(10), stats["total_conns"])
	assert.Equal(t, uint32(5), stats["idle_conns"])
	assert.Equal(t, 0, stats["low_stock_alerts"])
	assert.Equal(t, 0, stats["out_of_stock_count"])

	mockCache.AssertExpectations(t)
}
