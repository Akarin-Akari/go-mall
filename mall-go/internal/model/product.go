package model

import (
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// Category 商品分类
type Category struct {
	ID          uint   `gorm:"primarykey" json:"id"`
	Name        string `gorm:"not null;size:100;index" json:"name"`
	Description string `gorm:"size:500" json:"description"`
	ParentID    uint   `gorm:"default:0;index" json:"parent_id"`
	Level       int    `gorm:"default:1" json:"level"`
	Path        string `gorm:"size:500" json:"path"`  // 分类路径
	Icon        string `gorm:"size:255" json:"icon"`  // 分类图标
	Image       string `gorm:"size:255" json:"image"` // 分类图片
	Sort        int    `gorm:"default:0;index" json:"sort"`
	Status      string `gorm:"default:'active';size:20;index" json:"status"`

	// SEO信息
	SEOTitle       string `gorm:"size:255" json:"seo_title"`
	SEOKeywords    string `gorm:"size:500" json:"seo_keywords"`
	SEODescription string `gorm:"size:500" json:"seo_description"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联关系
	Parent   *Category  `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Children []Category `gorm:"foreignKey:ParentID" json:"children,omitempty"`
	Products []Product  `gorm:"foreignKey:CategoryID" json:"products,omitempty"`
}

// Product 商品模型
type Product struct {
	ID          uint   `gorm:"primarykey" json:"id"`
	Name        string `gorm:"not null;size:255;index" json:"name"`
	SubTitle    string `gorm:"size:500" json:"sub_title"`
	Description string `gorm:"type:text" json:"description"`
	Detail      string `gorm:"type:longtext" json:"detail"`
	CategoryID  uint   `gorm:"not null;index" json:"category_id"`
	BrandID     uint   `gorm:"index" json:"brand_id"`
	MerchantID  uint   `gorm:"not null;index" json:"merchant_id"`

	// 价格信息
	Price       decimal.Decimal `gorm:"type:decimal(10,2);not null" json:"price"`
	OriginPrice decimal.Decimal `gorm:"type:decimal(10,2)" json:"origin_price"`
	CostPrice   decimal.Decimal `gorm:"type:decimal(10,2)" json:"cost_price"`

	// 库存信息
	Stock     int `gorm:"not null;default:0" json:"stock"`
	MinStock  int `gorm:"default:0" json:"min_stock"`
	MaxStock  int `gorm:"default:0" json:"max_stock"`
	SoldCount int `gorm:"default:0" json:"sold_count"`
	Version   int `gorm:"not null;default:0" json:"version"` // 乐观锁版本号

	// 商品属性
	Weight decimal.Decimal `gorm:"type:decimal(8,3)" json:"weight"`
	Volume decimal.Decimal `gorm:"type:decimal(8,3)" json:"volume"`
	Unit   string          `gorm:"size:20" json:"unit"`

	// 状态信息
	Status      string `gorm:"size:20;default:'draft';index" json:"status"`
	IsHot       bool   `gorm:"default:false" json:"is_hot"`
	IsNew       bool   `gorm:"default:false" json:"is_new"`
	IsRecommend bool   `gorm:"default:false" json:"is_recommend"`

	// SEO信息
	SEOTitle       string `gorm:"size:255" json:"seo_title"`
	SEOKeywords    string `gorm:"size:500" json:"seo_keywords"`
	SEODescription string `gorm:"size:500" json:"seo_description"`

	// 排序和显示
	Sort      int `gorm:"default:0;index" json:"sort"`
	ViewCount int `gorm:"default:0" json:"view_count"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联关系
	Category   *Category       `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Brand      *Brand          `gorm:"foreignKey:BrandID" json:"brand,omitempty"`
	Merchant   *User           `gorm:"foreignKey:MerchantID" json:"merchant,omitempty"`
	Images     []ProductImage  `gorm:"foreignKey:ProductID" json:"images,omitempty"`
	Attributes []ProductAttr   `gorm:"foreignKey:ProductID" json:"attributes,omitempty"`
	SKUs       []ProductSKU    `gorm:"foreignKey:ProductID" json:"skus,omitempty"`
	Reviews    []ProductReview `gorm:"foreignKey:ProductID" json:"reviews,omitempty"`
}

// ProductImage 商品图片
type ProductImage struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	ProductID uint           `gorm:"not null;index" json:"product_id"`
	URL       string         `gorm:"not null;size:500" json:"url"`
	Alt       string         `gorm:"size:255" json:"alt"`
	Sort      int            `gorm:"default:0" json:"sort"`
	IsMain    bool           `gorm:"default:false" json:"is_main"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联关系
	Product *Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}

// Brand 品牌模型
type Brand struct {
	ID          uint   `gorm:"primarykey" json:"id"`
	Name        string `gorm:"size:100;not null;uniqueIndex" json:"name"`
	EnglishName string `gorm:"size:100" json:"english_name"`
	Logo        string `gorm:"size:255" json:"logo"`
	Description string `gorm:"type:text" json:"description"`
	Website     string `gorm:"size:255" json:"website"`
	Sort        int    `gorm:"default:0;index" json:"sort"`
	Status      string `gorm:"size:20;default:'active';index" json:"status"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联关系
	Products []Product `gorm:"foreignKey:BrandID" json:"products,omitempty"`
}

// ProductAttr 商品属性模型
type ProductAttr struct {
	ID        uint   `gorm:"primarykey" json:"id"`
	ProductID uint   `gorm:"not null;index" json:"product_id"`
	AttrName  string `gorm:"size:100;not null" json:"attr_name"`
	AttrValue string `gorm:"size:255;not null" json:"attr_value"`
	Sort      int    `gorm:"default:0" json:"sort"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联关系
	Product *Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}

// ProductSKU 商品SKU模型
type ProductSKU struct {
	ID        uint            `gorm:"primarykey" json:"id"`
	ProductID uint            `gorm:"not null;index" json:"product_id"`
	SKUCode   string          `gorm:"size:100;not null;uniqueIndex" json:"sku_code"`
	Name      string          `gorm:"size:255;not null" json:"name"`
	Price     decimal.Decimal `gorm:"type:decimal(10,2);not null" json:"price"`
	Stock     int             `gorm:"not null;default:0" json:"stock"`
	Version   int             `gorm:"not null;default:0" json:"version"` // 乐观锁版本号
	Image     string          `gorm:"size:500" json:"image"`
	Weight    decimal.Decimal `gorm:"type:decimal(8,3)" json:"weight"`
	Volume    decimal.Decimal `gorm:"type:decimal(8,3)" json:"volume"`
	Status    string          `gorm:"size:20;default:'active';index" json:"status"`

	// 规格属性（JSON格式存储）
	Attributes string `gorm:"type:json" json:"attributes"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联关系
	Product *Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}

// ProductReview 商品评价模型
type ProductReview struct {
	ID          uint   `gorm:"primarykey" json:"id"`
	ProductID   uint   `gorm:"not null;index" json:"product_id"`
	UserID      uint   `gorm:"not null;index" json:"user_id"`
	OrderID     uint   `gorm:"index" json:"order_id"`
	Rating      int    `gorm:"not null;default:5" json:"rating"`
	Content     string `gorm:"type:text" json:"content"`
	Images      string `gorm:"type:json" json:"images"`
	IsAnonymous bool   `gorm:"default:false" json:"is_anonymous"`
	Status      string `gorm:"size:20;default:'pending';index" json:"status"`

	// 商家回复
	ReplyContent string     `gorm:"type:text" json:"reply_content"`
	ReplyTime    *time.Time `gorm:"null" json:"reply_time"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联关系
	Product *Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	User    *User    `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName 指定表名
func (Product) TableName() string {
	return "products"
}

func (Category) TableName() string {
	return "categories"
}

func (ProductImage) TableName() string {
	return "product_images"
}

// ProductCreateRequest 创建商品请求
type ProductCreateRequest struct {
	Name        string   `json:"name" binding:"required,min=2,max=200"`
	Description string   `json:"description" binding:"required"`
	Price       float64  `json:"price" binding:"required,min=0"`
	Stock       int      `json:"stock" binding:"required,min=0"`
	CategoryID  uint     `json:"category_id" binding:"required"`
	Images      []string `json:"images"`
}

// ProductUpdateRequest 更新商品请求
type ProductUpdateRequest struct {
	Name        string   `json:"name" binding:"required,min=2,max=200"`
	Description string   `json:"description" binding:"required"`
	Price       float64  `json:"price" binding:"required,min=0"`
	Stock       int      `json:"stock" binding:"required,min=0"`
	CategoryID  uint     `json:"category_id" binding:"required"`
	Images      []string `json:"images"`
}

// ProductListRequest 商品列表请求
type ProductListRequest struct {
	Page       int    `form:"page" binding:"min=1"`
	PageSize   int    `form:"page_size" binding:"min=1,max=100"`
	CategoryID *uint  `form:"category_id"`
	Keyword    string `form:"keyword"`
	Status     string `form:"status"`
}

// TableName 方法
func (Brand) TableName() string {
	return "brands"
}

func (ProductAttr) TableName() string {
	return "product_attrs"
}

func (ProductSKU) TableName() string {
	return "product_skus"
}

func (ProductReview) TableName() string {
	return "product_reviews"
}

// 商品状态常量
const (
	ProductStatusDraft    = "draft"    // 草稿
	ProductStatusActive   = "active"   // 上架
	ProductStatusInactive = "inactive" // 下架
	ProductStatusDeleted  = "deleted"  // 已删除
)

// 分类状态常量
const (
	CategoryStatusActive   = "active"   // 启用
	CategoryStatusInactive = "inactive" // 禁用
)

// 品牌状态常量
const (
	BrandStatusActive   = "active"   // 启用
	BrandStatusInactive = "inactive" // 禁用
)

// SKU状态常量
const (
	SKUStatusActive   = "active"   // 启用
	SKUStatusInactive = "inactive" // 禁用
)

// 评价状态常量
const (
	ReviewStatusPending  = "pending"  // 待审核
	ReviewStatusApproved = "approved" // 已通过
	ReviewStatusRejected = "rejected" // 已拒绝
)

// 商品方法
func (p *Product) IsActive() bool {
	return p.Status == ProductStatusActive
}

func (p *Product) IsInStock() bool {
	return p.Stock > 0
}

func (p *Product) GetMainImage() string {
	for _, image := range p.Images {
		if image.IsMain {
			return image.URL
		}
	}
	if len(p.Images) > 0 {
		return p.Images[0].URL
	}
	return ""
}

func (p *Product) GetAverageRating() float64 {
	if len(p.Reviews) == 0 {
		return 0
	}

	total := 0
	approvedCount := 0
	for _, review := range p.Reviews {
		if review.Status == ReviewStatusApproved {
			total += review.Rating
			approvedCount++
		}
	}

	if approvedCount == 0 {
		return 0
	}

	return float64(total) / float64(approvedCount)
}

// 分类方法
func (c *Category) IsRoot() bool {
	return c.ParentID == 0
}

func (c *Category) GetFullPath() string {
	if c.Path != "" {
		return c.Path
	}
	return c.Name
}
