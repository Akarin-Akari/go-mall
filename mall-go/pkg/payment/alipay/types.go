package alipay

import (
	"fmt"
	"time"

	"mall-go/internal/model"

	"github.com/shopspring/decimal"
)

// PaymentRequest 支付请求
type PaymentRequest struct {
	OutTradeNo  string          `json:"out_trade_no"` // 商户订单号
	TotalAmount decimal.Decimal `json:"total_amount"` // 订单总金额
	Subject     string          `json:"subject"`      // 订单标题
	Body        string          `json:"body"`         // 订单描述
	NotifyURL   string          `json:"notify_url"`   // 异步通知地址
	ReturnURL   string          `json:"return_url"`   // 同步跳转地址
	TimeExpire  *time.Time      `json:"time_expire"`  // 订单过期时间
}

// PaymentResponse 支付响应
type PaymentResponse struct {
	OutTradeNo string `json:"out_trade_no"` // 商户订单号
	QRCode     string `json:"qr_code"`      // 二维码内容
	Success    bool   `json:"success"`      // 是否成功
	Message    string `json:"message"`      // 错误信息
}

// QueryResponse 查询响应
type QueryResponse struct {
	OutTradeNo  string              `json:"out_trade_no"` // 商户订单号
	TradeNo     string              `json:"trade_no"`     // 支付宝交易号
	TradeStatus string              `json:"trade_status"` // 交易状态
	TotalAmount decimal.Decimal     `json:"total_amount"` // 订单金额
	Status      model.PaymentStatus `json:"status"`       // 标准化状态
	Success     bool                `json:"success"`      // 是否成功
	Message     string              `json:"message"`      // 错误信息
}

// CallbackData 回调数据
type CallbackData struct {
	// 基本信息
	NotifyTime string `json:"notify_time"` // 通知时间
	NotifyType string `json:"notify_type"` // 通知类型
	NotifyID   string `json:"notify_id"`   // 通知校验ID
	AppID      string `json:"app_id"`      // 开发者的app_id
	Charset    string `json:"charset"`     // 编码格式
	Version    string `json:"version"`     // 调用的接口版本
	SignType   string `json:"sign_type"`   // 签名类型
	Sign       string `json:"sign"`        // 签名

	// 交易信息
	TradeNo      string `json:"trade_no"`       // 支付宝交易号
	OutTradeNo   string `json:"out_trade_no"`   // 商户订单号
	OutBizNo     string `json:"out_biz_no"`     // 商户业务号
	BuyerID      string `json:"buyer_id"`       // 买家支付宝用户号
	BuyerLogonID string `json:"buyer_logon_id"` // 买家支付宝账号
	SellerID     string `json:"seller_id"`      // 卖家支付宝用户号
	SellerEmail  string `json:"seller_email"`   // 卖家支付宝账号

	// 金额信息
	TotalAmount    decimal.Decimal `json:"total_amount"`     // 订单金额
	ReceiptAmount  decimal.Decimal `json:"receipt_amount"`   // 实收金额
	InvoiceAmount  decimal.Decimal `json:"invoice_amount"`   // 开票金额
	BuyerPayAmount decimal.Decimal `json:"buyer_pay_amount"` // 付款金额

	// 状态信息
	TradeStatus string `json:"trade_status"` // 交易状态
	Subject     string `json:"subject"`      // 订单标题
	Body        string `json:"body"`         // 商品描述

	// 时间信息
	GmtCreate  string `json:"gmt_create"`  // 交易创建时间
	GmtPayment string `json:"gmt_payment"` // 交易付款时间
	GmtRefund  string `json:"gmt_refund"`  // 交易退款时间
	GmtClose   string `json:"gmt_close"`   // 交易结束时间

	// 其他信息
	FundBillList      string `json:"fund_bill_list"`      // 支付金额信息
	Passbackparams    string `json:"passback_params"`     // 回传参数
	VoucherDetailList string `json:"voucher_detail_list"` // 优惠券信息
}

// RefundRequest 退款请求
type RefundRequest struct {
	OutTradeNo   string          `json:"out_trade_no"`   // 商户订单号
	TradeNo      string          `json:"trade_no"`       // 支付宝交易号
	RefundAmount decimal.Decimal `json:"refund_amount"`  // 退款金额
	RefundReason string          `json:"refund_reason"`  // 退款原因
	OutRequestNo string          `json:"out_request_no"` // 退款请求号
}

// RefundResponse 退款响应
type RefundResponse struct {
	OutTradeNo   string          `json:"out_trade_no"`   // 商户订单号
	TradeNo      string          `json:"trade_no"`       // 支付宝交易号
	OutRequestNo string          `json:"out_request_no"` // 退款请求号
	RefundFee    decimal.Decimal `json:"refund_fee"`     // 退款金额
	GmtRefundPay string          `json:"gmt_refund_pay"` // 退款时间
	Success      bool            `json:"success"`        // 是否成功
	Message      string          `json:"message"`        // 错误信息
}

// TradeStatus 交易状态常量
const (
	TradeStatusWaitBuyerPay = "WAIT_BUYER_PAY" // 交易创建，等待买家付款
	TradeStatusClosed       = "TRADE_CLOSED"   // 未付款交易超时关闭，或支付完成后全额退款
	TradeStatusSuccess      = "TRADE_SUCCESS"  // 交易支付成功
	TradeStatusFinished     = "TRADE_FINISHED" // 交易结束，不可退款
)

// NotifyType 通知类型常量
const (
	NotifyTypeTradePay    = "trade_status_sync" // 交易状态同步
	NotifyTypeTradeRefund = "trade_refund"      // 交易退款
)

// ToPaymentStatus 转换为标准支付状态
func (cd *CallbackData) ToPaymentStatus() model.PaymentStatus {
	switch cd.TradeStatus {
	case TradeStatusWaitBuyerPay:
		return model.PaymentStatusPending
	case TradeStatusSuccess, TradeStatusFinished:
		return model.PaymentStatusSuccess
	case TradeStatusClosed:
		return model.PaymentStatusCancelled
	default:
		return model.PaymentStatusFailed
	}
}

// IsPaymentSuccess 是否支付成功
func (cd *CallbackData) IsPaymentSuccess() bool {
	return cd.TradeStatus == TradeStatusSuccess || cd.TradeStatus == TradeStatusFinished
}

// GetPaymentTime 获取支付时间
func (cd *CallbackData) GetPaymentTime() (*time.Time, error) {
	if cd.GmtPayment == "" {
		return nil, nil
	}

	// 解析时间格式: 2006-01-02 15:04:05
	t, err := time.Parse("2006-01-02 15:04:05", cd.GmtPayment)
	if err != nil {
		return nil, err
	}

	return &t, nil
}

// Validate 验证回调数据
func (cd *CallbackData) Validate() error {
	if cd.OutTradeNo == "" {
		return fmt.Errorf("商户订单号不能为空")
	}

	if cd.TradeNo == "" {
		return fmt.Errorf("支付宝交易号不能为空")
	}

	if cd.TotalAmount.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("订单金额必须大于0")
	}

	if cd.TradeStatus == "" {
		return fmt.Errorf("交易状态不能为空")
	}

	return nil
}

// ToPaymentCallbackData 转换为通用回调数据
func (cd *CallbackData) ToPaymentCallbackData() *model.PaymentCallbackData {
	paidAt, _ := cd.GetPaymentTime()

	return &model.PaymentCallbackData{
		PaymentMethod: model.PaymentMethodAlipay,
		ThirdPartyID:  cd.TradeNo,
		PaymentNo:     cd.OutTradeNo,
		Amount:        cd.TotalAmount,
		PaymentStatus: cd.ToPaymentStatus(),
		PaidAt:        *paidAt,
		RawData:       "", // 需要在调用时设置
		Signature:     cd.Sign,
	}
}

// ErrorCode 错误码常量
const (
	ErrorCodeSuccess            = "10000" // 接口调用成功
	ErrorCodeServiceUnavailable = "20000" // 服务不可用
	ErrorCodeInvalidAuth        = "20001" // 授权权限不足
	ErrorCodeMissingParam       = "40001" // 缺少必选参数
	ErrorCodeInvalidParam       = "40002" // 非法的参数
	ErrorCodeBusinessFailed     = "40004" // 业务处理失败
	ErrorCodePermissionDenied   = "40006" // 权限不足
)

// IsSuccess 判断是否成功
func IsSuccess(code string) bool {
	return code == ErrorCodeSuccess
}

// GetErrorMessage 获取错误信息
func GetErrorMessage(code, subCode, subMsg string) string {
	if code == ErrorCodeSuccess {
		return "成功"
	}

	if subMsg != "" {
		return fmt.Sprintf("%s: %s", subCode, subMsg)
	}

	switch code {
	case ErrorCodeServiceUnavailable:
		return "服务不可用"
	case ErrorCodeInvalidAuth:
		return "授权权限不足"
	case ErrorCodeMissingParam:
		return "缺少必选参数"
	case ErrorCodeInvalidParam:
		return "非法的参数"
	case ErrorCodeBusinessFailed:
		return "业务处理失败"
	case ErrorCodePermissionDenied:
		return "权限不足"
	default:
		return fmt.Sprintf("未知错误: %s", code)
	}
}
