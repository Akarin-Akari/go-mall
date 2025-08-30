package model

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// Cart 购物车模型
type Cart struct {
	ID        uint   `gorm:"primarykey" json:"id"`
	UserID    uint   `gorm:"index" json:"user_id"`                         // 用户ID，0表示游客
	SessionID string `gorm:"size:100;index" json:"session_id"`             // 会话ID，用于游客购物车
	Status    string `gorm:"size:20;default:'active';index" json:"status"` // 购物车状态

	// 统计信息
	ItemCount   int             `gorm:"default:0" json:"item_count"`                      // 商品种类数量
	TotalQty    int             `gorm:"default:0" json:"total_qty"`                       // 商品总数量
	TotalAmount decimal.Decimal `gorm:"type:decimal(10,2);default:0" json:"total_amount"` // 总金额

	// 时间戳
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联关系
	User  *User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Items []CartItem `gorm:"foreignKey:CartID" json:"items,omitempty"`
}

// CartItem 购物车商品项模型
type CartItem struct {
	ID        uint            `gorm:"primarykey" json:"id"`
	CartID    uint            `gorm:"not null;index" json:"cart_id"`
	ProductID uint            `gorm:"not null;index" json:"product_id"`
	SKUID     uint            `gorm:"index" json:"sku_id"`                      // SKU ID，0表示使用商品默认规格
	Quantity  int             `gorm:"not null;default:1" json:"quantity"`       // 数量
	Price     decimal.Decimal `gorm:"type:decimal(10,2);not null" json:"price"` // 加入购物车时的价格

	// 商品快照信息（避免商品信息变更影响购物车显示）
	ProductName  string `gorm:"size:255;not null" json:"product_name"`
	ProductImage string `gorm:"size:500" json:"product_image"`
	SKUName      string `gorm:"size:255" json:"sku_name"`
	SKUImage     string `gorm:"size:500" json:"sku_image"`
	SKUAttrs     string `gorm:"type:json" json:"sku_attrs"` // SKU属性JSON

	// 状态信息
	Selected bool   `gorm:"default:true" json:"selected"`                 // 是否选中
	Status   string `gorm:"size:20;default:'normal';index" json:"status"` // 商品状态：normal/invalid/out_of_stock

	// 时间戳
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联关系
	Cart    *Cart       `gorm:"foreignKey:CartID" json:"cart,omitempty"`
	Product *Product    `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	SKU     *ProductSKU `gorm:"foreignKey:SKUID" json:"sku,omitempty"`
}

// CartSummary 购物车汇总信息
type CartSummary struct {
	ItemCount      int             `json:"item_count"`      // 商品种类数量
	TotalQty       int             `json:"total_qty"`       // 商品总数量
	SelectedCount  int             `json:"selected_count"`  // 选中商品种类数量
	SelectedQty    int             `json:"selected_qty"`    // 选中商品总数量
	TotalAmount    decimal.Decimal `json:"total_amount"`    // 总金额
	SelectedAmount decimal.Decimal `json:"selected_amount"` // 选中商品总金额
	DiscountAmount decimal.Decimal `json:"discount_amount"` // 优惠金额
	ShippingFee    decimal.Decimal `json:"shipping_fee"`    // 运费
	FinalAmount    decimal.Decimal `json:"final_amount"`    // 最终金额
	InvalidItems   []CartItem      `json:"invalid_items"`   // 失效商品列表
}

// CartMergeLog 购物车合并日志
type CartMergeLog struct {
	ID            uint      `gorm:"primarykey" json:"id"`
	UserID        uint      `gorm:"not null;index" json:"user_id"`
	SessionID     string    `gorm:"size:100;not null" json:"session_id"`
	GuestCartID   uint      `gorm:"not null" json:"guest_cart_id"`   // 游客购物车ID
	UserCartID    uint      `gorm:"not null" json:"user_cart_id"`    // 用户购物车ID
	MergedItems   int       `gorm:"default:0" json:"merged_items"`   // 合并的商品数量
	ConflictItems int       `gorm:"default:0" json:"conflict_items"` // 冲突商品数量
	Status        string    `gorm:"size:20;default:'success'" json:"status"`
	CreatedAt     time.Time `json:"created_at"`

	// 关联关系
	User      *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
	GuestCart *Cart `gorm:"foreignKey:GuestCartID" json:"guest_cart,omitempty"`
	UserCart  *Cart `gorm:"foreignKey:UserCartID" json:"user_cart,omitempty"`
}

// TableName 指定表名
func (Cart) TableName() string {
	return "carts"
}

func (CartItem) TableName() string {
	return "cart_items"
}

func (CartMergeLog) TableName() string {
	return "cart_merge_logs"
}

// 购物车状态常量
const (
	CartStatusActive   = "active"   // 活跃
	CartStatusInactive = "inactive" // 非活跃
	CartStatusMerged   = "merged"   // 已合并
	CartStatusExpired  = "expired"  // 已过期
)

// 购物车商品状态常量
const (
	CartItemStatusNormal      = "normal"       // 正常
	CartItemStatusInvalid     = "invalid"      // 失效（商品已删除或下架）
	CartItemStatusOutOfStock  = "out_of_stock" // 库存不足
	CartItemStatusPriceChange = "price_change" // 价格变动
)

// 购物车合并状态常量
const (
	CartMergeStatusSuccess = "success" // 合并成功
	CartMergeStatusFailed  = "failed"  // 合并失败
)

// 购物车方法
func (c *Cart) IsActive() bool {
	return c.Status == CartStatusActive
}

func (c *Cart) IsUserCart() bool {
	return c.UserID > 0
}

func (c *Cart) IsGuestCart() bool {
	return c.UserID == 0 && c.SessionID != ""
}

func (c *Cart) GetIdentifier() string {
	if c.IsUserCart() {
		return fmt.Sprintf("user_%d", c.UserID)
	}
	return fmt.Sprintf("guest_%s", c.SessionID)
}

// 计算购物车总金额
func (c *Cart) CalculateTotalAmount() decimal.Decimal {
	total := decimal.Zero
	for _, item := range c.Items {
		if item.Selected && item.Status == CartItemStatusNormal {
			itemTotal := item.Price.Mul(decimal.NewFromInt(int64(item.Quantity)))
			total = total.Add(itemTotal)
		}
	}
	return total
}

// 计算选中商品数量
func (c *Cart) GetSelectedItemCount() int {
	count := 0
	for _, item := range c.Items {
		if item.Selected && item.Status == CartItemStatusNormal {
			count++
		}
	}
	return count
}

// 计算选中商品总数量
func (c *Cart) GetSelectedTotalQty() int {
	qty := 0
	for _, item := range c.Items {
		if item.Selected && item.Status == CartItemStatusNormal {
			qty += item.Quantity
		}
	}
	return qty
}

// 获取失效商品列表
func (c *Cart) GetInvalidItems() []CartItem {
	var invalidItems []CartItem
	for _, item := range c.Items {
		if item.Status != CartItemStatusNormal {
			invalidItems = append(invalidItems, item)
		}
	}
	return invalidItems
}

// 购物车商品方法
func (ci *CartItem) IsValid() bool {
	return ci.Status == CartItemStatusNormal
}

func (ci *CartItem) IsSelected() bool {
	return ci.Selected && ci.IsValid()
}

func (ci *CartItem) GetTotalPrice() decimal.Decimal {
	return ci.Price.Mul(decimal.NewFromInt(int64(ci.Quantity)))
}

func (ci *CartItem) GetDisplayName() string {
	if ci.SKUName != "" {
		return fmt.Sprintf("%s - %s", ci.ProductName, ci.SKUName)
	}
	return ci.ProductName
}

func (ci *CartItem) GetDisplayImage() string {
	if ci.SKUImage != "" {
		return ci.SKUImage
	}
	return ci.ProductImage
}

// 购物车请求结构体
type AddToCartRequest struct {
	ProductID uint `json:"product_id" binding:"required"`
	SKUID     uint `json:"sku_id"`
	Quantity  int  `json:"quantity" binding:"required,min=1"`
}

type UpdateCartItemRequest struct {
	Quantity int  `json:"quantity" binding:"required,min=1"`
	Selected bool `json:"selected"`
}

type BatchUpdateCartRequest struct {
	Items []struct {
		ID       uint `json:"id" binding:"required"`
		Quantity int  `json:"quantity" binding:"min=1"`
		Selected bool `json:"selected"`
	} `json:"items" binding:"required,min=1"`
}

type CartListRequest struct {
	IncludeInvalid bool `form:"include_invalid"` // 是否包含失效商品
}

// 购物车响应结构体
type CartResponse struct {
	Cart    *Cart        `json:"cart"`
	Summary *CartSummary `json:"summary"`
}

type CartItemResponse struct {
	*CartItem
	CurrentPrice    decimal.Decimal `json:"current_price"`    // 当前价格
	PriceChanged    bool            `json:"price_changed"`    // 价格是否变动
	StockAvailable  int             `json:"stock_available"`  // 可用库存
	StockSufficient bool            `json:"stock_sufficient"` // 库存是否充足
}

// 购物车统计请求
type CartStatsRequest struct {
	UserID    uint   `form:"user_id"`
	SessionID string `form:"session_id"`
	DateFrom  string `form:"date_from"`
	DateTo    string `form:"date_to"`
}

// 购物车统计响应
type CartStatsResponse struct {
	TotalCarts     int64           `json:"total_carts"`
	ActiveCarts    int64           `json:"active_carts"`
	AbandonedCarts int64           `json:"abandoned_carts"`
	AverageItems   float64         `json:"average_items"`
	AverageAmount  decimal.Decimal `json:"average_amount"`
	ConversionRate float64         `json:"conversion_rate"`
	TopProducts    []ProductStats  `json:"top_products"`
	CartsByHour    []HourlyStats   `json:"carts_by_hour"`
}

type ProductStats struct {
	ProductID   uint   `json:"product_id"`
	ProductName string `json:"product_name"`
	AddCount    int64  `json:"add_count"`
	TotalQty    int64  `json:"total_qty"`
}

type HourlyStats struct {
	Hour  int   `json:"hour"`
	Count int64 `json:"count"`
}
