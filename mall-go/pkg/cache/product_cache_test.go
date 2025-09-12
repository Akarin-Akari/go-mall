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

// MockCacheManager 模拟缓存管理器
type MockCacheManager struct {
	mock.Mock
}

func (m *MockCacheManager) Get(key string) (interface{}, error) {
	args := m.Called(key)
	return args.Get(0), args.Error(1)
}

func (m *MockCacheManager) Set(key string, value interface{}, ttl time.Duration) error {
	args := m.Called(key, value, ttl)
	return args.Error(0)
}

func (m *MockCacheManager) Delete(key string) error {
	args := m.Called(key)
	return args.Error(0)
}

func (m *MockCacheManager) Exists(key string) bool {
	args := m.Called(key)
	return args.Bool(0)
}

func (m *MockCacheManager) Expire(key string, ttl time.Duration) error {
	args := m.Called(key, ttl)
	return args.Error(0)
}

func (m *MockCacheManager) TTL(key string) (time.Duration, error) {
	args := m.Called(key)
	return args.Get(0).(time.Duration), args.Error(1)
}

func (m *MockCacheManager) MGet(keys []string) ([]interface{}, error) {
	args := m.Called(keys)
	return args.Get(0).([]interface{}), args.Error(1)
}

func (m *MockCacheManager) MSet(pairs map[string]interface{}, ttl time.Duration) error {
	args := m.Called(pairs, ttl)
	return args.Error(0)
}

func (m *MockCacheManager) MDelete(keys []string) error {
	args := m.Called(keys)
	return args.Error(0)
}

func (m *MockCacheManager) HGet(key, field string) (interface{}, error) {
	args := m.Called(key, field)
	return args.Get(0), args.Error(1)
}

func (m *MockCacheManager) HSet(key, field string, value interface{}) error {
	args := m.Called(key, field, value)
	return args.Error(0)
}

func (m *MockCacheManager) HMGet(key string, fields []string) ([]interface{}, error) {
	args := m.Called(key, fields)
	return args.Get(0).([]interface{}), args.Error(1)
}

func (m *MockCacheManager) HMSet(key string, pairs map[string]interface{}) error {
	args := m.Called(key, pairs)
	return args.Error(0)
}

func (m *MockCacheManager) HDelete(key string, fields []string) error {
	args := m.Called(key, fields)
	return args.Error(0)
}

func (m *MockCacheManager) HExists(key, field string) bool {
	args := m.Called(key, field)
	return args.Bool(0)
}

func (m *MockCacheManager) LPush(key string, values ...interface{}) error {
	args := m.Called(key, values)
	return args.Error(0)
}

func (m *MockCacheManager) RPush(key string, values ...interface{}) error {
	args := m.Called(key, values)
	return args.Error(0)
}

func (m *MockCacheManager) LPop(key string) (interface{}, error) {
	args := m.Called(key)
	return args.Get(0), args.Error(1)
}

func (m *MockCacheManager) RPop(key string) (interface{}, error) {
	args := m.Called(key)
	return args.Get(0), args.Error(1)
}

func (m *MockCacheManager) LRange(key string, start, stop int64) ([]interface{}, error) {
	args := m.Called(key, start, stop)
	return args.Get(0).([]interface{}), args.Error(1)
}

func (m *MockCacheManager) LLen(key string) (int64, error) {
	args := m.Called(key)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockCacheManager) SAdd(key string, members ...interface{}) error {
	args := m.Called(key, members)
	return args.Error(0)
}

func (m *MockCacheManager) SMembers(key string) ([]interface{}, error) {
	args := m.Called(key)
	return args.Get(0).([]interface{}), args.Error(1)
}

func (m *MockCacheManager) SIsMember(key string, member interface{}) bool {
	args := m.Called(key, member)
	return args.Bool(0)
}

func (m *MockCacheManager) SRem(key string, members ...interface{}) error {
	args := m.Called(key, members)
	return args.Error(0)
}

func (m *MockCacheManager) ZAdd(key string, members ...redis.Z) error {
	args := m.Called(key, members)
	return args.Error(0)
}

func (m *MockCacheManager) ZRange(key string, start, stop int64) ([]interface{}, error) {
	args := m.Called(key, start, stop)
	return args.Get(0).([]interface{}), args.Error(1)
}

func (m *MockCacheManager) ZRangeByScore(key string, min, max string) ([]interface{}, error) {
	args := m.Called(key, min, max)
	return args.Get(0).([]interface{}), args.Error(1)
}

func (m *MockCacheManager) ZRem(key string, members ...interface{}) error {
	args := m.Called(key, members)
	return args.Error(0)
}

func (m *MockCacheManager) ZScore(key string, member string) (float64, error) {
	args := m.Called(key, member)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockCacheManager) GetMetrics() *CacheMetrics {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*CacheMetrics)
}

func (m *MockCacheManager) GetConnectionStats() *redis.PoolStats {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*redis.PoolStats)
}

func (m *MockCacheManager) HealthCheck() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockCacheManager) FlushDB() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockCacheManager) FlushAll() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockCacheManager) Close() error {
	args := m.Called()
	return args.Error(0)
}

// 创建测试用的商品数据
func createTestProduct() *model.Product {
	return &model.Product{
		ID:          123,
		Name:        "测试商品",
		SubTitle:    "测试副标题",
		Description: "测试描述",
		Detail:      "测试详情",
		CategoryID:  1,
		BrandID:     1,
		MerchantID:  1,

		CategoryName: "测试分类",
		BrandName:    "测试品牌",
		MerchantName: "测试商家",

		Price:       decimal.NewFromFloat(99.99),
		OriginPrice: decimal.NewFromFloat(199.99),
		CostPrice:   decimal.NewFromFloat(50.00),

		Stock:     100,
		MinStock:  10,
		MaxStock:  1000,
		SoldCount: 50,
		Version:   1,

		Weight: decimal.NewFromFloat(1.5),
		Volume: decimal.NewFromFloat(0.5),
		Unit:   "件",

		Status:      "active",
		IsHot:       true,
		IsNew:       false,
		IsRecommend: true,

		SEOTitle:       "测试SEO标题",
		SEOKeywords:    "测试,商品,关键词",
		SEODescription: "测试SEO描述",

		Sort:      100,
		ViewCount: 1000,

		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func setupProductCacheService() (*ProductCacheService, *MockCacheManager) {
	mockCache := new(MockCacheManager)
	keyManager := NewCacheKeyManager("test")
	service := NewProductCacheService(mockCache, keyManager)
	return service, mockCache
}

func TestProductCacheService_GetProduct(t *testing.T) {
	service, mockCache := setupProductCacheService()

	// 测试缓存命中
	t.Run("缓存命中", func(t *testing.T) {
		productID := uint(123)
		expectedKey := "test:product:123"

		// 模拟缓存数据
		cacheData := `{"id":123,"name":"测试商品","price":"99.99","cached_at":"2025-01-10T10:00:00Z"}`

		mockCache.On("Get", expectedKey).Return(cacheData, nil)

		result, err := service.GetProduct(productID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, uint(123), result.ID)
		assert.Equal(t, "测试商品", result.Name)
		assert.Equal(t, "99.99", result.Price)

		mockCache.AssertExpectations(t)
	})

	// 测试缓存未命中
	t.Run("缓存未命中", func(t *testing.T) {
		productID := uint(456)
		expectedKey := "test:product:456"

		mockCache.On("Get", expectedKey).Return(nil, nil)

		result, err := service.GetProduct(productID)

		assert.NoError(t, err)
		assert.Nil(t, result)

		mockCache.AssertExpectations(t)
	})
}

func TestProductCacheService_SetProduct(t *testing.T) {
	service, mockCache := setupProductCacheService()

	product := createTestProduct()
	expectedKey := "test:product:123"

	mockCache.On("Set", expectedKey, mock.AnythingOfType("string"), 30*time.Minute).Return(nil)

	err := service.SetProduct(product)

	assert.NoError(t, err)
	mockCache.AssertExpectations(t)
}

func TestProductCacheService_DeleteProduct(t *testing.T) {
	service, mockCache := setupProductCacheService()

	productID := uint(123)
	expectedKey := "test:product:123"

	mockCache.On("Delete", expectedKey).Return(nil)

	err := service.DeleteProduct(productID)

	assert.NoError(t, err)
	mockCache.AssertExpectations(t)
}

func TestProductCacheService_ExistsProduct(t *testing.T) {
	service, mockCache := setupProductCacheService()

	productID := uint(123)
	expectedKey := "test:product:123"

	mockCache.On("Exists", expectedKey).Return(true)

	exists := service.ExistsProduct(productID)

	assert.True(t, exists)
	mockCache.AssertExpectations(t)
}

func TestProductCacheService_GetProducts(t *testing.T) {
	service, mockCache := setupProductCacheService()

	productIDs := []uint{123, 456, 789}
	expectedKeys := []string{"test:product:123", "test:product:456", "test:product:789"}

	// 模拟批量获取结果
	cacheValues := []interface{}{
		`{"id":123,"name":"商品1","price":"99.99"}`,
		nil, // 缓存未命中
		`{"id":789,"name":"商品3","price":"199.99"}`,
	}

	mockCache.On("MGet", expectedKeys).Return(cacheValues, nil)

	result, err := service.GetProducts(productIDs)

	assert.NoError(t, err)
	assert.Len(t, result, 2) // 只有2个命中
	assert.Contains(t, result, uint(123))
	assert.Contains(t, result, uint(789))
	assert.NotContains(t, result, uint(456)) // 未命中的不包含

	mockCache.AssertExpectations(t)
}

func TestProductCacheService_SetProducts(t *testing.T) {
	service, mockCache := setupProductCacheService()

	products := []*model.Product{
		createTestProduct(),
		{
			ID:    456,
			Name:  "商品2",
			Price: decimal.NewFromFloat(199.99),
		},
	}

	mockCache.On("MSet", mock.AnythingOfType("map[string]interface {}"), 30*time.Minute).Return(nil)

	err := service.SetProducts(products)

	assert.NoError(t, err)
	mockCache.AssertExpectations(t)
}

func TestProductCacheService_WarmupProducts(t *testing.T) {
	service, mockCache := setupProductCacheService()

	products := []*model.Product{createTestProduct()}

	mockCache.On("MSet", mock.AnythingOfType("map[string]interface {}"), 30*time.Minute).Return(nil)

	err := service.WarmupProducts(products)

	assert.NoError(t, err)
	mockCache.AssertExpectations(t)
}

func TestProductCacheService_HotProducts(t *testing.T) {
	service, mockCache := setupProductCacheService()

	// 测试设置热门商品
	t.Run("设置热门商品", func(t *testing.T) {
		productIDs := []uint{123, 456, 789}
		expectedKey := "test:hot:products"

		mockCache.On("Set", expectedKey, mock.AnythingOfType("string"), 60*time.Minute).Return(nil)

		err := service.SetHotProducts(productIDs)

		assert.NoError(t, err)
		mockCache.AssertExpectations(t)
	})

	// 测试获取热门商品
	t.Run("获取热门商品", func(t *testing.T) {
		expectedKey := "test:hot:products"
		cacheData := `[123,456,789]`

		mockCache.On("Get", expectedKey).Return(cacheData, nil)

		result, err := service.GetHotProducts()

		assert.NoError(t, err)
		assert.Equal(t, []uint{123, 456, 789}, result)
		mockCache.AssertExpectations(t)
	})
}

func TestProductCacheService_ViewCount(t *testing.T) {
	service, mockCache := setupProductCacheService()

	// 测试增加浏览量
	t.Run("增加浏览量-新计数器", func(t *testing.T) {
		productID := uint(123)
		expectedKey := "test:counter:view:123"
		field := "count"

		mockCache.On("HExists", expectedKey, field).Return(false)
		mockCache.On("HSet", expectedKey, field, "1").Return(nil)
		mockCache.On("Expire", expectedKey, 1*time.Hour).Return(nil)

		err := service.IncrementViewCount(productID)

		assert.NoError(t, err)
		mockCache.AssertExpectations(t)
	})

	// 测试增加浏览量-已存在计数器
	t.Run("增加浏览量-已存在计数器", func(t *testing.T) {
		productID := uint(456)
		expectedKey := "test:counter:view:456"
		field := "count"

		mockCache.On("HExists", expectedKey, field).Return(true)
		mockCache.On("HGet", expectedKey, field).Return("5", nil)
		mockCache.On("HSet", expectedKey, field, "6").Return(nil)
		mockCache.On("Expire", expectedKey, 1*time.Hour).Return(nil)

		err := service.IncrementViewCount(productID)

		assert.NoError(t, err)
		mockCache.AssertExpectations(t)
	})

	// 测试获取浏览量
	t.Run("获取浏览量", func(t *testing.T) {
		productID := uint(123)
		expectedKey := "test:counter:view:123"
		field := "count"

		mockCache.On("HGet", expectedKey, field).Return("10", nil)

		count, err := service.GetViewCount(productID)

		assert.NoError(t, err)
		assert.Equal(t, 10, count)
		mockCache.AssertExpectations(t)
	})
}

func TestProductCacheService_GetCacheStats(t *testing.T) {
	service, mockCache := setupProductCacheService()

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

	mockCache.On("GetMetrics").Return(metrics)
	mockCache.On("GetConnectionStats").Return(connStats)

	stats := service.GetCacheStats()

	assert.NotNil(t, stats)
	assert.Equal(t, int64(1000), stats["total_ops"])
	assert.Equal(t, int64(800), stats["hit_count"])
	assert.Equal(t, float64(0.8), stats["hit_rate"])
	assert.Equal(t, uint32(10), stats["total_conns"])
	assert.Equal(t, uint32(5), stats["idle_conns"])

	mockCache.AssertExpectations(t)
}

func TestConvertToProductCacheData(t *testing.T) {
	product := createTestProduct()

	cacheData := ConvertToProductCacheData(product)

	assert.Equal(t, product.ID, cacheData.ID)
	assert.Equal(t, product.Name, cacheData.Name)
	assert.Equal(t, product.Price.String(), cacheData.Price)
	assert.Equal(t, product.Stock, cacheData.Stock)
	assert.Equal(t, product.Version, cacheData.Version)
	assert.Equal(t, product.Status, cacheData.Status)
	assert.NotZero(t, cacheData.CachedAt)
}
