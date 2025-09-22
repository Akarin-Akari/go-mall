package main

import (
	"fmt"
	"log"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Product 商品模型
type Product struct {
	ID          uint            `gorm:"primarykey" json:"id"`
	Name        string          `gorm:"size:255;not null" json:"name"`
	Description string          `gorm:"type:text" json:"description"`
	Price       decimal.Decimal `gorm:"type:decimal(10,2);not null" json:"price"`
	Stock       int             `gorm:"not null;default:0" json:"stock"`
	Status      string          `gorm:"size:20;default:'active';index" json:"status"`
	CategoryID  uint            `gorm:"index" json:"category_id"`
	BrandID     uint            `gorm:"index" json:"brand_id"`
	Images      string          `gorm:"type:json" json:"images"`
	Attributes  string          `gorm:"type:json" json:"attributes"`
	SoldCount   int             `gorm:"default:0" json:"sold_count"`
	ViewCount   int             `gorm:"default:0" json:"view_count"`
	Weight      decimal.Decimal `gorm:"type:decimal(8,3);default:0" json:"weight"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// Category 分类模型
type Category struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	Name        string    `gorm:"size:100;not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	ParentID    uint      `gorm:"default:0;index" json:"parent_id"`
	Sort        int       `gorm:"default:0" json:"sort"`
	Status      string    `gorm:"size:20;default:'active';index" json:"status"`
	Icon        string    `gorm:"size:255" json:"icon"`
	Image       string    `gorm:"size:255" json:"image"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Brand 品牌模型
type Brand struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	Name        string    `gorm:"size:100;not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	Logo        string    `gorm:"size:255" json:"logo"`
	Website     string    `gorm:"size:255" json:"website"`
	Status      string    `gorm:"size:20;default:'active';index" json:"status"`
	Sort        int       `gorm:"default:0" json:"sort"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func main() {
	fmt.Println("🚀 开始初始化商品数据...")

	// 连接数据库 - 使用mall-go项目的数据库
	db, err := gorm.Open(sqlite.Open("mall-go/mall.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("连接数据库失败:", err)
	}

	// 自动迁移表结构
	fmt.Println("📋 创建数据表...")
	err = db.AutoMigrate(&Category{}, &Brand{}, &Product{})
	if err != nil {
		log.Fatal("迁移数据表失败:", err)
	}

	// 初始化分类数据
	fmt.Println("📂 初始化分类数据...")
	categories := []Category{
		{
			ID:          1,
			Name:        "电子产品",
			Description: "各种电子设备和数码产品",
			Status:      "active",
			Sort:        1,
		},
		{
			ID:          2,
			Name:        "服装鞋帽",
			Description: "时尚服装和配饰",
			Status:      "active",
			Sort:        2,
		},
		{
			ID:          3,
			Name:        "家居用品",
			Description: "家庭生活用品",
			Status:      "active",
			Sort:        3,
		},
	}

	for _, category := range categories {
		var existingCategory Category
		if err := db.First(&existingCategory, category.ID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				db.Create(&category)
				fmt.Printf("   ✅ 创建分类: %s\n", category.Name)
			}
		} else {
			fmt.Printf("   ⚠️ 分类已存在: %s\n", category.Name)
		}
	}

	// 初始化品牌数据
	fmt.Println("🏷️ 初始化品牌数据...")
	brands := []Brand{
		{
			ID:          1,
			Name:        "Apple",
			Description: "苹果公司产品",
			Status:      "active",
			Sort:        1,
		},
		{
			ID:          2,
			Name:        "Nike",
			Description: "耐克运动品牌",
			Status:      "active",
			Sort:        2,
		},
		{
			ID:          3,
			Name:        "IKEA",
			Description: "宜家家居",
			Status:      "active",
			Sort:        3,
		},
	}

	for _, brand := range brands {
		var existingBrand Brand
		if err := db.First(&existingBrand, brand.ID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				db.Create(&brand)
				fmt.Printf("   ✅ 创建品牌: %s\n", brand.Name)
			}
		} else {
			fmt.Printf("   ⚠️ 品牌已存在: %s\n", brand.Name)
		}
	}

	// 初始化商品数据
	fmt.Println("📦 初始化商品数据...")
	products := []Product{
		{
			ID:          1,
			Name:        "iPhone 15 Pro",
			Description: "苹果最新款智能手机，配备A17 Pro芯片",
			Price:       decimal.NewFromFloat(8999.00),
			Stock:       100,
			Status:      "active",
			CategoryID:  1,
			BrandID:     1,
			Images:      `["https://example.com/iphone15pro.jpg"]`,
			Attributes:  `{"color": "深空黑", "storage": "256GB", "screen": "6.1英寸"}`,
			Weight:      decimal.NewFromFloat(0.187),
		},
		{
			ID:          2,
			Name:        "MacBook Pro 14",
			Description: "专业级笔记本电脑，搭载M3芯片",
			Price:       decimal.NewFromFloat(14999.00),
			Stock:       50,
			Status:      "active",
			CategoryID:  1,
			BrandID:     1,
			Images:      `["https://example.com/macbookpro14.jpg"]`,
			Attributes:  `{"color": "深空灰", "memory": "16GB", "storage": "512GB SSD"}`,
			Weight:      decimal.NewFromFloat(1.6),
		},
		{
			ID:          3,
			Name:        "Nike Air Max 270",
			Description: "经典运动鞋，舒适透气",
			Price:       decimal.NewFromFloat(899.00),
			Stock:       200,
			Status:      "active",
			CategoryID:  2,
			BrandID:     2,
			Images:      `["https://example.com/airmax270.jpg"]`,
			Attributes:  `{"color": "黑白", "size": "42", "material": "网布+合成革"}`,
			Weight:      decimal.NewFromFloat(0.8),
		},
		{
			ID:          4,
			Name:        "IKEA BILLY书架",
			Description: "简约现代书架，多种颜色可选",
			Price:       decimal.NewFromFloat(199.00),
			Stock:       150,
			Status:      "active",
			CategoryID:  3,
			BrandID:     3,
			Images:      `["https://example.com/billy-bookshelf.jpg"]`,
			Attributes:  `{"color": "白色", "material": "刨花板", "dimensions": "80x28x202cm"}`,
			Weight:      decimal.NewFromFloat(25.5),
		},
		{
			ID:          5,
			Name:        "AirPods Pro 2",
			Description: "主动降噪无线耳机",
			Price:       decimal.NewFromFloat(1899.00),
			Stock:       80,
			Status:      "active",
			CategoryID:  1,
			BrandID:     1,
			Images:      `["https://example.com/airpods-pro2.jpg"]`,
			Attributes:  `{"color": "白色", "battery": "6小时+24小时", "features": "主动降噪"}`,
			Weight:      decimal.NewFromFloat(0.056),
		},
	}

	for _, product := range products {
		var existingProduct Product
		if err := db.First(&existingProduct, product.ID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				db.Create(&product)
				fmt.Printf("   ✅ 创建商品: %s (¥%.2f)\n", product.Name, product.Price)
			}
		} else {
			fmt.Printf("   ⚠️ 商品已存在: %s\n", product.Name)
		}
	}

	fmt.Println("✅ 商品数据初始化完成！")
	fmt.Println()
	fmt.Println("📊 数据统计:")

	var categoryCount, brandCount, productCount int64
	db.Model(&Category{}).Count(&categoryCount)
	db.Model(&Brand{}).Count(&brandCount)
	db.Model(&Product{}).Count(&productCount)

	fmt.Printf("   分类数量: %d\n", categoryCount)
	fmt.Printf("   品牌数量: %d\n", brandCount)
	fmt.Printf("   商品数量: %d\n", productCount)

	fmt.Println()
	fmt.Println("🎉 初始化完成！现在可以测试API了。")
}
