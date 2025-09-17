package model

import (
	"errors"
	"time"

	"github.com/shopspring/decimal"
)

// PaymentCreateRequest 创建支付请求
type PaymentCreateRequest struct {
	OrderID        uint            `json:"order_id" binding:"required"`              // 订单ID
	PaymentMethod  PaymentMethod   `json:"payment_method" binding:"required"`        // 支付方式
	Amount         decimal.Decimal `json:"amount" binding:"required"`                // 支付金额
	Subject        string          `json:"subject" binding:"required,max=256"`       // 支付主题
	Description    string          `json:"description" binding:"max=500"`            // 支付描述
	ReturnURL      string          `json:"return_url" binding:"url,max=512"`         // 同步跳转地址
	NotifyURL      string          `json:"notify_url" binding:"url,max=512"`         // 异步通知地址
	ExpiredMinutes int             `json:"expired_minutes" binding:"min=1,max=1440"` // 过期时间(分钟)
}

// PaymentCreateResponse 创建支付响应
type PaymentCreateResponse struct {
	PaymentID     uint            `json:"payment_id"`             // 支付ID
	PaymentNo     string          `json:"payment_no"`             // 支付单号
	PaymentMethod PaymentMethod   `json:"payment_method"`         // 支付方式
	Amount        decimal.Decimal `json:"amount"`                 // 支付金额
	PaymentURL    string          `json:"payment_url,omitempty"`  // 支付链接
	QRCode        string          `json:"qr_code,omitempty"`      // 二维码内容
	PaymentData   interface{}     `json:"payment_data,omitempty"` // 支付数据
	ExpiredAt     time.Time       `json:"expired_at"`             // 过期时间
	CreatedAt     time.Time       `json:"created_at"`             // 创建时间
}

// PaymentQueryRequest 查询支付请求
type PaymentQueryRequest struct {
	PaymentID uint   `json:"payment_id" form:"payment_id"` // 支付ID
	PaymentNo string `json:"payment_no" form:"payment_no"` // 支付单号
	OrderID   uint   `json:"order_id" form:"order_id"`     // 订单ID
}

// PaymentQueryResponse 查询支付响应
type PaymentQueryResponse struct {
	PaymentID     uint            `json:"payment_id"`     // 支付ID
	PaymentNo     string          `json:"payment_no"`     // 支付单号
	OrderID       uint            `json:"order_id"`       // 订单ID
	PaymentMethod PaymentMethod   `json:"payment_method"` // 支付方式
	PaymentStatus PaymentStatus   `json:"payment_status"` // 支付状态
	PaymentType   PaymentType     `json:"payment_type"`   // 支付类型
	Amount        decimal.Decimal `json:"amount"`         // 支付金额
	ActualAmount  decimal.Decimal `json:"actual_amount"`  // 实际支付金额
	Currency      string          `json:"currency"`       // 货币类型
	Subject       string          `json:"subject"`        // 支付主题
	Description   string          `json:"description"`    // 支付描述
	ThirdPartyID  string          `json:"third_party_id"` // 第三方支付单号
	ExpiredAt     *time.Time      `json:"expired_at"`     // 过期时间
	PaidAt        *time.Time      `json:"paid_at"`        // 支付时间
	CreatedAt     time.Time       `json:"created_at"`     // 创建时间
	UpdatedAt     time.Time       `json:"updated_at"`     // 更新时间
}

// PaymentListRequest 支付列表请求
type PaymentListRequest struct {
	UserID        uint          `json:"user_id" form:"user_id"`                             // 用户ID
	OrderID       uint          `json:"order_id" form:"order_id"`                           // 订单ID
	PaymentMethod PaymentMethod `json:"payment_method" form:"payment_method"`               // 支付方式
	PaymentStatus PaymentStatus `json:"payment_status" form:"payment_status"`               // 支付状态
	PaymentType   PaymentType   `json:"payment_type" form:"payment_type"`                   // 支付类型
	StartTime     string        `json:"start_time" form:"start_time"`                       // 开始时间
	EndTime       string        `json:"end_time" form:"end_time"`                           // 结束时间
	Page          int           `json:"page" form:"page" binding:"min=1"`                   // 页码
	PageSize      int           `json:"page_size" form:"page_size" binding:"min=1,max=100"` // 每页数量
}

// PaymentRefundRequest 退款请求
type PaymentRefundRequest struct {
	PaymentID    uint            `json:"payment_id" binding:"required"`            // 支付ID
	RefundAmount decimal.Decimal `json:"refund_amount" binding:"required"`         // 退款金额
	RefundReason string          `json:"refund_reason" binding:"required,max=512"` // 退款原因
}

// PaymentRefundResponse 退款响应
type PaymentRefundResponse struct {
	RefundID     uint            `json:"refund_id"`     // 退款ID
	RefundNo     string          `json:"refund_no"`     // 退款单号
	PaymentID    uint            `json:"payment_id"`    // 支付ID
	RefundAmount decimal.Decimal `json:"refund_amount"` // 退款金额
	RefundStatus PaymentStatus   `json:"refund_status"` // 退款状态
	RefundReason string          `json:"refund_reason"` // 退款原因
	CreatedAt    time.Time       `json:"created_at"`    // 创建时间
}

// PaymentCallbackData 支付回调数据
type PaymentCallbackData struct {
	PaymentMethod PaymentMethod   `json:"payment_method"` // 支付方式
	ThirdPartyID  string          `json:"third_party_id"` // 第三方支付单号
	PaymentNo     string          `json:"payment_no"`     // 支付单号
	Amount        decimal.Decimal `json:"amount"`         // 支付金额
	PaymentStatus PaymentStatus   `json:"payment_status"` // 支付状态
	PaidAt        time.Time       `json:"paid_at"`        // 支付时间
	RawData       string          `json:"raw_data"`       // 原始数据
	Signature     string          `json:"signature"`      // 签名
}

// PaymentStatisticsRequest 支付统计请求
type PaymentStatisticsRequest struct {
	UserID        uint          `json:"user_id" form:"user_id"`               // 用户ID
	PaymentMethod PaymentMethod `json:"payment_method" form:"payment_method"` // 支付方式
	PaymentStatus PaymentStatus `json:"payment_status" form:"payment_status"` // 支付状态
	StartDate     string        `json:"start_date" form:"start_date"`         // 开始日期
	EndDate       string        `json:"end_date" form:"end_date"`             // 结束日期
	GroupBy       string        `json:"group_by" form:"group_by"`             // 分组方式: day, month, year
}

// PaymentStatisticsResponse 支付统计响应
type PaymentStatisticsResponse struct {
	TotalAmount   decimal.Decimal                     `json:"total_amount"`          // 总金额
	TotalCount    int64                               `json:"total_count"`           // 总笔数
	SuccessAmount decimal.Decimal                     `json:"success_amount"`        // 成功金额
	SuccessCount  int64                               `json:"success_count"`         // 成功笔数
	FailedCount   int64                               `json:"failed_count"`          // 失败笔数
	RefundAmount  decimal.Decimal                     `json:"refund_amount"`         // 退款金额
	RefundCount   int64                               `json:"refund_count"`          // 退款笔数
	MethodStats   map[PaymentMethod]PaymentMethodStat `json:"method_stats"`          // 支付方式统计
	DailyStats    []PaymentDailyStat                  `json:"daily_stats,omitempty"` // 日统计
}

// PaymentMethodStat 支付方式统计
type PaymentMethodStat struct {
	Method      PaymentMethod   `json:"method"`       // 支付方式
	Amount      decimal.Decimal `json:"amount"`       // 金额
	Count       int64           `json:"count"`        // 笔数
	SuccessRate float64         `json:"success_rate"` // 成功率
}

// PaymentDailyStat 日统计
type PaymentDailyStat struct {
	Date        string          `json:"date"`         // 日期
	Amount      decimal.Decimal `json:"amount"`       // 金额
	Count       int64           `json:"count"`        // 笔数
	SuccessRate float64         `json:"success_rate"` // 成功率
}

// PaymentConfigRequest 支付配置请求
type PaymentConfigRequest struct {
	PaymentMethod PaymentMethod   `json:"payment_method" binding:"required"`       // 支付方式
	IsEnabled     bool            `json:"is_enabled"`                              // 是否启用
	DisplayName   string          `json:"display_name" binding:"required,max=100"` // 显示名称
	DisplayOrder  int             `json:"display_order"`                           // 显示顺序
	Icon          string          `json:"icon" binding:"max=512"`                  // 图标URL
	Config        string          `json:"config"`                                  // 配置JSON
	MinAmount     decimal.Decimal `json:"min_amount"`                              // 最小金额
	MaxAmount     decimal.Decimal `json:"max_amount"`                              // 最大金额
}

// PaymentConfigResponse 支付配置响应
type PaymentConfigResponse struct {
	ID            uint            `json:"id"`             // 配置ID
	PaymentMethod PaymentMethod   `json:"payment_method"` // 支付方式
	IsEnabled     bool            `json:"is_enabled"`     // 是否启用
	DisplayName   string          `json:"display_name"`   // 显示名称
	DisplayOrder  int             `json:"display_order"`  // 显示顺序
	Icon          string          `json:"icon"`           // 图标URL
	MinAmount     decimal.Decimal `json:"min_amount"`     // 最小金额
	MaxAmount     decimal.Decimal `json:"max_amount"`     // 最大金额
	CreatedAt     time.Time       `json:"created_at"`     // 创建时间
	UpdatedAt     time.Time       `json:"updated_at"`     // 更新时间
}

// 验证方法
func (req *PaymentCreateRequest) Validate() error {
	if !req.PaymentMethod.IsValid() {
		return ErrInvalidPaymentMethod
	}

	if req.Amount.LessThanOrEqual(decimal.Zero) {
		return ErrInvalidAmount
	}

	if req.ExpiredMinutes <= 0 {
		req.ExpiredMinutes = 30 // 默认30分钟
	}

	return nil
}

func (req *PaymentRefundRequest) Validate() error {
	if req.RefundAmount.LessThanOrEqual(decimal.Zero) {
		return ErrInvalidAmount
	}

	return nil
}

// 错误定义
var (
	ErrInvalidPaymentMethod = errors.New("无效的支付方式")
	ErrInvalidAmount        = errors.New("无效的金额")
	ErrPaymentNotFound      = errors.New("支付记录不存在")
	ErrPaymentExpired       = errors.New("支付已过期")
	ErrPaymentAlreadyPaid   = errors.New("支付已完成")
	ErrInsufficientAmount   = errors.New("金额不足")
	ErrRefundFailed         = errors.New("退款失败")
)
