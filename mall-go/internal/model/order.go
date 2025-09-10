package model

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// Order 订单模型
type Order struct {
	ID            uint   `gorm:"primarykey" json:"id"`
	OrderNo       string `gorm:"uniqueIndex;not null;size:32" json:"order_no"`     // 订单号
	UserID        uint   `gorm:"not null;index" json:"user_id"`                    // 用户ID
	Status        string `gorm:"size:20;not null;index" json:"status"`             // 订单状态
	PaymentStatus string `gorm:"size:20;not null;index" json:"payment_status"`     // 支付状态
	OrderType     string `gorm:"size:20;default:'normal';index" json:"order_type"` // 订单类型
	PaymentType   string `gorm:"size:20" json:"payment_type"`                      // 支付方式

	// 金额信息
	TotalAmount    decimal.Decimal `gorm:"type:decimal(10,2);not null" json:"total_amount"`     // 订单总金额
	PayableAmount  decimal.Decimal `gorm:"type:decimal(10,2);not null" json:"payable_amount"`   // 应付金额
	PaidAmount     decimal.Decimal `gorm:"type:decimal(10,2);default:0" json:"paid_amount"`     // 已付金额
	DiscountAmount decimal.Decimal `gorm:"type:decimal(10,2);default:0" json:"discount_amount"` // 优惠金额
	ShippingFee    decimal.Decimal `gorm:"type:decimal(10,2);default:0" json:"shipping_fee"`    // 运费
	TaxAmount      decimal.Decimal `gorm:"type:decimal(10,2);default:0" json:"tax_amount"`      // 税费

	// 优惠信息
	CouponID     uint            `gorm:"index" json:"coupon_id"`                            // 优惠券ID
	CouponAmount decimal.Decimal `gorm:"type:decimal(10,2);default:0" json:"coupon_amount"` // 优惠券金额
	PointsUsed   int             `gorm:"default:0" json:"points_used"`                      // 使用积分
	PointsAmount decimal.Decimal `gorm:"type:decimal(10,2);default:0" json:"points_amount"` // 积分抵扣金额

	// 收货信息
	ReceiverName    string `gorm:"size:50;not null" json:"receiver_name"`     // 收货人姓名
	ReceiverPhone   string `gorm:"size:20;not null" json:"receiver_phone"`    // 收货人电话
	ReceiverAddress string `gorm:"size:500;not null" json:"receiver_address"` // 收货地址
	ReceiverZipCode string `gorm:"size:10" json:"receiver_zip_code"`          // 邮政编码
	Province        string `gorm:"size:50" json:"province"`                   // 省份
	City            string `gorm:"size:50" json:"city"`                       // 城市
	District        string `gorm:"size:50" json:"district"`                   // 区县

	// 物流信息
	ShippingMethod   string     `gorm:"size:50" json:"shipping_method"`  // 配送方式
	ShippingCompany  string     `gorm:"size:50" json:"shipping_company"` // 物流公司
	TrackingNumber   string     `gorm:"size:100" json:"tracking_number"` // 物流单号
	ShippingStatus   string     `gorm:"size:20" json:"shipping_status"`  // 物流状态
	EstimatedArrival *time.Time `json:"estimated_arrival"`               // 预计到达时间

	// 时间信息
	OrderTime    time.Time  `gorm:"not null" json:"order_time"` // 下单时间
	PayTime      *time.Time `json:"pay_time"`                   // 支付时间
	ShipTime     *time.Time `json:"ship_time"`                  // 发货时间
	DeliveryTime *time.Time `json:"delivery_time"`              // 配送时间
	ReceiveTime  *time.Time `json:"receive_time"`               // 收货时间
	FinishTime   *time.Time `json:"finish_time"`                // 完成时间
	CancelTime   *time.Time `json:"cancel_time"`                // 取消时间
	CloseTime    *time.Time `json:"close_time"`                 // 关闭时间

	// 超时设置
	PayExpireTime     *time.Time `json:"pay_expire_time"`     // 支付超时时间
	ReceiveExpireTime *time.Time `json:"receive_expire_time"` // 收货超时时间

	// 备注信息
	BuyerMessage  string `gorm:"size:500" json:"buyer_message"`  // 买家留言
	SellerMessage string `gorm:"size:500" json:"seller_message"` // 卖家备注
	AdminMessage  string `gorm:"size:500" json:"admin_message"`  // 管理员备注

	// 评价信息
	IsReviewed       bool       `gorm:"default:false" json:"is_reviewed"` // 是否已评价
	ReviewTime       *time.Time `json:"review_time"`                      // 评价时间
	ReviewExpireTime *time.Time `json:"review_expire_time"`               // 评价超时时间

	// 售后信息
	RefundStatus string          `gorm:"size:20" json:"refund_status"`                      // 退款状态
	RefundAmount decimal.Decimal `gorm:"type:decimal(10,2);default:0" json:"refund_amount"` // 退款金额
	RefundReason string          `gorm:"size:200" json:"refund_reason"`                     // 退款原因
	RefundTime   *time.Time      `json:"refund_time"`                                       // 退款时间

	// 乐观锁版本号
	Version int `gorm:"not null;default:1" json:"version"` // 乐观锁版本号

	// 时间戳
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联关系
	User       *User            `gorm:"foreignKey:UserID" json:"user,omitempty"`
	OrderItems []OrderItem      `gorm:"foreignKey:OrderID" json:"order_items,omitempty"`
	StatusLogs []OrderStatusLog `gorm:"foreignKey:OrderID" json:"status_logs,omitempty"`
	Payments   []OrderPayment   `gorm:"foreignKey:OrderID" json:"payments,omitempty"`
	Shipments  []OrderShipment  `gorm:"foreignKey:OrderID" json:"shipments,omitempty"`
	AfterSales []OrderAfterSale `gorm:"foreignKey:OrderID" json:"after_sales,omitempty"`
}

// OrderItem 订单商品项模型
type OrderItem struct {
	ID        uint `gorm:"primarykey" json:"id"`
	OrderID   uint `gorm:"not null;index" json:"order_id"`
	ProductID uint `gorm:"not null;index" json:"product_id"`
	SKUID     uint `gorm:"index" json:"sku_id"`
	Quantity  int  `gorm:"not null" json:"quantity"`

	// 商品快照信息（下单时的商品信息）
	ProductName  string          `gorm:"size:255;not null" json:"product_name"`
	ProductImage string          `gorm:"size:500" json:"product_image"`
	SKUName      string          `gorm:"size:255" json:"sku_name"`
	SKUImage     string          `gorm:"size:500" json:"sku_image"`
	SKUAttrs     string          `gorm:"type:json" json:"sku_attrs"`
	Price        decimal.Decimal `gorm:"type:decimal(10,2);not null" json:"price"`
	TotalPrice   decimal.Decimal `gorm:"type:decimal(10,2);not null" json:"total_price"`

	// 售后状态
	RefundStatus   string          `gorm:"size:20;default:'none'" json:"refund_status"`
	RefundQuantity int             `gorm:"default:0" json:"refund_quantity"`
	RefundAmount   decimal.Decimal `gorm:"type:decimal(10,2);default:0" json:"refund_amount"`

	// 时间戳
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联关系
	Order   *Order      `gorm:"foreignKey:OrderID" json:"order,omitempty"`
	Product *Product    `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	SKU     *ProductSKU `gorm:"foreignKey:SKUID" json:"sku,omitempty"`
}

// OrderStatusLog 订单状态流转日志
type OrderStatusLog struct {
	ID           uint      `gorm:"primarykey" json:"id"`
	OrderID      uint      `gorm:"not null;index" json:"order_id"`
	FromStatus   string    `gorm:"size:20" json:"from_status"`
	ToStatus     string    `gorm:"size:20;not null" json:"to_status"`
	OperatorID   uint      `gorm:"index" json:"operator_id"`
	OperatorType string    `gorm:"size:20;not null" json:"operator_type"` // user, admin, system
	Reason       string    `gorm:"size:200" json:"reason"`
	Remark       string    `gorm:"size:500" json:"remark"`
	CreatedAt    time.Time `json:"created_at"`

	// 关联关系
	Order    *Order `gorm:"foreignKey:OrderID" json:"order,omitempty"`
	Operator *User  `gorm:"foreignKey:OperatorID" json:"operator,omitempty"`
}

// OrderPayment 订单支付记录
type OrderPayment struct {
	ID             uint   `gorm:"primarykey" json:"id"`
	OrderID        uint   `gorm:"not null;index" json:"order_id"`
	PaymentNo      string `gorm:"uniqueIndex;not null;size:32" json:"payment_no"`
	PaymentMethod  string `gorm:"size:20;not null" json:"payment_method"` // alipay, wechat, balance
	PaymentChannel string `gorm:"size:50" json:"payment_channel"`         // 支付渠道

	Amount decimal.Decimal `gorm:"type:decimal(10,2);not null" json:"amount"`
	Status string          `gorm:"size:20;not null" json:"status"`

	// 第三方支付信息
	ThirdPartyNo   string `gorm:"size:100" json:"third_party_no"`    // 第三方交易号
	ThirdPartyData string `gorm:"type:json" json:"third_party_data"` // 第三方返回数据

	PayTime    *time.Time `json:"pay_time"`
	ExpireTime *time.Time `json:"expire_time"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`

	// 关联关系
	Order *Order `gorm:"foreignKey:OrderID" json:"order,omitempty"`
}

// OrderShipment 订单物流记录
type OrderShipment struct {
	ID              uint   `gorm:"primarykey" json:"id"`
	OrderID         uint   `gorm:"not null;index" json:"order_id"`
	ShipmentNo      string `gorm:"uniqueIndex;not null;size:32" json:"shipment_no"`
	ShippingCompany string `gorm:"size:50;not null" json:"shipping_company"`
	TrackingNumber  string `gorm:"size:100;not null" json:"tracking_number"`
	Status          string `gorm:"size:20;not null" json:"status"`

	// 发货信息
	SenderName    string `gorm:"size:50" json:"sender_name"`
	SenderPhone   string `gorm:"size:20" json:"sender_phone"`
	SenderAddress string `gorm:"size:500" json:"sender_address"`

	// 收货信息
	ReceiverName    string `gorm:"size:50" json:"receiver_name"`
	ReceiverPhone   string `gorm:"size:20" json:"receiver_phone"`
	ReceiverAddress string `gorm:"size:500" json:"receiver_address"`

	// 物流轨迹
	TrackingData string `gorm:"type:json" json:"tracking_data"`

	ShipTime     *time.Time `json:"ship_time"`
	DeliveryTime *time.Time `json:"delivery_time"`
	ReceiveTime  *time.Time `json:"receive_time"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`

	// 关联关系
	Order *Order `gorm:"foreignKey:OrderID" json:"order,omitempty"`
}

// OrderAfterSale 订单售后记录
type OrderAfterSale struct {
	ID          uint   `gorm:"primarykey" json:"id"`
	OrderID     uint   `gorm:"not null;index" json:"order_id"`
	OrderItemID uint   `gorm:"index" json:"order_item_id"`
	AfterSaleNo string `gorm:"uniqueIndex;not null;size:32" json:"after_sale_no"`
	Type        string `gorm:"size:20;not null" json:"type"`   // refund, return, exchange
	Status      string `gorm:"size:20;not null" json:"status"` // pending, approved, rejected, completed

	// 申请信息
	ApplyUserID uint            `gorm:"not null;index" json:"apply_user_id"`
	Reason      string          `gorm:"size:200;not null" json:"reason"`
	Description string          `gorm:"size:1000" json:"description"`
	Images      string          `gorm:"type:json" json:"images"`
	Amount      decimal.Decimal `gorm:"type:decimal(10,2)" json:"amount"`
	Quantity    int             `gorm:"default:1" json:"quantity"`

	// 处理信息
	HandleUserID uint       `gorm:"index" json:"handle_user_id"`
	HandleRemark string     `gorm:"size:500" json:"handle_remark"`
	HandleTime   *time.Time `json:"handle_time"`

	// 退款信息
	RefundMethod string          `gorm:"size:20" json:"refund_method"`
	RefundAmount decimal.Decimal `gorm:"type:decimal(10,2)" json:"refund_amount"`
	RefundTime   *time.Time      `json:"refund_time"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联关系
	Order      *Order     `gorm:"foreignKey:OrderID" json:"order,omitempty"`
	OrderItem  *OrderItem `gorm:"foreignKey:OrderItemID" json:"order_item,omitempty"`
	ApplyUser  *User      `gorm:"foreignKey:ApplyUserID" json:"apply_user,omitempty"`
	HandleUser *User      `gorm:"foreignKey:HandleUserID" json:"handle_user,omitempty"`
}

// TableName 指定表名
func (Order) TableName() string {
	return "orders"
}

func (OrderItem) TableName() string {
	return "order_items"
}

func (OrderStatusLog) TableName() string {
	return "order_status_logs"
}

func (OrderPayment) TableName() string {
	return "order_payments"
}

func (OrderShipment) TableName() string {
	return "order_shipments"
}

func (OrderAfterSale) TableName() string {
	return "order_after_sales"
}

// 订单状态常量
const (
	OrderStatusPending   = "pending"   // 待支付
	OrderStatusPaid      = "paid"      // 已支付
	OrderStatusShipped   = "shipped"   // 已发货
	OrderStatusDelivered = "delivered" // 已配送
	OrderStatusReceived  = "received"  // 已收货
	OrderStatusCompleted = "completed" // 已完成
	OrderStatusCancelled = "cancelled" // 已取消
	OrderStatusClosed    = "closed"    // 已关闭
	OrderStatusRefunding = "refunding" // 退款中
	OrderStatusRefunded  = "refunded"  // 已退款
)

// 订单类型常量
const (
	OrderTypeNormal   = "normal"   // 普通订单
	OrderTypePresale  = "presale"  // 预售订单
	OrderTypeGroup    = "group"    // 团购订单
	OrderTypeSeckill  = "seckill"  // 秒杀订单
	OrderTypeExchange = "exchange" // 换货订单
)

// 支付方式常量
const (
	PaymentTypeAlipay  = "alipay"  // 支付宝
	PaymentTypeWechat  = "wechat"  // 微信支付
	PaymentTypeBalance = "balance" // 余额支付
	PaymentTypeBank    = "bank"    // 银行卡
	PaymentTypeCOD     = "cod"     // 货到付款
)

// 物流状态常量
const (
	ShippingStatusPending   = "pending"   // 待发货
	ShippingStatusShipped   = "shipped"   // 已发货
	ShippingStatusTransit   = "transit"   // 运输中
	ShippingStatusDelivered = "delivered" // 已配送
	ShippingStatusReceived  = "received"  // 已签收
	ShippingStatusReturned  = "returned"  // 已退回
)

// 售后类型常量
const (
	AfterSaleTypeRefund   = "refund"   // 仅退款
	AfterSaleTypeReturn   = "return"   // 退货退款
	AfterSaleTypeExchange = "exchange" // 换货
)

// 售后状态常量
const (
	AfterSaleStatusPending   = "pending"   // 待处理
	AfterSaleStatusApproved  = "approved"  // 已同意
	AfterSaleStatusRejected  = "rejected"  // 已拒绝
	AfterSaleStatusReturning = "returning" // 退货中
	AfterSaleStatusCompleted = "completed" // 已完成
	AfterSaleStatusCancelled = "cancelled" // 已取消
)

// 退款状态常量
const (
	RefundStatusNone      = "none"      // 无退款
	RefundStatusPending   = "pending"   // 退款中
	RefundStatusCompleted = "completed" // 退款完成
	RefundStatusFailed    = "failed"    // 退款失败
	RefundStatusPartial   = "partial"   // 部分退款
)

// 操作者类型常量
const (
	OperatorTypeUser   = "user"   // 用户操作
	OperatorTypeAdmin  = "admin"  // 管理员操作
	OperatorTypeSystem = "system" // 系统操作
)

// 订单方法
func (o *Order) IsPaid() bool {
	return o.Status == OrderStatusPaid || o.Status == OrderStatusShipped ||
		o.Status == OrderStatusDelivered || o.Status == OrderStatusReceived ||
		o.Status == OrderStatusCompleted
}

func (o *Order) CanCancel() bool {
	return o.Status == OrderStatusPending
}

func (o *Order) CanPay() bool {
	return o.Status == OrderStatusPending && (o.PayExpireTime == nil || o.PayExpireTime.After(time.Now()))
}

func (o *Order) CanShip() bool {
	return o.Status == OrderStatusPaid
}

func (o *Order) CanReceive() bool {
	return o.Status == OrderStatusDelivered
}

func (o *Order) CanRefund() bool {
	return o.IsPaid() && o.Status != OrderStatusRefunded && o.Status != OrderStatusCancelled
}

func (o *Order) GenerateOrderNo() string {
	return fmt.Sprintf("ORD%d%d", time.Now().Unix(), o.UserID)
}

func (o *Order) IsExpired() bool {
	return o.PayExpireTime != nil && o.PayExpireTime.Before(time.Now())
}

func (o *Order) GetStatusText() string {
	statusMap := map[string]string{
		OrderStatusPending:   "待支付",
		OrderStatusPaid:      "已支付",
		OrderStatusShipped:   "已发货",
		OrderStatusDelivered: "已配送",
		OrderStatusReceived:  "已收货",
		OrderStatusCompleted: "已完成",
		OrderStatusCancelled: "已取消",
		OrderStatusClosed:    "已关闭",
		OrderStatusRefunding: "退款中",
		OrderStatusRefunded:  "已退款",
	}
	if text, exists := statusMap[o.Status]; exists {
		return text
	}
	return "未知状态"
}

// 订单请求结构体
type OrderCreateRequest struct {
	CartItemIDs     []uint `json:"cart_item_ids" binding:"required,min=1"`
	CouponID        uint   `json:"coupon_id"`
	PointsUsed      int    `json:"points_used"`
	ReceiverName    string `json:"receiver_name" binding:"required"`
	ReceiverPhone   string `json:"receiver_phone" binding:"required"`
	ReceiverAddress string `json:"receiver_address" binding:"required"`
	ReceiverZipCode string `json:"receiver_zip_code"`
	Province        string `json:"province" binding:"required"`
	City            string `json:"city" binding:"required"`
	District        string `json:"district" binding:"required"`
	ShippingMethod  string `json:"shipping_method"`
	BuyerMessage    string `json:"buyer_message"`
}

type OrderUpdateStatusRequest struct {
	Status string `json:"status" binding:"required"`
	Reason string `json:"reason"`
	Remark string `json:"remark"`
}

type OrderListRequest struct {
	Page      int    `form:"page" binding:"min=1"`
	PageSize  int    `form:"page_size" binding:"min=1,max=100"`
	Status    string `form:"status"`
	OrderType string `form:"order_type"`
	StartTime string `form:"start_time"`
	EndTime   string `form:"end_time"`
	Keyword   string `form:"keyword"`
}

type OrderShipRequest struct {
	ShippingCompany string `json:"shipping_company" binding:"required"`
	TrackingNumber  string `json:"tracking_number" binding:"required"`
	ShippingMethod  string `json:"shipping_method"`
	SenderName      string `json:"sender_name"`
	SenderPhone     string `json:"sender_phone"`
	SenderAddress   string `json:"sender_address"`
}

type OrderRefundRequest struct {
	OrderItemID uint            `json:"order_item_id"`
	Reason      string          `json:"reason" binding:"required"`
	Description string          `json:"description"`
	Amount      decimal.Decimal `json:"amount"`
	Quantity    int             `json:"quantity" binding:"min=1"`
	Images      []string        `json:"images"`
}

// 订单响应结构体
type OrderResponse struct {
	*Order
	StatusText string `json:"status_text"`
	CanCancel  bool   `json:"can_cancel"`
	CanPay     bool   `json:"can_pay"`
	CanShip    bool   `json:"can_ship"`
	CanReceive bool   `json:"can_receive"`
	CanRefund  bool   `json:"can_refund"`
}

type OrderListResponse struct {
	Orders     []*OrderResponse `json:"orders"`
	Total      int64            `json:"total"`
	Page       int              `json:"page"`
	PageSize   int              `json:"page_size"`
	TotalPages int              `json:"total_pages"`
}

type OrderStatisticsResponse struct {
	TotalOrders     int64           `json:"total_orders"`
	PendingOrders   int64           `json:"pending_orders"`
	PaidOrders      int64           `json:"paid_orders"`
	ShippedOrders   int64           `json:"shipped_orders"`
	CompletedOrders int64           `json:"completed_orders"`
	CancelledOrders int64           `json:"cancelled_orders"`
	TotalAmount     decimal.Decimal `json:"total_amount"`
	TodayOrders     int64           `json:"today_orders"`
	TodayAmount     decimal.Decimal `json:"today_amount"`
}

// 订单相关错误定义
var (
	ErrOrderNotFound      = fmt.Errorf("订单不存在")
	ErrOrderAlreadyPaid   = fmt.Errorf("订单已支付")
	ErrOrderCannotCancel  = fmt.Errorf("订单无法取消")
	ErrOrderCannotRefund  = fmt.Errorf("订单无法退款")
	ErrInvalidOrderStatus = fmt.Errorf("无效的订单状态")
)
