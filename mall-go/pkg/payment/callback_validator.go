package payment

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"mall-go/internal/model"
	"mall-go/pkg/logger"

	"github.com/go-redis/redis/v8"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// CallbackValidator 支付回调验证器
type CallbackValidator struct {
	db  *gorm.DB
	rdb *redis.Client
}

// NewCallbackValidator 创建回调验证器
func NewCallbackValidator(db *gorm.DB, rdb *redis.Client) *CallbackValidator {
	return &CallbackValidator{
		db:  db,
		rdb: rdb,
	}
}

// ValidationResult 验证结果
type ValidationResult struct {
	Valid        bool   `json:"valid"`
	ErrorCode    string `json:"error_code,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
	PaymentID    uint   `json:"payment_id,omitempty"`
}

// AlipayCallbackData 支付宝回调数据
type AlipayCallbackData struct {
	AppID        string `json:"app_id"`
	TradeNo      string `json:"trade_no"`
	OutTradeNo   string `json:"out_trade_no"`
	TradeStatus  string `json:"trade_status"`
	TotalAmount  string `json:"total_amount"`
	BuyerID      string `json:"buyer_id"`
	GmtCreate    string `json:"gmt_create"`
	GmtPayment   string `json:"gmt_payment"`
	Sign         string `json:"sign"`
	SignType     string `json:"sign_type"`
	NotifyTime   string `json:"notify_time"`
	NotifyType   string `json:"notify_type"`
	NotifyID     string `json:"notify_id"`
	Version      string `json:"version"`
	Charset      string `json:"charset"`
}

// WechatCallbackData 微信回调数据
type WechatCallbackData struct {
	AppID         string `json:"appid"`
	MchID         string `json:"mch_id"`
	OutTradeNo    string `json:"out_trade_no"`
	TransactionID string `json:"transaction_id"`
	TradeType     string `json:"trade_type"`
	TradeState    string `json:"trade_state"`
	BankType      string `json:"bank_type"`
	TotalFee      string `json:"total_fee"`
	CashFee       string `json:"cash_fee"`
	TimeEnd       string `json:"time_end"`
	Sign          string `json:"sign"`
	SignType      string `json:"sign_type"`
	Nonce         string `json:"nonce_str"`
}

// ValidateAlipayCallback 验证支付宝回调
func (cv *CallbackValidator) ValidateAlipayCallback(data *AlipayCallbackData, secretKey string) *ValidationResult {
	result := &ValidationResult{}

	// 1. 验证必要参数
	if data.OutTradeNo == "" || data.TradeNo == "" || data.TotalAmount == "" {
		result.ErrorCode = "MISSING_PARAMS"
		result.ErrorMessage = "缺少必要参数"
		return result
	}

	// 2. 验证通知时间（防重放攻击）
	if !cv.validateNotifyTime(data.NotifyTime) {
		result.ErrorCode = "INVALID_TIME"
		result.ErrorMessage = "通知时间无效或已过期"
		return result
	}

	// 3. 验证通知ID唯一性（防重复通知）
	if !cv.validateNotifyID("alipay", data.NotifyID) {
		result.ErrorCode = "DUPLICATE_NOTIFY"
		result.ErrorMessage = "重复的通知"
		return result
	}

	// 4. 验证签名
	if !cv.validateAlipaySign(data, secretKey) {
		result.ErrorCode = "INVALID_SIGN"
		result.ErrorMessage = "签名验证失败"
		return result
	}

	// 5. 验证支付记录存在性
	payment, err := cv.getPaymentByNo(data.OutTradeNo)
	if err != nil {
		result.ErrorCode = "PAYMENT_NOT_FOUND"
		result.ErrorMessage = "支付记录不存在"
		return result
	}

	// 6. 验证金额一致性
	if !cv.validateAmount(payment, data.TotalAmount) {
		result.ErrorCode = "AMOUNT_MISMATCH"
		result.ErrorMessage = "金额不匹配"
		return result
	}

	// 7. 验证支付状态
	if !cv.validatePaymentStatus(payment, data.TradeStatus) {
		result.ErrorCode = "INVALID_STATUS"
		result.ErrorMessage = "支付状态无效"
		return result
	}

	// 8. 验证幂等性（防止重复处理）
	if !cv.validateIdempotency("alipay", data.OutTradeNo, data.TradeNo) {
		result.ErrorCode = "ALREADY_PROCESSED"
		result.ErrorMessage = "该支付已处理"
		return result
	}

	result.Valid = true
	result.PaymentID = payment.ID
	return result
}

// ValidateWechatCallback 验证微信回调
func (cv *CallbackValidator) ValidateWechatCallback(data *WechatCallbackData, secretKey string) *ValidationResult {
	result := &ValidationResult{}

	// 1. 验证必要参数
	if data.OutTradeNo == "" || data.TransactionID == "" || data.TotalFee == "" {
		result.ErrorCode = "MISSING_PARAMS"
		result.ErrorMessage = "缺少必要参数"
		return result
	}

	// 2. 验证签名
	if !cv.validateWechatSign(data, secretKey) {
		result.ErrorCode = "INVALID_SIGN"
		result.ErrorMessage = "签名验证失败"
		return result
	}

	// 3. 验证支付记录存在性
	payment, err := cv.getPaymentByNo(data.OutTradeNo)
	if err != nil {
		result.ErrorCode = "PAYMENT_NOT_FOUND"
		result.ErrorMessage = "支付记录不存在"
		return result
	}

	// 4. 验证金额一致性（微信金额单位是分）
	totalFee, _ := strconv.Atoi(data.TotalFee)
	expectedAmount := decimal.NewFromFloat(float64(totalFee) / 100)
	if !payment.Amount.Equal(expectedAmount) {
		result.ErrorCode = "AMOUNT_MISMATCH"
		result.ErrorMessage = "金额不匹配"
		return result
	}

	// 5. 验证支付状态
	if !cv.validateWechatPaymentStatus(payment, data.TradeState) {
		result.ErrorCode = "INVALID_STATUS"
		result.ErrorMessage = "支付状态无效"
		return result
	}

	// 6. 验证幂等性
	if !cv.validateIdempotency("wechat", data.OutTradeNo, data.TransactionID) {
		result.ErrorCode = "ALREADY_PROCESSED"
		result.ErrorMessage = "该支付已处理"
		return result
	}

	result.Valid = true
	result.PaymentID = payment.ID
	return result
}

// validateNotifyTime 验证通知时间
func (cv *CallbackValidator) validateNotifyTime(notifyTime string) bool {
	if notifyTime == "" {
		return false
	}

	// 解析时间
	t, err := time.Parse("2006-01-02 15:04:05", notifyTime)
	if err != nil {
		return false
	}

	// 检查时间是否在合理范围内（5分钟内）
	now := time.Now()
	if t.After(now) || now.Sub(t) > 5*time.Minute {
		return false
	}

	return true
}

// validateNotifyID 验证通知ID唯一性
func (cv *CallbackValidator) validateNotifyID(platform, notifyID string) bool {
	if notifyID == "" {
		return true // 某些平台可能没有notify_id
	}

	key := fmt.Sprintf("callback_notify:%s:%s", platform, notifyID)
	
	// 检查是否已存在
	exists, err := cv.rdb.Exists(cv.rdb.Context(), key).Result()
	if err != nil {
		logger.Error("检查通知ID失败", zap.Error(err))
		return false
	}

	if exists > 0 {
		return false // 已存在，重复通知
	}

	// 设置标记，过期时间1小时
	cv.rdb.Set(cv.rdb.Context(), key, "1", time.Hour)
	return true
}

// validateAlipaySign 验证支付宝签名
func (cv *CallbackValidator) validateAlipaySign(data *AlipayCallbackData, secretKey string) bool {
	// 构建待签名字符串
	params := map[string]string{
		"app_id":       data.AppID,
		"trade_no":     data.TradeNo,
		"out_trade_no": data.OutTradeNo,
		"trade_status": data.TradeStatus,
		"total_amount": data.TotalAmount,
		"buyer_id":     data.BuyerID,
		"gmt_create":   data.GmtCreate,
		"gmt_payment":  data.GmtPayment,
		"notify_time":  data.NotifyTime,
		"notify_type":  data.NotifyType,
		"notify_id":    data.NotifyID,
		"version":      data.Version,
		"charset":      data.Charset,
	}

	// 排序并构建签名字符串
	signStr := cv.buildSignString(params)
	
	// 计算签名
	var expectedSign string
	if data.SignType == "MD5" {
		expectedSign = cv.md5Sign(signStr + secretKey)
	} else {
		expectedSign = cv.sha256Sign(signStr + secretKey)
	}

	return strings.ToUpper(expectedSign) == strings.ToUpper(data.Sign)
}

// validateWechatSign 验证微信签名
func (cv *CallbackValidator) validateWechatSign(data *WechatCallbackData, secretKey string) bool {
	params := map[string]string{
		"appid":          data.AppID,
		"mch_id":         data.MchID,
		"out_trade_no":   data.OutTradeNo,
		"transaction_id": data.TransactionID,
		"trade_type":     data.TradeType,
		"trade_state":    data.TradeState,
		"bank_type":      data.BankType,
		"total_fee":      data.TotalFee,
		"cash_fee":       data.CashFee,
		"time_end":       data.TimeEnd,
		"nonce_str":      data.Nonce,
	}

	// 构建签名字符串
	signStr := cv.buildSignString(params) + "&key=" + secretKey
	
	// 计算MD5签名
	expectedSign := cv.md5Sign(signStr)
	
	return strings.ToUpper(expectedSign) == strings.ToUpper(data.Sign)
}

// buildSignString 构建签名字符串
func (cv *CallbackValidator) buildSignString(params map[string]string) string {
	var keys []string
	for k, v := range params {
		if v != "" && k != "sign" && k != "sign_type" {
			keys = append(keys, k)
		}
	}
	
	sort.Strings(keys)
	
	var parts []string
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", k, params[k]))
	}
	
	return strings.Join(parts, "&")
}

// md5Sign MD5签名
func (cv *CallbackValidator) md5Sign(data string) string {
	h := md5.New()
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// sha256Sign SHA256签名
func (cv *CallbackValidator) sha256Sign(data string) string {
	h := sha256.New()
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// getPaymentByNo 根据支付单号获取支付记录
func (cv *CallbackValidator) getPaymentByNo(paymentNo string) (*model.Payment, error) {
	var payment model.Payment
	if err := cv.db.Where("payment_no = ?", paymentNo).First(&payment).Error; err != nil {
		return nil, err
	}
	return &payment, nil
}

// validateAmount 验证金额一致性
func (cv *CallbackValidator) validateAmount(payment *model.Payment, amountStr string) bool {
	amount, err := decimal.NewFromString(amountStr)
	if err != nil {
		return false
	}
	return payment.Amount.Equal(amount)
}

// validatePaymentStatus 验证支付状态
func (cv *CallbackValidator) validatePaymentStatus(payment *model.Payment, tradeStatus string) bool {
	// 只有待支付状态的订单才能接收支付成功回调
	if payment.PaymentStatus != model.PaymentStatusPending {
		return false
	}

	// 验证交易状态
	validStatuses := []string{"TRADE_SUCCESS", "TRADE_FINISHED"}
	for _, status := range validStatuses {
		if status == tradeStatus {
			return true
		}
	}
	return false
}

// validateWechatPaymentStatus 验证微信支付状态
func (cv *CallbackValidator) validateWechatPaymentStatus(payment *model.Payment, tradeState string) bool {
	if payment.PaymentStatus != model.PaymentStatusPending {
		return false
	}
	return tradeState == "SUCCESS"
}

// validateIdempotency 验证幂等性
func (cv *CallbackValidator) validateIdempotency(platform, outTradeNo, thirdPartyID string) bool {
	key := fmt.Sprintf("callback_processed:%s:%s:%s", platform, outTradeNo, thirdPartyID)
	
	// 检查是否已处理
	exists, err := cv.rdb.Exists(cv.rdb.Context(), key).Result()
	if err != nil {
		logger.Error("检查幂等性失败", zap.Error(err))
		return false
	}

	if exists > 0 {
		return false // 已处理
	}

	// 设置处理标记，过期时间24小时
	cv.rdb.Set(cv.rdb.Context(), key, "1", 24*time.Hour)
	return true
}

// MarkCallbackProcessed 标记回调已处理
func (cv *CallbackValidator) MarkCallbackProcessed(platform, outTradeNo, thirdPartyID string) {
	key := fmt.Sprintf("callback_processed:%s:%s:%s", platform, outTradeNo, thirdPartyID)
	cv.rdb.Set(cv.rdb.Context(), key, "1", 24*time.Hour)
}

// IsCallbackProcessed 检查回调是否已处理
func (cv *CallbackValidator) IsCallbackProcessed(platform, outTradeNo, thirdPartyID string) bool {
	key := fmt.Sprintf("callback_processed:%s:%s:%s", platform, outTradeNo, thirdPartyID)
	exists, _ := cv.rdb.Exists(cv.rdb.Context(), key).Result()
	return exists > 0
}
