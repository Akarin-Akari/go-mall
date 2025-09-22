package product

import (
	"fmt"
	"strings"

	"mall-go/internal/model"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// SearchService 商品搜索服务
type SearchService struct {
	db *gorm.DB
}

// NewSearchService 创建商品搜索服务
func NewSearchService(db *gorm.DB) *SearchService {
	return &SearchService{
		db: db,
	}
}

// SearchRequest 搜索请求
type SearchRequest struct {
	Keyword     string   `form:"keyword"`
	CategoryID  *uint    `form:"category_id"`
	BrandID     *uint    `form:"brand_id"`
	MinPrice    *string  `form:"min_price"`
	MaxPrice    *string  `form:"max_price"`
	Tags        []string `form:"tags"`
	IsHot       *bool    `form:"is_hot"`
	IsNew       *bool    `form:"is_new"`
	IsRecommend *bool    `form:"is_recommend"`
	InStock     *bool    `form:"in_stock"`
	SortBy      string   `form:"sort_by"` // relevance, price_asc, price_desc, sales_desc, rating_desc, created_desc
	Page        int      `form:"page" binding:"min=1"`
	PageSize    int      `form:"page_size" binding:"min=1,max=100"`
}

// SearchResponse 搜索响应
type SearchResponse struct {
	Products    []*model.Product `json:"products"`
	Total       int64            `json:"total"`
	Page        int              `json:"page"`
	PageSize    int              `json:"page_size"`
	TotalPages  int              `json:"total_pages"`
	Keyword     string           `json:"keyword"`
	Suggestions []string         `json:"suggestions"`
	Filters     *SearchFilters   `json:"filters"`
}

// SearchFilters 搜索过滤器
type SearchFilters struct {
	Categories []CategoryFilter `json:"categories"`
	Brands     []BrandFilter    `json:"brands"`
	PriceRange *PriceRange      `json:"price_range"`
	Tags       []string         `json:"tags"`
}

// CategoryFilter 分类过滤器
type CategoryFilter struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Count int64  `json:"count"`
}

// BrandFilter 品牌过滤器
type BrandFilter struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Count int64  `json:"count"`
}

// PriceRange 价格范围
type PriceRange struct {
	Min decimal.Decimal `json:"min"`
	Max decimal.Decimal `json:"max"`
}

// HotSearchKeyword 热门搜索关键词
type HotSearchKeyword struct {
	ID        uint   `gorm:"primarykey" json:"id"`
	Keyword   string `gorm:"size:100;not null;uniqueIndex" json:"keyword"`
	Count     int64  `gorm:"default:0" json:"count"`
	Sort      int    `gorm:"default:0" json:"sort"`
	Status    string `gorm:"size:20;default:'active'" json:"status"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// TableName 指定表名
func (HotSearchKeyword) TableName() string {
	return "hot_search_keywords"
}

// SearchHistory 搜索历史
type SearchHistory struct {
	ID        uint   `gorm:"primarykey" json:"id"`
	UserID    uint   `gorm:"not null;index" json:"user_id"`
	Keyword   string `gorm:"size:100;not null" json:"keyword"`
	Results   int64  `gorm:"default:0" json:"results"`
	CreatedAt string `json:"created_at"`

	// 关联关系
	User *model.User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName 指定表名
func (SearchHistory) TableName() string {
	return "search_histories"
}

// SearchProducts 搜索商品
func (ss *SearchService) SearchProducts(req *SearchRequest) (*SearchResponse, error) {
	query := ss.db.Model(&model.Product{}).Where("status = ?", model.ProductStatusActive)

	// 关键词搜索
	if req.Keyword != "" {
		// 记录搜索关键词
		ss.recordSearchKeyword(req.Keyword)

		// 多字段模糊搜索
		keywords := strings.Fields(req.Keyword)
		for _, keyword := range keywords {
			query = query.Where("(name LIKE ? OR description LIKE ? OR seo_keywords LIKE ?)",
				"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
		}
	}

	// 分类筛选
	if req.CategoryID != nil {
		query = query.Where("category_id = ?", *req.CategoryID)
	}

	// 品牌筛选
	if req.BrandID != nil {
		query = query.Where("brand_id = ?", *req.BrandID)
	}

	// 价格范围筛选
	if req.MinPrice != nil && *req.MinPrice != "" {
		if minPrice, err := decimal.NewFromString(*req.MinPrice); err == nil {
			query = query.Where("price >= ?", minPrice)
		}
	}

	if req.MaxPrice != nil && *req.MaxPrice != "" {
		if maxPrice, err := decimal.NewFromString(*req.MaxPrice); err == nil {
			query = query.Where("price <= ?", maxPrice)
		}
	}

	// 标签筛选
	if len(req.Tags) > 0 {
		for _, tag := range req.Tags {
			query = query.Where("seo_keywords LIKE ?", "%"+tag+"%")
		}
	}

	// 特殊属性筛选
	if req.IsHot != nil {
		query = query.Where("is_hot = ?", *req.IsHot)
	}

	if req.IsNew != nil {
		query = query.Where("is_new = ?", *req.IsNew)
	}

	if req.IsRecommend != nil {
		query = query.Where("is_recommend = ?", *req.IsRecommend)
	}

	// 库存筛选
	if req.InStock != nil && *req.InStock {
		query = query.Where("stock > 0")
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("查询商品总数失败: %v", err)
	}

	// 排序
	orderBy := ss.buildOrderBy(req.SortBy, req.Keyword)
	query = query.Order(orderBy)

	// 分页查询
	var products []*model.Product
	offset := (req.Page - 1) * req.PageSize
	if err := query.Preload("Category").
		Preload("Brand").
		Preload("Images", func(db *gorm.DB) *gorm.DB {
			return db.Where("is_main = ?", true).Order("sort ASC").Limit(1)
		}).
		Offset(offset).
		Limit(req.PageSize).
		Find(&products).Error; err != nil {
		return nil, fmt.Errorf("查询商品列表失败: %v", err)
	}

	// 构建响应
	totalPages := int((total + int64(req.PageSize) - 1) / int64(req.PageSize))

	response := &SearchResponse{
		Products:   products,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
		Keyword:    req.Keyword,
	}

	// 获取搜索建议
	if req.Keyword != "" {
		suggestions, _ := ss.GetSearchSuggestions(req.Keyword, 5)
		response.Suggestions = suggestions
	}

	// 获取搜索过滤器
	filters, _ := ss.buildSearchFilters(req)
	response.Filters = filters

	return response, nil
}

// GetSearchSuggestions 获取搜索建议
func (ss *SearchService) GetSearchSuggestions(keyword string, limit int) ([]string, error) {
	var suggestions []string

	// 从商品名称中获取建议
	var productNames []string
	if err := ss.db.Model(&model.Product{}).
		Select("DISTINCT name").
		Where("status = ? AND name LIKE ?", model.ProductStatusActive, "%"+keyword+"%").
		Limit(limit).
		Pluck("name", &productNames).Error; err == nil {
		suggestions = append(suggestions, productNames...)
	}

	// 从热门搜索关键词中获取建议
	if len(suggestions) < limit {
		var hotKeywords []string
		remaining := limit - len(suggestions)
		if err := ss.db.Model(&HotSearchKeyword{}).
			Select("keyword").
			Where("status = 'active' AND keyword LIKE ?", "%"+keyword+"%").
			Order("count DESC, sort ASC").
			Limit(remaining).
			Pluck("keyword", &hotKeywords).Error; err == nil {
			suggestions = append(suggestions, hotKeywords...)
		}
	}

	return suggestions, nil
}

// GetHotSearchKeywords 获取热门搜索关键词
func (ss *SearchService) GetHotSearchKeywords(limit int) ([]HotSearchKeyword, error) {
	var keywords []HotSearchKeyword
	if err := ss.db.Where("status = 'active'").
		Order("sort ASC, count DESC").
		Limit(limit).
		Find(&keywords).Error; err != nil {
		return nil, fmt.Errorf("查询热门搜索关键词失败: %v", err)
	}

	return keywords, nil
}

// SaveSearchHistory 保存搜索历史
func (ss *SearchService) SaveSearchHistory(userID uint, keyword string, results int64) error {
	if userID == 0 || keyword == "" {
		return nil
	}

	history := &SearchHistory{
		UserID:  userID,
		Keyword: keyword,
		Results: results,
	}

	if err := ss.db.Create(history).Error; err != nil {
		return fmt.Errorf("保存搜索历史失败: %v", err)
	}

	return nil
}

// GetUserSearchHistory 获取用户搜索历史
func (ss *SearchService) GetUserSearchHistory(userID uint, limit int) ([]SearchHistory, error) {
	var histories []SearchHistory
	if err := ss.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&histories).Error; err != nil {
		return nil, fmt.Errorf("查询搜索历史失败: %v", err)
	}

	return histories, nil
}

// ClearUserSearchHistory 清空用户搜索历史
func (ss *SearchService) ClearUserSearchHistory(userID uint) error {
	if err := ss.db.Where("user_id = ?", userID).Delete(&SearchHistory{}).Error; err != nil {
		return fmt.Errorf("清空搜索历史失败: %v", err)
	}

	return nil
}

// GetRelatedProducts 获取相关商品
func (ss *SearchService) GetRelatedProducts(productID uint, limit int) ([]*model.Product, error) {
	// 获取当前商品信息
	var product model.Product
	if err := ss.db.First(&product, productID).Error; err != nil {
		return nil, fmt.Errorf("商品不存在")
	}

	var products []*model.Product

	// 优先推荐同分类商品
	if err := ss.db.Where("category_id = ? AND id != ? AND status = ?",
		product.CategoryID, productID, model.ProductStatusActive).
		Preload("Images", func(db *gorm.DB) *gorm.DB {
			return db.Where("is_main = ?", true).Order("sort ASC").Limit(1)
		}).
		Order("is_recommend DESC, sold_count DESC").
		Limit(limit).
		Find(&products).Error; err != nil {
		return nil, fmt.Errorf("查询相关商品失败: %v", err)
	}

	// 如果同分类商品不够，补充同品牌商品
	if len(products) < limit && product.BrandID > 0 {
		var brandProducts []*model.Product
		remaining := limit - len(products)

		existingIDs := []uint{productID}
		for _, p := range products {
			existingIDs = append(existingIDs, p.ID)
		}

		if err := ss.db.Where("brand_id = ? AND id NOT IN ? AND status = ?",
			product.BrandID, existingIDs, model.ProductStatusActive).
			Preload("Images", func(db *gorm.DB) *gorm.DB {
				return db.Where("is_main = ?", true).Order("sort ASC").Limit(1)
			}).
			Order("is_recommend DESC, sold_count DESC").
			Limit(remaining).
			Find(&brandProducts).Error; err == nil {
			products = append(products, brandProducts...)
		}
	}

	return products, nil
}

// recordSearchKeyword 记录搜索关键词
func (ss *SearchService) recordSearchKeyword(keyword string) {
	// 异步更新搜索次数
	go func() {
		var hotKeyword HotSearchKeyword
		if err := ss.db.Where("keyword = ?", keyword).First(&hotKeyword).Error; err != nil {
			// 创建新记录
			hotKeyword = HotSearchKeyword{
				Keyword: keyword,
				Count:   1,
				Status:  "active",
			}
			ss.db.Create(&hotKeyword)
		} else {
			// 更新计数
			ss.db.Model(&hotKeyword).UpdateColumn("count", gorm.Expr("count + ?", 1))
		}
	}()
}

// buildOrderBy 构建排序条件
func (ss *SearchService) buildOrderBy(sortBy, keyword string) string {
	switch sortBy {
	case "price_asc":
		return "price ASC"
	case "price_desc":
		return "price DESC"
	case "sales_desc":
		return "sold_count DESC"
	case "rating_desc":
		return "view_count DESC" // 暂时用浏览量代替评分
	case "created_desc":
		return "created_at DESC"
	case "relevance":
		if keyword != "" {
			// 相关性排序：名称匹配 > 描述匹配 > 关键词匹配
			return "CASE WHEN name LIKE '%" + keyword + "%' THEN 1 WHEN description LIKE '%" + keyword + "%' THEN 2 ELSE 3 END, sold_count DESC"
		}
		return "is_recommend DESC, sold_count DESC"
	default:
		return "is_recommend DESC, sold_count DESC, sort ASC"
	}
}

// buildSearchFilters 构建搜索过滤器
func (ss *SearchService) buildSearchFilters(req *SearchRequest) (*SearchFilters, error) {
	filters := &SearchFilters{}

	// 构建基础查询
	baseQuery := ss.db.Model(&model.Product{}).Where("status = ?", model.ProductStatusActive)

	if req.Keyword != "" {
		keywords := strings.Fields(req.Keyword)
		for _, keyword := range keywords {
			baseQuery = baseQuery.Where("(name LIKE ? OR description LIKE ? OR seo_keywords LIKE ?)",
				"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
		}
	}

	// 获取分类过滤器
	var categoryFilters []CategoryFilter
	ss.db.Table("products p").
		Select("c.id, c.name, COUNT(*) as count").
		Joins("JOIN categories c ON p.category_id = c.id").
		Where("p.status = ?", model.ProductStatusActive).
		Group("c.id, c.name").
		Order("count DESC").
		Limit(10).
		Scan(&categoryFilters)
	filters.Categories = categoryFilters

	// 获取品牌过滤器
	var brandFilters []BrandFilter
	ss.db.Table("products p").
		Select("b.id, b.name, COUNT(*) as count").
		Joins("JOIN brands b ON p.brand_id = b.id").
		Where("p.status = ? AND p.brand_id > 0", model.ProductStatusActive).
		Group("b.id, b.name").
		Order("count DESC").
		Limit(10).
		Scan(&brandFilters)
	filters.Brands = brandFilters

	// 获取价格范围
	var priceRange PriceRange
	ss.db.Model(&model.Product{}).
		Select("MIN(price) as min, MAX(price) as max").
		Where("status = ?", model.ProductStatusActive).
		Scan(&priceRange)
	filters.PriceRange = &priceRange

	return filters, nil
}

// 全局搜索服务实例
var globalSearchService *SearchService

// InitGlobalSearchService 初始化全局搜索服务
func InitGlobalSearchService(db *gorm.DB) {
	globalSearchService = NewSearchService(db)
}

// GetGlobalSearchService 获取全局搜索服务
func GetGlobalSearchService() *SearchService {
	return globalSearchService
}
