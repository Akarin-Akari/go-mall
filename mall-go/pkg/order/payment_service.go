package order

import (
	"encoding/json"
	"fmt"
	"time"

	"mall-go/internal/model"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// PaymentService 订单支付服务
type PaymentService struct {
	db            *gorm.DB
	statusService *StatusService
}

// NewPaymentService 创建订单支付服务
func NewPaymentService(db *gorm.DB, statusService *StatusService) *PaymentService {
	return &PaymentService{
		db:            db,
		statusService: statusService,
	}
}

// PaymentRequest 支付请求
type PaymentRequest struct {
	OrderID       uint   `json:"order_id" binding:"required"`
	PaymentMethod string `json:"payment_method" binding:"required,oneof=alipay wechat balance"`
	ReturnURL     string `json:"return_url"`
	NotifyURL     string `json:"notify_url"`
	ClientIP      string `json:"client_ip"`
}

// PaymentResponse 支付响应
type PaymentResponse struct {
	PaymentNo     string                 `json:"payment_no"`
	PaymentMethod string                 `json:"payment_method"`
	Amount        decimal.Decimal        `json:"amount"`
	Status        string                 `json:"status"`
	PaymentData   map[string]interface{} `json:"payment_data"`
	ExpireTime    *time.Time             `json:"expire_time"`
}

// PaymentNotifyRequest 支付回调请求
type PaymentNotifyRequest struct {
	PaymentNo      string                 `json:"payment_no"`
	ThirdPartyNo   string                 `json:"third_party_no"`
	Status         string                 `json:"status"`
	Amount         decimal.Decimal        `json:"amount"`
	PayTime        *time.Time             `json:"pay_time"`
	ThirdPartyData map[string]interface{} `json:"third_party_data"`
}

// CreatePayment 创建支付
func (ps *PaymentService) CreatePayment(userID uint, req *PaymentRequest) (*PaymentResponse, error) {
	// 开始事务
	tx := ps.db.Begin()
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

	// 检查订单状态
	if !order.CanPay() {
		tx.Rollback()
		return nil, fmt.Errorf("订单状态不允许支付")
	}

	// 检查是否已有待支付的支付记录
	var existingPayment model.OrderPayment
	if err := tx.Where("order_id = ? AND status IN ?", req.OrderID, 
		[]string{model.PaymentStatusPending}).First(&existingPayment).Error; err == nil {
		tx.Rollback()
		return nil, fmt.Errorf("订单已有待支付记录")
	}

	// 创建支付记录
	payment := &model.OrderPayment{
		OrderID:        req.OrderID,
		PaymentNo:      ps.generatePaymentNo(),
		PaymentMethod:  req.PaymentMethod,
		PaymentChannel: ps.getPaymentChannel(req.PaymentMethod),
		Amount:         order.PayableAmount,
		Status:         model.PaymentStatusPending,
		ExpireTime:     order.PayExpireTime,
	}

	if err := tx.Create(payment).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("创建支付记录失败: %v", err)
	}

	tx.Commit()

	// 调用第三方支付接口
	paymentData, err := ps.callThirdPartyPayment(payment, req)
	if err != nil {
		// 更新支付状态为失败
		ps.db.Model(payment).Updates(map[string]interface{}{
			"status": model.PaymentStatusFailed,
			"third_party_data": fmt.Sprintf(`{"error": "%s"}`, err.Error()),
		})
		return nil, fmt.Errorf("调用支付接口失败: %v", err)
	}

	// 更新支付记录的第三方数据
	thirdPartyDataJSON, _ := json.Marshal(paymentData)
	ps.db.Model(payment).Update("third_party_data", string(thirdPartyDataJSON))

	return &PaymentResponse{
		PaymentNo:     payment.PaymentNo,
		PaymentMethod: payment.PaymentMethod,
		Amount:        payment.Amount,
		Status:        payment.Status,
		PaymentData:   paymentData,
		ExpireTime:    payment.ExpireTime,
	}, nil
}

// callThirdPartyPayment 调用第三方支付接口
func (ps *PaymentService) callThirdPartyPayment(payment *model.OrderPayment, req *PaymentRequest) (map[string]interface{}, error) {
	switch payment.PaymentMethod {
	case model.PaymentTypeAlipay:
		return ps.callAlipayAPI(payment, req)
	case model.PaymentTypeWechat:
		return ps.callWechatPayAPI(payment, req)
	case model.PaymentTypeBalance:
		return ps.processBalancePayment(payment, req)
	default:
		return nil, fmt.Errorf("不支持的支付方式: %s", payment.PaymentMethod)
	}
}

// callAlipayAPI 调用支付宝支付接口
func (ps *PaymentService) callAlipayAPI(payment *model.OrderPayment, req *PaymentRequest) (map[string]interface{}, error) {
	// 这里应该调用真实的支付宝SDK
	// 为了演示，返回模拟数据
	
	paymentData := map[string]interface{}{
		"app_id":      "2021000000000000", // 应用ID
		"method":      "alipay.trade.app.pay",
		"charset":     "utf-8",
		"sign_type":   "RSA2",
		"timestamp":   time.Now().Format("2006-01-02 15:04:05"),
		"version":     "1.0",
		"notify_url":  req.NotifyURL,
		"return_url":  req.ReturnURL,
		"biz_content": map[string]interface{}{
			"out_trade_no": payment.PaymentNo,
			"total_amount": payment.Amount.String(),
			"subject":      "商城订单支付",
			"product_code": "QUICK_MSECURITY_PAY",
		},
	}

	// 模拟生成支付宝支付字符串
	paymentData["payment_string"] = "alipay_sdk=alipay-sdk-java-4.22.110.ALL&app_id=2021000000000000&biz_content=%7B%22out_trade_no%22%3A%22" + payment.PaymentNo + "%22%7D"

	return paymentData, nil
}

// callWechatPayAPI 调用微信支付接口
func (ps *PaymentService) callWechatPayAPI(payment *model.OrderPayment, req *PaymentRequest) (map[string]interface{}, error) {
	// 这里应该调用真实的微信支付SDK
	// 为了演示，返回模拟数据
	
	paymentData := map[string]interface{}{
		"appid":            "wx1234567890123456", // 应用ID
		"mch_id":           "1234567890",         // 商户号
		"nonce_str":        ps.generateNonceStr(),
		"body":             "商城订单支付",
		"out_trade_no":     payment.PaymentNo,
		"total_fee":        payment.Amount.Mul(decimal.NewFromInt(100)).IntPart(), // 微信支付金额单位为分
		"spbill_create_ip": req.ClientIP,
		"notify_url":       req.NotifyURL,
		"trade_type":       "APP",
	}

	// 模拟生成预支付交易会话标识
	paymentData["prepay_id"] = "wx" + fmt.Sprintf("%d", time.Now().Unix()) + "1234567890"

	return paymentData, nil
}

// processBalancePayment 处理余额支付
func (ps *PaymentService) processBalancePayment(payment *model.OrderPayment, req *PaymentRequest) (map[string]interface{}, error) {
	// 获取用户余额
	var user model.User
	if err := ps.db.Select("balance").First(&user, payment.Order.UserID).Error; err != nil {
		return nil, fmt.Errorf("获取用户信息失败: %v", err)
	}

	// 检查余额是否足够
	if user.Balance.LessThan(payment.Amount) {
		return nil, fmt.Errorf("余额不足，当前余额：%.2f", user.Balance.InexactFloat64())
	}

	// 扣减余额
	if err := ps.db.Model(&user).UpdateColumn("balance", 
		gorm.Expr("balance - ?", payment.Amount)).Error; err != nil {
		return nil, fmt.Errorf("扣减余额失败: %v", err)
	}

	// 直接标记支付成功
	now := time.Now()
	ps.db.Model(payment).Updates(map[string]interface{}{
		"status":   model.PaymentStatusPaid,
		"pay_time": &now,
	})

	// 更新订单状态
	ps.statusService.UpdateOrderStatus(payment.OrderID, model.OrderStatusPaid, 
		payment.Order.UserID, model.OperatorTypeUser, "余额支付", "余额支付成功")

	return map[string]interface{}{
		"payment_method": "balance",
		"status":         "success",
		"message":        "余额支付成功",
	}, nil
}

// HandlePaymentNotify 处理支付回调
func (ps *PaymentService) HandlePaymentNotify(req *PaymentNotifyRequest) error {
	// 开始事务
	tx := ps.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 获取支付记录
	var payment model.OrderPayment
	if err := tx.Preload("Order").Where("payment_no = ?", req.PaymentNo).First(&payment).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("支付记录不存在")
	}

	// 检查支付状态
	if payment.Status == model.PaymentStatusPaid {
		tx.Rollback()
		return nil // 已经处理过，直接返回成功
	}

	// 验证支付金额
	if !req.Amount.Equal(payment.Amount) {
		tx.Rollback()
		return fmt.Errorf("支付金额不匹配")
	}

	// 更新支付记录
	thirdPartyDataJSON, _ := json.Marshal(req.ThirdPartyData)
	updates := map[string]interface{}{
		"status":           req.Status,
		"third_party_no":   req.ThirdPartyNo,
		"third_party_data": string(thirdPartyDataJSON),
	}

	if req.PayTime != nil {
		updates["pay_time"] = req.PayTime
	}

	if err := tx.Model(&payment).Updates(updates).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("更新支付记录失败: %v", err)
	}

	// 如果支付成功，更新订单状态和金额
	if req.Status == model.PaymentStatusPaid {
		orderUpdates := map[string]interface{}{
			"paid_amount": gorm.Expr("paid_amount + ?", req.Amount),
		}

		if req.PayTime != nil {
			orderUpdates["pay_time"] = req.PayTime
		}

		if err := tx.Model(&payment.Order).Updates(orderUpdates).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("更新订单支付金额失败: %v", err)
		}

		// 检查是否完成支付
		var updatedOrder model.Order
		if err := tx.First(&updatedOrder, payment.OrderID).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("获取更新后订单失败: %v", err)
		}

		if updatedOrder.PaidAmount.GreaterThanOrEqual(updatedOrder.PayableAmount) {
			// 更新订单状态为已支付
			if err := ps.statusService.UpdateOrderStatus(payment.OrderID, model.OrderStatusPaid, 
				updatedOrder.UserID, model.OperatorTypeSystem, "支付成功", "第三方支付回调"); err != nil {
				tx.Rollback()
				return fmt.Errorf("更新订单状态失败: %v", err)
			}
		}
	}

	tx.Commit()
	return nil
}

// QueryPaymentStatus 查询支付状态
func (ps *PaymentService) QueryPaymentStatus(paymentNo string) (*PaymentResponse, error) {
	var payment model.OrderPayment
	if err := ps.db.Where("payment_no = ?", paymentNo).First(&payment).Error; err != nil {
		return nil, fmt.Errorf("支付记录不存在")
	}

	// 解析第三方数据
	var thirdPartyData map[string]interface{}
	if payment.ThirdPartyData != "" {
		json.Unmarshal([]byte(payment.ThirdPartyData), &thirdPartyData)
	}

	return &PaymentResponse{
		PaymentNo:     payment.PaymentNo,
		PaymentMethod: payment.PaymentMethod,
		Amount:        payment.Amount,
		Status:        payment.Status,
		PaymentData:   thirdPartyData,
		ExpireTime:    payment.ExpireTime,
	}, nil
}

// CancelPayment 取消支付
func (ps *PaymentService) CancelPayment(paymentNo string, reason string) error {
	var payment model.OrderPayment
	if err := ps.db.Where("payment_no = ?", paymentNo).First(&payment).Error; err != nil {
		return fmt.Errorf("支付记录不存在")
	}

	if payment.Status != model.PaymentStatusPending {
		return fmt.Errorf("支付状态不允许取消")
	}

	// 更新支付状态
	if err := ps.db.Model(&payment).Updates(map[string]interface{}{
		"status": model.PaymentStatusCancelled,
		"third_party_data": fmt.Sprintf(`{"cancel_reason": "%s"}`, reason),
	}).Error; err != nil {
		return fmt.Errorf("取消支付失败: %v", err)
	}

	return nil
}

// GetPaymentHistory 获取支付历史
func (ps *PaymentService) GetPaymentHistory(orderID uint) ([]model.OrderPayment, error) {
	var payments []model.OrderPayment
	if err := ps.db.Where("order_id = ?", orderID).
		Order("created_at DESC").
		Find(&payments).Error; err != nil {
		return nil, fmt.Errorf("获取支付历史失败: %v", err)
	}

	return payments, nil
}

// generatePaymentNo 生成支付单号
func (ps *PaymentService) generatePaymentNo() string {
	return fmt.Sprintf("PAY%d", time.Now().UnixNano())
}

// getPaymentChannel 获取支付渠道
func (ps *PaymentService) getPaymentChannel(paymentMethod string) string {
	channelMap := map[string]string{
		model.PaymentTypeAlipay:  "支付宝",
		model.PaymentTypeWechat:  "微信支付",
		model.PaymentTypeBalance: "余额支付",
		model.PaymentTypeBank:    "银行卡",
		model.PaymentTypeCOD:     "货到付款",
	}

	if channel, exists := channelMap[paymentMethod]; exists {
		return channel
	}
	return paymentMethod
}

// generateNonceStr 生成随机字符串
func (ps *PaymentService) generateNonceStr() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// RefundPayment 退款
func (ps *PaymentService) RefundPayment(paymentNo string, refundAmount decimal.Decimal, reason string) error {
	// 开始事务
	tx := ps.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 获取支付记录
	var payment model.OrderPayment
	if err := tx.Preload("Order").Where("payment_no = ? AND status = ?", 
		paymentNo, model.PaymentStatusPaid).First(&payment).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("支付记录不存在或状态不正确")
	}

	// 检查退款金额
	if refundAmount.GreaterThan(payment.Amount) {
		tx.Rollback()
		return fmt.Errorf("退款金额不能大于支付金额")
	}

	// 调用第三方退款接口
	if err := ps.callThirdPartyRefund(&payment, refundAmount, reason); err != nil {
		tx.Rollback()
		return fmt.Errorf("调用退款接口失败: %v", err)
	}

	// 更新支付状态
	if err := tx.Model(&payment).Update("status", model.PaymentStatusRefunded).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("更新支付状态失败: %v", err)
	}

	// 更新订单退款信息
	if err := tx.Model(&payment.Order).Updates(map[string]interface{}{
		"refund_amount": gorm.Expr("refund_amount + ?", refundAmount),
		"refund_status": model.RefundStatusCompleted,
		"refund_time":   time.Now(),
	}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("更新订单退款信息失败: %v", err)
	}

	tx.Commit()
	return nil
}

// callThirdPartyRefund 调用第三方退款接口
func (ps *PaymentService) callThirdPartyRefund(payment *model.OrderPayment, refundAmount decimal.Decimal, reason string) error {
	switch payment.PaymentMethod {
	case model.PaymentTypeAlipay:
		return ps.callAlipayRefund(payment, refundAmount, reason)
	case model.PaymentTypeWechat:
		return ps.callWechatRefund(payment, refundAmount, reason)
	case model.PaymentTypeBalance:
		return ps.processBalanceRefund(payment, refundAmount, reason)
	default:
		return fmt.Errorf("不支持的退款方式: %s", payment.PaymentMethod)
	}
}

// callAlipayRefund 调用支付宝退款
func (ps *PaymentService) callAlipayRefund(payment *model.OrderPayment, refundAmount decimal.Decimal, reason string) error {
	// 这里应该调用真实的支付宝退款接口
	// 为了演示，直接返回成功
	return nil
}

// callWechatRefund 调用微信退款
func (ps *PaymentService) callWechatRefund(payment *model.OrderPayment, refundAmount decimal.Decimal, reason string) error {
	// 这里应该调用真实的微信退款接口
	// 为了演示，直接返回成功
	return nil
}

// processBalanceRefund 处理余额退款
func (ps *PaymentService) processBalanceRefund(payment *model.OrderPayment, refundAmount decimal.Decimal, reason string) error {
	// 退款到用户余额
	if err := ps.db.Model(&model.User{}).
		Where("id = ?", payment.Order.UserID).
		UpdateColumn("balance", gorm.Expr("balance + ?", refundAmount)).Error; err != nil {
		return fmt.Errorf("退款到余额失败: %v", err)
	}

	return nil
}

// 全局订单支付服务实例
var globalPaymentService *PaymentService

// InitGlobalPaymentService 初始化全局订单支付服务
func InitGlobalPaymentService(db *gorm.DB, statusService *StatusService) {
	globalPaymentService = NewPaymentService(db, statusService)
}

// GetGlobalPaymentService 获取全局订单支付服务
func GetGlobalPaymentService() *PaymentService {
	return globalPaymentService
}
