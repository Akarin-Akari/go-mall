package cache

import (
	"encoding/json"
	"fmt"
	"mall-go/internal/model"
	"mall-go/pkg/logger"
	"time"
)

// UserPreferenceCacheService 用户偏好缓存服务
type UserPreferenceCacheService struct {
	cache      CacheManager
	keyManager *CacheKeyManager
}

// NewUserPreferenceCacheService 创建用户偏好缓存服务
func NewUserPreferenceCacheService(cache CacheManager, keyManager *CacheKeyManager) *UserPreferenceCacheService {
	return &UserPreferenceCacheService{
		cache:      cache,
		keyManager: keyManager,
	}
}

// UserBrowseHistoryCacheData 用户浏览历史缓存数据结构
type UserBrowseHistoryCacheData struct {
	UserID    uint                `json:"user_id"`
	Items     []BrowseHistoryItem `json:"items"`
	TotalView int                 `json:"total_view"`
	CachedAt  time.Time           `json:"cached_at"`
	UpdatedAt time.Time           `json:"updated_at"`
}

// BrowseHistoryItem 浏览历史项
type BrowseHistoryItem struct {
	ProductID    uint      `json:"product_id"`
	ProductName  string    `json:"product_name"`
	ProductImage string    `json:"product_image"`
	CategoryID   uint      `json:"category_id"`
	CategoryName string    `json:"category_name"`
	Price        string    `json:"price"`         // 使用string保持decimal精度
	ViewCount    int       `json:"view_count"`    // 浏览次数
	LastViewAt   time.Time `json:"last_view_at"`  // 最后浏览时间
	ViewDuration int       `json:"view_duration"` // 浏览时长(秒)
}

// UserFavoriteCacheData 用户收藏缓存数据结构
type UserFavoriteCacheData struct {
	UserID      uint           `json:"user_id"`
	Items       []FavoriteItem `json:"items"`
	TotalCount  int            `json:"total_count"`
	CategoryMap map[uint]int   `json:"category_map"` // 分类收藏统计
	CachedAt    time.Time      `json:"cached_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// FavoriteItem 收藏项
type FavoriteItem struct {
	ID           uint      `json:"id"`
	ProductID    uint      `json:"product_id"`
	ProductName  string    `json:"product_name"`
	ProductImage string    `json:"product_image"`
	CategoryID   uint      `json:"category_id"`
	CategoryName string    `json:"category_name"`
	Price        string    `json:"price"`       // 使用string保持decimal精度
	Status       string    `json:"status"`      // 商品状态
	FavoriteAt   time.Time `json:"favorite_at"` // 收藏时间
	Tags         []string  `json:"tags"`        // 收藏标签
}

// UserRecommendationCacheData 用户推荐缓存数据结构
type UserRecommendationCacheData struct {
	UserID             uint                 `json:"user_id"`
	PersonalBased      []RecommendationItem `json:"personal_based"`      // 个人偏好推荐
	BehaviorBased      []RecommendationItem `json:"behavior_based"`      // 行为推荐
	CategoryBased      []RecommendationItem `json:"category_based"`      // 分类推荐
	CollaborativeBased []RecommendationItem `json:"collaborative_based"` // 协同过滤推荐
	HotProducts        []RecommendationItem `json:"hot_products"`        // 热门商品推荐
	NewProducts        []RecommendationItem `json:"new_products"`        // 新品推荐
	AlgorithmVersion   string               `json:"algorithm_version"`   // 算法版本
	GeneratedAt        time.Time            `json:"generated_at"`        // 生成时间
	CachedAt           time.Time            `json:"cached_at"`
	UpdatedAt          time.Time            `json:"updated_at"`
}

// RecommendationItem 推荐项
type RecommendationItem struct {
	ProductID     uint      `json:"product_id"`
	ProductName   string    `json:"product_name"`
	ProductImage  string    `json:"product_image"`
	CategoryID    uint      `json:"category_id"`
	CategoryName  string    `json:"category_name"`
	Price         string    `json:"price"`          // 使用string保持decimal精度
	Score         float64   `json:"score"`          // 推荐分数
	Reason        string    `json:"reason"`         // 推荐理由
	RecommendType string    `json:"recommend_type"` // 推荐类型
	Weight        float64   `json:"weight"`         // 权重
	CreatedAt     time.Time `json:"created_at"`
}

// UserBehaviorCacheData 用户行为缓存数据结构
type UserBehaviorCacheData struct {
	UserID            uint                  `json:"user_id"`
	SearchHistory     []SearchHistoryItem   `json:"search_history"`     // 搜索历史
	ClickBehavior     []ClickBehaviorItem   `json:"click_behavior"`     // 点击行为
	PurchasePattern   PurchasePatternData   `json:"purchase_pattern"`   // 购买模式
	PreferenceProfile PreferenceProfileData `json:"preference_profile"` // 偏好画像
	ActivityStats     ActivityStatsData     `json:"activity_stats"`     // 活动统计
	CachedAt          time.Time             `json:"cached_at"`
	UpdatedAt         time.Time             `json:"updated_at"`
}

// SearchHistoryItem 搜索历史项
type SearchHistoryItem struct {
	Keyword     string    `json:"keyword"`
	SearchCount int       `json:"search_count"`
	ResultCount int64     `json:"result_count"`
	LastSearch  time.Time `json:"last_search"`
	ClickRate   float64   `json:"click_rate"` // 点击率
}

// ClickBehaviorItem 点击行为项
type ClickBehaviorItem struct {
	ProductID    uint      `json:"product_id"`
	CategoryID   uint      `json:"category_id"`
	ClickCount   int       `json:"click_count"`
	LastClick    time.Time `json:"last_click"`
	ClickSource  string    `json:"click_source"`  // 点击来源
	ClickContext string    `json:"click_context"` // 点击上下文
}

// PurchasePatternData 购买模式数据
type PurchasePatternData struct {
	PreferredCategories []uint             `json:"preferred_categories"` // 偏好分类
	PreferredBrands     []string           `json:"preferred_brands"`     // 偏好品牌
	PriceRange          PriceRange         `json:"price_range"`          // 价格区间
	PurchaseFrequency   string             `json:"purchase_frequency"`   // 购买频率
	SeasonalPattern     map[string]float64 `json:"seasonal_pattern"`     // 季节性模式
	TimePattern         map[string]float64 `json:"time_pattern"`         // 时间模式
}

// PriceRange 价格区间
type PriceRange struct {
	Min     string `json:"min"`     // 最小价格
	Max     string `json:"max"`     // 最大价格
	Average string `json:"average"` // 平均价格
}

// PreferenceProfileData 偏好画像数据
type PreferenceProfileData struct {
	CategoryWeights map[uint]float64   `json:"category_weights"` // 分类权重
	BrandWeights    map[string]float64 `json:"brand_weights"`    // 品牌权重
	PriceWeight     float64            `json:"price_weight"`     // 价格敏感度
	QualityWeight   float64            `json:"quality_weight"`   // 质量偏好
	TrendWeight     float64            `json:"trend_weight"`     // 潮流偏好
	Tags            []string           `json:"tags"`             // 用户标签
}

// ActivityStatsData 活动统计数据
type ActivityStatsData struct {
	DailyActiveHours     []int     `json:"daily_active_hours"`     // 每日活跃时段
	WeeklyActivePattern  []float64 `json:"weekly_active_pattern"`  // 周活跃模式
	MonthlyActivePattern []float64 `json:"monthly_active_pattern"` // 月活跃模式
	AvgSessionDuration   int       `json:"avg_session_duration"`   // 平均会话时长
	PageViewsPerSession  float64   `json:"page_views_per_session"` // 每会话页面浏览数
	BounceRate           float64   `json:"bounce_rate"`            // 跳出率
}

// GetUserBrowseHistory 获取用户浏览历史缓存
func (upcs *UserPreferenceCacheService) GetUserBrowseHistory(userID uint) (*UserBrowseHistoryCacheData, error) {
	key := upcs.keyManager.GenerateUserBrowseHistoryKey(userID)

	data, err := upcs.cache.Get(key)
	if err != nil {
		logger.Error(fmt.Sprintf("获取用户浏览历史缓存失败: UserID=%d, Error=%v", userID, err))
		return nil, fmt.Errorf("获取用户浏览历史缓存失败: %w", err)
	}

	if data == nil {
		return nil, nil // 缓存未命中
	}

	var browseHistory UserBrowseHistoryCacheData
	if err := json.Unmarshal([]byte(data.(string)), &browseHistory); err != nil {
		logger.Error(fmt.Sprintf("用户浏览历史缓存数据反序列化失败: UserID=%d, Error=%v", userID, err))
		return nil, fmt.Errorf("用户浏览历史缓存数据反序列化失败: %w", err)
	}

	return &browseHistory, nil
}

// SetUserBrowseHistory 设置用户浏览历史缓存
func (upcs *UserPreferenceCacheService) SetUserBrowseHistory(userID uint, browseHistory *UserBrowseHistoryCacheData) error {
	key := upcs.keyManager.GenerateUserBrowseHistoryKey(userID)

	// 设置缓存时间戳
	browseHistory.CachedAt = time.Now()
	browseHistory.UpdatedAt = time.Now()

	data, err := json.Marshal(browseHistory)
	if err != nil {
		logger.Error(fmt.Sprintf("用户浏览历史缓存数据序列化失败: UserID=%d, Error=%v", userID, err))
		return fmt.Errorf("用户浏览历史缓存数据序列化失败: %w", err)
	}

	// 浏览历史TTL: 7天
	ttl := CacheTTL["browse_history"]
	if err := upcs.cache.Set(key, string(data), ttl); err != nil {
		logger.Error(fmt.Sprintf("设置用户浏览历史缓存失败: UserID=%d, Error=%v", userID, err))
		return fmt.Errorf("设置用户浏览历史缓存失败: %w", err)
	}

	logger.Info(fmt.Sprintf("设置用户浏览历史缓存成功: UserID=%d, ItemsCount=%d", userID, len(browseHistory.Items)))
	return nil
}

// GetUserFavorite 获取用户收藏缓存
func (upcs *UserPreferenceCacheService) GetUserFavorite(userID uint) (*UserFavoriteCacheData, error) {
	key := upcs.keyManager.GenerateUserFavoriteKey(userID)

	data, err := upcs.cache.Get(key)
	if err != nil {
		logger.Error(fmt.Sprintf("获取用户收藏缓存失败: UserID=%d, Error=%v", userID, err))
		return nil, fmt.Errorf("获取用户收藏缓存失败: %w", err)
	}

	if data == nil {
		return nil, nil // 缓存未命中
	}

	var favorite UserFavoriteCacheData
	if err := json.Unmarshal([]byte(data.(string)), &favorite); err != nil {
		logger.Error(fmt.Sprintf("用户收藏缓存数据反序列化失败: UserID=%d, Error=%v", userID, err))
		return nil, fmt.Errorf("用户收藏缓存数据反序列化失败: %w", err)
	}

	return &favorite, nil
}

// SetUserFavorite 设置用户收藏缓存
func (upcs *UserPreferenceCacheService) SetUserFavorite(userID uint, favorite *UserFavoriteCacheData) error {
	key := upcs.keyManager.GenerateUserFavoriteKey(userID)

	// 设置缓存时间戳
	favorite.CachedAt = time.Now()
	favorite.UpdatedAt = time.Now()

	data, err := json.Marshal(favorite)
	if err != nil {
		logger.Error(fmt.Sprintf("用户收藏缓存数据序列化失败: UserID=%d, Error=%v", userID, err))
		return fmt.Errorf("用户收藏缓存数据序列化失败: %w", err)
	}

	// 收藏数据TTL: 永久（使用较长时间）
	ttl := CacheTTL["favorite"]
	if err := upcs.cache.Set(key, string(data), ttl); err != nil {
		logger.Error(fmt.Sprintf("设置用户收藏缓存失败: UserID=%d, Error=%v", userID, err))
		return fmt.Errorf("设置用户收藏缓存失败: %w", err)
	}

	logger.Info(fmt.Sprintf("设置用户收藏缓存成功: UserID=%d, ItemsCount=%d", userID, len(favorite.Items)))
	return nil
}

// GetUserRecommendation 获取用户推荐缓存
func (upcs *UserPreferenceCacheService) GetUserRecommendation(userID uint) (*UserRecommendationCacheData, error) {
	key := upcs.keyManager.GenerateUserRecommendationKey(userID)

	data, err := upcs.cache.Get(key)
	if err != nil {
		logger.Error(fmt.Sprintf("获取用户推荐缓存失败: UserID=%d, Error=%v", userID, err))
		return nil, fmt.Errorf("获取用户推荐缓存失败: %w", err)
	}

	if data == nil {
		return nil, nil // 缓存未命中
	}

	var recommendation UserRecommendationCacheData
	if err := json.Unmarshal([]byte(data.(string)), &recommendation); err != nil {
		logger.Error(fmt.Sprintf("用户推荐缓存数据反序列化失败: UserID=%d, Error=%v", userID, err))
		return nil, fmt.Errorf("用户推荐缓存数据反序列化失败: %w", err)
	}

	return &recommendation, nil
}

// SetUserRecommendation 设置用户推荐缓存
func (upcs *UserPreferenceCacheService) SetUserRecommendation(userID uint, recommendation *UserRecommendationCacheData) error {
	key := upcs.keyManager.GenerateUserRecommendationKey(userID)

	// 设置缓存时间戳
	recommendation.CachedAt = time.Now()
	recommendation.UpdatedAt = time.Now()

	data, err := json.Marshal(recommendation)
	if err != nil {
		logger.Error(fmt.Sprintf("用户推荐缓存数据序列化失败: UserID=%d, Error=%v", userID, err))
		return fmt.Errorf("用户推荐缓存数据序列化失败: %w", err)
	}

	// 推荐数据TTL: 4小时
	ttl := CacheTTL["recommendation"]
	if err := upcs.cache.Set(key, string(data), ttl); err != nil {
		logger.Error(fmt.Sprintf("设置用户推荐缓存失败: UserID=%d, Error=%v", userID, err))
		return fmt.Errorf("设置用户推荐缓存失败: %w", err)
	}

	logger.Info(fmt.Sprintf("设置用户推荐缓存成功: UserID=%d, PersonalCount=%d", userID, len(recommendation.PersonalBased)))
	return nil
}

// GetUserBehavior 获取用户行为缓存
func (upcs *UserPreferenceCacheService) GetUserBehavior(userID uint) (*UserBehaviorCacheData, error) {
	key := upcs.keyManager.GenerateUserBehaviorKey(userID)

	data, err := upcs.cache.Get(key)
	if err != nil {
		logger.Error(fmt.Sprintf("获取用户行为缓存失败: UserID=%d, Error=%v", userID, err))
		return nil, fmt.Errorf("获取用户行为缓存失败: %w", err)
	}

	if data == nil {
		return nil, nil // 缓存未命中
	}

	var behavior UserBehaviorCacheData
	if err := json.Unmarshal([]byte(data.(string)), &behavior); err != nil {
		logger.Error(fmt.Sprintf("用户行为缓存数据反序列化失败: UserID=%d, Error=%v", userID, err))
		return nil, fmt.Errorf("用户行为缓存数据反序列化失败: %w", err)
	}

	return &behavior, nil
}

// SetUserBehavior 设置用户行为缓存
func (upcs *UserPreferenceCacheService) SetUserBehavior(userID uint, behavior *UserBehaviorCacheData) error {
	key := upcs.keyManager.GenerateUserBehaviorKey(userID)

	// 设置缓存时间戳
	behavior.CachedAt = time.Now()
	behavior.UpdatedAt = time.Now()

	data, err := json.Marshal(behavior)
	if err != nil {
		logger.Error(fmt.Sprintf("用户行为缓存数据序列化失败: UserID=%d, Error=%v", userID, err))
		return fmt.Errorf("用户行为缓存数据序列化失败: %w", err)
	}

	// 用户行为TTL: 30天
	ttl := CacheTTL["behavior"]
	if err := upcs.cache.Set(key, string(data), ttl); err != nil {
		logger.Error(fmt.Sprintf("设置用户行为缓存失败: UserID=%d, Error=%v", userID, err))
		return fmt.Errorf("设置用户行为缓存失败: %w", err)
	}

	logger.Info(fmt.Sprintf("设置用户行为缓存成功: UserID=%d, SearchCount=%d", userID, len(behavior.SearchHistory)))
	return nil
}

// AddBrowseHistory 添加浏览历史
func (upcs *UserPreferenceCacheService) AddBrowseHistory(userID uint, product *model.Product, viewDuration int) error {
	// 获取现有浏览历史
	browseHistory, err := upcs.GetUserBrowseHistory(userID)
	if err != nil {
		return fmt.Errorf("获取用户浏览历史失败: %w", err)
	}

	// 如果没有缓存，创建新的
	if browseHistory == nil {
		browseHistory = &UserBrowseHistoryCacheData{
			UserID:    userID,
			Items:     []BrowseHistoryItem{},
			TotalView: 0,
		}
	}

	// 查找是否已存在该商品的浏览记录
	found := false
	for i := range browseHistory.Items {
		if browseHistory.Items[i].ProductID == product.ID {
			// 更新现有记录
			browseHistory.Items[i].ViewCount++
			browseHistory.Items[i].LastViewAt = time.Now()
			browseHistory.Items[i].ViewDuration = viewDuration
			browseHistory.Items[i].Price = product.Price.String()
			found = true
			break
		}
	}

	// 如果不存在，添加新记录
	if !found {
		newItem := BrowseHistoryItem{
			ProductID:    product.ID,
			ProductName:  product.Name,
			ProductImage: getMainProductImage(product),
			CategoryID:   product.CategoryID,
			CategoryName: getCategoryName(product.CategoryID),
			Price:        product.Price.String(),
			ViewCount:    1,
			LastViewAt:   time.Now(),
			ViewDuration: viewDuration,
		}

		// 添加到列表开头
		browseHistory.Items = append([]BrowseHistoryItem{newItem}, browseHistory.Items...)

		// 限制历史记录数量（最多保留100条）
		if len(browseHistory.Items) > 100 {
			browseHistory.Items = browseHistory.Items[:100]
		}
	}

	browseHistory.TotalView++

	// 保存更新后的浏览历史
	return upcs.SetUserBrowseHistory(userID, browseHistory)
}

// AddFavoriteItem 添加收藏商品
func (upcs *UserPreferenceCacheService) AddFavoriteItem(userID uint, product *model.Product, tags []string) error {
	// 获取现有收藏
	favorite, err := upcs.GetUserFavorite(userID)
	if err != nil {
		return fmt.Errorf("获取用户收藏失败: %w", err)
	}

	// 如果没有缓存，创建新的
	if favorite == nil {
		favorite = &UserFavoriteCacheData{
			UserID:      userID,
			Items:       []FavoriteItem{},
			TotalCount:  0,
			CategoryMap: make(map[uint]int),
		}
	}

	// 检查是否已收藏
	for _, item := range favorite.Items {
		if item.ProductID == product.ID {
			return fmt.Errorf("商品已在收藏列表中")
		}
	}

	// 添加新收藏
	newItem := FavoriteItem{
		ID:           uint(time.Now().UnixNano()), // 临时ID
		ProductID:    product.ID,
		ProductName:  product.Name,
		ProductImage: getMainProductImage(product),
		CategoryID:   product.CategoryID,
		CategoryName: getCategoryName(product.CategoryID),
		Price:        product.Price.String(),
		Status:       product.Status,
		FavoriteAt:   time.Now(),
		Tags:         tags,
	}

	favorite.Items = append([]FavoriteItem{newItem}, favorite.Items...)
	favorite.TotalCount++

	// 更新分类统计
	favorite.CategoryMap[product.CategoryID]++

	// 保存更新后的收藏
	return upcs.SetUserFavorite(userID, favorite)
}

// RemoveFavoriteItem 移除收藏商品
func (upcs *UserPreferenceCacheService) RemoveFavoriteItem(userID uint, productID uint) error {
	// 获取现有收藏
	favorite, err := upcs.GetUserFavorite(userID)
	if err != nil {
		return fmt.Errorf("获取用户收藏失败: %w", err)
	}

	if favorite == nil {
		return fmt.Errorf("用户收藏列表为空")
	}

	// 查找并移除商品
	found := false
	var categoryID uint
	for i, item := range favorite.Items {
		if item.ProductID == productID {
			categoryID = item.CategoryID
			favorite.Items = append(favorite.Items[:i], favorite.Items[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("商品不在收藏列表中")
	}

	favorite.TotalCount--

	// 更新分类统计
	if favorite.CategoryMap[categoryID] > 0 {
		favorite.CategoryMap[categoryID]--
		if favorite.CategoryMap[categoryID] == 0 {
			delete(favorite.CategoryMap, categoryID)
		}
	}

	// 保存更新后的收藏
	return upcs.SetUserFavorite(userID, favorite)
}

// UpdateRecommendationScore 更新推荐分数
func (upcs *UserPreferenceCacheService) UpdateRecommendationScore(userID uint, productID uint, scoreChange float64, reason string) error {
	// 获取现有推荐
	recommendation, err := upcs.GetUserRecommendation(userID)
	if err != nil {
		return fmt.Errorf("获取用户推荐失败: %w", err)
	}

	if recommendation == nil {
		return nil // 没有推荐缓存，忽略更新
	}

	// 更新各类推荐中的分数
	updateRecommendationItems := func(items []RecommendationItem) []RecommendationItem {
		for i := range items {
			if items[i].ProductID == productID {
				items[i].Score += scoreChange
				if items[i].Score < 0 {
					items[i].Score = 0
				}
				items[i].Reason = fmt.Sprintf("%s; %s", items[i].Reason, reason)
			}
		}
		return items
	}

	recommendation.PersonalBased = updateRecommendationItems(recommendation.PersonalBased)
	recommendation.BehaviorBased = updateRecommendationItems(recommendation.BehaviorBased)
	recommendation.CategoryBased = updateRecommendationItems(recommendation.CategoryBased)
	recommendation.CollaborativeBased = updateRecommendationItems(recommendation.CollaborativeBased)
	recommendation.HotProducts = updateRecommendationItems(recommendation.HotProducts)
	recommendation.NewProducts = updateRecommendationItems(recommendation.NewProducts)

	// 保存更新后的推荐
	return upcs.SetUserRecommendation(userID, recommendation)
}

// AddSearchHistory 添加搜索历史
func (upcs *UserPreferenceCacheService) AddSearchHistory(userID uint, keyword string, resultCount int64, clickRate float64) error {
	// 获取现有行为数据
	behavior, err := upcs.GetUserBehavior(userID)
	if err != nil {
		return fmt.Errorf("获取用户行为失败: %w", err)
	}

	// 如果没有缓存，创建新的
	if behavior == nil {
		behavior = &UserBehaviorCacheData{
			UserID:        userID,
			SearchHistory: []SearchHistoryItem{},
			ClickBehavior: []ClickBehaviorItem{},
		}
	}

	// 查找是否已存在该关键词
	found := false
	for i := range behavior.SearchHistory {
		if behavior.SearchHistory[i].Keyword == keyword {
			// 更新现有记录
			behavior.SearchHistory[i].SearchCount++
			behavior.SearchHistory[i].ResultCount = resultCount
			behavior.SearchHistory[i].LastSearch = time.Now()
			behavior.SearchHistory[i].ClickRate = clickRate
			found = true
			break
		}
	}

	// 如果不存在，添加新记录
	if !found {
		newItem := SearchHistoryItem{
			Keyword:     keyword,
			SearchCount: 1,
			ResultCount: resultCount,
			LastSearch:  time.Now(),
			ClickRate:   clickRate,
		}

		// 添加到列表开头
		behavior.SearchHistory = append([]SearchHistoryItem{newItem}, behavior.SearchHistory...)

		// 限制搜索历史数量（最多保留50条）
		if len(behavior.SearchHistory) > 50 {
			behavior.SearchHistory = behavior.SearchHistory[:50]
		}
	}

	// 保存更新后的行为数据
	return upcs.SetUserBehavior(userID, behavior)
}

// AddClickBehavior 添加点击行为
func (upcs *UserPreferenceCacheService) AddClickBehavior(userID uint, productID uint, categoryID uint, clickSource string, clickContext string) error {
	// 获取现有行为数据
	behavior, err := upcs.GetUserBehavior(userID)
	if err != nil {
		return fmt.Errorf("获取用户行为失败: %w", err)
	}

	// 如果没有缓存，创建新的
	if behavior == nil {
		behavior = &UserBehaviorCacheData{
			UserID:        userID,
			SearchHistory: []SearchHistoryItem{},
			ClickBehavior: []ClickBehaviorItem{},
		}
	}

	// 查找是否已存在该商品的点击记录
	found := false
	for i := range behavior.ClickBehavior {
		if behavior.ClickBehavior[i].ProductID == productID {
			// 更新现有记录
			behavior.ClickBehavior[i].ClickCount++
			behavior.ClickBehavior[i].LastClick = time.Now()
			behavior.ClickBehavior[i].ClickSource = clickSource
			behavior.ClickBehavior[i].ClickContext = clickContext
			found = true
			break
		}
	}

	// 如果不存在，添加新记录
	if !found {
		newItem := ClickBehaviorItem{
			ProductID:    productID,
			CategoryID:   categoryID,
			ClickCount:   1,
			LastClick:    time.Now(),
			ClickSource:  clickSource,
			ClickContext: clickContext,
		}

		// 添加到列表开头
		behavior.ClickBehavior = append([]ClickBehaviorItem{newItem}, behavior.ClickBehavior...)

		// 限制点击行为数量（最多保留200条）
		if len(behavior.ClickBehavior) > 200 {
			behavior.ClickBehavior = behavior.ClickBehavior[:200]
		}
	}

	// 保存更新后的行为数据
	return upcs.SetUserBehavior(userID, behavior)
}

// DeleteUserPreference 删除用户偏好缓存
func (upcs *UserPreferenceCacheService) DeleteUserPreference(userID uint) error {
	keys := []string{
		upcs.keyManager.GenerateUserBrowseHistoryKey(userID),
		upcs.keyManager.GenerateUserFavoriteKey(userID),
		upcs.keyManager.GenerateUserRecommendationKey(userID),
		upcs.keyManager.GenerateUserBehaviorKey(userID),
	}

	for _, key := range keys {
		if err := upcs.cache.Delete(key); err != nil {
			logger.Error(fmt.Sprintf("删除用户偏好缓存失败: UserID=%d, Key=%s, Error=%v", userID, key, err))
			return fmt.Errorf("删除用户偏好缓存失败: %w", err)
		}
	}

	logger.Info(fmt.Sprintf("删除用户偏好缓存成功: UserID=%d", userID))
	return nil
}

// ExistsUserBrowseHistory 检查用户浏览历史缓存是否存在
func (upcs *UserPreferenceCacheService) ExistsUserBrowseHistory(userID uint) bool {
	key := upcs.keyManager.GenerateUserBrowseHistoryKey(userID)
	return upcs.cache.Exists(key)
}

// ExistsUserFavorite 检查用户收藏缓存是否存在
func (upcs *UserPreferenceCacheService) ExistsUserFavorite(userID uint) bool {
	key := upcs.keyManager.GenerateUserFavoriteKey(userID)
	return upcs.cache.Exists(key)
}

// ExistsUserRecommendation 检查用户推荐缓存是否存在
func (upcs *UserPreferenceCacheService) ExistsUserRecommendation(userID uint) bool {
	key := upcs.keyManager.GenerateUserRecommendationKey(userID)
	return upcs.cache.Exists(key)
}

// ExistsUserBehavior 检查用户行为缓存是否存在
func (upcs *UserPreferenceCacheService) ExistsUserBehavior(userID uint) bool {
	key := upcs.keyManager.GenerateUserBehaviorKey(userID)
	return upcs.cache.Exists(key)
}

// GetUserBrowseHistoryTTL 获取用户浏览历史缓存TTL
func (upcs *UserPreferenceCacheService) GetUserBrowseHistoryTTL(userID uint) (time.Duration, error) {
	key := upcs.keyManager.GenerateUserBrowseHistoryKey(userID)
	return upcs.cache.TTL(key)
}

// RefreshUserBrowseHistoryTTL 刷新用户浏览历史缓存TTL
func (upcs *UserPreferenceCacheService) RefreshUserBrowseHistoryTTL(userID uint) error {
	key := upcs.keyManager.GenerateUserBrowseHistoryKey(userID)
	ttl := CacheTTL["browse_history"]
	return upcs.cache.Expire(key, ttl)
}

// GetUserFavoriteTTL 获取用户收藏缓存TTL
func (upcs *UserPreferenceCacheService) GetUserFavoriteTTL(userID uint) (time.Duration, error) {
	key := upcs.keyManager.GenerateUserFavoriteKey(userID)
	return upcs.cache.TTL(key)
}

// RefreshUserFavoriteTTL 刷新用户收藏缓存TTL
func (upcs *UserPreferenceCacheService) RefreshUserFavoriteTTL(userID uint) error {
	key := upcs.keyManager.GenerateUserFavoriteKey(userID)
	ttl := CacheTTL["favorite"]
	return upcs.cache.Expire(key, ttl)
}

// GetUserRecommendationTTL 获取用户推荐缓存TTL
func (upcs *UserPreferenceCacheService) GetUserRecommendationTTL(userID uint) (time.Duration, error) {
	key := upcs.keyManager.GenerateUserRecommendationKey(userID)
	return upcs.cache.TTL(key)
}

// RefreshUserRecommendationTTL 刷新用户推荐缓存TTL
func (upcs *UserPreferenceCacheService) RefreshUserRecommendationTTL(userID uint) error {
	key := upcs.keyManager.GenerateUserRecommendationKey(userID)
	ttl := CacheTTL["recommendation"]
	return upcs.cache.Expire(key, ttl)
}

// GetUserBehaviorTTL 获取用户行为缓存TTL
func (upcs *UserPreferenceCacheService) GetUserBehaviorTTL(userID uint) (time.Duration, error) {
	key := upcs.keyManager.GenerateUserBehaviorKey(userID)
	return upcs.cache.TTL(key)
}

// RefreshUserBehaviorTTL 刷新用户行为缓存TTL
func (upcs *UserPreferenceCacheService) RefreshUserBehaviorTTL(userID uint) error {
	key := upcs.keyManager.GenerateUserBehaviorKey(userID)
	ttl := CacheTTL["behavior"]
	return upcs.cache.Expire(key, ttl)
}

// GetUserPreferenceStats 获取用户偏好统计信息
func (upcs *UserPreferenceCacheService) GetUserPreferenceStats(userID uint) (*UserPreferenceStats, error) {
	stats := &UserPreferenceStats{
		UserID: userID,
	}

	// 获取浏览历史统计
	browseHistory, err := upcs.GetUserBrowseHistory(userID)
	if err == nil && browseHistory != nil {
		stats.BrowseHistoryCount = len(browseHistory.Items)
		stats.TotalViewCount = browseHistory.TotalView
		stats.HasBrowseHistory = true
	}

	// 获取收藏统计
	favorite, err := upcs.GetUserFavorite(userID)
	if err == nil && favorite != nil {
		stats.FavoriteCount = favorite.TotalCount
		stats.FavoriteCategoryCount = len(favorite.CategoryMap)
		stats.HasFavorite = true
	}

	// 获取推荐统计
	recommendation, err := upcs.GetUserRecommendation(userID)
	if err == nil && recommendation != nil {
		stats.PersonalRecommendationCount = len(recommendation.PersonalBased)
		stats.BehaviorRecommendationCount = len(recommendation.BehaviorBased)
		stats.HasRecommendation = true
		stats.RecommendationAlgorithmVersion = recommendation.AlgorithmVersion
	}

	// 获取行为统计
	behavior, err := upcs.GetUserBehavior(userID)
	if err == nil && behavior != nil {
		stats.SearchHistoryCount = len(behavior.SearchHistory)
		stats.ClickBehaviorCount = len(behavior.ClickBehavior)
		stats.HasBehavior = true
	}

	return stats, nil
}

// UserPreferenceStats 用户偏好统计信息
type UserPreferenceStats struct {
	UserID                         uint   `json:"user_id"`
	BrowseHistoryCount             int    `json:"browse_history_count"`
	TotalViewCount                 int    `json:"total_view_count"`
	FavoriteCount                  int    `json:"favorite_count"`
	FavoriteCategoryCount          int    `json:"favorite_category_count"`
	PersonalRecommendationCount    int    `json:"personal_recommendation_count"`
	BehaviorRecommendationCount    int    `json:"behavior_recommendation_count"`
	SearchHistoryCount             int    `json:"search_history_count"`
	ClickBehaviorCount             int    `json:"click_behavior_count"`
	HasBrowseHistory               bool   `json:"has_browse_history"`
	HasFavorite                    bool   `json:"has_favorite"`
	HasRecommendation              bool   `json:"has_recommendation"`
	HasBehavior                    bool   `json:"has_behavior"`
	RecommendationAlgorithmVersion string `json:"recommendation_algorithm_version"`
}

// 辅助函数
func getMainProductImage(product *model.Product) string {
	if len(product.Images) > 0 {
		for _, img := range product.Images {
			if img.IsMain {
				return img.URL
			}
		}
		return product.Images[0].URL
	}
	return ""
}

func getCategoryName(categoryID uint) string {
	// 这里应该从缓存或数据库获取分类名称
	// 简化处理，返回默认值
	return fmt.Sprintf("分类%d", categoryID)
}
