package cache

import (
	"mall-go/internal/model"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// 创建测试用的商品数据
func createTestProductForUserPref(id uint, name string, price float64) *model.Product {
	return &model.Product{
		ID:         id,
		Name:       name,
		Price:      decimal.NewFromFloat(price),
		CategoryID: 1,
		Status:     model.ProductStatusActive,
		Images: []model.ProductImage{
			{
				ID:     1,
				URL:    "https://example.com/product.jpg",
				IsMain: true,
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// 创建测试用的浏览历史数据
func createTestBrowseHistory(userID uint) *UserBrowseHistoryCacheData {
	return &UserBrowseHistoryCacheData{
		UserID: userID,
		Items: []BrowseHistoryItem{
			{
				ProductID:    101,
				ProductName:  "iPhone 15 Pro",
				ProductImage: "https://example.com/iphone.jpg",
				CategoryID:   1,
				CategoryName: "手机",
				Price:        "8999.00",
				ViewCount:    3,
				LastViewAt:   time.Now(),
				ViewDuration: 120,
			},
			{
				ProductID:    102,
				ProductName:  "MacBook Pro",
				ProductImage: "https://example.com/macbook.jpg",
				CategoryID:   2,
				CategoryName: "电脑",
				Price:        "12999.00",
				ViewCount:    1,
				LastViewAt:   time.Now().Add(-1 * time.Hour),
				ViewDuration: 300,
			},
		},
		TotalView: 4,
		CachedAt:  time.Now(),
		UpdatedAt: time.Now(),
	}
}

// 创建测试用的收藏数据
func createTestFavorite(userID uint) *UserFavoriteCacheData {
	return &UserFavoriteCacheData{
		UserID: userID,
		Items: []FavoriteItem{
			{
				ID:           1,
				ProductID:    201,
				ProductName:  "AirPods Pro",
				ProductImage: "https://example.com/airpods.jpg",
				CategoryID:   3,
				CategoryName: "耳机",
				Price:        "1999.00",
				Status:       model.ProductStatusActive,
				FavoriteAt:   time.Now(),
				Tags:         []string{"音频", "苹果"},
			},
		},
		TotalCount: 1,
		CategoryMap: map[uint]int{
			3: 1,
		},
		CachedAt:  time.Now(),
		UpdatedAt: time.Now(),
	}
}

// 创建测试用的推荐数据
func createTestRecommendation(userID uint) *UserRecommendationCacheData {
	return &UserRecommendationCacheData{
		UserID: userID,
		PersonalBased: []RecommendationItem{
			{
				ProductID:     301,
				ProductName:   "iPad Pro",
				ProductImage:  "https://example.com/ipad.jpg",
				CategoryID:    4,
				CategoryName:  "平板",
				Price:         "6999.00",
				Score:         0.95,
				Reason:        "基于浏览历史推荐",
				RecommendType: "personal",
				Weight:        1.0,
				CreatedAt:     time.Now(),
			},
		},
		BehaviorBased: []RecommendationItem{
			{
				ProductID:     302,
				ProductName:   "Apple Watch",
				ProductImage:  "https://example.com/watch.jpg",
				CategoryID:    5,
				CategoryName:  "手表",
				Price:         "2999.00",
				Score:         0.88,
				Reason:        "基于行为分析推荐",
				RecommendType: "behavior",
				Weight:        0.8,
				CreatedAt:     time.Now(),
			},
		},
		CategoryBased:      []RecommendationItem{},
		CollaborativeBased: []RecommendationItem{},
		HotProducts:        []RecommendationItem{},
		NewProducts:        []RecommendationItem{},
		AlgorithmVersion:   "v1.0",
		GeneratedAt:        time.Now(),
		CachedAt:           time.Now(),
		UpdatedAt:          time.Now(),
	}
}

// 创建测试用的行为数据
func createTestBehavior(userID uint) *UserBehaviorCacheData {
	return &UserBehaviorCacheData{
		UserID: userID,
		SearchHistory: []SearchHistoryItem{
			{
				Keyword:     "iPhone",
				SearchCount: 5,
				ResultCount: 120,
				LastSearch:  time.Now(),
				ClickRate:   0.15,
			},
			{
				Keyword:     "MacBook",
				SearchCount: 2,
				ResultCount: 45,
				LastSearch:  time.Now().Add(-2 * time.Hour),
				ClickRate:   0.22,
			},
		},
		ClickBehavior: []ClickBehaviorItem{
			{
				ProductID:    101,
				CategoryID:   1,
				ClickCount:   8,
				LastClick:    time.Now(),
				ClickSource:  "search",
				ClickContext: "keyword:iPhone",
			},
		},
		PurchasePattern: PurchasePatternData{
			PreferredCategories: []uint{1, 2, 3},
			PreferredBrands:     []string{"Apple", "Samsung"},
			PriceRange: PriceRange{
				Min:     "1000.00",
				Max:     "15000.00",
				Average: "8000.00",
			},
			PurchaseFrequency: "monthly",
			SeasonalPattern:   map[string]float64{"spring": 0.3, "summer": 0.2, "autumn": 0.3, "winter": 0.2},
			TimePattern:       map[string]float64{"morning": 0.2, "afternoon": 0.4, "evening": 0.4},
		},
		PreferenceProfile: PreferenceProfileData{
			CategoryWeights: map[uint]float64{1: 0.4, 2: 0.3, 3: 0.3},
			BrandWeights:    map[string]float64{"Apple": 0.6, "Samsung": 0.4},
			PriceWeight:     0.7,
			QualityWeight:   0.8,
			TrendWeight:     0.5,
			Tags:            []string{"科技爱好者", "苹果粉丝"},
		},
		ActivityStats: ActivityStatsData{
			DailyActiveHours:     []int{9, 10, 11, 14, 15, 16, 20, 21},
			WeeklyActivePattern:  []float64{0.1, 0.15, 0.15, 0.15, 0.15, 0.2, 0.1},
			MonthlyActivePattern: []float64{0.08, 0.08, 0.08, 0.08, 0.08, 0.08, 0.08, 0.08, 0.08, 0.08, 0.1, 0.18},
			AvgSessionDuration:   1800, // 30分钟
			PageViewsPerSession:  12.5,
			BounceRate:           0.25,
		},
		CachedAt:  time.Now(),
		UpdatedAt: time.Now(),
	}
}

func setupUserPreferenceCacheService() (*UserPreferenceCacheService, *MockCacheManager) {
	mockCache := new(MockCacheManager)
	keyManager := NewCacheKeyManager("test")
	service := NewUserPreferenceCacheService(mockCache, keyManager)
	return service, mockCache
}

func TestUserPreferenceCacheService_BrowseHistory(t *testing.T) {
	service, mockCache := setupUserPreferenceCacheService()

	// 测试获取浏览历史缓存命中
	t.Run("获取用户浏览历史缓存命中", func(t *testing.T) {
		userID := uint(1001)

		// 模拟缓存数据
		cacheData := `{"user_id":1001,"items":[{"product_id":101,"product_name":"iPhone 15 Pro","product_image":"https://example.com/iphone.jpg","category_id":1,"category_name":"手机","price":"8999.00","view_count":3,"last_view_at":"2025-01-10T10:00:00Z","view_duration":120}],"total_view":3,"cached_at":"2025-01-10T10:00:00Z","updated_at":"2025-01-10T10:00:00Z"}`

		mockCache.On("Get", mock.AnythingOfType("string")).Return(cacheData, nil)

		result, err := service.GetUserBrowseHistory(userID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, userID, result.UserID)
		assert.Equal(t, 3, result.TotalView)
		assert.Len(t, result.Items, 1)
		assert.Equal(t, uint(101), result.Items[0].ProductID)
		assert.Equal(t, "iPhone 15 Pro", result.Items[0].ProductName)

		mockCache.AssertExpectations(t)
	})

	// 测试设置浏览历史缓存
	t.Run("设置用户浏览历史缓存", func(t *testing.T) {
		// 创建新的mock实例避免冲突
		newMockCache := new(MockCacheManager)
		newKeyManager := NewCacheKeyManager("test")
		newService := NewUserPreferenceCacheService(newMockCache, newKeyManager)

		userID := uint(1001)
		browseHistory := createTestBrowseHistory(userID)

		newMockCache.On("Set", mock.AnythingOfType("string"), mock.AnythingOfType("string"), 7*24*time.Hour).Return(nil)

		err := newService.SetUserBrowseHistory(userID, browseHistory)

		assert.NoError(t, err)
		newMockCache.AssertExpectations(t)
	})

	// 测试添加浏览历史
	t.Run("添加浏览历史", func(t *testing.T) {
		// 创建新的mock实例避免冲突
		newMockCache := new(MockCacheManager)
		newKeyManager := NewCacheKeyManager("test")
		newService := NewUserPreferenceCacheService(newMockCache, newKeyManager)

		userID := uint(1001)
		product := createTestProductForUserPref(103, "iPad Air", 4999.00)

		// 模拟获取现有浏览历史（返回nil表示没有缓存）
		newMockCache.On("Get", mock.AnythingOfType("string")).Return(nil, nil)
		// 模拟设置新的浏览历史
		newMockCache.On("Set", mock.AnythingOfType("string"), mock.AnythingOfType("string"), 7*24*time.Hour).Return(nil)

		err := newService.AddBrowseHistory(userID, product, 180)

		assert.NoError(t, err)
		newMockCache.AssertExpectations(t)
	})
}

func TestUserPreferenceCacheService_Favorite(t *testing.T) {
	service, mockCache := setupUserPreferenceCacheService()

	// 测试获取收藏缓存命中
	t.Run("获取用户收藏缓存命中", func(t *testing.T) {
		userID := uint(1001)

		// 模拟缓存数据
		cacheData := `{"user_id":1001,"items":[{"id":1,"product_id":201,"product_name":"AirPods Pro","product_image":"https://example.com/airpods.jpg","category_id":3,"category_name":"耳机","price":"1999.00","status":"active","favorite_at":"2025-01-10T10:00:00Z","tags":["音频","苹果"]}],"total_count":1,"category_map":{"3":1},"cached_at":"2025-01-10T10:00:00Z","updated_at":"2025-01-10T10:00:00Z"}`

		mockCache.On("Get", mock.AnythingOfType("string")).Return(cacheData, nil)

		result, err := service.GetUserFavorite(userID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, userID, result.UserID)
		assert.Equal(t, 1, result.TotalCount)
		assert.Len(t, result.Items, 1)
		assert.Equal(t, uint(201), result.Items[0].ProductID)
		assert.Equal(t, "AirPods Pro", result.Items[0].ProductName)
		assert.Contains(t, result.Items[0].Tags, "音频")

		mockCache.AssertExpectations(t)
	})

	// 测试设置收藏缓存
	t.Run("设置用户收藏缓存", func(t *testing.T) {
		// 创建新的mock实例避免冲突
		newMockCache := new(MockCacheManager)
		newKeyManager := NewCacheKeyManager("test")
		newService := NewUserPreferenceCacheService(newMockCache, newKeyManager)

		userID := uint(1001)
		favorite := createTestFavorite(userID)

		newMockCache.On("Set", mock.AnythingOfType("string"), mock.AnythingOfType("string"), 30*24*time.Hour).Return(nil)

		err := newService.SetUserFavorite(userID, favorite)

		assert.NoError(t, err)
		newMockCache.AssertExpectations(t)
	})

	// 测试添加收藏商品
	t.Run("添加收藏商品", func(t *testing.T) {
		// 创建新的mock实例避免冲突
		newMockCache := new(MockCacheManager)
		newKeyManager := NewCacheKeyManager("test")
		newService := NewUserPreferenceCacheService(newMockCache, newKeyManager)

		userID := uint(1001)
		product := createTestProductForUserPref(202, "Magic Mouse", 799.00)
		tags := []string{"配件", "鼠标"}

		// 模拟获取现有收藏（返回nil表示没有缓存）
		newMockCache.On("Get", mock.AnythingOfType("string")).Return(nil, nil)
		// 模拟设置新的收藏
		newMockCache.On("Set", mock.AnythingOfType("string"), mock.AnythingOfType("string"), 30*24*time.Hour).Return(nil)

		err := newService.AddFavoriteItem(userID, product, tags)

		assert.NoError(t, err)
		newMockCache.AssertExpectations(t)
	})
}

func TestUserPreferenceCacheService_Recommendation(t *testing.T) {
	service, mockCache := setupUserPreferenceCacheService()

	// 测试获取推荐缓存命中
	t.Run("获取用户推荐缓存命中", func(t *testing.T) {
		userID := uint(1001)

		// 模拟缓存数据
		cacheData := `{"user_id":1001,"personal_based":[{"product_id":301,"product_name":"iPad Pro","product_image":"https://example.com/ipad.jpg","category_id":4,"category_name":"平板","price":"6999.00","score":0.95,"reason":"基于浏览历史推荐","recommend_type":"personal","weight":1.0,"created_at":"2025-01-10T10:00:00Z"}],"behavior_based":[],"category_based":[],"collaborative_based":[],"hot_products":[],"new_products":[],"algorithm_version":"v1.0","generated_at":"2025-01-10T10:00:00Z","cached_at":"2025-01-10T10:00:00Z","updated_at":"2025-01-10T10:00:00Z"}`

		mockCache.On("Get", mock.AnythingOfType("string")).Return(cacheData, nil)

		result, err := service.GetUserRecommendation(userID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, userID, result.UserID)
		assert.Equal(t, "v1.0", result.AlgorithmVersion)
		assert.Len(t, result.PersonalBased, 1)
		assert.Equal(t, uint(301), result.PersonalBased[0].ProductID)
		assert.Equal(t, "iPad Pro", result.PersonalBased[0].ProductName)
		assert.Equal(t, 0.95, result.PersonalBased[0].Score)

		mockCache.AssertExpectations(t)
	})

	// 测试设置推荐缓存
	t.Run("设置用户推荐缓存", func(t *testing.T) {
		// 创建新的mock实例避免冲突
		newMockCache := new(MockCacheManager)
		newKeyManager := NewCacheKeyManager("test")
		newService := NewUserPreferenceCacheService(newMockCache, newKeyManager)

		userID := uint(1001)
		recommendation := createTestRecommendation(userID)

		newMockCache.On("Set", mock.AnythingOfType("string"), mock.AnythingOfType("string"), 4*time.Hour).Return(nil)

		err := newService.SetUserRecommendation(userID, recommendation)

		assert.NoError(t, err)
		newMockCache.AssertExpectations(t)
	})
}

func TestUserPreferenceCacheService_Behavior(t *testing.T) {
	service, mockCache := setupUserPreferenceCacheService()

	// 测试获取行为缓存命中
	t.Run("获取用户行为缓存命中", func(t *testing.T) {
		userID := uint(1001)

		// 模拟缓存数据
		cacheData := `{"user_id":1001,"search_history":[{"keyword":"iPhone","search_count":5,"result_count":120,"last_search":"2025-01-10T10:00:00Z","click_rate":0.15}],"click_behavior":[{"product_id":101,"category_id":1,"click_count":8,"last_click":"2025-01-10T10:00:00Z","click_source":"search","click_context":"keyword:iPhone"}],"purchase_pattern":{"preferred_categories":[1,2,3],"preferred_brands":["Apple","Samsung"],"price_range":{"min":"1000.00","max":"15000.00","average":"8000.00"},"purchase_frequency":"monthly","seasonal_pattern":{"spring":0.3,"summer":0.2,"autumn":0.3,"winter":0.2},"time_pattern":{"morning":0.2,"afternoon":0.4,"evening":0.4}},"preference_profile":{"category_weights":{"1":0.4,"2":0.3,"3":0.3},"brand_weights":{"Apple":0.6,"Samsung":0.4},"price_weight":0.7,"quality_weight":0.8,"trend_weight":0.5,"tags":["科技爱好者","苹果粉丝"]},"activity_stats":{"daily_active_hours":[9,10,11,14,15,16,20,21],"weekly_active_pattern":[0.1,0.15,0.15,0.15,0.15,0.2,0.1],"monthly_active_pattern":[0.08,0.08,0.08,0.08,0.08,0.08,0.08,0.08,0.08,0.08,0.1,0.18],"avg_session_duration":1800,"page_views_per_session":12.5,"bounce_rate":0.25},"cached_at":"2025-01-10T10:00:00Z","updated_at":"2025-01-10T10:00:00Z"}`

		mockCache.On("Get", mock.AnythingOfType("string")).Return(cacheData, nil)

		result, err := service.GetUserBehavior(userID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, userID, result.UserID)
		assert.Len(t, result.SearchHistory, 1)
		assert.Equal(t, "iPhone", result.SearchHistory[0].Keyword)
		assert.Equal(t, 5, result.SearchHistory[0].SearchCount)
		assert.Len(t, result.ClickBehavior, 1)
		assert.Equal(t, uint(101), result.ClickBehavior[0].ProductID)
		assert.Equal(t, 8, result.ClickBehavior[0].ClickCount)

		mockCache.AssertExpectations(t)
	})

	// 测试设置行为缓存
	t.Run("设置用户行为缓存", func(t *testing.T) {
		// 创建新的mock实例避免冲突
		newMockCache := new(MockCacheManager)
		newKeyManager := NewCacheKeyManager("test")
		newService := NewUserPreferenceCacheService(newMockCache, newKeyManager)

		userID := uint(1001)
		behavior := createTestBehavior(userID)

		newMockCache.On("Set", mock.AnythingOfType("string"), mock.AnythingOfType("string"), 30*24*time.Hour).Return(nil)

		err := newService.SetUserBehavior(userID, behavior)

		assert.NoError(t, err)
		newMockCache.AssertExpectations(t)
	})
}

func TestUserPreferenceCacheService_ExistsOperations(t *testing.T) {
	service, mockCache := setupUserPreferenceCacheService()

	userID := uint(1001)

	// 测试检查浏览历史存在
	t.Run("检查用户浏览历史存在", func(t *testing.T) {
		mockCache.On("Exists", mock.AnythingOfType("string")).Return(true)

		exists := service.ExistsUserBrowseHistory(userID)

		assert.True(t, exists)
		mockCache.AssertExpectations(t)
	})

	// 测试检查收藏存在
	t.Run("检查用户收藏存在", func(t *testing.T) {
		// 创建新的mock实例避免冲突
		newMockCache := new(MockCacheManager)
		newKeyManager := NewCacheKeyManager("test")
		newService := NewUserPreferenceCacheService(newMockCache, newKeyManager)

		newMockCache.On("Exists", mock.AnythingOfType("string")).Return(true)

		exists := newService.ExistsUserFavorite(userID)

		assert.True(t, exists)
		newMockCache.AssertExpectations(t)
	})
}

func TestUserPreferenceCacheService_TTLOperations(t *testing.T) {
	service, mockCache := setupUserPreferenceCacheService()

	userID := uint(1001)

	// 测试获取浏览历史TTL
	t.Run("获取用户浏览历史TTL", func(t *testing.T) {
		expectedTTL := 7 * 24 * time.Hour

		mockCache.On("TTL", mock.AnythingOfType("string")).Return(expectedTTL, nil)

		ttl, err := service.GetUserBrowseHistoryTTL(userID)

		assert.NoError(t, err)
		assert.Equal(t, expectedTTL, ttl)
		mockCache.AssertExpectations(t)
	})

	// 测试刷新推荐TTL
	t.Run("刷新用户推荐TTL", func(t *testing.T) {
		// 创建新的mock实例避免冲突
		newMockCache := new(MockCacheManager)
		newKeyManager := NewCacheKeyManager("test")
		newService := NewUserPreferenceCacheService(newMockCache, newKeyManager)

		newMockCache.On("Expire", mock.AnythingOfType("string"), 4*time.Hour).Return(nil)

		err := newService.RefreshUserRecommendationTTL(userID)

		assert.NoError(t, err)
		newMockCache.AssertExpectations(t)
	})
}

func TestUserPreferenceCacheService_DeleteOperations(t *testing.T) {
	service, mockCache := setupUserPreferenceCacheService()

	// 测试删除用户偏好缓存
	t.Run("删除用户偏好缓存", func(t *testing.T) {
		userID := uint(1001)

		// 模拟删除4个缓存键
		mockCache.On("Delete", mock.AnythingOfType("string")).Return(nil).Times(4)

		err := service.DeleteUserPreference(userID)

		assert.NoError(t, err)
		mockCache.AssertExpectations(t)
	})
}

func TestUserPreferenceCacheService_GetUserPreferenceStats(t *testing.T) {
	service, mockCache := setupUserPreferenceCacheService()

	// 测试获取用户偏好统计信息
	t.Run("获取用户偏好统计信息", func(t *testing.T) {
		userID := uint(1001)

		// 模拟各种缓存数据的获取
		browseHistoryData := `{"user_id":1001,"items":[{"product_id":101,"product_name":"iPhone 15 Pro","product_image":"https://example.com/iphone.jpg","category_id":1,"category_name":"手机","price":"8999.00","view_count":3,"last_view_at":"2025-01-10T10:00:00Z","view_duration":120}],"total_view":5,"cached_at":"2025-01-10T10:00:00Z","updated_at":"2025-01-10T10:00:00Z"}`
		favoriteData := `{"user_id":1001,"items":[{"id":1,"product_id":201,"product_name":"AirPods Pro","product_image":"https://example.com/airpods.jpg","category_id":3,"category_name":"耳机","price":"1999.00","status":"active","favorite_at":"2025-01-10T10:00:00Z","tags":["音频","苹果"]}],"total_count":1,"category_map":{"3":1},"cached_at":"2025-01-10T10:00:00Z","updated_at":"2025-01-10T10:00:00Z"}`
		recommendationData := `{"user_id":1001,"personal_based":[{"product_id":301,"product_name":"iPad Pro","product_image":"https://example.com/ipad.jpg","category_id":4,"category_name":"平板","price":"6999.00","score":0.95,"reason":"基于浏览历史推荐","recommend_type":"personal","weight":1.0,"created_at":"2025-01-10T10:00:00Z"}],"behavior_based":[{"product_id":302,"product_name":"Apple Watch","product_image":"https://example.com/watch.jpg","category_id":5,"category_name":"手表","price":"2999.00","score":0.88,"reason":"基于行为分析推荐","recommend_type":"behavior","weight":0.8,"created_at":"2025-01-10T10:00:00Z"}],"category_based":[],"collaborative_based":[],"hot_products":[],"new_products":[],"algorithm_version":"v1.0","generated_at":"2025-01-10T10:00:00Z","cached_at":"2025-01-10T10:00:00Z","updated_at":"2025-01-10T10:00:00Z"}`
		behaviorData := `{"user_id":1001,"search_history":[{"keyword":"iPhone","search_count":5,"result_count":120,"last_search":"2025-01-10T10:00:00Z","click_rate":0.15},{"keyword":"MacBook","search_count":2,"result_count":45,"last_search":"2025-01-10T08:00:00Z","click_rate":0.22}],"click_behavior":[{"product_id":101,"category_id":1,"click_count":8,"last_click":"2025-01-10T10:00:00Z","click_source":"search","click_context":"keyword:iPhone"}],"purchase_pattern":{"preferred_categories":[1,2,3],"preferred_brands":["Apple","Samsung"],"price_range":{"min":"1000.00","max":"15000.00","average":"8000.00"},"purchase_frequency":"monthly","seasonal_pattern":{"spring":0.3,"summer":0.2,"autumn":0.3,"winter":0.2},"time_pattern":{"morning":0.2,"afternoon":0.4,"evening":0.4}},"preference_profile":{"category_weights":{"1":0.4,"2":0.3,"3":0.3},"brand_weights":{"Apple":0.6,"Samsung":0.4},"price_weight":0.7,"quality_weight":0.8,"trend_weight":0.5,"tags":["科技爱好者","苹果粉丝"]},"activity_stats":{"daily_active_hours":[9,10,11,14,15,16,20,21],"weekly_active_pattern":[0.1,0.15,0.15,0.15,0.15,0.2,0.1],"monthly_active_pattern":[0.08,0.08,0.08,0.08,0.08,0.08,0.08,0.08,0.08,0.08,0.1,0.18],"avg_session_duration":1800,"page_views_per_session":12.5,"bounce_rate":0.25},"cached_at":"2025-01-10T10:00:00Z","updated_at":"2025-01-10T10:00:00Z"}`

		// 按顺序模拟4次Get调用
		mockCache.On("Get", mock.AnythingOfType("string")).Return(browseHistoryData, nil).Once()
		mockCache.On("Get", mock.AnythingOfType("string")).Return(favoriteData, nil).Once()
		mockCache.On("Get", mock.AnythingOfType("string")).Return(recommendationData, nil).Once()
		mockCache.On("Get", mock.AnythingOfType("string")).Return(behaviorData, nil).Once()

		stats, err := service.GetUserPreferenceStats(userID)

		assert.NoError(t, err)
		assert.NotNil(t, stats)
		assert.Equal(t, userID, stats.UserID)
		assert.Equal(t, 1, stats.BrowseHistoryCount)
		assert.Equal(t, 5, stats.TotalViewCount)
		assert.Equal(t, 1, stats.FavoriteCount)
		assert.Equal(t, 1, stats.FavoriteCategoryCount)
		assert.Equal(t, 1, stats.PersonalRecommendationCount)
		assert.Equal(t, 1, stats.BehaviorRecommendationCount)
		assert.Equal(t, 2, stats.SearchHistoryCount)
		assert.Equal(t, 1, stats.ClickBehaviorCount)
		assert.True(t, stats.HasBrowseHistory)
		assert.True(t, stats.HasFavorite)
		assert.True(t, stats.HasRecommendation)
		assert.True(t, stats.HasBehavior)
		assert.Equal(t, "v1.0", stats.RecommendationAlgorithmVersion)

		mockCache.AssertExpectations(t)
	})
}
