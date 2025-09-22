package order

import (
	"net/http"
	"strconv"

	"mall-go/internal/model"
	"mall-go/pkg/cart"
	"mall-go/pkg/inventory"
	"mall-go/pkg/logger"
	"mall-go/pkg/order"
	"mall-go/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// OrderHandler 订单处理器
type OrderHandler struct {
	db               *gorm.DB
	orderService     *order.OrderService
	statusService    *order.StatusService
	paymentService   *order.PaymentService
	shippingService  *order.ShippingService
	afterSaleService *order.AfterSaleService
	cacheService     *order.CacheService
}

// NewOrderHandler 创建订单处理器
func NewOrderHandler(db *gorm.DB, rdb *redis.Client) *OrderHandler {
	// 创建购物车服务（订单服务依赖）
	cartService := cart.NewCartService(db)
	calculationService := cart.NewCalculationService(db)

	// 创建库存服务（订单服务依赖）
	inventoryService := inventory.NewInventoryService(db, rdb)

	// 创建订单相关服务
	orderService := order.NewOrderService(db, cartService, calculationService, inventoryService)
	statusService := order.NewStatusService(db)
	paymentService := order.NewPaymentService(db, statusService)
	shippingService := order.NewShippingService(db, statusService)
	afterSaleService := order.NewAfterSaleService(db, statusService, paymentService)
	cacheService := order.NewCacheService(rdb, orderService)

	return &OrderHandler{
		db:               db,
		orderService:     orderService,
		statusService:    statusService,
		paymentService:   paymentService,
		shippingService:  shippingService,
		afterSaleService: afterSaleService,
		cacheService:     cacheService,
	}
}

// Handler 保持向后兼容
type Handler = OrderHandler

// NewHandler 保持向后兼容
func NewHandler(db *gorm.DB) *Handler {
	return NewOrderHandler(db, nil)
}

// GetOrderList 获取订单列表
func (h *OrderHandler) GetOrderList(c *gin.Context) {
	var req model.OrderListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}

	// 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	userID := h.getUserID(c)

	// 使用缓存获取订单列表
	orders, total, err := h.cacheService.GetUserOrdersWithCache(userID, req.Status, req.Page, req.PageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	// 构建响应数据
	var orderResponses []*model.OrderResponse
	for _, order := range orders {
		orderResponses = append(orderResponses, &model.OrderResponse{
			Order:      order,
			StatusText: order.GetStatusText(),
			CanCancel:  order.CanCancel(),
			CanPay:     order.CanPay(),
			CanShip:    order.CanShip(),
			CanReceive: order.CanReceive(),
			CanRefund:  order.CanRefund(),
		})
	}

	totalPages := int((total + int64(req.PageSize) - 1) / int64(req.PageSize))

	listResponse := &model.OrderListResponse{
		Orders:     orderResponses,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	}

	response.Success(c, "获取订单列表成功", listResponse)
}

// List 保持向后兼容
func (h *Handler) List(c *gin.Context) {
	h.GetOrderList(c)
}

// Get 获取订单详情
// @Summary 获取订单详情
// @Description 根据订单ID获取订单详细信息
// @Tags 订单管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "订单ID"
// @Success 200 {object} model.Order
// @Failure 404 {object} map[string]interface{}
// @Router /orders/{id} [get]
func (h *Handler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的订单ID")
		return
	}

	userID := h.getUserID(c)
	if userID == 0 {
		response.Error(c, http.StatusUnauthorized, "用户未登录")
		return
	}

	var order model.Order
	if err := h.db.Preload("User").Preload("OrderItems.Product").Where("id = ? AND user_id = ?", id, userID).First(&order).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.NotFound(c, "订单不存在")
			return
		}
		logger.Error("查询订单详情失败", zap.Error(err))
		response.ServerError(c, "查询订单详情失败")
		return
	}

	response.SuccessWithData(c, order)
}

// Create 创建订单
// @Summary 创建订单
// @Description 创建新订单
// @Tags 订单管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param order body model.OrderCreateRequest true "订单信息"
// @Success 200 {object} model.Order
// @Failure 400 {object} map[string]interface{}
// @Router /orders [post]
func (h *Handler) Create(c *gin.Context) {
	var req model.OrderCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	userID := h.getUserID(c)
	if userID == 0 {
		response.Error(c, http.StatusUnauthorized, "用户未登录")
		return
	}

	// 使用订单服务创建订单
	order, err := h.orderService.CreateOrder(userID, &req)
	if err != nil {
		logger.Error("创建订单失败", zap.Error(err))
		response.ServerError(c, "创建订单失败: "+err.Error())
		return
	}

	response.Success(c, "订单创建成功", gin.H{
		"order_id": order.ID,
		"order_no": order.OrderNo,
	})
}

// UpdateStatus 更新订单状态
// @Summary 更新订单状态
// @Description 更新订单状态（仅限管理员或订单所有者）
// @Tags 订单管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "订单ID"
// @Param status body model.OrderUpdateStatusRequest true "订单状态"
// @Success 200 {object} model.Order
// @Failure 400 {object} map[string]interface{}
// @Router /orders/{id}/status [put]
func (h *Handler) UpdateStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的订单ID")
		return
	}

	var req model.OrderUpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	userID := h.getUserID(c)
	if userID == 0 {
		response.Error(c, http.StatusUnauthorized, "用户未登录")
		return
	}

	isAdmin := h.isAdmin(c)

	var order model.Order
	if err := h.db.First(&order, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.NotFound(c, "订单不存在")
			return
		}
		response.ServerError(c, "查询订单失败")
		return
	}

	// 检查权限（只有订单所有者或管理员可以更新状态）
	if order.UserID != userID && !isAdmin {
		response.Forbidden(c, "无权限更新此订单")
		return
	}

	// 更新订单状态
	order.Status = req.Status

	// 如果订单状态变为已支付，更新支付状态
	if req.Status == model.OrderStatusPaid {
		order.PaymentStatus = string(model.PaymentStatusPaid)
	}

	if err := h.db.Save(&order).Error; err != nil {
		logger.Error("更新订单状态失败", zap.Error(err))
		response.ServerError(c, "更新订单状态失败")
		return
	}

	// 重新查询订单（包含关联数据）
	h.db.Preload("User").Preload("OrderItems.Product").First(&order, order.ID)

	response.Success(c, "订单状态更新成功", order)
}

// getUserID 获取用户ID
func (h *OrderHandler) getUserID(c *gin.Context) uint {
	if uid, exists := c.Get("user_id"); exists {
		return uid.(uint)
	}
	return 0 // 返回0表示未认证用户
}

// isAdmin 检查是否为管理员
func (h *OrderHandler) isAdmin(c *gin.Context) bool {
	role, exists := c.Get("user_role")
	if !exists {
		return false
	}
	return role == model.RoleAdmin
}

// CreateOrder 创建订单
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req model.OrderCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}

	userID := h.getUserID(c)
	if userID == 0 {
		response.Error(c, http.StatusUnauthorized, "用户未登录")
		return
	}

	// 获取订单锁
	lockValue, err := h.cacheService.AcquireOrderLock(userID)
	if err != nil {
		response.Error(c, http.StatusTooManyRequests, err.Error())
		return
	}
	defer h.cacheService.ReleaseOrderLock(userID, lockValue)

	// 创建订单
	order, err := h.orderService.CreateOrder(userID, &req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	// 清除相关缓存
	h.cacheService.InvalidateUserOrdersCache(userID)
	h.cacheService.InvalidateOrderStatsCache()

	response.Success(c, "创建订单成功", order)
}

// GetOrder 获取订单详情
func (h *OrderHandler) GetOrder(c *gin.Context) {
	orderID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "订单ID格式错误")
		return
	}

	userID := h.getUserID(c)

	// 使用缓存获取订单
	order, err := h.cacheService.GetOrderWithCache(uint(orderID))
	if err != nil {
		response.Error(c, http.StatusNotFound, "订单不存在")
		return
	}

	// 检查权限（用户只能查看自己的订单）
	if !h.isAdmin(c) && order.UserID != userID {
		response.Error(c, http.StatusForbidden, "无权访问此订单")
		return
	}

	// 构建响应数据
	orderResponse := &model.OrderResponse{
		Order:      order,
		StatusText: order.GetStatusText(),
		CanCancel:  order.CanCancel(),
		CanPay:     order.CanPay(),
		CanShip:    order.CanShip(),
		CanReceive: order.CanReceive(),
		CanRefund:  order.CanRefund(),
	}

	response.Success(c, "获取订单成功", orderResponse)
}

// UpdateOrderStatus 更新订单状态
func (h *OrderHandler) UpdateOrderStatus(c *gin.Context) {
	orderID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "订单ID格式错误")
		return
	}

	var req model.OrderUpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}

	// 检查管理员权限
	if !h.isAdmin(c) {
		response.Error(c, http.StatusForbidden, "权限不足")
		return
	}

	operatorID := h.getUserID(c)

	// 更新订单状态
	err = h.statusService.UpdateOrderStatus(uint(orderID), req.Status, operatorID,
		model.OperatorTypeAdmin, req.Reason, req.Remark)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	// 清除缓存
	h.cacheService.InvalidateOrderCache(uint(orderID))
	h.cacheService.InvalidateOrderStatsCache()

	response.Success(c, "更新订单状态成功", nil)
}

// CancelOrder 取消订单
func (h *OrderHandler) CancelOrder(c *gin.Context) {
	orderID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "订单ID格式错误")
		return
	}

	userID := h.getUserID(c)
	if userID == 0 {
		response.Error(c, http.StatusUnauthorized, "用户未登录")
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	c.ShouldBindJSON(&req)

	// 获取订单锁
	lockValue, err := h.cacheService.AcquireOrderLock(uint(orderID))
	if err != nil {
		response.Error(c, http.StatusTooManyRequests, err.Error())
		return
	}
	defer h.cacheService.ReleaseOrderLock(uint(orderID), lockValue)

	// 更新订单状态为已取消
	err = h.statusService.UpdateOrderStatus(uint(orderID), model.OrderStatusCancelled,
		userID, model.OperatorTypeUser, req.Reason, "用户取消订单")
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	// 清除缓存
	h.cacheService.InvalidateOrderCache(uint(orderID))
	h.cacheService.InvalidateUserOrdersCache(userID)
	h.cacheService.InvalidateOrderStatsCache()

	response.Success(c, "取消订单成功", nil)
}

// GetOrderByNo 根据订单号获取订单
func (h *OrderHandler) GetOrderByNo(c *gin.Context) {
	orderNo := c.Param("orderNo")
	if orderNo == "" {
		response.Error(c, http.StatusBadRequest, "订单号不能为空")
		return
	}

	userID := h.getUserID(c)

	var order model.Order
	if err := h.db.Preload("OrderItems.Product").Where("order_no = ? AND user_id = ?", orderNo, userID).First(&order).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(c, http.StatusNotFound, "订单不存在")
			return
		}
		response.Error(c, http.StatusInternalServerError, "查询订单失败")
		return
	}

	// 构建响应数据
	orderResponse := &model.OrderResponse{
		Order:      &order,
		StatusText: order.GetStatusText(),
		CanCancel:  order.CanCancel(),
		CanPay:     order.CanPay(),
		CanShip:    order.CanShip(),
		CanReceive: order.CanReceive(),
		CanRefund:  order.CanRefund(),
	}

	response.Success(c, "获取订单成功", orderResponse)
}

// CreatePayment 创建支付
func (h *OrderHandler) CreatePayment(c *gin.Context) {
	var req struct {
		OrderID       uint   `json:"order_id" binding:"required"`
		PaymentMethod string `json:"payment_method" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}

	userID := h.getUserID(c)
	if userID == 0 {
		response.Error(c, http.StatusUnauthorized, "用户未登录")
		return
	}

	// TODO: 集成支付服务
	response.Success(c, "支付功能开发中", gin.H{
		"order_id":       req.OrderID,
		"payment_method": req.PaymentMethod,
		"user_id":        userID,
		"message":        "支付功能待实现",
	})
}

// PaymentNotify 支付回调
func (h *OrderHandler) PaymentNotify(c *gin.Context) {
	// TODO: 实现支付回调逻辑
	response.Success(c, "支付回调功能开发中", gin.H{
		"message": "支付回调功能待实现",
	})
}

// GetShippingInfo 获取物流信息
func (h *OrderHandler) GetShippingInfo(c *gin.Context) {
	orderID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "订单ID格式错误")
		return
	}

	userID := h.getUserID(c)

	var order model.Order
	if err := h.db.Where("id = ? AND user_id = ?", orderID, userID).First(&order).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(c, http.StatusNotFound, "订单不存在")
			return
		}
		response.Error(c, http.StatusInternalServerError, "查询订单失败")
		return
	}

	// TODO: 获取实际物流信息
	response.Success(c, "获取物流信息成功", gin.H{
		"order_id":  orderID,
		"status":    order.Status,
		"message":   "物流查询功能待实现",
		"logistics": []string{},
	})
}

// ConfirmReceipt 确认收货
func (h *OrderHandler) ConfirmReceipt(c *gin.Context) {
	orderID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "订单ID格式错误")
		return
	}

	userID := h.getUserID(c)

	// 更新订单状态为已收货
	err = h.statusService.UpdateOrderStatus(uint(orderID), model.OrderStatusDelivered,
		userID, model.OperatorTypeUser, "用户确认收货", "")
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	// 清除缓存
	h.cacheService.InvalidateOrderCache(uint(orderID))
	h.cacheService.InvalidateUserOrdersCache(userID)
	h.cacheService.InvalidateOrderStatsCache()

	response.Success(c, "确认收货成功", nil)
}

// CreateAfterSale 创建售后申请
func (h *OrderHandler) CreateAfterSale(c *gin.Context) {
	var req struct {
		OrderID     uint   `json:"order_id" binding:"required"`
		Type        string `json:"type" binding:"required"` // refund/return
		Reason      string `json:"reason" binding:"required"`
		Description string `json:"description"`
		Amount      string `json:"amount"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}

	userID := h.getUserID(c)
	if userID == 0 {
		response.Error(c, http.StatusUnauthorized, "用户未登录")
		return
	}

	// TODO: 实现售后申请逻辑
	response.Success(c, "售后申请创建成功", gin.H{
		"order_id":    req.OrderID,
		"type":        req.Type,
		"reason":      req.Reason,
		"description": req.Description,
		"user_id":     userID,
		"message":     "售后功能待实现",
	})
}

// GetOrderStatistics 获取订单统计
func (h *OrderHandler) GetOrderStatistics(c *gin.Context) {
	// TODO: 实现订单统计逻辑
	response.Success(c, "获取订单统计成功", gin.H{
		"total_orders":     0,
		"pending_orders":   0,
		"paid_orders":      0,
		"shipped_orders":   0,
		"delivered_orders": 0,
		"cancelled_orders": 0,
		"message":          "订单统计功能待实现",
	})
}

// CreateShipment 创建发货
func (h *OrderHandler) CreateShipment(c *gin.Context) {
	orderID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "订单ID格式错误")
		return
	}

	var req struct {
		TrackingNumber  string `json:"tracking_number" binding:"required"`
		ShippingMethod  string `json:"shipping_method" binding:"required"`
		ShippingCompany string `json:"shipping_company" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}

	operatorID := h.getUserID(c)

	// 更新订单状态为已发货
	err = h.statusService.UpdateOrderStatus(uint(orderID), model.OrderStatusShipped,
		operatorID, model.OperatorTypeAdmin, "管理员发货", req.TrackingNumber)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	// TODO: 保存物流信息到数据库

	response.Success(c, "发货成功", gin.H{
		"order_id":         orderID,
		"tracking_number":  req.TrackingNumber,
		"shipping_method":  req.ShippingMethod,
		"shipping_company": req.ShippingCompany,
	})
}

// HandleAfterSale 处理售后申请
func (h *OrderHandler) HandleAfterSale(c *gin.Context) {
	afterSaleID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "售后申请ID格式错误")
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"` // approved/rejected
		Remark string `json:"remark"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}

	operatorID := h.getUserID(c)

	// TODO: 实现售后申请处理逻辑
	response.Success(c, "售后申请处理成功", gin.H{
		"aftersale_id": afterSaleID,
		"status":       req.Status,
		"remark":       req.Remark,
		"operator_id":  operatorID,
		"message":      "售后处理功能待实现",
	})
}

// GetAfterSaleList 获取售后申请列表
func (h *OrderHandler) GetAfterSaleList(c *gin.Context) {
	var req struct {
		Page     int    `form:"page"`
		PageSize int    `form:"page_size"`
		Status   string `form:"status"`
	}

	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	response.Success(c, "获取售后申请列表成功", gin.H{
		"list":      []interface{}{},
		"total":     0,
		"page":      req.Page,
		"page_size": req.PageSize,
		"message":   "售后列表功能待实现",
	})
}

// GetAfterSaleDetail 获取售后申请详情
func (h *OrderHandler) GetAfterSaleDetail(c *gin.Context) {
	afterSaleID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "售后申请ID格式错误")
		return
	}

	response.Success(c, "获取售后申请详情成功", gin.H{
		"aftersale_id": afterSaleID,
		"message":      "售后详情功能待实现",
	})
}

// GetMerchantOrderList 获取商家订单列表
func (h *OrderHandler) GetMerchantOrderList(c *gin.Context) {
	var req model.OrderListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	merchantID := h.getUserID(c)

	response.Success(c, "获取商家订单列表成功", gin.H{
		"orders":      []interface{}{},
		"total":       0,
		"page":        req.Page,
		"page_size":   req.PageSize,
		"merchant_id": merchantID,
		"message":     "商家订单列表功能待实现",
	})
}

// GetMerchantAfterSaleList 获取商家售后申请列表
func (h *OrderHandler) GetMerchantAfterSaleList(c *gin.Context) {
	var req struct {
		Page     int    `form:"page"`
		PageSize int    `form:"page_size"`
		Status   string `form:"status"`
	}

	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	merchantID := h.getUserID(c)

	response.Success(c, "获取商家售后申请列表成功", gin.H{
		"list":        []interface{}{},
		"total":       0,
		"page":        req.Page,
		"page_size":   req.PageSize,
		"merchant_id": merchantID,
		"message":     "商家售后列表功能待实现",
	})
}

// Webhook相关方法
// AlipayNotify 支付宝回调
func (h *OrderHandler) AlipayNotify(c *gin.Context) {
	response.Success(c, "支付宝回调处理成功", gin.H{
		"message": "支付宝回调功能待实现",
	})
}

// WechatNotify 微信支付回调
func (h *OrderHandler) WechatNotify(c *gin.Context) {
	response.Success(c, "微信支付回调处理成功", gin.H{
		"message": "微信支付回调功能待实现",
	})
}

// TrackShipment 物流跟踪
func (h *OrderHandler) TrackShipment(c *gin.Context) {
	trackingNumber := c.Param("trackingNumber")
	if trackingNumber == "" {
		response.Error(c, http.StatusBadRequest, "物流单号不能为空")
		return
	}

	response.Success(c, "物流跟踪查询成功", gin.H{
		"tracking_number": trackingNumber,
		"status":          "运输中",
		"events":          []interface{}{},
		"message":         "物流跟踪功能待实现",
	})
}

// GetShippingCompanies 获取物流公司列表
func (h *OrderHandler) GetShippingCompanies(c *gin.Context) {
	companies := []gin.H{
		{"code": "SF", "name": "顺丰速运", "enabled": true},
		{"code": "YTO", "name": "圆通快递", "enabled": true},
		{"code": "ZTO", "name": "中通快递", "enabled": true},
		{"code": "STO", "name": "申通快递", "enabled": true},
		{"code": "YD", "name": "韵达快递", "enabled": true},
	}

	response.Success(c, "获取物流公司列表成功", gin.H{
		"companies": companies,
		"message":   "物流公司列表",
	})
}

// Webhook处理方法
// AlipayWebhook 支付宝Webhook
func (h *OrderHandler) AlipayWebhook(c *gin.Context) {
	response.Success(c, "支付宝Webhook处理成功", gin.H{
		"message": "支付宝Webhook功能待实现",
	})
}

// WechatWebhook 微信支付Webhook
func (h *OrderHandler) WechatWebhook(c *gin.Context) {
	response.Success(c, "微信支付Webhook处理成功", gin.H{
		"message": "微信支付Webhook功能待实现",
	})
}

// StripeWebhook Stripe Webhook
func (h *OrderHandler) StripeWebhook(c *gin.Context) {
	response.Success(c, "Stripe Webhook处理成功", gin.H{
		"message": "Stripe Webhook功能待实现",
	})
}

// PaypalWebhook PayPal Webhook
func (h *OrderHandler) PaypalWebhook(c *gin.Context) {
	response.Success(c, "PayPal Webhook处理成功", gin.H{
		"message": "PayPal Webhook功能待实现",
	})
}

// 物流Webhook方法
// SFWebhook 顺丰Webhook
func (h *OrderHandler) SFWebhook(c *gin.Context) {
	response.Success(c, "顺丰Webhook处理成功", gin.H{
		"message": "顺丰Webhook功能待实现",
	})
}

// YTWebhook 圆通Webhook
func (h *OrderHandler) YTWebhook(c *gin.Context) {
	response.Success(c, "圆通Webhook处理成功", gin.H{
		"message": "圆通Webhook功能待实现",
	})
}

// ZTWebhook 中通Webhook
func (h *OrderHandler) ZTWebhook(c *gin.Context) {
	response.Success(c, "中通Webhook处理成功", gin.H{
		"message": "中通Webhook功能待实现",
	})
}

// STWebhook 申通Webhook
func (h *OrderHandler) STWebhook(c *gin.Context) {
	response.Success(c, "申通Webhook处理成功", gin.H{
		"message": "申通Webhook功能待实现",
	})
}

// YDWebhook 韵达Webhook
func (h *OrderHandler) YDWebhook(c *gin.Context) {
	response.Success(c, "韵达Webhook处理成功", gin.H{
		"message": "韵达Webhook功能待实现",
	})
}

// 服务Webhook方法
// SMSWebhook 短信服务Webhook
func (h *OrderHandler) SMSWebhook(c *gin.Context) {
	response.Success(c, "短信服务Webhook处理成功", gin.H{
		"message": "短信服务Webhook功能待实现",
	})
}

// EmailWebhook 邮件服务Webhook
func (h *OrderHandler) EmailWebhook(c *gin.Context) {
	response.Success(c, "邮件服务Webhook处理成功", gin.H{
		"message": "邮件服务Webhook功能待实现",
	})
}

// PushWebhook 推送服务Webhook
func (h *OrderHandler) PushWebhook(c *gin.Context) {
	response.Success(c, "推送服务Webhook处理成功", gin.H{
		"message": "推送服务Webhook功能待实现",
	})
}
