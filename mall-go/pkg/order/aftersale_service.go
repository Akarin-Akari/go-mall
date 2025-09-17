package order

import (
	"encoding/json"
	"fmt"
	"time"

	"mall-go/internal/model"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// AfterSaleService 订单售后服务
type AfterSaleService struct {
	db             *gorm.DB
	statusService  *StatusService
	paymentService *PaymentService
}

// NewAfterSaleService 创建订单售后服务
func NewAfterSaleService(db *gorm.DB, statusService *StatusService, paymentService *PaymentService) *AfterSaleService {
	return &AfterSaleService{
		db:             db,
		statusService:  statusService,
		paymentService: paymentService,
	}
}

// AfterSaleRequest 售后申请请求
type AfterSaleRequest struct {
	OrderID     uint            `json:"order_id" binding:"required"`
	OrderItemID uint            `json:"order_item_id"`
	Type        string          `json:"type" binding:"required,oneof=refund return exchange"`
	Reason      string          `json:"reason" binding:"required"`
	Description string          `json:"description"`
	Images      []string        `json:"images"`
	Amount      decimal.Decimal `json:"amount"`
	Quantity    int             `json:"quantity" binding:"min=1"`
}

// AfterSaleResponse 售后申请响应
type AfterSaleResponse struct {
	AfterSaleNo  string          `json:"after_sale_no"`
	Type         string          `json:"type"`
	Status       string          `json:"status"`
	Amount       decimal.Decimal `json:"amount"`
	Quantity     int             `json:"quantity"`
	Reason       string          `json:"reason"`
	Description  string          `json:"description"`
	Images       []string        `json:"images"`
	CreatedAt    time.Time       `json:"created_at"`
	HandleTime   *time.Time      `json:"handle_time"`
	HandleRemark string          `json:"handle_remark"`
}

// CreateAfterSale 创建售后申请
func (as *AfterSaleService) CreateAfterSale(userID uint, req *AfterSaleRequest) (*AfterSaleResponse, error) {
	// 开始事务
	tx := as.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 获取订单
	var order model.Order
	if err := tx.Where("id = ? AND user_id = ?", req.OrderID, userID).First(&order).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("订单不存在")
	}

	// 检查订单是否可以申请售后
	if !order.CanRefund() {
		tx.Rollback()
		return nil, fmt.Errorf("订单状态不允许申请售后")
	}

	// 如果指定了订单商品项，验证商品项
	if req.OrderItemID > 0 {
		var orderItem model.OrderItem
		if err := tx.Where("id = ? AND order_id = ?", req.OrderItemID, req.OrderID).First(&orderItem).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("订单商品不存在")
		}

		// 检查申请数量
		if req.Quantity > orderItem.Quantity-orderItem.RefundQuantity {
			tx.Rollback()
			return nil, fmt.Errorf("申请数量超过可退货数量")
		}

		// 如果没有指定金额，计算退款金额
		if req.Amount.IsZero() {
			req.Amount = orderItem.Price.Mul(decimal.NewFromInt(int64(req.Quantity)))
		}
	} else {
		// 整单退款
		if req.Amount.IsZero() {
			req.Amount = order.PayableAmount.Sub(order.RefundAmount)
		}
		req.Quantity = 1
	}

	// 检查退款金额
	maxRefundAmount := order.PayableAmount.Sub(order.RefundAmount)
	if req.Amount.GreaterThan(maxRefundAmount) {
		tx.Rollback()
		return nil, fmt.Errorf("退款金额超过可退款金额：%.2f", maxRefundAmount.InexactFloat64())
	}

	// 检查是否已有相同的售后申请
	var existingAfterSale model.OrderAfterSale
	query := tx.Where("order_id = ? AND status IN ?", req.OrderID,
		[]string{model.AfterSaleStatusPending, model.AfterSaleStatusApproved})

	if req.OrderItemID > 0 {
		query = query.Where("order_item_id = ?", req.OrderItemID)
	}

	if err := query.First(&existingAfterSale).Error; err == nil {
		tx.Rollback()
		return nil, fmt.Errorf("已存在待处理的售后申请")
	}

	// 创建售后申请
	imagesJSON, _ := json.Marshal(req.Images)
	afterSale := &model.OrderAfterSale{
		OrderID:     req.OrderID,
		OrderItemID: req.OrderItemID,
		AfterSaleNo: as.generateAfterSaleNo(),
		Type:        req.Type,
		Status:      model.AfterSaleStatusPending,
		ApplyUserID: userID,
		Reason:      req.Reason,
		Description: req.Description,
		Images:      string(imagesJSON),
		Amount:      req.Amount,
		Quantity:    req.Quantity,
	}

	if err := tx.Create(afterSale).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("创建售后申请失败: %v", err)
	}

	// 如果是仅退款且订单已收货，可以直接处理
	if req.Type == model.AfterSaleTypeRefund && order.Status == model.OrderStatusReceived {
		// 自动同意退款申请
		if err := as.handleAfterSaleApproval(tx, afterSale, 0, model.OperatorTypeSystem,
			"自动同意", "仅退款申请自动处理"); err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("处理退款申请失败: %v", err)
		}
	}

	tx.Commit()

	return &AfterSaleResponse{
		AfterSaleNo: afterSale.AfterSaleNo,
		Type:        afterSale.Type,
		Status:      afterSale.Status,
		Amount:      afterSale.Amount,
		Quantity:    afterSale.Quantity,
		Reason:      afterSale.Reason,
		Description: afterSale.Description,
		Images:      req.Images,
		CreatedAt:   afterSale.CreatedAt,
	}, nil
}

// HandleAfterSale 处理售后申请
func (as *AfterSaleService) HandleAfterSale(afterSaleID uint, action string, handleUserID uint, remark string) error {
	// 开始事务
	tx := as.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 获取售后申请
	var afterSale model.OrderAfterSale
	if err := tx.Preload("Order").Preload("OrderItem").First(&afterSale, afterSaleID).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("售后申请不存在")
	}

	// 检查售后申请状态
	if afterSale.Status != model.AfterSaleStatusPending {
		tx.Rollback()
		return fmt.Errorf("售后申请状态不允许处理")
	}

	switch action {
	case "approve":
		if err := as.handleAfterSaleApproval(tx, &afterSale, handleUserID, model.OperatorTypeAdmin, "同意申请", remark); err != nil {
			tx.Rollback()
			return err
		}
	case "reject":
		if err := as.handleAfterSaleRejection(tx, &afterSale, handleUserID, remark); err != nil {
			tx.Rollback()
			return err
		}
	default:
		tx.Rollback()
		return fmt.Errorf("无效的处理动作: %s", action)
	}

	tx.Commit()
	return nil
}

// handleAfterSaleApproval 处理售后申请同意
func (as *AfterSaleService) handleAfterSaleApproval(tx *gorm.DB, afterSale *model.OrderAfterSale, handleUserID uint, operatorType, reason, remark string) error {
	now := time.Now()

	// 更新售后申请状态
	updates := map[string]interface{}{
		"status":         model.AfterSaleStatusApproved,
		"handle_user_id": handleUserID,
		"handle_remark":  remark,
		"handle_time":    &now,
	}

	if err := tx.Model(afterSale).Updates(updates).Error; err != nil {
		return fmt.Errorf("更新售后申请状态失败: %v", err)
	}

	// 根据售后类型处理
	switch afterSale.Type {
	case model.AfterSaleTypeRefund:
		return as.processRefund(tx, afterSale)
	case model.AfterSaleTypeReturn:
		return as.processReturn(tx, afterSale)
	case model.AfterSaleTypeExchange:
		return as.processExchange(tx, afterSale)
	default:
		return fmt.Errorf("不支持的售后类型: %s", afterSale.Type)
	}
}

// handleAfterSaleRejection 处理售后申请拒绝
func (as *AfterSaleService) handleAfterSaleRejection(tx *gorm.DB, afterSale *model.OrderAfterSale, handleUserID uint, remark string) error {
	now := time.Now()

	// 更新售后申请状态
	updates := map[string]interface{}{
		"status":         model.AfterSaleStatusRejected,
		"handle_user_id": handleUserID,
		"handle_remark":  remark,
		"handle_time":    &now,
	}

	if err := tx.Model(afterSale).Updates(updates).Error; err != nil {
		return fmt.Errorf("更新售后申请状态失败: %v", err)
	}

	return nil
}

// processRefund 处理退款
func (as *AfterSaleService) processRefund(tx *gorm.DB, afterSale *model.OrderAfterSale) error {
	// 获取订单的支付记录
	var payment model.OrderPayment
	if err := tx.Where("order_id = ? AND status = ?", afterSale.OrderID, model.PaymentStatusPaid).
		Order("created_at DESC").First(&payment).Error; err != nil {
		return fmt.Errorf("未找到有效的支付记录")
	}

	// 调用退款接口
	if err := as.paymentService.RefundPayment(payment.PaymentNo, afterSale.Amount, afterSale.Reason); err != nil {
		return fmt.Errorf("退款处理失败: %v", err)
	}

	// 更新售后申请状态
	now := time.Now()
	updates := map[string]interface{}{
		"status":        model.AfterSaleStatusCompleted,
		"refund_method": payment.PaymentMethod,
		"refund_amount": afterSale.Amount,
		"refund_time":   &now,
	}

	if err := tx.Model(afterSale).Updates(updates).Error; err != nil {
		return fmt.Errorf("更新售后申请失败: %v", err)
	}

	// 更新订单商品项退款信息
	if afterSale.OrderItemID > 0 {
		if err := tx.Model(&model.OrderItem{}).Where("id = ?", afterSale.OrderItemID).Updates(map[string]interface{}{
			"refund_status":   model.RefundStatusCompleted,
			"refund_quantity": gorm.Expr("refund_quantity + ?", afterSale.Quantity),
			"refund_amount":   gorm.Expr("refund_amount + ?", afterSale.Amount),
		}).Error; err != nil {
			return fmt.Errorf("更新订单商品退款信息失败: %v", err)
		}
	}

	// 更新订单退款信息
	if err := tx.Model(&model.Order{}).Where("id = ?", afterSale.OrderID).Updates(map[string]interface{}{
		"refund_amount": gorm.Expr("refund_amount + ?", afterSale.Amount),
		"refund_status": model.RefundStatusCompleted,
		"refund_time":   &now,
	}).Error; err != nil {
		return fmt.Errorf("更新订单退款信息失败: %v", err)
	}

	// 检查是否需要更新订单状态
	var order model.Order
	if err := tx.First(&order, afterSale.OrderID).Error; err != nil {
		return fmt.Errorf("获取订单信息失败: %v", err)
	}

	// 如果全额退款，更新订单状态为已退款
	if order.RefundAmount.GreaterThanOrEqual(order.PayableAmount) {
		as.statusService.UpdateOrderStatus(afterSale.OrderID, model.OrderStatusRefunded,
			0, model.OperatorTypeSystem, "全额退款", "售后退款完成")
	}

	return nil
}

// processReturn 处理退货
func (as *AfterSaleService) processReturn(tx *gorm.DB, afterSale *model.OrderAfterSale) error {
	// 更新售后申请状态为退货中
	if err := tx.Model(afterSale).Update("status", model.AfterSaleStatusReturning).Error; err != nil {
		return fmt.Errorf("更新售后申请状态失败: %v", err)
	}

	// 这里应该生成退货地址和退货单号
	// 实际实现中需要调用物流接口生成退货单

	return nil
}

// processExchange 处理换货
func (as *AfterSaleService) processExchange(tx *gorm.DB, afterSale *model.OrderAfterSale) error {
	// 换货逻辑比较复杂，需要处理新商品的发货
	// 这里简化处理，直接标记为处理中
	if err := tx.Model(afterSale).Update("status", model.AfterSaleStatusReturning).Error; err != nil {
		return fmt.Errorf("更新售后申请状态失败: %v", err)
	}

	return nil
}

// ConfirmReturn 确认退货
func (as *AfterSaleService) ConfirmReturn(afterSaleID uint, handleUserID uint, remark string) error {
	// 开始事务
	tx := as.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 获取售后申请
	var afterSale model.OrderAfterSale
	if err := tx.Preload("Order").First(&afterSale, afterSaleID).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("售后申请不存在")
	}

	// 检查售后申请状态
	if afterSale.Status != model.AfterSaleStatusReturning {
		tx.Rollback()
		return fmt.Errorf("售后申请状态不正确")
	}

	// 处理退款
	if err := as.processRefund(tx, &afterSale); err != nil {
		tx.Rollback()
		return fmt.Errorf("处理退款失败: %v", err)
	}

	// 恢复库存
	if afterSale.OrderItemID > 0 {
		var orderItem model.OrderItem
		if err := tx.First(&orderItem, afterSale.OrderItemID).Error; err == nil {
			if orderItem.SKUID > 0 {
				// 恢复SKU库存
				tx.Model(&model.ProductSKU{}).Where("id = ?", orderItem.SKUID).
					UpdateColumn("stock", gorm.Expr("stock + ?", afterSale.Quantity))
			} else {
				// 恢复商品库存
				tx.Model(&model.Product{}).Where("id = ?", orderItem.ProductID).
					UpdateColumn("stock", gorm.Expr("stock + ?", afterSale.Quantity))
			}
		}
	}

	tx.Commit()
	return nil
}

// GetAfterSaleList 获取售后申请列表
func (as *AfterSaleService) GetAfterSaleList(userID uint, page, pageSize int, status string) ([]*AfterSaleResponse, int64, error) {
	query := as.db.Model(&model.OrderAfterSale{})

	if userID > 0 {
		query = query.Where("apply_user_id = ?", userID)
	}

	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("获取售后申请总数失败: %v", err)
	}

	// 获取列表
	var afterSales []model.OrderAfterSale
	offset := (page - 1) * pageSize
	if err := query.Preload("Order").Preload("OrderItem").
		Offset(offset).Limit(pageSize).
		Order("created_at DESC").
		Find(&afterSales).Error; err != nil {
		return nil, 0, fmt.Errorf("获取售后申请列表失败: %v", err)
	}

	// 转换为响应格式
	var responses []*AfterSaleResponse
	for _, afterSale := range afterSales {
		var images []string
		if afterSale.Images != "" {
			json.Unmarshal([]byte(afterSale.Images), &images)
		}

		responses = append(responses, &AfterSaleResponse{
			AfterSaleNo:  afterSale.AfterSaleNo,
			Type:         afterSale.Type,
			Status:       afterSale.Status,
			Amount:       afterSale.Amount,
			Quantity:     afterSale.Quantity,
			Reason:       afterSale.Reason,
			Description:  afterSale.Description,
			Images:       images,
			CreatedAt:    afterSale.CreatedAt,
			HandleTime:   afterSale.HandleTime,
			HandleRemark: afterSale.HandleRemark,
		})
	}

	return responses, total, nil
}

// GetAfterSaleDetail 获取售后申请详情
func (as *AfterSaleService) GetAfterSaleDetail(afterSaleNo string) (*AfterSaleResponse, error) {
	var afterSale model.OrderAfterSale
	if err := as.db.Preload("Order").Preload("OrderItem").
		Where("after_sale_no = ?", afterSaleNo).First(&afterSale).Error; err != nil {
		return nil, fmt.Errorf("售后申请不存在")
	}

	var images []string
	if afterSale.Images != "" {
		json.Unmarshal([]byte(afterSale.Images), &images)
	}

	return &AfterSaleResponse{
		AfterSaleNo:  afterSale.AfterSaleNo,
		Type:         afterSale.Type,
		Status:       afterSale.Status,
		Amount:       afterSale.Amount,
		Quantity:     afterSale.Quantity,
		Reason:       afterSale.Reason,
		Description:  afterSale.Description,
		Images:       images,
		CreatedAt:    afterSale.CreatedAt,
		HandleTime:   afterSale.HandleTime,
		HandleRemark: afterSale.HandleRemark,
	}, nil
}

// GetAfterSaleStatistics 获取售后统计信息
func (as *AfterSaleService) GetAfterSaleStatistics() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 统计各状态的售后申请数量
	var statusStats []struct {
		Status string `json:"status"`
		Count  int64  `json:"count"`
	}

	if err := as.db.Model(&model.OrderAfterSale{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Scan(&statusStats).Error; err != nil {
		return nil, fmt.Errorf("获取售后状态统计失败: %v", err)
	}

	stats["status_stats"] = statusStats

	// 统计各类型的售后申请数量
	var typeStats []struct {
		Type  string `json:"type"`
		Count int64  `json:"count"`
	}

	if err := as.db.Model(&model.OrderAfterSale{}).
		Select("type, COUNT(*) as count").
		Group("type").
		Scan(&typeStats).Error; err != nil {
		return nil, fmt.Errorf("获取售后类型统计失败: %v", err)
	}

	stats["type_stats"] = typeStats

	// 统计今日售后申请数量
	var todayAfterSales int64
	today := time.Now().Format("2006-01-02")
	if err := as.db.Model(&model.OrderAfterSale{}).
		Where("DATE(created_at) = ?", today).
		Count(&todayAfterSales).Error; err != nil {
		return nil, fmt.Errorf("获取今日售后统计失败: %v", err)
	}

	stats["today_after_sales"] = todayAfterSales

	return stats, nil
}

// generateAfterSaleNo 生成售后单号
func (as *AfterSaleService) generateAfterSaleNo() string {
	return fmt.Sprintf("AS%d", time.Now().UnixNano())
}

// 全局订单售后服务实例
var globalAfterSaleService *AfterSaleService

// InitGlobalAfterSaleService 初始化全局订单售后服务
func InitGlobalAfterSaleService(db *gorm.DB, statusService *StatusService, paymentService *PaymentService) {
	globalAfterSaleService = NewAfterSaleService(db, statusService, paymentService)
}

// GetGlobalAfterSaleService 获取全局订单售后服务
func GetGlobalAfterSaleService() *AfterSaleService {
	return globalAfterSaleService
}
