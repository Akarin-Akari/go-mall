package cart

import (
	"fmt"

	"mall-go/internal/model"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// RecommendationService 购物车推荐服务
type RecommendationService struct {
	db *gorm.DB
}

// NewRecommendationService 创建购物车推荐服务
func NewRecommendationService(db *gorm.DB) *RecommendationService {
	return &RecommendationService{
		db: db,
	}
}

// RecommendationResult 推荐结果
type RecommendationResult struct {
	Products      []*model.Product `json:"products"`
	Reason        string           `json:"reason"`
	RecommendType string           `json:"recommend_type"`
	Score         float64          `json:"score"`
}

// RecommendationResponse 推荐响应
type RecommendationResponse struct {
	CartBased       []*RecommendationResult `json:"cart_based"`       // 基于购物车的推荐
	CategoryBased   []*RecommendationResult `json:"category_based"`   // 基于分类的推荐
	BrandBased      []*RecommendationResult `json:"brand_based"`      // 基于品牌的推荐
	PriceBased      []*RecommendationResult `json:"price_based"`      // 基于价格的推荐
	Complementary   []*RecommendationResult `json:"complementary"`    // 互补商品推荐
	Promotional     []*RecommendationResult `json:"promotional"`      // 促销商品推荐
	PersonalHistory []*RecommendationResult `json:"personal_history"` // 个人历史推荐
}

// ProductScore 商品评分
type ProductScore struct {
	Product *model.Product
	Score   float64
	Reasons []string
}

// GetCartRecommendations 获取购物车推荐
func (rs *RecommendationService) GetCartRecommendations(userID uint, sessionID string, limit int) (*RecommendationResponse, error) {
	// 获取购物车
	cart, err := rs.getCart(userID, sessionID)
	if err != nil {
		return rs.getDefaultRecommendations(limit), nil
	}

	response := &RecommendationResponse{
		CartBased:       []*RecommendationResult{},
		CategoryBased:   []*RecommendationResult{},
		BrandBased:      []*RecommendationResult{},
		PriceBased:      []*RecommendationResult{},
		Complementary:   []*RecommendationResult{},
		Promotional:     []*RecommendationResult{},
		PersonalHistory: []*RecommendationResult{},
	}

	// 基于购物车内容的推荐
	cartBased, err := rs.getCartBasedRecommendations(cart, limit)
	if err == nil {
		response.CartBased = cartBased
	}

	// 基于分类的推荐
	categoryBased, err := rs.getCategoryBasedRecommendations(cart, limit)
	if err == nil {
		response.CategoryBased = categoryBased
	}

	// 基于品牌的推荐
	brandBased, err := rs.getBrandBasedRecommendations(cart, limit)
	if err == nil {
		response.BrandBased = brandBased
	}

	// 基于价格的推荐
	priceBased, err := rs.getPriceBasedRecommendations(cart, limit)
	if err == nil {
		response.PriceBased = priceBased
	}

	// 互补商品推荐
	complementary, err := rs.getComplementaryRecommendations(cart, limit)
	if err == nil {
		response.Complementary = complementary
	}

	// 促销商品推荐
	promotional, err := rs.getPromotionalRecommendations(cart, limit)
	if err == nil {
		response.Promotional = promotional
	}

	// 个人历史推荐
	if userID > 0 {
		personalHistory, err := rs.getPersonalHistoryRecommendations(userID, cart, limit)
		if err == nil {
			response.PersonalHistory = personalHistory
		}
	}

	return response, nil
}

// getCartBasedRecommendations 基于购物车内容的推荐
func (rs *RecommendationService) getCartBasedRecommendations(cart *model.Cart, limit int) ([]*RecommendationResult, error) {
	if len(cart.Items) == 0 {
		return []*RecommendationResult{}, nil
	}

	// 获取购物车中的商品ID
	productIDs := make([]uint, 0)
	for _, item := range cart.Items {
		if item.Status == model.CartItemStatusNormal {
			productIDs = append(productIDs, item.ProductID)
		}
	}

	if len(productIDs) == 0 {
		return []*RecommendationResult{}, nil
	}

	// 查找经常一起购买的商品
	var relatedProducts []*model.Product
	if err := rs.db.Raw(`
		SELECT p.* FROM products p
		WHERE p.id IN (
			SELECT DISTINCT ci2.product_id 
			FROM cart_items ci1
			JOIN cart_items ci2 ON ci1.cart_id = ci2.cart_id AND ci1.product_id != ci2.product_id
			WHERE ci1.product_id IN ? AND ci2.product_id NOT IN ?
			GROUP BY ci2.product_id
			ORDER BY COUNT(*) DESC
			LIMIT ?
		) AND p.status = ?
	`, productIDs, productIDs, limit, model.ProductStatusActive).
		Preload("Images", func(db *gorm.DB) *gorm.DB {
			return db.Where("is_main = ?", true).Order("sort ASC").Limit(1)
		}).
		Find(&relatedProducts).Error; err != nil {
		return []*RecommendationResult{}, err
	}

	result := &RecommendationResult{
		Products:      relatedProducts,
		Reason:        "经常一起购买的商品",
		RecommendType: "cart_based",
		Score:         0.9,
	}

	return []*RecommendationResult{result}, nil
}

// getCategoryBasedRecommendations 基于分类的推荐
func (rs *RecommendationService) getCategoryBasedRecommendations(cart *model.Cart, limit int) ([]*RecommendationResult, error) {
	// 统计购物车中的分类
	categoryCount := make(map[uint]int)
	for _, item := range cart.Items {
		if item.Status == model.CartItemStatusNormal && item.Product != nil {
			categoryCount[item.Product.CategoryID]++
		}
	}

	if len(categoryCount) == 0 {
		return []*RecommendationResult{}, nil
	}

	// 找到最多的分类
	var topCategoryID uint
	maxCount := 0
	for categoryID, count := range categoryCount {
		if count > maxCount {
			maxCount = count
			topCategoryID = categoryID
		}
	}

	// 获取该分类下的热销商品
	var categoryProducts []*model.Product
	if err := rs.db.Where("category_id = ? AND status = ?", topCategoryID, model.ProductStatusActive).
		Where("id NOT IN (SELECT product_id FROM cart_items WHERE cart_id = ?)", cart.ID).
		Preload("Images", func(db *gorm.DB) *gorm.DB {
			return db.Where("is_main = ?", true).Order("sort ASC").Limit(1)
		}).
		Order("sold_count DESC, is_recommend DESC").
		Limit(limit).
		Find(&categoryProducts).Error; err != nil {
		return []*RecommendationResult{}, err
	}

	result := &RecommendationResult{
		Products:      categoryProducts,
		Reason:        "同分类热销商品",
		RecommendType: "category_based",
		Score:         0.8,
	}

	return []*RecommendationResult{result}, nil
}

// getBrandBasedRecommendations 基于品牌的推荐
func (rs *RecommendationService) getBrandBasedRecommendations(cart *model.Cart, limit int) ([]*RecommendationResult, error) {
	// 统计购物车中的品牌
	brandCount := make(map[uint]int)
	for _, item := range cart.Items {
		if item.Status == model.CartItemStatusNormal && item.Product != nil && item.Product.BrandID > 0 {
			brandCount[item.Product.BrandID]++
		}
	}

	if len(brandCount) == 0 {
		return []*RecommendationResult{}, nil
	}

	// 找到最多的品牌
	var topBrandID uint
	maxCount := 0
	for brandID, count := range brandCount {
		if count > maxCount {
			maxCount = count
			topBrandID = brandID
		}
	}

	// 获取该品牌下的其他商品
	var brandProducts []*model.Product
	if err := rs.db.Where("brand_id = ? AND status = ?", topBrandID, model.ProductStatusActive).
		Where("id NOT IN (SELECT product_id FROM cart_items WHERE cart_id = ?)", cart.ID).
		Preload("Images", func(db *gorm.DB) *gorm.DB {
			return db.Where("is_main = ?", true).Order("sort ASC").Limit(1)
		}).
		Order("sold_count DESC, is_recommend DESC").
		Limit(limit).
		Find(&brandProducts).Error; err != nil {
		return []*RecommendationResult{}, err
	}

	result := &RecommendationResult{
		Products:      brandProducts,
		Reason:        "同品牌推荐商品",
		RecommendType: "brand_based",
		Score:         0.7,
	}

	return []*RecommendationResult{result}, nil
}

// getPriceBasedRecommendations 基于价格的推荐
func (rs *RecommendationService) getPriceBasedRecommendations(cart *model.Cart, limit int) ([]*RecommendationResult, error) {
	// 计算购物车平均价格
	totalPrice := decimal.Zero
	itemCount := 0
	for _, item := range cart.Items {
		if item.Status == model.CartItemStatusNormal {
			totalPrice = totalPrice.Add(item.Price)
			itemCount++
		}
	}

	if itemCount == 0 {
		return []*RecommendationResult{}, nil
	}

	avgPrice := totalPrice.Div(decimal.NewFromInt(int64(itemCount)))
	minPrice := avgPrice.Mul(decimal.NewFromFloat(0.8)) // 平均价格的80%
	maxPrice := avgPrice.Mul(decimal.NewFromFloat(1.2)) // 平均价格的120%

	// 获取相似价格区间的商品
	var priceProducts []*model.Product
	if err := rs.db.Where("price BETWEEN ? AND ? AND status = ?", minPrice, maxPrice, model.ProductStatusActive).
		Where("id NOT IN (SELECT product_id FROM cart_items WHERE cart_id = ?)", cart.ID).
		Preload("Images", func(db *gorm.DB) *gorm.DB {
			return db.Where("is_main = ?", true).Order("sort ASC").Limit(1)
		}).
		Order("is_recommend DESC, sold_count DESC").
		Limit(limit).
		Find(&priceProducts).Error; err != nil {
		return []*RecommendationResult{}, err
	}

	result := &RecommendationResult{
		Products:      priceProducts,
		Reason:        fmt.Sprintf("相似价位商品(%.2f-%.2f元)", minPrice.InexactFloat64(), maxPrice.InexactFloat64()),
		RecommendType: "price_based",
		Score:         0.6,
	}

	return []*RecommendationResult{result}, nil
}

// getComplementaryRecommendations 互补商品推荐
func (rs *RecommendationService) getComplementaryRecommendations(cart *model.Cart, limit int) ([]*RecommendationResult, error) {
	// 这里简化处理，实际应该有商品关联表
	// 根据购物车中的商品推荐互补商品

	// 示例：如果购物车中有手机，推荐手机壳、充电器等
	complementaryMap := map[string][]string{
		"手机":  {"手机壳", "充电器", "耳机", "钢化膜"},
		"电脑":  {"鼠标", "键盘", "音响", "摄像头"},
		"相机":  {"镜头", "三脚架", "存储卡", "相机包"},
		"运动鞋": {"运动袜", "鞋垫", "运动服", "运动包"},
	}

	var complementaryKeywords []string
	for _, item := range cart.Items {
		if item.Status == model.CartItemStatusNormal {
			for keyword, complements := range complementaryMap {
				if contains(item.ProductName, keyword) {
					complementaryKeywords = append(complementaryKeywords, complements...)
				}
			}
		}
	}

	if len(complementaryKeywords) == 0 {
		return []*RecommendationResult{}, nil
	}

	// 查找互补商品
	var complementaryProducts []*model.Product
	query := rs.db.Where("status = ?", model.ProductStatusActive).
		Where("id NOT IN (SELECT product_id FROM cart_items WHERE cart_id = ?)", cart.ID)

	// 构建关键词查询
	for i, keyword := range complementaryKeywords {
		if i == 0 {
			query = query.Where("name LIKE ?", "%"+keyword+"%")
		} else {
			query = query.Or("name LIKE ?", "%"+keyword+"%")
		}
	}

	if err := query.Preload("Images", func(db *gorm.DB) *gorm.DB {
		return db.Where("is_main = ?", true).Order("sort ASC").Limit(1)
	}).
		Order("is_recommend DESC, sold_count DESC").
		Limit(limit).
		Find(&complementaryProducts).Error; err != nil {
		return []*RecommendationResult{}, err
	}

	result := &RecommendationResult{
		Products:      complementaryProducts,
		Reason:        "互补商品推荐",
		RecommendType: "complementary",
		Score:         0.8,
	}

	return []*RecommendationResult{result}, nil
}

// getPromotionalRecommendations 促销商品推荐
func (rs *RecommendationService) getPromotionalRecommendations(cart *model.Cart, limit int) ([]*RecommendationResult, error) {
	// 获取正在促销的商品
	var promotionalProducts []*model.Product
	if err := rs.db.Where("status = ? AND (is_hot = ? OR is_recommend = ?)",
		model.ProductStatusActive, true, true).
		Where("id NOT IN (SELECT product_id FROM cart_items WHERE cart_id = ?)", cart.ID).
		Preload("Images", func(db *gorm.DB) *gorm.DB {
			return db.Where("is_main = ?", true).Order("sort ASC").Limit(1)
		}).
		Order("is_hot DESC, is_recommend DESC, sold_count DESC").
		Limit(limit).
		Find(&promotionalProducts).Error; err != nil {
		return []*RecommendationResult{}, err
	}

	result := &RecommendationResult{
		Products:      promotionalProducts,
		Reason:        "热销推荐商品",
		RecommendType: "promotional",
		Score:         0.7,
	}

	return []*RecommendationResult{result}, nil
}

// getPersonalHistoryRecommendations 个人历史推荐
func (rs *RecommendationService) getPersonalHistoryRecommendations(userID uint, cart *model.Cart, limit int) ([]*RecommendationResult, error) {
	// 获取用户历史购买的商品分类和品牌
	var historyProducts []*model.Product

	// 这里简化处理，实际应该查询订单表
	// 暂时基于用户之前的购物车记录
	if err := rs.db.Raw(`
		SELECT DISTINCT p.* FROM products p
		JOIN cart_items ci ON p.id = ci.product_id
		JOIN carts c ON ci.cart_id = c.id
		WHERE c.user_id = ? AND c.id != ? AND p.status = ?
		ORDER BY ci.created_at DESC
		LIMIT ?
	`, userID, cart.ID, model.ProductStatusActive, limit).
		Preload("Images", func(db *gorm.DB) *gorm.DB {
			return db.Where("is_main = ?", true).Order("sort ASC").Limit(1)
		}).
		Find(&historyProducts).Error; err != nil {
		return []*RecommendationResult{}, err
	}

	result := &RecommendationResult{
		Products:      historyProducts,
		Reason:        "基于购买历史推荐",
		RecommendType: "personal_history",
		Score:         0.9,
	}

	return []*RecommendationResult{result}, nil
}

// getDefaultRecommendations 获取默认推荐
func (rs *RecommendationService) getDefaultRecommendations(limit int) *RecommendationResponse {
	// 获取热销商品作为默认推荐
	var hotProducts []*model.Product
	rs.db.Where("status = ? AND is_hot = ?", model.ProductStatusActive, true).
		Preload("Images", func(db *gorm.DB) *gorm.DB {
			return db.Where("is_main = ?", true).Order("sort ASC").Limit(1)
		}).
		Order("sold_count DESC").
		Limit(limit).
		Find(&hotProducts)

	result := &RecommendationResult{
		Products:      hotProducts,
		Reason:        "热销商品推荐",
		RecommendType: "default",
		Score:         0.5,
	}

	return &RecommendationResponse{
		Promotional: []*RecommendationResult{result},
	}
}

// getCart 获取购物车
func (rs *RecommendationService) getCart(userID uint, sessionID string) (*model.Cart, error) {
	var cart model.Cart
	query := rs.db.Where("status = ?", model.CartStatusActive)

	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	} else if sessionID != "" {
		query = query.Where("user_id = 0 AND session_id = ?", sessionID)
	} else {
		return nil, fmt.Errorf("用户ID和会话ID不能同时为空")
	}

	if err := query.Preload("Items.Product").First(&cart).Error; err != nil {
		return nil, err
	}

	return &cart, nil
}

// contains 检查字符串是否包含子字符串
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) &&
			(s[:len(substr)] == substr ||
				s[len(s)-len(substr):] == substr ||
				containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// 全局购物车推荐服务实例
var globalRecommendationService *RecommendationService

// InitGlobalRecommendationService 初始化全局购物车推荐服务
func InitGlobalRecommendationService(db *gorm.DB) {
	globalRecommendationService = NewRecommendationService(db)
}

// GetGlobalRecommendationService 获取全局购物车推荐服务
func GetGlobalRecommendationService() *RecommendationService {
	return globalRecommendationService
}
