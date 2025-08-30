package payment

import (
	"net/http"
	"strconv"

	"mall-go/internal/model"
	"mall-go/pkg/logger"
	"mall-go/pkg/payment"
	"mall-go/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Handler 支付处理器
type Handler struct {
	db             *gorm.DB
	paymentService *payment.Service
}

// NewHandler 创建支付处理器
func NewHandler(db *gorm.DB, paymentService *payment.Service) *Handler {
	return &Handler{
		db:             db,
		paymentService: paymentService,
	}
}

// CreatePayment 创建支付
// @Summary 创建支付订单
// @Description 创建支付订单，支持支付宝、微信支付等多种支付方式
// @Tags 支付管理
// @Accept json
// @Produce json
// @Param request body model.PaymentCreateRequest true "支付创建请求"
// @Success 200 {object} response.Response{data=model.PaymentCreateResponse} "创建成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 401 {object} response.Response "未授权"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /api/v1/payments [post]
// @Security ApiKeyAuth
func (h *Handler) CreatePayment(c *gin.Context) {
	var req model.PaymentCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("绑定支付创建请求失败", zap.Error(err))
		response.Error(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	logger.Info("创建支付请求",
		zap.Uint("order_id", req.OrderID),
		zap.String("payment_method", string(req.PaymentMethod)),
		zap.String("amount", req.Amount.String()))

	// 创建支付
	resp, err := h.paymentService.CreatePayment(&req)
	if err != nil {
		logger.Error("创建支付失败", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "创建支付失败", err.Error())
		return
	}

	response.Success(c, "创建支付成功", resp)
}

// QueryPayment 查询支付状态
// @Summary 查询支付状态
// @Description 根据支付ID、支付单号或订单ID查询支付状态
// @Tags 支付管理
// @Accept json
// @Produce json
// @Param payment_id query uint false "支付ID"
// @Param payment_no query string false "支付单号"
// @Param order_id query uint false "订单ID"
// @Success 200 {object} response.Response{data=model.PaymentQueryResponse} "查询成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "支付记录不存在"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /api/v1/payments/query [get]
// @Security ApiKeyAuth
func (h *Handler) QueryPayment(c *gin.Context) {
	var req model.PaymentQueryRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		logger.Error("绑定支付查询请求失败", zap.Error(err))
		response.Error(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	logger.Info("查询支付状态",
		zap.Uint("payment_id", req.PaymentID),
		zap.String("payment_no", req.PaymentNo),
		zap.Uint("order_id", req.OrderID))

	// 查询支付
	resp, err := h.paymentService.QueryPayment(&req)
	if err != nil {
		if err == model.ErrPaymentNotFound {
			response.Error(c, http.StatusNotFound, "支付记录不存在", err.Error())
			return
		}
		logger.Error("查询支付失败", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "查询支付失败", err.Error())
		return
	}

	response.Success(c, "查询支付成功", resp)
}

// GetPaymentByID 根据ID获取支付详情
// @Summary 获取支付详情
// @Description 根据支付ID获取支付详细信息
// @Tags 支付管理
// @Accept json
// @Produce json
// @Param id path uint true "支付ID"
// @Success 200 {object} response.Response{data=model.PaymentQueryResponse} "获取成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "支付记录不存在"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /api/v1/payments/{id} [get]
// @Security ApiKeyAuth
func (h *Handler) GetPaymentByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的支付ID", err.Error())
		return
	}

	req := &model.PaymentQueryRequest{
		PaymentID: uint(id),
	}

	resp, err := h.paymentService.QueryPayment(req)
	if err != nil {
		if err == model.ErrPaymentNotFound {
			response.Error(c, http.StatusNotFound, "支付记录不存在", err.Error())
			return
		}
		logger.Error("获取支付详情失败", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "获取支付详情失败", err.Error())
		return
	}

	response.Success(c, "获取支付详情成功", resp)
}

// ListPayments 获取支付列表
// @Summary 获取支付列表
// @Description 分页获取支付列表，支持多种筛选条件
// @Tags 支付管理
// @Accept json
// @Produce json
// @Param user_id query uint false "用户ID"
// @Param order_id query uint false "订单ID"
// @Param payment_method query string false "支付方式"
// @Param payment_status query string false "支付状态"
// @Param payment_type query string false "支付类型"
// @Param start_time query string false "开始时间"
// @Param end_time query string false "结束时间"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Success 200 {object} response.Response{data=response.PageResult{list=[]model.PaymentQueryResponse}} "获取成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /api/v1/payments [get]
// @Security ApiKeyAuth
func (h *Handler) ListPayments(c *gin.Context) {
	var req model.PaymentListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		logger.Error("绑定支付列表请求失败", zap.Error(err))
		response.Error(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	// 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	logger.Info("获取支付列表",
		zap.Int("page", req.Page),
		zap.Int("page_size", req.PageSize),
		zap.String("payment_method", string(req.PaymentMethod)),
		zap.String("payment_status", string(req.PaymentStatus)))

	// 构建查询
	query := h.db.Model(&model.Payment{})

	// 添加筛选条件
	if req.UserID > 0 {
		query = query.Where("user_id = ?", req.UserID)
	}
	if req.OrderID > 0 {
		query = query.Where("order_id = ?", req.OrderID)
	}
	if req.PaymentMethod != "" {
		query = query.Where("payment_method = ?", req.PaymentMethod)
	}
	if req.PaymentStatus != "" {
		query = query.Where("payment_status = ?", req.PaymentStatus)
	}
	if req.PaymentType != "" {
		query = query.Where("payment_type = ?", req.PaymentType)
	}
	if req.StartTime != "" {
		query = query.Where("created_at >= ?", req.StartTime)
	}
	if req.EndTime != "" {
		query = query.Where("created_at <= ?", req.EndTime)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		logger.Error("获取支付总数失败", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "获取支付总数失败", err.Error())
		return
	}

	// 分页查询
	var payments []model.Payment
	offset := (req.Page - 1) * req.PageSize
	if err := query.Offset(offset).Limit(req.PageSize).Order("created_at DESC").Find(&payments).Error; err != nil {
		logger.Error("获取支付列表失败", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "获取支付列表失败", err.Error())
		return
	}

	// 转换为响应格式
	var list []model.PaymentQueryResponse
	for _, payment := range payments {
		list = append(list, model.PaymentQueryResponse{
			PaymentID:     payment.ID,
			PaymentNo:     payment.PaymentNo,
			OrderID:       payment.OrderID,
			PaymentMethod: payment.PaymentMethod,
			PaymentStatus: payment.PaymentStatus,
			PaymentType:   payment.PaymentType,
			Amount:        payment.Amount,
			ActualAmount:  payment.ActualAmount,
			Currency:      payment.Currency,
			Subject:       payment.Subject,
			Description:   payment.Description,
			ThirdPartyID:  payment.ThirdPartyID,
			ExpiredAt:     payment.ExpiredAt,
			PaidAt:        payment.PaidAt,
			CreatedAt:     payment.CreatedAt,
			UpdatedAt:     payment.UpdatedAt,
		})
	}

	result := response.PageResult{
		List:     list,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}

	response.Success(c, "获取支付列表成功", result)
}

// RefundPayment 申请退款
// @Summary 申请退款
// @Description 对已支付的订单申请退款
// @Tags 支付管理
// @Accept json
// @Produce json
// @Param request body model.PaymentRefundRequest true "退款申请请求"
// @Success 200 {object} response.Response{data=model.PaymentRefundResponse} "申请成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "支付记录不存在"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /api/v1/payments/refund [post]
// @Security ApiKeyAuth
func (h *Handler) RefundPayment(c *gin.Context) {
	var req model.PaymentRefundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("绑定退款请求失败", zap.Error(err))
		response.Error(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	logger.Info("申请退款",
		zap.Uint("payment_id", req.PaymentID),
		zap.String("refund_amount", req.RefundAmount.String()),
		zap.String("refund_reason", req.RefundReason))

	// 申请退款
	resp, err := h.paymentService.RefundPayment(&req)
	if err != nil {
		if err == model.ErrPaymentNotFound {
			response.Error(c, http.StatusNotFound, "支付记录不存在", err.Error())
			return
		}
		logger.Error("申请退款失败", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "申请退款失败", err.Error())
		return
	}

	response.Success(c, "申请退款成功", resp)
}

// GetPaymentMethods 获取支付方式列表
// @Summary 获取支付方式列表
// @Description 获取所有可用的支付方式配置
// @Tags 支付管理
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=[]model.PaymentConfigResponse} "获取成功"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /api/v1/payments/methods [get]
func (h *Handler) GetPaymentMethods(c *gin.Context) {
	logger.Info("获取支付方式列表")

	// 获取所有支付方式配置
	var configs []model.PaymentConfig
	if err := h.db.Where("is_enabled = ?", true).Order("display_order ASC").Find(&configs).Error; err != nil {
		logger.Error("获取支付方式配置失败", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "获取支付方式配置失败")
		return
	}

	// 转换为响应格式
	var methods []model.PaymentConfigResponse
	for _, config := range configs {
		methods = append(methods, model.PaymentConfigResponse{
			ID:            config.ID,
			PaymentMethod: config.PaymentMethod,
			IsEnabled:     config.IsEnabled,
			DisplayName:   config.DisplayName,
			DisplayOrder:  config.DisplayOrder,
			Icon:          config.Icon,
			MinAmount:     config.MinAmount,
			MaxAmount:     config.MaxAmount,
			CreatedAt:     config.CreatedAt,
			UpdatedAt:     config.UpdatedAt,
		})
	}

	response.Success(c, "获取支付方式列表成功", methods)
}

// GetPaymentStatistics 获取支付统计
// @Summary 获取支付统计
// @Description 获取支付统计数据，支持按时间范围和支付方式筛选
// @Tags 支付管理
// @Accept json
// @Produce json
// @Param user_id query uint false "用户ID"
// @Param payment_method query string false "支付方式"
// @Param payment_status query string false "支付状态"
// @Param start_date query string false "开始日期(YYYY-MM-DD)"
// @Param end_date query string false "结束日期(YYYY-MM-DD)"
// @Param group_by query string false "分组方式(day/month/year)" default(day)
// @Success 200 {object} response.Response{data=model.PaymentStatisticsResponse} "获取成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /api/v1/payments/statistics [get]
// @Security ApiKeyAuth
func (h *Handler) GetPaymentStatistics(c *gin.Context) {
	var req model.PaymentStatisticsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		logger.Error("绑定统计请求失败", zap.Error(err))
		response.Error(c, http.StatusBadRequest, "请求参数错误")
		return
	}

	// 设置默认值
	if req.GroupBy == "" {
		req.GroupBy = "day"
	}

	logger.Info("获取支付统计",
		zap.String("start_date", req.StartDate),
		zap.String("end_date", req.EndDate),
		zap.String("group_by", req.GroupBy))

	// 构建基础查询
	query := h.db.Model(&model.Payment{})

	// 添加筛选条件
	if req.UserID > 0 {
		query = query.Where("user_id = ?", req.UserID)
	}
	if req.PaymentMethod != "" {
		query = query.Where("payment_method = ?", req.PaymentMethod)
	}
	if req.PaymentStatus != "" {
		query = query.Where("payment_status = ?", req.PaymentStatus)
	}
	if req.StartDate != "" {
		query = query.Where("DATE(created_at) >= ?", req.StartDate)
	}
	if req.EndDate != "" {
		query = query.Where("DATE(created_at) <= ?", req.EndDate)
	}

	// 获取总体统计
	var stats struct {
		TotalAmount   float64 `gorm:"column:total_amount"`
		TotalCount    int64   `gorm:"column:total_count"`
		SuccessAmount float64 `gorm:"column:success_amount"`
		SuccessCount  int64   `gorm:"column:success_count"`
		FailedCount   int64   `gorm:"column:failed_count"`
	}

	err := query.Select(`
		COALESCE(SUM(amount), 0) as total_amount,
		COUNT(*) as total_count,
		COALESCE(SUM(CASE WHEN payment_status = 'success' THEN amount ELSE 0 END), 0) as success_amount,
		COUNT(CASE WHEN payment_status = 'success' THEN 1 END) as success_count,
		COUNT(CASE WHEN payment_status = 'failed' THEN 1 END) as failed_count
	`).Scan(&stats).Error

	if err != nil {
		logger.Error("获取支付统计失败", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "获取支付统计失败")
		return
	}

	// 获取退款统计
	var refundStats struct {
		RefundAmount float64 `gorm:"column:refund_amount"`
		RefundCount  int64   `gorm:"column:refund_count"`
	}

	refundQuery := h.db.Model(&model.PaymentRefund{}).
		Joins("JOIN payments ON payment_refunds.payment_id = payments.id")

	if req.UserID > 0 {
		refundQuery = refundQuery.Where("payments.user_id = ?", req.UserID)
	}
	if req.StartDate != "" {
		refundQuery = refundQuery.Where("DATE(payment_refunds.created_at) >= ?", req.StartDate)
	}
	if req.EndDate != "" {
		refundQuery = refundQuery.Where("DATE(payment_refunds.created_at) <= ?", req.EndDate)
	}

	err = refundQuery.Select(`
		COALESCE(SUM(refund_amount), 0) as refund_amount,
		COUNT(*) as refund_count
	`).Where("refund_status = ?", "success").Scan(&refundStats).Error

	if err != nil {
		logger.Error("获取退款统计失败", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "获取退款统计失败")
		return
	}

	// 构建响应
	resp := model.PaymentStatisticsResponse{
		TotalAmount:   decimal.NewFromFloat(stats.TotalAmount),
		TotalCount:    stats.TotalCount,
		SuccessAmount: decimal.NewFromFloat(stats.SuccessAmount),
		SuccessCount:  stats.SuccessCount,
		FailedCount:   stats.FailedCount,
		RefundAmount:  decimal.NewFromFloat(refundStats.RefundAmount),
		RefundCount:   refundStats.RefundCount,
		MethodStats:   make(map[model.PaymentMethod]model.PaymentMethodStat),
	}

	response.Success(c, "获取支付统计成功", resp)
}
