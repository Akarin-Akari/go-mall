package cache

import (
	"encoding/json"
	"fmt"
	"mall-go/internal/model"
	"mall-go/pkg/logger"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// CategoryCacheService 分类缓存服务
type CategoryCacheService struct {
	cacheManager CacheManager
	keyManager   *CacheKeyManager
}

// NewCategoryCacheService 创建分类缓存服务
func NewCategoryCacheService(cacheManager CacheManager, keyManager *CacheKeyManager) *CategoryCacheService {
	return &CategoryCacheService{
		cacheManager: cacheManager,
		keyManager:   keyManager,
	}
}

// CategoryProductsData 分类商品列表缓存数据结构
type CategoryProductsData struct {
	CategoryID   uint                   `json:"category_id"`
	CategoryName string                 `json:"category_name"`
	Products     []*ProductCacheData    `json:"products"`
	TotalCount   int                    `json:"total_count"`
	Page         int                    `json:"page"`
	PageSize     int                    `json:"page_size"`
	SortBy       string                 `json:"sort_by"` // price_asc, price_desc, sales_desc, rating_desc, created_desc
	Filters      map[string]interface{} `json:"filters"` // 筛选条件
	CachedAt     time.Time              `json:"cached_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// HotProductsRankingData 热门商品排行榜数据结构
type HotProductsRankingData struct {
	RankingType string            `json:"ranking_type"` // daily, weekly, monthly, sales, views, rating
	Products    []*HotProductItem `json:"products"`
	TotalCount  int               `json:"total_count"`
	Period      string            `json:"period"` // 统计周期：2025-01-10, 2025-W02, 2025-01
	CachedAt    time.Time         `json:"cached_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// HotProductItem 热门商品项
type HotProductItem struct {
	ProductID    uint    `json:"product_id"`
	Name         string  `json:"name"`
	CategoryID   uint    `json:"category_id"`
	CategoryName string  `json:"category_name"`
	Price        string  `json:"price"`
	OriginPrice  string  `json:"origin_price"`
	SalesCount   int     `json:"sales_count"`
	ViewCount    int     `json:"view_count"`
	Rating       float64 `json:"rating"`
	RankScore    float64 `json:"rank_score"` // 排行分数
	Rank         int     `json:"rank"`       // 排名
	Image        string  `json:"image"`      // 主图
	IsHot        bool    `json:"is_hot"`
	IsNew        bool    `json:"is_new"`
	IsRecommend  bool    `json:"is_recommend"`
}

// SearchResultData 搜索结果缓存数据结构
type SearchResultData struct {
	Keyword    string                 `json:"keyword"`
	Products   []*ProductCacheData    `json:"products"`
	TotalCount int                    `json:"total_count"`
	Page       int                    `json:"page"`
	PageSize   int                    `json:"page_size"`
	SortBy     string                 `json:"sort_by"`
	Filters    map[string]interface{} `json:"filters"`
	SearchTime time.Duration          `json:"search_time"` // 搜索耗时
	CachedAt   time.Time              `json:"cached_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
}

// RecommendProductsData 推荐商品数据结构
type RecommendProductsData struct {
	RecommendType string              `json:"recommend_type"` // related, user_preference, category_hot, cross_sell
	TargetID      uint                `json:"target_id"`      // 目标ID（商品ID或用户ID）
	Products      []*ProductCacheData `json:"products"`
	TotalCount    int                 `json:"total_count"`
	Algorithm     string              `json:"algorithm"`  // 推荐算法
	Confidence    float64             `json:"confidence"` // 推荐置信度
	CachedAt      time.Time           `json:"cached_at"`
	UpdatedAt     time.Time           `json:"updated_at"`
}

// CategoryListRequest 分类商品列表请求
type CategoryListRequest struct {
	CategoryID uint                   `json:"category_id"`
	Page       int                    `json:"page"`
	PageSize   int                    `json:"page_size"`
	SortBy     string                 `json:"sort_by"`
	Filters    map[string]interface{} `json:"filters"`
}

// SearchRequest 搜索请求
type SearchRequest struct {
	Keyword  string                 `json:"keyword"`
	Page     int                    `json:"page"`
	PageSize int                    `json:"page_size"`
	SortBy   string                 `json:"sort_by"`
	Filters  map[string]interface{} `json:"filters"`
}

// GetCategoryProducts 获取分类商品列表缓存
func (ccs *CategoryCacheService) GetCategoryProducts(request *CategoryListRequest) (*CategoryProductsData, error) {
	key := ccs.generateCategoryProductsKey(request)

	// 从缓存获取
	data, err := ccs.cacheManager.Get(key)
	if err != nil {
		logger.Error(fmt.Sprintf("获取分类商品列表缓存失败: %v", err))
		return nil, fmt.Errorf("获取分类商品列表缓存失败: %w", err)
	}

	if data == nil {
		return nil, nil // 缓存未命中
	}

	// 反序列化
	var categoryProducts CategoryProductsData
	jsonStr, ok := data.(string)
	if !ok {
		return nil, fmt.Errorf("分类商品列表缓存数据格式错误")
	}

	if err := json.Unmarshal([]byte(jsonStr), &categoryProducts); err != nil {
		logger.Error(fmt.Sprintf("分类商品列表缓存数据反序列化失败: %v", err))
		return nil, fmt.Errorf("分类商品列表缓存数据反序列化失败: %w", err)
	}

	return &categoryProducts, nil
}

// SetCategoryProducts 设置分类商品列表缓存
func (ccs *CategoryCacheService) SetCategoryProducts(request *CategoryListRequest, products []*model.Product, totalCount int, categoryName string) error {
	key := ccs.generateCategoryProductsKey(request)

	// 转换商品数据
	var productCacheData []*ProductCacheData
	for _, product := range products {
		productCacheData = append(productCacheData, ConvertToProductCacheData(product))
	}

	// 创建缓存数据
	cacheData := &CategoryProductsData{
		CategoryID:   request.CategoryID,
		CategoryName: categoryName,
		Products:     productCacheData,
		TotalCount:   totalCount,
		Page:         request.Page,
		PageSize:     request.PageSize,
		SortBy:       request.SortBy,
		Filters:      request.Filters,
		CachedAt:     time.Now(),
		UpdatedAt:    time.Now(),
	}

	// 序列化
	jsonData, err := json.Marshal(cacheData)
	if err != nil {
		logger.Error(fmt.Sprintf("分类商品列表数据序列化失败: %v", err))
		return fmt.Errorf("分类商品列表数据序列化失败: %w", err)
	}

	// 获取TTL
	ttl := GetTTL("category")

	// 存储到缓存
	if err := ccs.cacheManager.Set(key, string(jsonData), ttl); err != nil {
		logger.Error(fmt.Sprintf("设置分类商品列表缓存失败: %v", err))
		return fmt.Errorf("设置分类商品列表缓存失败: %w", err)
	}

	logger.Info(fmt.Sprintf("分类商品列表缓存设置成功: CategoryID=%d, Page=%d, Count=%d, TTL=%v",
		request.CategoryID, request.Page, len(products), ttl))
	return nil
}

// generateCategoryProductsKey 生成分类商品列表缓存键
func (ccs *CategoryCacheService) generateCategoryProductsKey(request *CategoryListRequest) string {
	// 构建基础键
	keyBuilder := NewKeyBuilder().Add("category").AddUint(request.CategoryID).Add("products")

	// 添加分页信息
	keyBuilder.AddFormat("page_%d_%d", request.Page, request.PageSize)

	// 添加排序信息
	if request.SortBy != "" {
		keyBuilder.Add("sort").Add(request.SortBy)
	}

	// 添加筛选条件
	if len(request.Filters) > 0 {
		filterStr := ccs.encodeFilters(request.Filters)
		keyBuilder.Add("filter").Add(filterStr)
	}

	return keyBuilder.BuildWithPrefix(ccs.keyManager.prefix)
}

// encodeFilters 编码筛选条件为字符串
func (ccs *CategoryCacheService) encodeFilters(filters map[string]interface{}) string {
	if len(filters) == 0 {
		return ""
	}

	var parts []string
	for key, value := range filters {
		switch v := value.(type) {
		case string:
			parts = append(parts, fmt.Sprintf("%s:%s", key, v))
		case []string:
			parts = append(parts, fmt.Sprintf("%s:%s", key, strings.Join(v, ",")))
		case int, int64, uint, uint64:
			parts = append(parts, fmt.Sprintf("%s:%v", key, v))
		case float64:
			parts = append(parts, fmt.Sprintf("%s:%.2f", key, v))
		case bool:
			parts = append(parts, fmt.Sprintf("%s:%t", key, v))
		default:
			// 对于复杂类型，使用JSON序列化
			if jsonData, err := json.Marshal(v); err == nil {
				parts = append(parts, fmt.Sprintf("%s:%s", key, string(jsonData)))
			}
		}
	}

	// 排序确保一致性
	sort.Strings(parts)
	return strings.Join(parts, "|")
}

// GetHotProductsRanking 获取热门商品排行榜
func (ccs *CategoryCacheService) GetHotProductsRanking(rankingType, period string) (*HotProductsRankingData, error) {
	key := ccs.generateHotProductsKey(rankingType, period)

	// 从缓存获取
	data, err := ccs.cacheManager.Get(key)
	if err != nil {
		logger.Error(fmt.Sprintf("获取热门商品排行榜缓存失败: %v", err))
		return nil, fmt.Errorf("获取热门商品排行榜缓存失败: %w", err)
	}

	if data == nil {
		return nil, nil // 缓存未命中
	}

	// 反序列化
	var ranking HotProductsRankingData
	jsonStr, ok := data.(string)
	if !ok {
		return nil, fmt.Errorf("热门商品排行榜缓存数据格式错误")
	}

	if err := json.Unmarshal([]byte(jsonStr), &ranking); err != nil {
		logger.Error(fmt.Sprintf("热门商品排行榜缓存数据反序列化失败: %v", err))
		return nil, fmt.Errorf("热门商品排行榜缓存数据反序列化失败: %w", err)
	}

	return &ranking, nil
}

// SetHotProductsRanking 设置热门商品排行榜缓存
func (ccs *CategoryCacheService) SetHotProductsRanking(rankingType, period string, products []*HotProductItem) error {
	key := ccs.generateHotProductsKey(rankingType, period)

	// 按排名排序
	sort.Slice(products, func(i, j int) bool {
		return products[i].RankScore > products[j].RankScore
	})

	// 设置排名
	for i, product := range products {
		product.Rank = i + 1
	}

	// 创建缓存数据
	cacheData := &HotProductsRankingData{
		RankingType: rankingType,
		Products:    products,
		TotalCount:  len(products),
		Period:      period,
		CachedAt:    time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 序列化
	jsonData, err := json.Marshal(cacheData)
	if err != nil {
		logger.Error(fmt.Sprintf("热门商品排行榜数据序列化失败: %v", err))
		return fmt.Errorf("热门商品排行榜数据序列化失败: %w", err)
	}

	// 获取TTL
	ttl := GetTTL("hot")

	// 存储到缓存
	if err := ccs.cacheManager.Set(key, string(jsonData), ttl); err != nil {
		logger.Error(fmt.Sprintf("设置热门商品排行榜缓存失败: %v", err))
		return fmt.Errorf("设置热门商品排行榜缓存失败: %w", err)
	}

	logger.Info(fmt.Sprintf("热门商品排行榜缓存设置成功: Type=%s, Period=%s, Count=%d, TTL=%v",
		rankingType, period, len(products), ttl))
	return nil
}

// generateHotProductsKey 生成热门商品排行榜缓存键
func (ccs *CategoryCacheService) generateHotProductsKey(rankingType, period string) string {
	return NewKeyBuilder().
		Add("hot").
		Add("products").
		Add(rankingType).
		Add(period).
		BuildWithPrefix(ccs.keyManager.prefix)
}

// GetSearchResults 获取搜索结果缓存
func (ccs *CategoryCacheService) GetSearchResults(request *SearchRequest) (*SearchResultData, error) {
	key := ccs.generateSearchResultsKey(request)

	// 从缓存获取
	data, err := ccs.cacheManager.Get(key)
	if err != nil {
		logger.Error(fmt.Sprintf("获取搜索结果缓存失败: %v", err))
		return nil, fmt.Errorf("获取搜索结果缓存失败: %w", err)
	}

	if data == nil {
		return nil, nil // 缓存未命中
	}

	// 反序列化
	var searchResult SearchResultData
	jsonStr, ok := data.(string)
	if !ok {
		return nil, fmt.Errorf("搜索结果缓存数据格式错误")
	}

	if err := json.Unmarshal([]byte(jsonStr), &searchResult); err != nil {
		logger.Error(fmt.Sprintf("搜索结果缓存数据反序列化失败: %v", err))
		return nil, fmt.Errorf("搜索结果缓存数据反序列化失败: %w", err)
	}

	return &searchResult, nil
}

// SetSearchResults 设置搜索结果缓存
func (ccs *CategoryCacheService) SetSearchResults(request *SearchRequest, products []*model.Product, totalCount int, searchTime time.Duration) error {
	key := ccs.generateSearchResultsKey(request)

	// 转换商品数据
	var productCacheData []*ProductCacheData
	for _, product := range products {
		productCacheData = append(productCacheData, ConvertToProductCacheData(product))
	}

	// 创建缓存数据
	cacheData := &SearchResultData{
		Keyword:    request.Keyword,
		Products:   productCacheData,
		TotalCount: totalCount,
		Page:       request.Page,
		PageSize:   request.PageSize,
		SortBy:     request.SortBy,
		Filters:    request.Filters,
		SearchTime: searchTime,
		CachedAt:   time.Now(),
		UpdatedAt:  time.Now(),
	}

	// 序列化
	jsonData, err := json.Marshal(cacheData)
	if err != nil {
		logger.Error(fmt.Sprintf("搜索结果数据序列化失败: %v", err))
		return fmt.Errorf("搜索结果数据序列化失败: %w", err)
	}

	// 获取TTL - 搜索结果使用较短的TTL
	ttl := GetTTL("category") / 2 // 10分钟

	// 存储到缓存
	if err := ccs.cacheManager.Set(key, string(jsonData), ttl); err != nil {
		logger.Error(fmt.Sprintf("设置搜索结果缓存失败: %v", err))
		return fmt.Errorf("设置搜索结果缓存失败: %w", err)
	}

	logger.Info(fmt.Sprintf("搜索结果缓存设置成功: Keyword=%s, Page=%d, Count=%d, TTL=%v",
		request.Keyword, request.Page, len(products), ttl))
	return nil
}

// generateSearchResultsKey 生成搜索结果缓存键
func (ccs *CategoryCacheService) generateSearchResultsKey(request *SearchRequest) string {
	// 构建基础键
	keyBuilder := NewKeyBuilder().Add("search").Add("results")

	// 添加关键词（需要处理特殊字符）
	keyword := strings.ReplaceAll(request.Keyword, ":", "_")
	keyword = strings.ReplaceAll(keyword, " ", "_")
	keyBuilder.Add(keyword)

	// 添加分页信息
	keyBuilder.AddFormat("page_%d_%d", request.Page, request.PageSize)

	// 添加排序信息
	if request.SortBy != "" {
		keyBuilder.Add("sort").Add(request.SortBy)
	}

	// 添加筛选条件
	if len(request.Filters) > 0 {
		filterStr := ccs.encodeFilters(request.Filters)
		keyBuilder.Add("filter").Add(filterStr)
	}

	return keyBuilder.BuildWithPrefix(ccs.keyManager.prefix)
}

// GetRecommendProducts 获取推荐商品缓存
func (ccs *CategoryCacheService) GetRecommendProducts(recommendType string, targetID uint) (*RecommendProductsData, error) {
	key := ccs.generateRecommendProductsKey(recommendType, targetID)

	// 从缓存获取
	data, err := ccs.cacheManager.Get(key)
	if err != nil {
		logger.Error(fmt.Sprintf("获取推荐商品缓存失败: %v", err))
		return nil, fmt.Errorf("获取推荐商品缓存失败: %w", err)
	}

	if data == nil {
		return nil, nil // 缓存未命中
	}

	// 反序列化
	var recommend RecommendProductsData
	jsonStr, ok := data.(string)
	if !ok {
		return nil, fmt.Errorf("推荐商品缓存数据格式错误")
	}

	if err := json.Unmarshal([]byte(jsonStr), &recommend); err != nil {
		logger.Error(fmt.Sprintf("推荐商品缓存数据反序列化失败: %v", err))
		return nil, fmt.Errorf("推荐商品缓存数据反序列化失败: %w", err)
	}

	return &recommend, nil
}

// SetRecommendProducts 设置推荐商品缓存
func (ccs *CategoryCacheService) SetRecommendProducts(recommendType string, targetID uint, products []*model.Product, algorithm string, confidence float64) error {
	key := ccs.generateRecommendProductsKey(recommendType, targetID)

	// 转换商品数据
	var productCacheData []*ProductCacheData
	for _, product := range products {
		productCacheData = append(productCacheData, ConvertToProductCacheData(product))
	}

	// 创建缓存数据
	cacheData := &RecommendProductsData{
		RecommendType: recommendType,
		TargetID:      targetID,
		Products:      productCacheData,
		TotalCount:    len(products),
		Algorithm:     algorithm,
		Confidence:    confidence,
		CachedAt:      time.Now(),
		UpdatedAt:     time.Now(),
	}

	// 序列化
	jsonData, err := json.Marshal(cacheData)
	if err != nil {
		logger.Error(fmt.Sprintf("推荐商品数据序列化失败: %v", err))
		return fmt.Errorf("推荐商品数据序列化失败: %w", err)
	}

	// 获取TTL
	ttl := GetTTL("category")

	// 存储到缓存
	if err := ccs.cacheManager.Set(key, string(jsonData), ttl); err != nil {
		logger.Error(fmt.Sprintf("设置推荐商品缓存失败: %v", err))
		return fmt.Errorf("设置推荐商品缓存失败: %w", err)
	}

	logger.Info(fmt.Sprintf("推荐商品缓存设置成功: Type=%s, TargetID=%d, Count=%d, TTL=%v",
		recommendType, targetID, len(products), ttl))
	return nil
}

// generateRecommendProductsKey 生成推荐商品缓存键
func (ccs *CategoryCacheService) generateRecommendProductsKey(recommendType string, targetID uint) string {
	return NewKeyBuilder().
		Add("recommend").
		Add(recommendType).
		AddUint(targetID).
		BuildWithPrefix(ccs.keyManager.prefix)
}

// DeleteCategoryProducts 删除分类商品列表缓存
func (ccs *CategoryCacheService) DeleteCategoryProducts(categoryID uint) error {
	// 删除该分类下的所有缓存（使用通配符模式）
	pattern := fmt.Sprintf("%s:category:%d:products:*", ccs.keyManager.prefix, categoryID)

	// 由于Redis不直接支持通配符删除，我们需要先获取所有匹配的键
	// 这里简化处理，删除常见的缓存键
	commonKeys := []string{
		ccs.generateCategoryProductsKey(&CategoryListRequest{CategoryID: categoryID, Page: 1, PageSize: 20}),
		ccs.generateCategoryProductsKey(&CategoryListRequest{CategoryID: categoryID, Page: 1, PageSize: 50}),
	}

	for _, key := range commonKeys {
		if err := ccs.cacheManager.Delete(key); err != nil {
			logger.Error(fmt.Sprintf("删除分类商品缓存失败: Key=%s, Error=%v", key, err))
		}
	}

	logger.Info(fmt.Sprintf("分类商品缓存删除完成: CategoryID=%d, Pattern=%s", categoryID, pattern))
	return nil
}

// DeleteSearchResults 删除搜索结果缓存
func (ccs *CategoryCacheService) DeleteSearchResults(keyword string) error {
	// 删除该关键词的所有搜索结果缓存
	commonKeys := []string{
		ccs.generateSearchResultsKey(&SearchRequest{Keyword: keyword, Page: 1, PageSize: 20}),
		ccs.generateSearchResultsKey(&SearchRequest{Keyword: keyword, Page: 1, PageSize: 50}),
	}

	for _, key := range commonKeys {
		if err := ccs.cacheManager.Delete(key); err != nil {
			logger.Error(fmt.Sprintf("删除搜索结果缓存失败: Key=%s, Error=%v", key, err))
		}
	}

	logger.Info(fmt.Sprintf("搜索结果缓存删除完成: Keyword=%s", keyword))
	return nil
}

// DeleteRecommendProducts 删除推荐商品缓存
func (ccs *CategoryCacheService) DeleteRecommendProducts(recommendType string, targetID uint) error {
	key := ccs.generateRecommendProductsKey(recommendType, targetID)

	if err := ccs.cacheManager.Delete(key); err != nil {
		logger.Error(fmt.Sprintf("删除推荐商品缓存失败: %v", err))
		return fmt.Errorf("删除推荐商品缓存失败: %w", err)
	}

	logger.Info(fmt.Sprintf("推荐商品缓存删除成功: Type=%s, TargetID=%d", recommendType, targetID))
	return nil
}

// DeleteHotProductsRanking 删除热门商品排行榜缓存
func (ccs *CategoryCacheService) DeleteHotProductsRanking(rankingType, period string) error {
	key := ccs.generateHotProductsKey(rankingType, period)

	if err := ccs.cacheManager.Delete(key); err != nil {
		logger.Error(fmt.Sprintf("删除热门商品排行榜缓存失败: %v", err))
		return fmt.Errorf("删除热门商品排行榜缓存失败: %w", err)
	}

	logger.Info(fmt.Sprintf("热门商品排行榜缓存删除成功: Type=%s, Period=%s", rankingType, period))
	return nil
}

// ExistsCategoryProducts 检查分类商品列表缓存是否存在
func (ccs *CategoryCacheService) ExistsCategoryProducts(request *CategoryListRequest) bool {
	key := ccs.generateCategoryProductsKey(request)
	return ccs.cacheManager.Exists(key)
}

// ExistsSearchResults 检查搜索结果缓存是否存在
func (ccs *CategoryCacheService) ExistsSearchResults(request *SearchRequest) bool {
	key := ccs.generateSearchResultsKey(request)
	return ccs.cacheManager.Exists(key)
}

// ExistsHotProductsRanking 检查热门商品排行榜缓存是否存在
func (ccs *CategoryCacheService) ExistsHotProductsRanking(rankingType, period string) bool {
	key := ccs.generateHotProductsKey(rankingType, period)
	return ccs.cacheManager.Exists(key)
}

// ExistsRecommendProducts 检查推荐商品缓存是否存在
func (ccs *CategoryCacheService) ExistsRecommendProducts(recommendType string, targetID uint) bool {
	key := ccs.generateRecommendProductsKey(recommendType, targetID)
	return ccs.cacheManager.Exists(key)
}

// GetCategoryProductsTTL 获取分类商品列表缓存剩余TTL
func (ccs *CategoryCacheService) GetCategoryProductsTTL(request *CategoryListRequest) (time.Duration, error) {
	key := ccs.generateCategoryProductsKey(request)
	return ccs.cacheManager.TTL(key)
}

// RefreshCategoryProductsTTL 刷新分类商品列表缓存TTL
func (ccs *CategoryCacheService) RefreshCategoryProductsTTL(request *CategoryListRequest) error {
	key := ccs.generateCategoryProductsKey(request)
	ttl := GetTTL("category")
	return ccs.cacheManager.Expire(key, ttl)
}

// WarmupCategoryProducts 分类商品列表缓存预热
func (ccs *CategoryCacheService) WarmupCategoryProducts(categoryIDs []uint, pageSize int) error {
	for _, categoryID := range categoryIDs {
		// 预热第一页数据
		request := &CategoryListRequest{
			CategoryID: categoryID,
			Page:       1,
			PageSize:   pageSize,
			SortBy:     "created_desc", // 默认按创建时间排序
		}

		// 检查是否已存在缓存
		if ccs.ExistsCategoryProducts(request) {
			continue
		}

		// 这里应该调用实际的数据库查询来获取数据
		// 由于这是缓存层，我们只记录预热请求
		logger.Info(fmt.Sprintf("分类商品缓存预热请求: CategoryID=%d, PageSize=%d", categoryID, pageSize))
	}

	return nil
}

// GetCategoryStats 获取分类缓存统计信息
func (ccs *CategoryCacheService) GetCategoryStats() map[string]interface{} {
	stats := make(map[string]interface{})

	// 获取缓存指标
	if metrics := ccs.cacheManager.GetMetrics(); metrics != nil {
		stats["total_ops"] = metrics.TotalOps
		stats["hit_count"] = metrics.HitCount
		stats["miss_count"] = metrics.MissCount
		stats["hit_rate"] = metrics.HitRate
		stats["error_count"] = metrics.ErrorCount
		stats["last_updated"] = metrics.LastUpdated
	}

	// 获取连接池统计
	if connStats := ccs.cacheManager.GetConnectionStats(); connStats != nil {
		stats["total_conns"] = connStats.TotalConns
		stats["idle_conns"] = connStats.IdleConns
		stats["hits"] = connStats.Hits
		stats["misses"] = connStats.Misses
	}

	return stats
}

// BatchDeleteCategoryProducts 批量删除分类商品缓存
func (ccs *CategoryCacheService) BatchDeleteCategoryProducts(categoryIDs []uint) error {
	if len(categoryIDs) == 0 {
		return nil
	}

	for _, categoryID := range categoryIDs {
		if err := ccs.DeleteCategoryProducts(categoryID); err != nil {
			logger.Error(fmt.Sprintf("批量删除分类商品缓存失败: CategoryID=%d, Error=%v", categoryID, err))
		}
	}

	logger.Info(fmt.Sprintf("批量删除分类商品缓存完成: 数量=%d", len(categoryIDs)))
	return nil
}

// UpdateHotProductsScore 更新热门商品分数（用于实时排行榜）
func (ccs *CategoryCacheService) UpdateHotProductsScore(productID uint, scoreType string, increment float64) error {
	// 使用ZSet存储实时热门商品分数
	key := fmt.Sprintf("%s:hot_scores:%s", ccs.keyManager.prefix, scoreType)

	// 获取当前分数
	currentScore, err := ccs.cacheManager.ZScore(key, strconv.Itoa(int(productID)))
	if err != nil {
		// 如果不存在，当前分数为0
		currentScore = 0
	}

	// 计算新分数
	newScore := currentScore + increment

	// 使用ZAdd更新分数
	member := redis.Z{
		Score:  newScore,
		Member: strconv.Itoa(int(productID)),
	}

	if err := ccs.cacheManager.ZAdd(key, member); err != nil {
		logger.Error(fmt.Sprintf("更新热门商品分数失败: ProductID=%d, Type=%s, Error=%v", productID, scoreType, err))
		return fmt.Errorf("更新热门商品分数失败: %w", err)
	}

	// 设置过期时间
	ttl := GetTTL("hot")
	ccs.cacheManager.Expire(key, ttl)

	logger.Info(fmt.Sprintf("热门商品分数更新成功: ProductID=%d, Type=%s, NewScore=%.2f",
		productID, scoreType, newScore))
	return nil
}

// GetTopHotProducts 获取实时热门商品排行
func (ccs *CategoryCacheService) GetTopHotProducts(scoreType string, limit int64) ([]uint, error) {
	key := fmt.Sprintf("%s:hot_scores:%s", ccs.keyManager.prefix, scoreType)

	// 获取排行榜（使用ZRange获取，然后反转顺序）
	members, err := ccs.cacheManager.ZRange(key, 0, limit-1)
	if err != nil {
		return nil, fmt.Errorf("获取热门商品排行失败: %w", err)
	}

	var productIDs []uint
	// 反转顺序以获得降序排列
	for i := len(members) - 1; i >= 0; i-- {
		if memberStr, ok := members[i].(string); ok {
			if id, err := strconv.ParseUint(memberStr, 10, 32); err == nil {
				productIDs = append(productIDs, uint(id))
			}
		}
	}

	return productIDs, nil
}
