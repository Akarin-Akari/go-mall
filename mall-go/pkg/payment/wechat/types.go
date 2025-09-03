package wechat

import (
	"encoding/xml"
	"fmt"
	"strconv"
	"time"

	"mall-go/internal/model"

	"github.com/shopspring/decimal"
)

// PaymentRequest 支付请求
type PaymentRequest struct {
	Body        string          `json:"body"`         // 商品描述
	Detail      string          `json:"detail"`       // 商品详情
	Attach      string          `json:"attach"`       // 附加数据
	OutTradeNo  string          `json:"out_trade_no"` // 商户订单号
	TotalFee    decimal.Decimal `json:"total_fee"`    // 订单总金额(元)
	NotifyURL   string          `json:"notify_url"`   // 异步通知地址
	TimeExpire  *time.Time      `json:"time_expire"`  // 订单过期时间
}

// PaymentResponse 支付响应
type PaymentResponse struct {
	PrepayID  string `json:"prepay_id"`  // 预支付交易会话标识
	CodeURL   string `json:"code_url"`   // 二维码链接
	TradeType string `json:"trade_type"` // 交易类型
	Success   bool   `json:"success"`    // 是否成功
	Message   string `json:"message"`    // 错误信息
}

// QueryResponse 查询响应
type QueryResponse struct {
	OutTradeNo    string                `json:"out_trade_no"`    // 商户订单号
	TransactionID string                `json:"transaction_id"`  // 微信支付订单号
	TradeState    string                `json:"trade_state"`     // 交易状态
	TotalFee      decimal.Decimal       `json:"total_fee"`       // 订单金额
	Status        model.PaymentStatus   `json:"status"`          // 标准化状态
	TimeEnd       string                `json:"time_end"`        // 支付完成时间
	Success       bool                  `json:"success"`         // 是否成功
	Message       string                `json:"message"`         // 错误信息
}

// CallbackData 回调数据
type CallbackData struct {
	XMLName xml.Name `xml:"xml"`
	
	// 基本信息
	ReturnCode string `xml:"return_code"` // 返回状态码
	ReturnMsg  string `xml:"return_msg"`  // 返回信息
	ResultCode string `xml:"result_code"` // 业务结果
	ErrCode    string `xml:"err_code"`    // 错误代码
	ErrCodeDes string `xml:"err_code_des"` // 错误代码描述
	
	// 应用信息
	AppID    string `xml:"appid"`     // 公众账号ID
	MchID    string `xml:"mch_id"`    // 商户号
	NonceStr string `xml:"nonce_str"` // 随机字符串
	Sign     string `xml:"sign"`      // 签名
	SignType string `xml:"sign_type"` // 签名类型
	
	// 交易信息
	OpenID        string `xml:"openid"`         // 用户标识
	IsSubscribe   string `xml:"is_subscribe"`   // 是否关注公众账号
	TradeType     string `xml:"trade_type"`     // 交易类型
	BankType      string `xml:"bank_type"`      // 付款银行
	TotalFee      string `xml:"total_fee"`      // 订单金额(分)
	SettlementTotalFee string `xml:"settlement_total_fee"` // 应结订单金额
	FeeType       string `xml:"fee_type"`       // 货币种类
	CashFee       string `xml:"cash_fee"`       // 现金支付金额
	CashFeeType   string `xml:"cash_fee_type"`  // 现金支付货币类型
	
	// 订单信息
	TransactionID string `xml:"transaction_id"` // 微信支付订单号
	OutTradeNo    string `xml:"out_trade_no"`   // 商户订单号
	Attach        string `xml:"attach"`         // 商家数据包
	TimeEnd       string `xml:"time_end"`       // 支付完成时间
	
	// 优惠信息
	CouponFee   string `xml:"coupon_fee"`   // 代金券金额
	CouponCount string `xml:"coupon_count"` // 代金券使用数量
}

// RefundRequest 退款请求
type RefundRequest struct {
	OutTradeNo    string          `json:"out_trade_no"`    // 商户订单号
	TransactionID string          `json:"transaction_id"`  // 微信订单号
	OutRefundNo   string          `json:"out_refund_no"`   // 商户退款单号
	TotalFee      decimal.Decimal `json:"total_fee"`       // 订单金额
	RefundFee     decimal.Decimal `json:"refund_fee"`      // 退款金额
	RefundDesc    string          `json:"refund_desc"`     // 退款原因
	NotifyURL     string          `json:"notify_url"`      // 退款通知地址
}

// RefundResponse 退款响应
type RefundResponse struct {
	OutTradeNo      string          `json:"out_trade_no"`      // 商户订单号
	TransactionID   string          `json:"transaction_id"`    // 微信订单号
	OutRefundNo     string          `json:"out_refund_no"`     // 商户退款单号
	RefundID        string          `json:"refund_id"`         // 微信退款单号
	RefundFee       decimal.Decimal `json:"refund_fee"`        // 退款金额
	SettlementRefundFee decimal.Decimal `json:"settlement_refund_fee"` // 应结退款金额
	TotalFee        decimal.Decimal `json:"total_fee"`         // 订单金额
	SettlementTotalFee decimal.Decimal `json:"settlement_total_fee"` // 应结订单金额
	CashFee         decimal.Decimal `json:"cash_fee"`          // 现金支付金额
	CashRefundFee   decimal.Decimal `json:"cash_refund_fee"`   // 现金退款金额
	Success         bool            `json:"success"`           // 是否成功
	Message         string          `json:"message"`           // 错误信息
}

// TradeState 交易状态常量
const (
	TradeStateSuccess    = "SUCCESS"    // 支付成功
	TradeStateRefund     = "REFUND"     // 转入退款
	TradeStateNotPay     = "NOTPAY"     // 未支付
	TradeStateClosed     = "CLOSED"     // 已关闭
	TradeStateRevoked    = "REVOKED"    // 已撤销（刷卡支付）
	TradeStateUserPaying = "USERPAYING" // 用户支付中
	TradeStatePayError   = "PAYERROR"   // 支付失败
)

// ReturnCode 返回状态码常量
const (
	ReturnCodeSuccess = "SUCCESS" // 成功
	ReturnCodeFail    = "FAIL"    // 失败
)

// ResultCode 业务结果常量
const (
	ResultCodeSuccess = "SUCCESS" // 成功
	ResultCodeFail    = "FAIL"    // 失败
)

// ToPaymentStatus 转换为标准支付状态
func (cd *CallbackData) ToPaymentStatus() model.PaymentStatus {
	switch cd.GetTradeState() {
	case TradeStateSuccess:
		return model.PaymentStatusSuccess
	case TradeStateRefund:
		return model.PaymentStatusRefunded
	case TradeStateNotPay:
		return model.PaymentStatusPending
	case TradeStateClosed, TradeStateRevoked:
		return model.PaymentStatusCancelled
	case TradeStateUserPaying:
		return model.PaymentStatusPaying
	case TradeStatePayError:
		return model.PaymentStatusFailed
	default:
		return model.PaymentStatusFailed
	}
}

// GetTradeState 获取交易状态（从回调数据中推断）
func (cd *CallbackData) GetTradeState() string {
	if cd.ResultCode == ResultCodeSuccess {
		return TradeStateSuccess
	}
	return TradeStatePayError
}

// IsPaymentSuccess 是否支付成功
func (cd *CallbackData) IsPaymentSuccess() bool {
	return cd.ReturnCode == ReturnCodeSuccess && cd.ResultCode == ResultCodeSuccess
}

// GetTotalAmount 获取订单金额（转换为元）
func (cd *CallbackData) GetTotalAmount() decimal.Decimal {
	totalFee, _ := strconv.ParseInt(cd.TotalFee, 10, 64)
	return decimal.NewFromInt(totalFee).Div(decimal.NewFromInt(100))
}

// GetPaymentTime 获取支付时间
func (cd *CallbackData) GetPaymentTime() (*time.Time, error) {
	if cd.TimeEnd == "" {
		return nil, nil
	}
	
	// 解析时间格式: 20060102150405
	t, err := time.Parse("20060102150405", cd.TimeEnd)
	if err != nil {
		return nil, err
	}
	
	return &t, nil
}

// Validate 验证回调数据
func (cd *CallbackData) Validate() error {
	if cd.ReturnCode != ReturnCodeSuccess {
		return fmt.Errorf("回调返回失败: %s", cd.ReturnMsg)
	}
	
	if cd.ResultCode != ResultCodeSuccess {
		return fmt.Errorf("业务处理失败: %s - %s", cd.ErrCode, cd.ErrCodeDes)
	}
	
	if cd.OutTradeNo == "" {
		return fmt.Errorf("商户订单号不能为空")
	}
	
	if cd.TransactionID == "" {
		return fmt.Errorf("微信支付订单号不能为空")
	}
	
	if cd.TotalFee == "" {
		return fmt.Errorf("订单金额不能为空")
	}
	
	return nil
}

// ToPaymentCallbackData 转换为通用回调数据
func (cd *CallbackData) ToPaymentCallbackData() *model.PaymentCallbackData {
	paidAt, _ := cd.GetPaymentTime()
	if paidAt == nil {
		now := time.Now()
		paidAt = &now
	}
	
	return &model.PaymentCallbackData{
		PaymentMethod: model.PaymentMethodWechat,
		ThirdPartyID:  cd.TransactionID,
		PaymentNo:     cd.OutTradeNo,
		Amount:        cd.GetTotalAmount(),
		PaymentStatus: cd.ToPaymentStatus(),
		PaidAt:        *paidAt,
		RawData:       "", // 需要在调用时设置
		Signature:     cd.Sign,
	}
}

// ErrorCode 错误码常量
const (
	ErrorCodeOrderNotExist    = "ORDERNOTEXIST"    // 订单不存在
	ErrorCodeSystemError      = "SYSTEMERROR"      // 系统错误
	ErrorCodeSignError        = "SIGNERROR"        // 签名错误
	ErrorCodeParamError       = "PARAM_ERROR"      // 参数错误
	ErrorCodeNotEnoughFunds   = "NOTENOUGH"        // 余额不足
	ErrorCodeOrderPaid        = "ORDERPAID"        // 商户订单已支付
	ErrorCodeOrderClosed      = "ORDERCLOSED"      // 订单已关闭
	ErrorCodeAuthCodeExpire   = "AUTHCODEEXPIRE"   // 授权码过期
	ErrorCodeAuthCodeInvalid  = "AUTHCODEINVALID"  // 授权码无效
)

// GetErrorMessage 获取错误信息
func GetErrorMessage(errCode, errCodeDes string) string {
	if errCodeDes != "" {
		return errCodeDes
	}
	
	switch errCode {
	case ErrorCodeOrderNotExist:
		return "订单不存在"
	case ErrorCodeSystemError:
		return "系统错误"
	case ErrorCodeSignError:
		return "签名错误"
	case ErrorCodeParamError:
		return "参数错误"
	case ErrorCodeNotEnoughFunds:
		return "余额不足"
	case ErrorCodeOrderPaid:
		return "订单已支付"
	case ErrorCodeOrderClosed:
		return "订单已关闭"
	case ErrorCodeAuthCodeExpire:
		return "授权码过期"
	case ErrorCodeAuthCodeInvalid:
		return "授权码无效"
	default:
		return fmt.Sprintf("未知错误: %s", errCode)
	}
}

// BuildSuccessResponse 构建成功响应XML
func BuildSuccessResponse() string {
	return `<xml><return_code><![CDATA[SUCCESS]]></return_code><return_msg><![CDATA[OK]]></return_msg></xml>`
}

// BuildFailResponse 构建失败响应XML
func BuildFailResponse(msg string) string {
	return fmt.Sprintf(`<xml><return_code><![CDATA[FAIL]]></return_code><return_msg><![CDATA[%s]]></return_msg></xml>`, msg)
}
