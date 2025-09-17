package payment

import (
	"fmt"
	"time"

	"mall-go/internal/model"
	"mall-go/pkg/logger"
	"mall-go/pkg/payment/alipay"
	"mall-go/pkg/payment/wechat"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Service 支付服务
type Service struct {
	db            *gorm.DB
	configManager *ConfigManager
	alipayClient  *alipay.Client
	wechatClient  *wechat.Client
}

// NewService 创建支付服务
func NewService(db *gorm.DB, config *PaymentConfig) (*Service, error) {
	service := &Service{
		db:            db,
		configManager: NewConfigManager(db, config),
	}

	// TODO: 重新启用支付客户端初始化
	// 初始化支付宝客户端
	// if config.Alipay.Enabled {
	// 	// 将PaymentConfig的AlipayConfig转换为config包的AlipayConfig
	// 	alipayConfig := &config.AlipayConfig{
	// 		AppID:        config.Alipay.AppID,
	// 		PrivateKey:   config.Alipay.PrivateKey,
	// 		PublicKey:    config.Alipay.PublicKey,
	// 		SignType:     config.Alipay.SignType,
	// 		Format:       config.Alipay.Format,
	// 		Charset:      config.Alipay.Charset,
	// 		GatewayURL:   config.Alipay.GatewayURL,
	// 		Timeout:      config.Alipay.Timeout,
	// 	}
	// 	client, err := alipay.NewClient(alipayConfig)
	// 	if err != nil {
	// 		return nil, fmt.Errorf("初始化支付宝客户端失败: %v", err)
	// 	}
	// 	service.alipayClient = client
	// }

	// 初始化微信支付客户端
	// if config.Wechat.Enabled {
	// 	// 将PaymentConfig的WechatConfig转换为config包的WechatConfig
	// 	wechatConfig := &config.WechatConfig{
	// 		AppID:        config.Wechat.AppID,
	// 		MchID:        config.Wechat.MchID,
	// 		APIKey:       config.Wechat.Key,
	// 		SignType:     config.Wechat.SignType,
	// 		GatewayURL:   config.Wechat.GatewayURL,
	// 		Timeout:      config.Wechat.Timeout,
	// 	}
	// 	service.wechatClient = wechat.NewClient(wechatConfig)
	// }

	return service, nil
}

// CreatePayment 创建支付
func (s *Service) CreatePayment(req *model.PaymentCreateRequest) (*model.PaymentCreateResponse, error) {
	logger.Info("创建支付订单",
		zap.Uint("order_id", req.OrderID),
		zap.String("payment_method", string(req.PaymentMethod)),
		zap.String("amount", req.Amount.String()))

	// 验证请求
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// 检查支付方式是否启用
	if !s.configManager.IsMethodEnabled(req.PaymentMethod) {
		return nil, fmt.Errorf("支付方式 %s 未启用", req.PaymentMethod)
	}

	// 验证金额限制
	if err := s.configManager.ValidateAmount(req.PaymentMethod, req.Amount); err != nil {
		return nil, err
	}

	// 查询订单信息
	var order model.Order
	if err := s.db.First(&order, req.OrderID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, model.ErrOrderNotFound
		}
		return nil, fmt.Errorf("查询订单失败: %v", err)
	}

	// 检查订单状态
	if order.PaymentStatus == string(model.PaymentStatusPaid) {
		return nil, model.ErrPaymentAlreadyPaid
	}

	// 检查金额是否匹配
	if !req.Amount.Equal(order.TotalAmount) {
		return nil, model.ErrInvalidAmount
	}

	// 开启事务
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 生成支付单号
	paymentNo := s.generatePaymentNo()

	// 创建支付记录
	payment := &model.Payment{
		PaymentNo:     paymentNo,
		OrderID:       req.OrderID,
		UserID:        order.UserID,
		PaymentType:   model.PaymentTypeOrder,
		PaymentMethod: req.PaymentMethod,
		PaymentStatus: model.PaymentStatusPending,
		Amount:        req.Amount,
		ActualAmount:  req.Amount,
		Currency:      "CNY",
		Subject:       req.Subject,
		Description:   req.Description,
		NotifyURL:     req.NotifyURL,
		ReturnURL:     req.ReturnURL,
		ExpiredAt:     s.calculateExpiredTime(req.ExpiredMinutes),
	}

	if err := tx.Create(payment).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("创建支付记录失败: %v", err)
	}

	// 调用第三方支付
	paymentData, err := s.callThirdPartyPayment(payment)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("调用第三方支付失败: %v", err)
	}

	// 更新支付记录
	payment.PaymentStatus = model.PaymentStatusPaying
	if err := tx.Save(payment).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("更新支付记录失败: %v", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("提交事务失败: %v", err)
	}

	// 记录支付日志
	s.logPaymentAction(payment.ID, "CREATE", "SUCCESS", "支付订单创建成功", "", "")

	return &model.PaymentCreateResponse{
		PaymentID:     payment.ID,
		PaymentNo:     payment.PaymentNo,
		PaymentMethod: payment.PaymentMethod,
		Amount:        payment.Amount,
		PaymentData:   paymentData,
		ExpiredAt:     *payment.ExpiredAt,
		CreatedAt:     payment.CreatedAt,
	}, nil
}

// callThirdPartyPayment 调用第三方支付
func (s *Service) callThirdPartyPayment(payment *model.Payment) (interface{}, error) {
	switch payment.PaymentMethod {
	case model.PaymentMethodAlipay:
		return s.createAlipayPayment(payment)
	case model.PaymentMethodWechat:
		return s.createWechatPayment(payment)
	default:
		return nil, fmt.Errorf("不支持的支付方式: %s", payment.PaymentMethod)
	}
}

// createAlipayPayment 创建支付宝支付
func (s *Service) createAlipayPayment(payment *model.Payment) (interface{}, error) {
	if s.alipayClient == nil {
		return nil, fmt.Errorf("支付宝客户端未初始化")
	}

	req := &alipay.PaymentRequest{
		OutTradeNo:  payment.PaymentNo,
		TotalAmount: payment.Amount,
		Subject:     payment.Subject,
		Body:        payment.Description,
		NotifyURL:   payment.NotifyURL,
		TimeExpire:  payment.ExpiredAt,
	}

	resp, err := s.alipayClient.CreatePayment(req)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"qr_code":     resp.QRCode,
		"payment_url": fmt.Sprintf("https://qr.alipay.com/bax%s", resp.QRCode),
	}, nil
}

// createWechatPayment 创建微信支付
func (s *Service) createWechatPayment(payment *model.Payment) (interface{}, error) {
	if s.wechatClient == nil {
		return nil, fmt.Errorf("微信支付客户端未初始化")
	}

	req := &wechat.PaymentRequest{
		Body:       payment.Subject,
		Detail:     payment.Description,
		OutTradeNo: payment.PaymentNo,
		TotalFee:   payment.Amount,
		NotifyURL:  payment.NotifyURL,
		TimeExpire: payment.ExpiredAt,
	}

	resp, err := s.wechatClient.CreatePayment(req)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"code_url":  resp.CodeURL,
		"prepay_id": resp.PrepayID,
	}, nil
}

// QueryPayment 查询支付状态
func (s *Service) QueryPayment(req *model.PaymentQueryRequest) (*model.PaymentQueryResponse, error) {
	var payment model.Payment
	query := s.db.Model(&model.Payment{})

	if req.PaymentID > 0 {
		query = query.Where("id = ?", req.PaymentID)
	} else if req.PaymentNo != "" {
		query = query.Where("payment_no = ?", req.PaymentNo)
	} else if req.OrderID > 0 {
		query = query.Where("order_id = ?", req.OrderID)
	} else {
		return nil, fmt.Errorf("查询参数不能为空")
	}

	if err := query.First(&payment).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, model.ErrPaymentNotFound
		}
		return nil, fmt.Errorf("查询支付记录失败: %v", err)
	}

	// 如果支付状态为进行中，查询第三方状态
	if payment.PaymentStatus == model.PaymentStatusPaying || payment.PaymentStatus == model.PaymentStatusPending {
		if err := s.syncPaymentStatus(&payment); err != nil {
			logger.Error("同步支付状态失败", zap.Error(err))
		}
	}

	return &model.PaymentQueryResponse{
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
	}, nil
}

// syncPaymentStatus 同步支付状态
func (s *Service) syncPaymentStatus(payment *model.Payment) error {
	switch payment.PaymentMethod {
	case model.PaymentMethodAlipay:
		return s.syncAlipayStatus(payment)
	case model.PaymentMethodWechat:
		return s.syncWechatStatus(payment)
	default:
		return fmt.Errorf("不支持的支付方式: %s", payment.PaymentMethod)
	}
}

// syncAlipayStatus 同步支付宝状态
func (s *Service) syncAlipayStatus(payment *model.Payment) error {
	if s.alipayClient == nil {
		return fmt.Errorf("支付宝客户端未初始化")
	}

	resp, err := s.alipayClient.QueryPayment(payment.PaymentNo)
	if err != nil {
		return err
	}

	if resp.Status != payment.PaymentStatus {
		// 更新支付状态
		payment.PaymentStatus = resp.Status
		payment.ThirdPartyID = resp.TradeNo

		if resp.Status == model.PaymentStatusSuccess {
			now := time.Now()
			payment.PaidAt = &now
		}

		if err := s.db.Save(payment).Error; err != nil {
			return fmt.Errorf("更新支付状态失败: %v", err)
		}

		// 处理支付成功
		if resp.Status == model.PaymentStatusSuccess {
			if err := s.handlePaymentSuccess(payment); err != nil {
				logger.Error("处理支付成功失败", zap.Error(err))
			}
		}
	}

	return nil
}

// syncWechatStatus 同步微信支付状态
func (s *Service) syncWechatStatus(payment *model.Payment) error {
	if s.wechatClient == nil {
		return fmt.Errorf("微信支付客户端未初始化")
	}

	resp, err := s.wechatClient.QueryPayment(payment.PaymentNo)
	if err != nil {
		return err
	}

	if resp.Status != payment.PaymentStatus {
		// 更新支付状态
		payment.PaymentStatus = resp.Status
		payment.ThirdPartyID = resp.TransactionID

		if resp.Status == model.PaymentStatusSuccess {
			paidAt, _ := time.Parse("20060102150405", resp.TimeEnd)
			payment.PaidAt = &paidAt
		}

		if err := s.db.Save(payment).Error; err != nil {
			return fmt.Errorf("更新支付状态失败: %v", err)
		}

		// 处理支付成功
		if resp.Status == model.PaymentStatusSuccess {
			if err := s.handlePaymentSuccess(payment); err != nil {
				logger.Error("处理支付成功失败", zap.Error(err))
			}
		}
	}

	return nil
}

// handlePaymentSuccess 处理支付成功
func (s *Service) handlePaymentSuccess(payment *model.Payment) error {
	logger.Info("处理支付成功",
		zap.Uint("payment_id", payment.ID),
		zap.String("payment_no", payment.PaymentNo))

	// 开启事务
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 更新订单状态
	if err := tx.Model(&model.Order{}).Where("id = ?", payment.OrderID).Updates(map[string]interface{}{
		"status":         model.OrderStatusPaid,
		"payment_status": model.PaymentStatusPaid,
		"payment_method": payment.PaymentMethod,
	}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("更新订单状态失败: %v", err)
	}

	// 记录支付日志
	s.logPaymentAction(payment.ID, "SUCCESS", "SUCCESS", "支付成功", "", "")

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}

	return nil
}

// generatePaymentNo 生成支付单号
func (s *Service) generatePaymentNo() string {
	return fmt.Sprintf("PAY%d", time.Now().UnixNano())
}

// calculateExpiredTime 计算过期时间
func (s *Service) calculateExpiredTime(minutes int) *time.Time {
	if minutes <= 0 {
		minutes = 30 // 默认30分钟
	}
	expiredAt := time.Now().Add(time.Duration(minutes) * time.Minute)
	return &expiredAt
}

// ProcessCallback 处理支付回调
func (s *Service) ProcessCallback(method model.PaymentMethod, data []byte) error {
	logger.Info("处理支付回调",
		zap.String("payment_method", string(method)),
		zap.Int("data_size", len(data)))

	switch method {
	case model.PaymentMethodAlipay:
		return s.processAlipayCallback(data)
	case model.PaymentMethodWechat:
		return s.processWechatCallback(data)
	default:
		return fmt.Errorf("不支持的支付方式: %s", method)
	}
}

// processAlipayCallback 处理支付宝回调
func (s *Service) processAlipayCallback(data []byte) error {
	if s.alipayClient == nil {
		return fmt.Errorf("支付宝客户端未初始化")
	}

	// 解析回调参数
	params := make(map[string]string)
	// 这里需要根据实际的回调数据格式解析
	// 简化处理，实际应该解析POST参数

	// 验证签名
	if err := s.alipayClient.VerifyCallback(params); err != nil {
		return fmt.Errorf("支付宝回调签名验证失败: %v", err)
	}

	// 查询支付记录
	outTradeNo := params["out_trade_no"]
	var payment model.Payment
	if err := s.db.Where("payment_no = ?", outTradeNo).First(&payment).Error; err != nil {
		return fmt.Errorf("查询支付记录失败: %v", err)
	}

	// 更新支付状态
	tradeStatus := params["trade_status"]
	if tradeStatus == "TRADE_SUCCESS" || tradeStatus == "TRADE_FINISHED" {
		payment.PaymentStatus = model.PaymentStatusSuccess
		payment.ThirdPartyID = params["trade_no"]
		now := time.Now()
		payment.PaidAt = &now

		if err := s.db.Save(&payment).Error; err != nil {
			return fmt.Errorf("更新支付状态失败: %v", err)
		}

		// 处理支付成功
		return s.handlePaymentSuccess(&payment)
	}

	return nil
}

// processWechatCallback 处理微信支付回调
func (s *Service) processWechatCallback(data []byte) error {
	if s.wechatClient == nil {
		return fmt.Errorf("微信支付客户端未初始化")
	}

	// 解析回调数据
	callback, err := s.wechatClient.ParseCallback(data)
	if err != nil {
		return fmt.Errorf("解析微信回调数据失败: %v", err)
	}

	// 验证回调数据
	if err := callback.Validate(); err != nil {
		return fmt.Errorf("微信回调数据验证失败: %v", err)
	}

	// 验证签名
	params := map[string]string{
		"return_code":    callback.ReturnCode,
		"result_code":    callback.ResultCode,
		"out_trade_no":   callback.OutTradeNo,
		"transaction_id": callback.TransactionID,
		"total_fee":      callback.TotalFee,
		"time_end":       callback.TimeEnd,
	}

	if err := s.wechatClient.VerifyCallback(params); err != nil {
		return fmt.Errorf("微信回调签名验证失败: %v", err)
	}

	// 查询支付记录
	var payment model.Payment
	if err := s.db.Where("payment_no = ?", callback.OutTradeNo).First(&payment).Error; err != nil {
		return fmt.Errorf("查询支付记录失败: %v", err)
	}

	// 更新支付状态
	if callback.IsPaymentSuccess() {
		payment.PaymentStatus = model.PaymentStatusSuccess
		payment.ThirdPartyID = callback.TransactionID
		paidAt, _ := callback.GetPaymentTime()
		if paidAt != nil {
			payment.PaidAt = paidAt
		}

		if err := s.db.Save(&payment).Error; err != nil {
			return fmt.Errorf("更新支付状态失败: %v", err)
		}

		// 处理支付成功
		return s.handlePaymentSuccess(&payment)
	}

	return nil
}

// RefundPayment 退款
func (s *Service) RefundPayment(req *model.PaymentRefundRequest) (*model.PaymentRefundResponse, error) {
	logger.Info("处理退款请求",
		zap.Uint("payment_id", req.PaymentID),
		zap.String("refund_amount", req.RefundAmount.String()))

	// 验证请求
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// 查询支付记录
	var payment model.Payment
	if err := s.db.First(&payment, req.PaymentID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, model.ErrPaymentNotFound
		}
		return nil, fmt.Errorf("查询支付记录失败: %v", err)
	}

	// 检查支付状态
	if payment.PaymentStatus != model.PaymentStatusSuccess {
		return nil, fmt.Errorf("只有支付成功的订单才能退款")
	}

	// 检查退款金额
	if req.RefundAmount.GreaterThan(payment.Amount) {
		return nil, fmt.Errorf("退款金额不能大于支付金额")
	}

	// 开启事务
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 生成退款单号
	refundNo := s.generateRefundNo()

	// 创建退款记录
	refund := &model.PaymentRefund{
		RefundNo:     refundNo,
		PaymentID:    payment.ID,
		UserID:       payment.UserID,
		RefundAmount: req.RefundAmount,
		RefundReason: req.RefundReason,
		RefundStatus: model.PaymentStatusPending,
	}

	if err := tx.Create(refund).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("创建退款记录失败: %v", err)
	}

	// 调用第三方退款
	err := s.callThirdPartyRefund(&payment, refund)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("调用第三方退款失败: %v", err)
	}

	// 更新退款状态
	refund.RefundStatus = model.PaymentStatusSuccess
	now := time.Now()
	refund.RefundedAt = &now

	if err := tx.Save(refund).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("更新退款状态失败: %v", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("提交事务失败: %v", err)
	}

	return &model.PaymentRefundResponse{
		RefundID:     refund.ID,
		RefundNo:     refund.RefundNo,
		PaymentID:    refund.PaymentID,
		RefundAmount: refund.RefundAmount,
		RefundStatus: refund.RefundStatus,
		RefundReason: refund.RefundReason,
		CreatedAt:    refund.CreatedAt,
	}, nil
}

// callThirdPartyRefund 调用第三方退款
func (s *Service) callThirdPartyRefund(payment *model.Payment, refund *model.PaymentRefund) error {
	switch payment.PaymentMethod {
	case model.PaymentMethodAlipay:
		return s.refundAlipayPayment(payment, refund)
	case model.PaymentMethodWechat:
		return s.refundWechatPayment(payment, refund)
	default:
		return fmt.Errorf("不支持的支付方式: %s", payment.PaymentMethod)
	}
}

// refundAlipayPayment 支付宝退款
func (s *Service) refundAlipayPayment(payment *model.Payment, refund *model.PaymentRefund) error {
	// 这里应该调用支付宝退款API
	// 简化处理，实际需要实现完整的退款逻辑
	logger.Info("执行支付宝退款", zap.String("payment_no", payment.PaymentNo))
	return nil
}

// refundWechatPayment 微信支付退款
func (s *Service) refundWechatPayment(payment *model.Payment, refund *model.PaymentRefund) error {
	// 这里应该调用微信支付退款API
	// 简化处理，实际需要实现完整的退款逻辑
	logger.Info("执行微信支付退款", zap.String("payment_no", payment.PaymentNo))
	return nil
}

// generateRefundNo 生成退款单号
func (s *Service) generateRefundNo() string {
	return fmt.Sprintf("REF%d", time.Now().UnixNano())
}

// logPaymentAction 记录支付日志
func (s *Service) logPaymentAction(paymentID uint, action, status, message, requestData, responseData string) {
	log := &model.PaymentLog{
		PaymentID:    paymentID,
		Action:       action,
		Status:       status,
		Message:      message,
		RequestData:  requestData,
		ResponseData: responseData,
	}

	if err := s.db.Create(log).Error; err != nil {
		logger.Error("记录支付日志失败", zap.Error(err))
	}
}
