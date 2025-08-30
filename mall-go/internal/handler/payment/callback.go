package payment

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"mall-go/internal/model"
	"mall-go/pkg/logger"
	"mall-go/pkg/order"
	"mall-go/pkg/payment"
	"mall-go/pkg/payment/alipay"
	"mall-go/pkg/payment/wechat"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// CallbackHandler 回调处理器
type CallbackHandler struct {
	db                 *gorm.DB
	paymentService     *payment.Service
	orderStatusManager *order.OrderStatusManager
	callbackValidator  *payment.CallbackValidator
	alipayClient       *alipay.Client
	wechatClient       *wechat.Client
}

// NewCallbackHandler 创建回调处理器
func NewCallbackHandler(db *gorm.DB, paymentService *payment.Service, orderStatusManager *order.OrderStatusManager, callbackValidator *payment.CallbackValidator, alipayClient *alipay.Client, wechatClient *wechat.Client) *CallbackHandler {
	return &CallbackHandler{
		db:                 db,
		paymentService:     paymentService,
		orderStatusManager: orderStatusManager,
		callbackValidator:  callbackValidator,
		alipayClient:       alipayClient,
		wechatClient:       wechatClient,
	}
}

// AlipayCallback 支付宝回调处理
// @Summary 支付宝支付回调
// @Description 处理支付宝异步通知回调
// @Tags 支付回调
// @Accept application/x-www-form-urlencoded
// @Produce plain
// @Success 200 {string} string "success"
// @Failure 400 {string} string "fail"
// @Router /api/v1/payments/callback/alipay [post]
func (h *CallbackHandler) AlipayCallback(c *gin.Context) {
	logger.Info("收到支付宝回调通知")

	// 解析POST参数
	if err := c.Request.ParseForm(); err != nil {
		logger.Error("解析支付宝回调参数失败", zap.Error(err))
		c.String(http.StatusBadRequest, "fail")
		return
	}

	// 转换为map
	params := make(map[string]string)
	for key, values := range c.Request.PostForm {
		if len(values) > 0 {
			params[key] = values[0]
		}
	}

	logger.Info("支付宝回调参数",
		zap.String("out_trade_no", params["out_trade_no"]),
		zap.String("trade_no", params["trade_no"]),
		zap.String("trade_status", params["trade_status"]))

	// 验证签名
	if h.alipayClient != nil {
		if err := h.alipayClient.VerifyCallback(params); err != nil {
			logger.Error("支付宝回调签名验证失败", zap.Error(err))
			c.String(http.StatusBadRequest, "fail")
			return
		}
	}

	// 处理回调数据
	if err := h.processAlipayCallback(params); err != nil {
		logger.Error("处理支付宝回调失败", zap.Error(err))
		c.String(http.StatusBadRequest, "fail")
		return
	}

	// 返回成功响应
	c.String(http.StatusOK, "success")
}

// processAlipayCallback 处理支付宝回调数据
func (h *CallbackHandler) processAlipayCallback(params map[string]string) error {
	// 构建回调数据结构
	callbackData := &payment.AlipayCallbackData{
		AppID:       params["app_id"],
		TradeNo:     params["trade_no"],
		OutTradeNo:  params["out_trade_no"],
		TradeStatus: params["trade_status"],
		TotalAmount: params["total_amount"],
		BuyerID:     params["buyer_id"],
		GmtCreate:   params["gmt_create"],
		GmtPayment:  params["gmt_payment"],
		Sign:        params["sign"],
		SignType:    params["sign_type"],
		NotifyTime:  params["notify_time"],
		NotifyType:  params["notify_type"],
		NotifyID:    params["notify_id"],
		Version:     params["version"],
		Charset:     params["charset"],
	}

	logger.Info("处理支付宝回调",
		zap.String("out_trade_no", callbackData.OutTradeNo),
		zap.String("trade_no", callbackData.TradeNo),
		zap.String("trade_status", callbackData.TradeStatus))

	// 验证回调数据
	secretKey := "your_alipay_secret_key" // 应该从配置中获取
	validationResult := h.callbackValidator.ValidateAlipayCallback(callbackData, secretKey)

	if !validationResult.Valid {
		logger.Error("支付宝回调验证失败",
			zap.String("error_code", validationResult.ErrorCode),
			zap.String("error_message", validationResult.ErrorMessage),
			zap.String("out_trade_no", callbackData.OutTradeNo))
		return fmt.Errorf("回调验证失败: %s", validationResult.ErrorMessage)
	}

	// 获取支付记录
	var payment model.Payment
	if err := h.db.First(&payment, validationResult.PaymentID).Error; err != nil {
		return fmt.Errorf("获取支付记录失败: %v", err)
	}

	// 检查支付状态是否需要更新
	if payment.PaymentStatus == model.PaymentStatusSuccess {
		logger.Info("支付已成功，跳过处理", zap.String("payment_no", callbackData.OutTradeNo))
		return nil
	}

	// 开启事务
	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 更新支付状态
	switch callbackData.TradeStatus {
	case "TRADE_SUCCESS", "TRADE_FINISHED":
		payment.PaymentStatus = model.PaymentStatusSuccess
		payment.ThirdPartyID = callbackData.TradeNo

		// 提交支付记录更新事务
		if err := tx.Commit().Error; err != nil {
			return err
		}

		// 使用订单状态管理器更新订单状态
		statusReq := &order.StatusUpdateRequest{
			OrderID:      payment.OrderID,
			ToStatus:     model.OrderStatusPaid,
			OperatorID:   0, // 系统操作
			OperatorType: "system",
			Reason:       "支付成功",
			Remark:       fmt.Sprintf("支付宝交易号: %s", callbackData.TradeNo),
		}

		result, err := h.orderStatusManager.UpdateOrderStatusWithLock(statusReq)
		if err != nil || !result.Success {
			logger.Error("更新订单状态失败",
				zap.Uint("order_id", payment.OrderID),
				zap.String("error", err.Error()))
			return fmt.Errorf("更新订单状态失败: %v", err)
		}

	case "TRADE_CLOSED":
		payment.PaymentStatus = model.PaymentStatusCancelled
	default:
		payment.PaymentStatus = model.PaymentStatusFailed
	}

	// 保存支付记录
	if err := tx.Save(&payment).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 记录回调日志
	callbackLog := &model.PaymentLog{
		PaymentID:    payment.ID,
		Action:       "CALLBACK",
		Status:       "SUCCESS",
		Message:      "支付宝回调处理成功",
		RequestData:  h.mapToString(params),
		ResponseData: "success",
	}

	if err := tx.Create(callbackLog).Error; err != nil {
		logger.Error("记录回调日志失败", zap.Error(err))
		// 不影响主流程
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return err
	}

	logger.Info("支付宝回调处理成功",
		zap.String("payment_no", callbackData.OutTradeNo),
		zap.String("trade_no", callbackData.TradeNo),
		zap.String("status", string(payment.PaymentStatus)))

	return nil
}

// WechatCallback 微信支付回调处理
// @Summary 微信支付回调
// @Description 处理微信支付异步通知回调
// @Tags 支付回调
// @Accept application/xml
// @Produce application/xml
// @Success 200 {string} string "success xml"
// @Failure 400 {string} string "fail xml"
// @Router /api/v1/payments/callback/wechat [post]
func (h *CallbackHandler) WechatCallback(c *gin.Context) {
	logger.Info("收到微信支付回调通知")

	// 读取请求体
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		logger.Error("读取微信回调数据失败", zap.Error(err))
		c.Data(http.StatusBadRequest, "application/xml", []byte(wechat.BuildFailResponse("读取数据失败")))
		return
	}

	logger.Info("微信回调原始数据", zap.String("body", string(body)))

	// 解析回调数据
	if h.wechatClient == nil {
		logger.Error("微信支付客户端未初始化")
		c.Data(http.StatusBadRequest, "application/xml", []byte(wechat.BuildFailResponse("客户端未初始化")))
		return
	}

	callback, err := h.wechatClient.ParseCallback(body)
	if err != nil {
		logger.Error("解析微信回调数据失败", zap.Error(err))
		c.Data(http.StatusBadRequest, "application/xml", []byte(wechat.BuildFailResponse("解析数据失败")))
		return
	}

	logger.Info("微信回调数据",
		zap.String("out_trade_no", callback.OutTradeNo),
		zap.String("transaction_id", callback.TransactionID),
		zap.String("result_code", callback.ResultCode))

	// 验证回调数据
	if err := callback.Validate(); err != nil {
		logger.Error("微信回调数据验证失败", zap.Error(err))
		c.Data(http.StatusBadRequest, "application/xml", []byte(wechat.BuildFailResponse("数据验证失败")))
		return
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

	if err := h.wechatClient.VerifyCallback(params); err != nil {
		logger.Error("微信回调签名验证失败", zap.Error(err))
		c.Data(http.StatusBadRequest, "application/xml", []byte(wechat.BuildFailResponse("签名验证失败")))
		return
	}

	// 处理回调数据
	if err := h.processWechatCallback(callback); err != nil {
		logger.Error("处理微信回调失败", zap.Error(err))
		c.Data(http.StatusBadRequest, "application/xml", []byte(wechat.BuildFailResponse("处理失败")))
		return
	}

	// 返回成功响应
	c.Data(http.StatusOK, "application/xml", []byte(wechat.BuildSuccessResponse()))
}

// processWechatCallback 处理微信回调数据
func (h *CallbackHandler) processWechatCallback(callback *wechat.CallbackData) error {
	outTradeNo := callback.OutTradeNo
	transactionID := callback.TransactionID

	// 查询支付记录
	var payment model.Payment
	if err := h.db.Where("payment_no = ?", outTradeNo).First(&payment).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.Warn("支付记录不存在", zap.String("out_trade_no", outTradeNo))
			return nil // 返回成功，避免重复通知
		}
		return err
	}

	// 检查支付状态是否需要更新
	if payment.PaymentStatus == model.PaymentStatusSuccess {
		logger.Info("支付已成功，跳过处理", zap.String("payment_no", outTradeNo))
		return nil
	}

	// 开启事务
	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 更新支付状态
	if callback.IsPaymentSuccess() {
		payment.PaymentStatus = model.PaymentStatusSuccess
		payment.ThirdPartyID = transactionID

		// 设置支付时间
		if paidAt, err := callback.GetPaymentTime(); err == nil && paidAt != nil {
			payment.PaidAt = paidAt
		}

		// 提交支付记录更新事务
		if err := tx.Commit().Error; err != nil {
			return err
		}

		// 使用订单状态管理器更新订单状态
		statusReq := &order.StatusUpdateRequest{
			OrderID:      payment.OrderID,
			ToStatus:     model.OrderStatusPaid,
			OperatorID:   0, // 系统操作
			OperatorType: "system",
			Reason:       "支付成功",
			Remark:       fmt.Sprintf("微信交易号: %s", transactionID),
		}

		result, err := h.orderStatusManager.UpdateOrderStatusWithLock(statusReq)
		if err != nil || !result.Success {
			logger.Error("更新订单状态失败",
				zap.Uint("order_id", payment.OrderID),
				zap.String("error", err.Error()))
			return fmt.Errorf("更新订单状态失败: %v", err)
		}
	} else {
		payment.PaymentStatus = model.PaymentStatusFailed
	}

	// 保存支付记录
	if err := tx.Save(&payment).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 记录回调日志
	callbackLog := &model.PaymentLog{
		PaymentID:    payment.ID,
		Action:       "CALLBACK",
		Status:       "SUCCESS",
		Message:      "微信支付回调处理成功",
		RequestData:  string(callback.OutTradeNo), // 简化存储
		ResponseData: wechat.BuildSuccessResponse(),
	}

	if err := tx.Create(callbackLog).Error; err != nil {
		logger.Error("记录回调日志失败", zap.Error(err))
		// 不影响主流程
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return err
	}

	logger.Info("微信支付回调处理成功",
		zap.String("payment_no", outTradeNo),
		zap.String("transaction_id", transactionID),
		zap.String("status", string(payment.PaymentStatus)))

	return nil
}

// mapToString 将map转换为字符串
func (h *CallbackHandler) mapToString(params map[string]string) string {
	var parts []string
	for key, value := range params {
		parts = append(parts, key+"="+url.QueryEscape(value))
	}
	return strings.Join(parts, "&")
}
