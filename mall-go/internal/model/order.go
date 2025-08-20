package model

import (
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// Order 订单模型
type Order struct {
	ID            uint            `gorm:"primarykey" json:"id"`
	OrderNo       string          `gorm:"uniqueIndex;not null;size:50" json:"order_no"`
	UserID        uint            `gorm:"not null" json:"user_id"`
	User          User            `json:"user"`
	TotalAmount   decimal.Decimal `gorm:"type:decimal(10,2);not null" json:"total_amount"`
	Status        string          `gorm:"default:'pending';size:20" json:"status"`
	PaymentMethod string          `gorm:"size:50" json:"payment_method"`
	PaymentStatus string          `gorm:"default:'unpaid';size:20" json:"payment_status"`
	OrderItems    []OrderItem     `json:"order_items"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
	DeletedAt     gorm.DeletedAt  `gorm:"index" json:"-"`
}

// OrderItem 订单项
type OrderItem struct {
	ID        uint            `gorm:"primarykey" json:"id"`
	OrderID   uint            `gorm:"not null" json:"order_id"`
	ProductID uint            `gorm:"not null" json:"product_id"`
	Product   Product         `json:"product"`
	Quantity  int             `gorm:"not null;min=1" json:"quantity"`
	Price     decimal.Decimal `gorm:"type:decimal(10,2);not null" json:"price"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// TableName 指定表名
func (Order) TableName() string {
	return "orders"
}

func (OrderItem) TableName() string {
	return "order_items"
}

// OrderCreateRequest 创建订单请求
type OrderCreateRequest struct {
	ProductID uint `json:"product_id" binding:"required"`
	Quantity  int  `json:"quantity" binding:"required,min=1"`
}

// OrderUpdateStatusRequest 更新订单状态请求
type OrderUpdateStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=pending paid shipped delivered cancelled"`
}

// OrderListRequest 订单列表请求
type OrderListRequest struct {
	Page     int    `form:"page" binding:"min=1"`
	PageSize int    `form:"page_size" binding:"min=1,max=100"`
	Status   string `form:"status"`
}

// OrderStatus 订单状态常量
const (
	OrderStatusPending   = "pending"   // 待付款
	OrderStatusPaid      = "paid"      // 已付款
	OrderStatusShipped   = "shipped"   // 已发货
	OrderStatusDelivered = "delivered" // 已送达
	OrderStatusCancelled = "cancelled" // 已取消
)

// PaymentStatus 支付状态常量
const (
	PaymentStatusUnpaid = "unpaid" // 未支付
	PaymentStatusPaid   = "paid"   // 已支付
	PaymentStatusFailed = "failed" // 支付失败
)
