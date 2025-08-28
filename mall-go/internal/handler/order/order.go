package order

import (
	"fmt"
	"mall-go/internal/model"
	"mall-go/pkg/logger"
	"mall-go/pkg/response"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Handler struct {
	db *gorm.DB
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{db: db}
}

// List 获取订单列表
// @Summary 获取订单列表
// @Description 分页获取当前用户的订单列表
// @Tags 订单管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param status query string false "订单状态"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /orders [get]
func (h *Handler) List(c *gin.Context) {
	var req model.OrderListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	// 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	// TODO: 从JWT中获取用户ID
	userID := uint(1) // 临时硬编码

	// 构建查询
	query := h.db.Model(&model.Order{}).Where("user_id = ?", userID).Preload("User").Preload("OrderItems.Product")

	// 状态筛选
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	// 统计总数
	var total int64
	query.Count(&total)

	// 分页查询
	var orders []model.Order
	offset := (req.Page - 1) * req.PageSize
	if err := query.Offset(offset).Limit(req.PageSize).Order("created_at DESC").Find(&orders).Error; err != nil {
		logger.Error("查询订单列表失败", zap.Error(err))
		response.ServerError(c, "查询订单列表失败")
		return
	}

	response.SuccessWithPage(c, "查询成功", orders, total, req.Page, req.PageSize)
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
