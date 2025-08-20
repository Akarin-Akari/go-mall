package model

import (
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// Category 商品分类
type Category struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	Name        string         `gorm:"not null;size:100" json:"name"`
	Description string         `gorm:"size:500" json:"description"`
	ParentID    *uint          `json:"parent_id"`
	Level       int            `gorm:"default:1" json:"level"`
	Sort        int            `gorm:"default:0" json:"sort"`
	Status      string         `gorm:"default:'active';size:20" json:"status"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// Product 商品模型
type Product struct {
	ID          uint            `gorm:"primarykey" json:"id"`
	Name        string          `gorm:"not null;size:200" json:"name"`
	Description string          `gorm:"type:text" json:"description"`
	Price       decimal.Decimal `gorm:"type:decimal(10,2);not null" json:"price"`
	Stock       int             `gorm:"default:0;not null" json:"stock"`
	CategoryID  uint            `gorm:"not null" json:"category_id"`
	Category    Category        `json:"category"`
	Images      []ProductImage  `json:"images"`
	Status      string          `gorm:"default:'active';size:20" json:"status"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	DeletedAt   gorm.DeletedAt  `gorm:"index" json:"-"`
}

// ProductImage 商品图片
type ProductImage struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	ProductID uint           `gorm:"not null" json:"product_id"`
	URL       string         `gorm:"not null;size:500" json:"url"`
	Sort      int            `gorm:"default:0" json:"sort"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
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
	Name        string  `json:"name" binding:"required,min=2,max=200"`
	Description string  `json:"description" binding:"required"`
	Price       float64 `json:"price" binding:"required,min=0"`
	Stock       int     `json:"stock" binding:"required,min=0"`
	CategoryID  uint    `json:"category_id" binding:"required"`
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
