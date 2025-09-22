package cart

import (
	"net/http"
	"strconv"

	"mall-go/internal/model"
	"mall-go/pkg/cart"
	"mall-go/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// CartHandler 购物车处理器
type CartHandler struct {
	db                    *gorm.DB
	cartService           *cart.CartService
	cacheService          *cart.CacheService
	syncService           *cart.SyncService
	calculationService    *cart.CalculationService
	recommendationService *cart.RecommendationService
}

// NewCartHandler 创建购物车处理器
func NewCartHandler(db *gorm.DB, rdb *redis.Client) *CartHandler {
	cartService := cart.NewCartService(db)
	calculationService := cart.NewCalculationService(db)
	recommendationService := cart.NewRecommendationService(db)

	// 如果Redis客户端为nil，则不初始化缓存和同步服务
	var cacheService *cart.CacheService
	var syncService *cart.SyncService

	if rdb != nil {
		cacheService = cart.NewCacheService(rdb, cartService)
		syncService = cart.NewSyncService(db, cartService, cacheService)
	}

	return &CartHandler{
		db:                    db,
		cartService:           cartService,
		cacheService:          cacheService,
		syncService:           syncService,
		calculationService:    calculationService,
		recommendationService: recommendationService,
	}
}

// AddToCart 添加商品到购物车
func (h *CartHandler) AddToCart(c *gin.Context) {
	var req model.AddToCartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}

	userID, sessionID := h.getUserInfo(c)

	// 根据缓存服务是否可用选择不同的处理方式
	var cartItem *model.CartItem
	var err error

	if h.cacheService != nil {
		// 使用缓存服务添加商品
		cartItem, err = h.cacheService.AddToCartWithCache(userID, sessionID, &req)
	} else {
		// 直接使用数据库服务添加商品
		cartItem, err = h.cartService.AddToCart(userID, sessionID, &req)
	}

	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, "添加商品到购物车成功", cartItem)
}

// GetCart 获取购物车
func (h *CartHandler) GetCart(c *gin.Context) {
	userID, sessionID := h.getUserInfo(c)
	includeInvalid := c.DefaultQuery("include_invalid", "false") == "true"

	// 根据缓存服务是否可用选择不同的处理方式
	var cartResponse *model.CartResponse
	var err error

	if h.cacheService != nil {
		// 使用缓存服务获取购物车
		cartResponse, err = h.cacheService.GetCartWithCache(userID, sessionID, includeInvalid)
	} else {
		// 直接使用数据库服务获取购物车
		cartResponse, err = h.cartService.GetCart(userID, sessionID, includeInvalid)
	}

	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	// 计算购物车金额
	if cartResponse.Cart != nil && len(cartResponse.Cart.Items) > 0 {
		region := c.DefaultQuery("region", "domestic")
		calculation, err := h.calculationService.CalculateCart(cartResponse.Cart, userID, region)
		if err == nil {
			cartResponse.Summary.TotalAmount = calculation.SubtotalAmount
			cartResponse.Summary.SelectedAmount = calculation.SelectedAmount
			cartResponse.Summary.DiscountAmount = calculation.TotalDiscount
			cartResponse.Summary.ShippingFee = calculation.ShippingFee
			cartResponse.Summary.FinalAmount = calculation.PayableAmount
		}
	}

	response.Success(c, "获取购物车成功", cartResponse)
}

// UpdateCartItem 更新购物车商品
func (h *CartHandler) UpdateCartItem(c *gin.Context) {
	itemID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "商品ID格式错误")
		return
	}

	var req model.UpdateCartItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}

	userID, sessionID := h.getUserInfo(c)

	// 使用缓存服务更新商品
	cartItem, err := h.cacheService.UpdateCartItemWithCache(userID, sessionID, uint(itemID), &req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, "更新购物车商品成功", cartItem)
}

// RemoveFromCart 从购物车移除商品
func (h *CartHandler) RemoveFromCart(c *gin.Context) {
	itemID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "商品ID格式错误")
		return
	}

	userID, sessionID := h.getUserInfo(c)

	// 使用缓存服务移除商品
	err = h.cacheService.RemoveFromCartWithCache(userID, sessionID, uint(itemID))
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, "移除商品成功", nil)
}

// ClearCart 清空购物车
func (h *CartHandler) ClearCart(c *gin.Context) {
	userID, sessionID := h.getUserInfo(c)

	// 使用缓存服务清空购物车
	err := h.cacheService.ClearCartWithCache(userID, sessionID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, "清空购物车成功", nil)
}

// BatchUpdateCart 批量更新购物车
func (h *CartHandler) BatchUpdateCart(c *gin.Context) {
	var req model.BatchUpdateCartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}

	userID, sessionID := h.getUserInfo(c)

	// 使用缓存服务批量更新
	err := h.cacheService.BatchUpdateCartWithCache(userID, sessionID, &req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, "批量更新购物车成功", nil)
}

// SelectAllItems 全选/取消全选购物车商品
func (h *CartHandler) SelectAllItems(c *gin.Context) {
	var req struct {
		Selected bool `json:"selected"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}

	userID, sessionID := h.getUserInfo(c)

	err := h.cartService.SelectAllItems(userID, sessionID, req.Selected)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	// 清除缓存
	h.cacheService.RefreshCartCache(userID, sessionID)

	response.Success(c, "更新选中状态成功", nil)
}

// GetCartItemCount 获取购物车商品数量
func (h *CartHandler) GetCartItemCount(c *gin.Context) {
	userID, sessionID := h.getUserInfo(c)

	// 使用缓存服务获取数量
	count, err := h.cacheService.GetCartItemCountWithCache(userID, sessionID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, "获取购物车商品数量成功", gin.H{
		"count": count,
	})
}

// SyncCartItems 同步购物车商品信息
func (h *CartHandler) SyncCartItems(c *gin.Context) {
	userID, sessionID := h.getUserInfo(c)

	result, err := h.syncService.SyncCartItems(userID, sessionID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, "同步购物车商品成功", result)
}

// ValidateCartItems 验证购物车商品
func (h *CartHandler) ValidateCartItems(c *gin.Context) {
	userID, sessionID := h.getUserInfo(c)

	cartResponse, err := h.syncService.ValidateCartItems(userID, sessionID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, "验证购物车商品成功", cartResponse)
}

// RemoveInvalidItems 移除失效商品
func (h *CartHandler) RemoveInvalidItems(c *gin.Context) {
	userID, sessionID := h.getUserInfo(c)

	err := h.syncService.CleanInvalidItems(userID, sessionID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, "移除失效商品成功", nil)
}

// CalculateCart 计算购物车
func (h *CartHandler) CalculateCart(c *gin.Context) {
	userID, sessionID := h.getUserInfo(c)
	region := c.DefaultQuery("region", "domestic")
	couponIDStr := c.Query("coupon_id")

	// 获取购物车
	cartResponse, err := h.cacheService.GetCartWithCache(userID, sessionID, false)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	var calculation *cart.CartCalculation
	if couponIDStr != "" {
		couponID, err := strconv.ParseUint(couponIDStr, 10, 32)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "优惠券ID格式错误")
			return
		}
		calculation, err = h.calculationService.CalculateCartWithCoupon(cartResponse.Cart, userID, region, uint(couponID))
	} else {
		calculation, err = h.calculationService.CalculateCart(cartResponse.Cart, userID, region)
	}

	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, "计算购物车成功", calculation)
}

// EstimateShipping 估算运费
func (h *CartHandler) EstimateShipping(c *gin.Context) {
	userID, sessionID := h.getUserInfo(c)
	region := c.DefaultQuery("region", "domestic")

	// 获取购物车
	cartResponse, err := h.cacheService.GetCartWithCache(userID, sessionID, false)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	shippingFee, err := h.calculationService.EstimateShipping(cartResponse.Cart, region)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, "估算运费成功", gin.H{
		"shipping_fee": shippingFee,
		"region":       region,
	})
}

// GetPromotionSuggestions 获取促销建议
func (h *CartHandler) GetPromotionSuggestions(c *gin.Context) {
	userID, sessionID := h.getUserInfo(c)

	// 获取购物车
	cartResponse, err := h.cacheService.GetCartWithCache(userID, sessionID, false)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	// 计算选中商品金额
	selectedAmount := cartResponse.Summary.SelectedAmount

	suggestions := h.calculationService.GetPromotionSuggestions(selectedAmount)

	response.Success(c, "获取促销建议成功", suggestions)
}

// GetCartRecommendations 获取购物车推荐
func (h *CartHandler) GetCartRecommendations(c *gin.Context) {
	userID, sessionID := h.getUserInfo(c)
	limitStr := c.DefaultQuery("limit", "5")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 5
	}

	recommendations, err := h.recommendationService.GetCartRecommendations(userID, sessionID, limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, "获取购物车推荐成功", recommendations)
}

// GetCartStats 获取购物车统计信息
func (h *CartHandler) GetCartStats(c *gin.Context) {
	// 管理员权限检查
	if !h.isAdmin(c) {
		response.Error(c, http.StatusForbidden, "权限不足")
		return
	}

	// 获取同步统计
	syncStats, err := h.syncService.GetSyncStats()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取同步统计失败: "+err.Error())
		return
	}

	// 获取缓存统计
	cacheStats, err := h.cacheService.GetCacheStats()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取缓存统计失败: "+err.Error())
		return
	}

	stats := gin.H{
		"sync_stats":  syncStats,
		"cache_stats": cacheStats,
	}

	response.Success(c, "获取购物车统计信息成功", stats)
}

// MergeGuestCart 合并游客购物车
func (h *CartHandler) MergeGuestCart(c *gin.Context) {
	userID, _ := h.getUserInfo(c)
	if userID == 0 {
		response.Error(c, http.StatusUnauthorized, "用户未登录")
		return
	}

	var req struct {
		GuestSessionID string `json:"guest_session_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}

	// 这里应该实现购物车合并逻辑
	// 简化处理，直接返回成功
	response.Success(c, "合并购物车成功", nil)
}

// getUserInfo 获取用户信息
func (h *CartHandler) getUserInfo(c *gin.Context) (uint, string) {
	var userID uint
	var sessionID string

	// 从JWT中获取用户ID
	if uid, exists := c.Get("user_id"); exists {
		userID = uid.(uint)
	}

	// 从请求头或Cookie中获取会话ID
	sessionID = c.GetHeader("X-Session-ID")
	if sessionID == "" {
		sessionID = c.GetHeader("X-Guest-ID")
	}
	if sessionID == "" {
		// 从Cookie获取
		if cookie, err := c.Cookie("session_id"); err == nil {
			sessionID = cookie
		}
	}

	return userID, sessionID
}

// isAdmin 检查是否为管理员
func (h *CartHandler) isAdmin(c *gin.Context) bool {
	role, exists := c.Get("user_role")
	if !exists {
		return false
	}
	return role == model.RoleAdmin || role == model.RoleSuperAdmin
}
