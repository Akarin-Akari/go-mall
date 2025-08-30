package order

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"mall-go/internal/model"
	"mall-go/pkg/cart"
	"mall-go/pkg/logger"
	"mall-go/pkg/order"
	"mall-go/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/shopspring/decimal"
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

	// 创建订单相关服务
	orderService := order.NewOrderService(db, cartService, calculationService)
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

	// TODO: 从JWT中获取用户ID
	userID := uint(1) // 临时硬编码

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

	// TODO: 从JWT中获取用户ID
	userID := uint(1) // 临时硬编码

	// 查询商品
	var product model.Product
	if err := h.db.First(&product, req.ProductID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.BadRequest(c, "商品不存在")
			return
		}
		response.ServerError(c, "查询商品失败")
		return
	}

	// 检查库存
	if product.Stock < req.Quantity {
		response.BadRequest(c, "商品库存不足")
		return
	}

	// 生成订单号
	orderNo := fmt.Sprintf("ORD%d", time.Now().UnixNano())

	// 计算总金额
	totalAmount := product.Price.Mul(decimal.NewFromInt(int64(req.Quantity)))

	// 开启事务
	tx := h.db.Begin()

	// 创建订单
	order := model.Order{
		OrderNo:       orderNo,
		UserID:        userID,
		TotalAmount:   totalAmount,
		Status:        model.OrderStatusPending,
		PaymentStatus: model.PaymentStatusUnpaid,
	}

	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		logger.Error("创建订单失败", zap.Error(err))
		response.ServerError(c, "创建订单失败")
		return
	}

	// 创建订单项
	orderItem := model.OrderItem{
		OrderID:   order.ID,
		ProductID: req.ProductID,
		Quantity:  req.Quantity,
		Price:     product.Price,
	}

	if err := tx.Create(&orderItem).Error; err != nil {
		tx.Rollback()
		logger.Error("创建订单项失败", zap.Error(err))
		response.ServerError(c, "创建订单项失败")
		return
	}

	// 更新商品库存
	if err := tx.Model(&product).Update("stock", product.Stock-req.Quantity).Error; err != nil {
		tx.Rollback()
		logger.Error("更新商品库存失败", zap.Error(err))
		response.ServerError(c, "更新商品库存失败")
		return
	}

	// 提交事务
	tx.Commit()

	// 重新查询订单（包含关联数据）
	h.db.Preload("User").Preload("OrderItems.Product").First(&order, order.ID)

	response.Success(c, "订单创建成功", order)
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

	// TODO: 从JWT中获取用户ID和角色
	userID := uint(1)  // 临时硬编码
	userRole := "user" // 临时硬编码

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
	if order.UserID != userID && userRole != "admin" {
		response.Forbidden(c, "无权限更新此订单")
		return
	}

	// 更新订单状态
	order.Status = req.Status

	// 如果订单状态变为已支付，更新支付状态
	if req.Status == model.OrderStatusPaid {
		order.PaymentStatus = model.PaymentStatusPaid
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
	return 1 // 临时硬编码，实际应该从JWT获取
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
