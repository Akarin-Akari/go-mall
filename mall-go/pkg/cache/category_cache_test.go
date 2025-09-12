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
func createTestProductForCategory(id uint, categoryID uint, name string, price float64) *model.Product {
	return &model.Product{
		ID:           id,
		Name:         name,
		CategoryID:   categoryID,
		CategoryName: fmt.Sprintf("分类%d", categoryID),
		Price:        decimal.NewFromFloat(price),
		OriginPrice:  decimal.NewFromFloat(price * 1.5),
		CostPrice:    decimal.NewFromFloat(price * 0.6),
		Stock:        100,
		SoldCount:    50,
		Version:      1,
		Status:       "active",
		IsHot:        id%2 == 0, // 偶数ID为热门商品
		IsNew:        id%3 == 0, // 3的倍数为新品
		IsRecommend:  id%5 == 0, // 5的倍数为推荐商品

		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func setupCategoryCacheService() (*CategoryCacheService, *MockCacheManager) {
	mockCache := new(MockCacheManager)
	keyManager := NewCacheKeyManager("test")
	service := NewCategoryCacheService(mockCache, keyManager)
	return service, mockCache
}

func TestCategoryCacheService_GetCategoryProducts(t *testing.T) {
	service, mockCache := setupCategoryCacheService()

	// 测试缓存命中
	t.Run("缓存命中", func(t *testing.T) {
		request := &CategoryListRequest{
			CategoryID: 1,
			Page:       1,
			PageSize:   20,
			SortBy:     "created_desc",
		}

		// 模拟缓存数据
		cacheData := `{"category_id":1,"category_name":"电子产品","products":[],"total_count":10,"page":1,"page_size":20,"sort_by":"created_desc","cached_at":"2025-01-10T10:00:00Z"}`

		mockCache.On("Get", mock.AnythingOfType("string")).Return(cacheData, nil)

		result, err := service.GetCategoryProducts(request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, uint(1), result.CategoryID)
		assert.Equal(t, "电子产品", result.CategoryName)
		assert.Equal(t, 10, result.TotalCount)
		assert.Equal(t, 1, result.Page)
		assert.Equal(t, 20, result.PageSize)

		mockCache.AssertExpectations(t)
	})

	// 测试缓存未命中
	t.Run("缓存未命中", func(t *testing.T) {
		request := &CategoryListRequest{
			CategoryID: 2,
			Page:       1,
			PageSize:   20,
		}

		mockCache.On("Get", mock.AnythingOfType("string")).Return("", nil) // 返回空字符串而不是nil

		result, err := service.GetCategoryProducts(request)

		assert.NoError(t, err)
		assert.Nil(t, result)

		mockCache.AssertExpectations(t)
	})
}

func TestCategoryCacheService_SetCategoryProducts(t *testing.T) {
	service, mockCache := setupCategoryCacheService()

	request := &CategoryListRequest{
		CategoryID: 1,
		Page:       1,
		PageSize:   20,
		SortBy:     "created_desc",
	}

	products := []*model.Product{
		createTestProductForCategory(1, 1, "iPhone 15", 999.99),
		createTestProductForCategory(2, 1, "MacBook Pro", 1999.99),
	}

	mockCache.On("Set", mock.AnythingOfType("string"), mock.AnythingOfType("string"), 60*time.Minute).Return(nil)

	err := service.SetCategoryProducts(request, products, 2, "电子产品")

	assert.NoError(t, err)
	mockCache.AssertExpectations(t)
}

func TestCategoryCacheService_GetHotProductsRanking(t *testing.T) {
	service, mockCache := setupCategoryCacheService()

	// 测试获取热门商品排行榜
	t.Run("获取热门商品排行榜", func(t *testing.T) {
		rankingType := "sales"
		period := "2025-01-10"

		// 模拟缓存数据
		cacheData := `{"ranking_type":"sales","products":[{"product_id":1,"name":"iPhone 15","rank":1,"rank_score":100.0}],"total_count":1,"period":"2025-01-10","cached_at":"2025-01-10T10:00:00Z"}`

		mockCache.On("Get", mock.AnythingOfType("string")).Return(cacheData, nil)

		result, err := service.GetHotProductsRanking(rankingType, period)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "sales", result.RankingType)
		assert.Equal(t, "2025-01-10", result.Period)
		assert.Equal(t, 1, result.TotalCount)
		assert.Len(t, result.Products, 1)

		mockCache.AssertExpectations(t)
	})
}

func TestCategoryCacheService_SetHotProductsRanking(t *testing.T) {
	service, mockCache := setupCategoryCacheService()

	rankingType := "sales"
	period := "2025-01-10"

	products := []*HotProductItem{
		{
			ProductID:    1,
			Name:         "iPhone 15",
			CategoryID:   1,
			CategoryName: "电子产品",
			Price:        "999.99",
			SalesCount:   100,
			RankScore:    100.0,
		},
		{
			ProductID:    2,
			Name:         "MacBook Pro",
			CategoryID:   1,
			CategoryName: "电子产品",
			Price:        "1999.99",
			SalesCount:   80,
			RankScore:    80.0,
		},
	}

	mockCache.On("Set", mock.AnythingOfType("string"), mock.AnythingOfType("string"), 60*time.Minute).Return(nil)

	err := service.SetHotProductsRanking(rankingType, period, products)

	assert.NoError(t, err)
	mockCache.AssertExpectations(t)
}

func TestCategoryCacheService_GetSearchResults(t *testing.T) {
	service, mockCache := setupCategoryCacheService()

	// 测试获取搜索结果
	t.Run("获取搜索结果", func(t *testing.T) {
		request := &SearchRequest{
			Keyword:  "iPhone",
			Page:     1,
			PageSize: 20,
			SortBy:   "relevance",
		}

		// 模拟缓存数据
		cacheData := `{"keyword":"iPhone","products":[],"total_count":5,"page":1,"page_size":20,"sort_by":"relevance","search_time":50000000,"cached_at":"2025-01-10T10:00:00Z"}`

		mockCache.On("Get", mock.AnythingOfType("string")).Return(cacheData, nil)

		result, err := service.GetSearchResults(request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "iPhone", result.Keyword)
		assert.Equal(t, 5, result.TotalCount)
		assert.Equal(t, 1, result.Page)
		assert.Equal(t, 20, result.PageSize)
		assert.Equal(t, "relevance", result.SortBy)

		mockCache.AssertExpectations(t)
	})
}

func TestCategoryCacheService_SetSearchResults(t *testing.T) {
	service, mockCache := setupCategoryCacheService()

	request := &SearchRequest{
		Keyword:  "iPhone",
		Page:     1,
		PageSize: 20,
		SortBy:   "relevance",
	}

	products := []*model.Product{
		createTestProductForCategory(1, 1, "iPhone 15", 999.99),
		createTestProductForCategory(2, 1, "iPhone 14", 799.99),
	}

	searchTime := 50 * time.Millisecond

	mockCache.On("Set", mock.AnythingOfType("string"), mock.AnythingOfType("string"), 30*time.Minute).Return(nil)

	err := service.SetSearchResults(request, products, 2, searchTime)

	assert.NoError(t, err)
	mockCache.AssertExpectations(t)
}

func TestCategoryCacheService_GetRecommendProducts(t *testing.T) {
	service, mockCache := setupCategoryCacheService()

	// 测试获取推荐商品
	t.Run("获取推荐商品", func(t *testing.T) {
		recommendType := "related"
		targetID := uint(1)

		// 模拟缓存数据
		cacheData := `{"recommend_type":"related","target_id":1,"products":[],"total_count":3,"algorithm":"collaborative_filtering","confidence":0.85,"cached_at":"2025-01-10T10:00:00Z"}`

		mockCache.On("Get", mock.AnythingOfType("string")).Return(cacheData, nil)

		result, err := service.GetRecommendProducts(recommendType, targetID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "related", result.RecommendType)
		assert.Equal(t, uint(1), result.TargetID)
		assert.Equal(t, 3, result.TotalCount)
		assert.Equal(t, "collaborative_filtering", result.Algorithm)
		assert.Equal(t, 0.85, result.Confidence)

		mockCache.AssertExpectations(t)
	})
}

func TestCategoryCacheService_SetRecommendProducts(t *testing.T) {
	service, mockCache := setupCategoryCacheService()

	recommendType := "related"
	targetID := uint(1)
	algorithm := "collaborative_filtering"
	confidence := 0.85

	products := []*model.Product{
		createTestProductForCategory(2, 1, "相关商品1", 599.99),
		createTestProductForCategory(3, 1, "相关商品2", 699.99),
	}

	mockCache.On("Set", mock.AnythingOfType("string"), mock.AnythingOfType("string"), 60*time.Minute).Return(nil)

	err := service.SetRecommendProducts(recommendType, targetID, products, algorithm, confidence)

	assert.NoError(t, err)
	mockCache.AssertExpectations(t)
}

func TestCategoryCacheService_DeleteOperations(t *testing.T) {
	service, mockCache := setupCategoryCacheService()

	// 测试删除分类商品缓存
	t.Run("删除分类商品缓存", func(t *testing.T) {
		categoryID := uint(1)

		mockCache.On("Delete", mock.AnythingOfType("string")).Return(nil).Times(2)

		err := service.DeleteCategoryProducts(categoryID)

		assert.NoError(t, err)
		mockCache.AssertExpectations(t)
	})

	// 测试删除搜索结果缓存
	t.Run("删除搜索结果缓存", func(t *testing.T) {
		keyword := "iPhone"

		mockCache.On("Delete", mock.AnythingOfType("string")).Return(nil).Times(2)

		err := service.DeleteSearchResults(keyword)

		assert.NoError(t, err)
		mockCache.AssertExpectations(t)
	})

	// 测试删除推荐商品缓存
	t.Run("删除推荐商品缓存", func(t *testing.T) {
		recommendType := "related"
		targetID := uint(1)

		mockCache.On("Delete", mock.AnythingOfType("string")).Return(nil)

		err := service.DeleteRecommendProducts(recommendType, targetID)

		assert.NoError(t, err)
		mockCache.AssertExpectations(t)
	})

	// 测试删除热门商品排行榜缓存
	t.Run("删除热门商品排行榜缓存", func(t *testing.T) {
		rankingType := "sales"
		period := "2025-01-10"

		mockCache.On("Delete", mock.AnythingOfType("string")).Return(nil)

		err := service.DeleteHotProductsRanking(rankingType, period)

		assert.NoError(t, err)
		mockCache.AssertExpectations(t)
	})
}

func TestCategoryCacheService_ExistsOperations(t *testing.T) {
	service, mockCache := setupCategoryCacheService()

	// 测试检查分类商品缓存存在
	t.Run("检查分类商品缓存存在", func(t *testing.T) {
		request := &CategoryListRequest{
			CategoryID: 1,
			Page:       1,
			PageSize:   20,
		}

		mockCache.On("Exists", mock.AnythingOfType("string")).Return(true)

		exists := service.ExistsCategoryProducts(request)

		assert.True(t, exists)
		mockCache.AssertExpectations(t)
	})

	// 测试检查搜索结果缓存存在
	t.Run("检查搜索结果缓存存在", func(t *testing.T) {
		request := &SearchRequest{
			Keyword:  "iPhone",
			Page:     1,
			PageSize: 20,
		}

		mockCache.On("Exists", mock.AnythingOfType("string")).Return(true)

		exists := service.ExistsSearchResults(request)

		assert.True(t, exists)
		mockCache.AssertExpectations(t)
	})
}

func TestCategoryCacheService_TTLOperations(t *testing.T) {
	service, mockCache := setupCategoryCacheService()

	// 测试获取TTL
	t.Run("获取TTL", func(t *testing.T) {
		request := &CategoryListRequest{
			CategoryID: 1,
			Page:       1,
			PageSize:   20,
		}
		expectedTTL := 30 * time.Minute

		mockCache.On("TTL", mock.AnythingOfType("string")).Return(expectedTTL, nil)

		ttl, err := service.GetCategoryProductsTTL(request)

		assert.NoError(t, err)
		assert.Equal(t, expectedTTL, ttl)
		mockCache.AssertExpectations(t)
	})

	// 测试刷新TTL
	t.Run("刷新TTL", func(t *testing.T) {
		request := &CategoryListRequest{
			CategoryID: 1,
			Page:       1,
			PageSize:   20,
		}

		mockCache.On("Expire", mock.AnythingOfType("string"), 60*time.Minute).Return(nil)

		err := service.RefreshCategoryProductsTTL(request)

		assert.NoError(t, err)
		mockCache.AssertExpectations(t)
	})
}

func TestCategoryCacheService_GetCategoryStats(t *testing.T) {
	service, mockCache := setupCategoryCacheService()

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

	stats := service.GetCategoryStats()

	assert.NotNil(t, stats)
	assert.Equal(t, int64(1000), stats["total_ops"])
	assert.Equal(t, int64(800), stats["hit_count"])
	assert.Equal(t, float64(0.8), stats["hit_rate"])
	assert.Equal(t, uint32(10), stats["total_conns"])
	assert.Equal(t, uint32(5), stats["idle_conns"])

	mockCache.AssertExpectations(t)
}

func TestCategoryCacheService_HotProductsScore(t *testing.T) {
	service, mockCache := setupCategoryCacheService()

	// 测试更新热门商品分数
	t.Run("更新热门商品分数", func(t *testing.T) {
		productID := uint(1)
		scoreType := "sales"
		increment := 10.5

		mockCache.On("ZScore", mock.AnythingOfType("string"), "1").Return(5.0, nil)
		mockCache.On("ZAdd", mock.AnythingOfType("string"), mock.MatchedBy(func(members []redis.Z) bool {
			return len(members) == 1 && members[0].Member == "1"
		})).Return(nil)
		mockCache.On("Expire", mock.AnythingOfType("string"), 60*time.Minute).Return(nil)

		err := service.UpdateHotProductsScore(productID, scoreType, increment)

		assert.NoError(t, err)
		mockCache.AssertExpectations(t)
	})

	// 测试获取热门商品排行
	t.Run("获取热门商品排行", func(t *testing.T) {
		scoreType := "sales"
		limit := int64(10)

		members := []interface{}{"1", "2", "3"}
		mockCache.On("ZRange", mock.AnythingOfType("string"), int64(0), int64(9)).Return(members, nil)

		productIDs, err := service.GetTopHotProducts(scoreType, limit)

		assert.NoError(t, err)
		assert.Len(t, productIDs, 3)
		assert.Equal(t, uint(3), productIDs[0]) // 反转后的顺序
		assert.Equal(t, uint(2), productIDs[1])
		assert.Equal(t, uint(1), productIDs[2])

		mockCache.AssertExpectations(t)
	})
}
