package model

import (
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// PaymentMethod 支付方式枚举
type PaymentMethod string

const (
	PaymentMethodAlipay    PaymentMethod = "alipay"     // 支付宝
	PaymentMethodWechat    PaymentMethod = "wechat"     // 微信支付
	PaymentMethodUnionPay  PaymentMethod = "unionpay"   // 银联支付
	PaymentMethodBalance   PaymentMethod = "balance"    // 余额支付
	PaymentMethodCredit    PaymentMethod = "credit"     // 信用支付
)

// PaymentStatus 支付状态枚举
type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "pending"   // 待支付
	PaymentStatusPaying    PaymentStatus = "paying"    // 支付中
	PaymentStatusSuccess   PaymentStatus = "success"   // 支付成功
	PaymentStatusFailed    PaymentStatus = "failed"    // 支付失败
	PaymentStatusCancelled PaymentStatus = "cancelled" // 已取消
	PaymentStatusRefunded  PaymentStatus = "refunded"  // 已退款
	PaymentStatusExpired   PaymentStatus = "expired"   // 已过期
)

// PaymentType 支付类型枚举
type PaymentType string

const (
	PaymentTypeOrder  PaymentType = "order"  // 订单支付
	PaymentTypeRecharge PaymentType = "recharge" // 充值
	PaymentTypeRefund PaymentType = "refund" // 退款
)

// Payment 支付记录模型
type Payment struct {
	ID              uint            `gorm:"primarykey" json:"id"`
	PaymentNo       string          `gorm:"uniqueIndex;not null;size:64" json:"payment_no"`        // 支付单号
	OrderID         uint            `gorm:"not null;index" json:"order_id"`                        // 关联订单ID
	Order           Order           `gorm:"foreignKey:OrderID" json:"order,omitempty"`             // 关联订单
	UserID          uint            `gorm:"not null;index" json:"user_id"`                         // 用户ID
	User            User            `gorm:"foreignKey:UserID" json:"user,omitempty"`               // 关联用户
	
	// 支付基本信息
	PaymentType     PaymentType     `gorm:"not null;size:20;default:'order'" json:"payment_type"`  // 支付类型
	PaymentMethod   PaymentMethod   `gorm:"not null;size:20" json:"payment_method"`                // 支付方式
	PaymentStatus   PaymentStatus   `gorm:"not null;size:20;default:'pending'" json:"payment_status"` // 支付状态
	
	// 金额信息
	Amount          decimal.Decimal `gorm:"type:decimal(10,2);not null" json:"amount"`             // 支付金额
	ActualAmount    decimal.Decimal `gorm:"type:decimal(10,2)" json:"actual_amount"`               // 实际支付金额
	Currency        string          `gorm:"size:3;default:'CNY'" json:"currency"`                  // 货币类型
	
	// 第三方支付信息
	ThirdPartyID    string          `gorm:"size:128;index" json:"third_party_id"`                  // 第三方支付单号
	ThirdPartyData  string          `gorm:"type:text" json:"third_party_data"`                     // 第三方返回数据
	
	// 支付详情
	Subject         string          `gorm:"size:256" json:"subject"`                               // 支付主题
	Description     string          `gorm:"type:text" json:"description"`                         // 支付描述
	NotifyURL       string          `gorm:"size:512" json:"notify_url"`                           // 异步通知地址
	ReturnURL       string          `gorm:"size:512" json:"return_url"`                           // 同步跳转地址
	
	// 时间信息
	ExpiredAt       *time.Time      `json:"expired_at"`                                            // 过期时间
	PaidAt          *time.Time      `json:"paid_at"`                                               // 支付时间
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
	DeletedAt       gorm.DeletedAt  `gorm:"index" json:"deleted_at,omitempty"`
}

// PaymentRefund 退款记录模型
type PaymentRefund struct {
	ID              uint            `gorm:"primarykey" json:"id"`
	RefundNo        string          `gorm:"uniqueIndex;not null;size:64" json:"refund_no"`         // 退款单号
	PaymentID       uint            `gorm:"not null;index" json:"payment_id"`                      // 关联支付ID
	Payment         Payment         `gorm:"foreignKey:PaymentID" json:"payment,omitempty"`         // 关联支付
	UserID          uint            `gorm:"not null;index" json:"user_id"`                         // 用户ID
	User            User            `gorm:"foreignKey:UserID" json:"user,omitempty"`               // 关联用户
	
	// 退款信息
	RefundAmount    decimal.Decimal `gorm:"type:decimal(10,2);not null" json:"refund_amount"`      // 退款金额
	RefundReason    string          `gorm:"size:512" json:"refund_reason"`                         // 退款原因
	RefundStatus    PaymentStatus   `gorm:"not null;size:20;default:'pending'" json:"refund_status"` // 退款状态
	
	// 第三方退款信息
	ThirdPartyRefundID string       `gorm:"size:128;index" json:"third_party_refund_id"`           // 第三方退款单号
	ThirdPartyData     string       `gorm:"type:text" json:"third_party_data"`                     // 第三方返回数据
	
	// 时间信息
	RefundedAt      *time.Time      `json:"refunded_at"`                                           // 退款时间
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
	DeletedAt       gorm.DeletedAt  `gorm:"index" json:"deleted_at,omitempty"`
}

// PaymentLog 支付日志模型
type PaymentLog struct {
	ID              uint            `gorm:"primarykey" json:"id"`
	PaymentID       uint            `gorm:"not null;index" json:"payment_id"`                      // 关联支付ID
	Payment         Payment         `gorm:"foreignKey:PaymentID" json:"payment,omitempty"`         // 关联支付
	
	// 日志信息
	Action          string          `gorm:"not null;size:50" json:"action"`                        // 操作类型
	Status          string          `gorm:"not null;size:20" json:"status"`                        // 操作状态
	Message         string          `gorm:"type:text" json:"message"`                              // 日志消息
	RequestData     string          `gorm:"type:text" json:"request_data"`                         // 请求数据
	ResponseData    string          `gorm:"type:text" json:"response_data"`                        // 响应数据
	ErrorMessage    string          `gorm:"type:text" json:"error_message"`                        // 错误信息
	
	// 时间信息
	CreatedAt       time.Time       `json:"created_at"`
}

// PaymentConfig 支付配置模型
type PaymentConfig struct {
	ID              uint            `gorm:"primarykey" json:"id"`
	PaymentMethod   PaymentMethod   `gorm:"not null;size:20;uniqueIndex" json:"payment_method"`    // 支付方式
	
	// 配置信息
	IsEnabled       bool            `gorm:"default:true" json:"is_enabled"`                        // 是否启用
	DisplayName     string          `gorm:"not null;size:100" json:"display_name"`                 // 显示名称
	DisplayOrder    int             `gorm:"default:0" json:"display_order"`                        // 显示顺序
	Icon            string          `gorm:"size:512" json:"icon"`                                   // 图标URL
	
	// 支付配置
	Config          string          `gorm:"type:text" json:"config"`                               // 配置JSON
	MinAmount       decimal.Decimal `gorm:"type:decimal(10,2);default:0.01" json:"min_amount"`     // 最小金额
	MaxAmount       decimal.Decimal `gorm:"type:decimal(10,2);default:50000" json:"max_amount"`    // 最大金额
	
	// 时间信息
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
	DeletedAt       gorm.DeletedAt  `gorm:"index" json:"deleted_at,omitempty"`
}

// TableName 指定表名
func (Payment) TableName() string {
	return "payments"
}

func (PaymentRefund) TableName() string {
	return "payment_refunds"
}

func (PaymentLog) TableName() string {
	return "payment_logs"
}

func (PaymentConfig) TableName() string {
	return "payment_configs"
}

// 支付状态检查方法
func (p *Payment) IsPending() bool {
	return p.PaymentStatus == PaymentStatusPending
}

func (p *Payment) IsPaying() bool {
	return p.PaymentStatus == PaymentStatusPaying
}

func (p *Payment) IsSuccess() bool {
	return p.PaymentStatus == PaymentStatusSuccess
}

func (p *Payment) IsFailed() bool {
	return p.PaymentStatus == PaymentStatusFailed
}

func (p *Payment) IsCancelled() bool {
	return p.PaymentStatus == PaymentStatusCancelled
}

func (p *Payment) IsRefunded() bool {
	return p.PaymentStatus == PaymentStatusRefunded
}

func (p *Payment) IsExpired() bool {
	return p.PaymentStatus == PaymentStatusExpired
}

// 支付状态转换方法
func (p *Payment) MarkAsPaying() {
	p.PaymentStatus = PaymentStatusPaying
}

func (p *Payment) MarkAsSuccess(paidAt time.Time, thirdPartyID string) {
	p.PaymentStatus = PaymentStatusSuccess
	p.PaidAt = &paidAt
	p.ThirdPartyID = thirdPartyID
}

func (p *Payment) MarkAsFailed(reason string) {
	p.PaymentStatus = PaymentStatusFailed
	p.Description = reason
}

func (p *Payment) MarkAsCancelled() {
	p.PaymentStatus = PaymentStatusCancelled
}

func (p *Payment) MarkAsExpired() {
	p.PaymentStatus = PaymentStatusExpired
}

// 支付方式验证
func (pm PaymentMethod) IsValid() bool {
	switch pm {
	case PaymentMethodAlipay, PaymentMethodWechat, PaymentMethodUnionPay, 
		 PaymentMethodBalance, PaymentMethodCredit:
		return true
	default:
		return false
	}
}

// 支付状态验证
func (ps PaymentStatus) IsValid() bool {
	switch ps {
	case PaymentStatusPending, PaymentStatusPaying, PaymentStatusSuccess,
		 PaymentStatusFailed, PaymentStatusCancelled, PaymentStatusRefunded, PaymentStatusExpired:
		return true
	default:
		return false
	}
}

// 支付类型验证
func (pt PaymentType) IsValid() bool {
	switch pt {
	case PaymentTypeOrder, PaymentTypeRecharge, PaymentTypeRefund:
		return true
	default:
		return false
	}
}
