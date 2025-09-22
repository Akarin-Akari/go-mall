package main

import (
	"fmt"
	"log"
	"mall-go/internal/config"
	"mall-go/internal/model"
	"mall-go/pkg/database"

	"github.com/shopspring/decimal"
)

func main() {
	fmt.Println("🚀 开始初始化商品数据...")

	// 初始化配置
	config.Load()

	// 初始化数据库连接
	db := database.Init()
	if db == nil {
		log.Fatal("数据库连接失败")
	}

	// 清理现有数据
	fmt.Println("🧹 清理现有数据...")
	db.Exec("DELETE FROM products")
	db.Exec("DELETE FROM categories")
	db.Exec("DELETE FROM brands")

	// 初始化分类数据
	fmt.Println("📂 初始化分类数据...")
	categories := []model.Category{
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
		{
			ID:          4,
			Name:        "图书文具",
			Description: "书籍和办公用品",
			Status:      "active",
			Sort:        4,
		},
	}

	for _, category := range categories {
		if err := db.Create(&category).Error; err != nil {
			fmt.Printf("   ⚠️ 创建分类失败: %s - %v\n", category.Name, err)
		} else {
			fmt.Printf("   ✅ 创建分类: %s\n", category.Name)
		}
	}

	// 初始化品牌数据
	fmt.Println("🏷️ 初始化品牌数据...")
	brands := []model.Brand{
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
		{
			ID:          4,
			Name:        "小米",
			Description: "小米科技产品",
			Status:      "active",
			Sort:        4,
		},
	}

	for _, brand := range brands {
		if err := db.Create(&brand).Error; err != nil {
			fmt.Printf("   ⚠️ 创建品牌失败: %s - %v\n", brand.Name, err)
		} else {
			fmt.Printf("   ✅ 创建品牌: %s\n", brand.Name)
		}
	}

	// 初始化商品数据
	fmt.Println("📦 初始化商品数据...")
	products := []model.Product{
		{
			ID:          1,
			Name:        "iPhone 15 Pro",
			Description: "苹果最新款智能手机，配备A17 Pro芯片",
			Price:       decimal.NewFromFloat(8999.00),
			Stock:       100,
			Status:      model.ProductStatusActive,
			CategoryID:  1,
			BrandID:     1,
			MerchantID:  1, // 添加商家ID
			Weight:      decimal.NewFromFloat(0.187),
		},
		{
			ID:          2,
			Name:        "MacBook Pro 14",
			Description: "专业级笔记本电脑，搭载M3芯片",
			Price:       decimal.NewFromFloat(14999.00),
			Stock:       50,
			Status:      model.ProductStatusActive,
			CategoryID:  1,
			BrandID:     1,
			MerchantID:  1,
			Weight:      decimal.NewFromFloat(1.6),
		},
		{
			ID:          3,
			Name:        "Nike Air Max 270",
			Description: "经典运动鞋，舒适透气",
			Price:       decimal.NewFromFloat(899.00),
			Stock:       200,
			Status:      model.ProductStatusActive,
			CategoryID:  2,
			BrandID:     2,
			MerchantID:  1,
			Weight:      decimal.NewFromFloat(0.8),
		},
		{
			ID:          4,
			Name:        "IKEA BILLY书架",
			Description: "简约现代书架，多种颜色可选",
			Price:       decimal.NewFromFloat(199.00),
			Stock:       150,
			Status:      model.ProductStatusActive,
			CategoryID:  3,
			BrandID:     3,
			MerchantID:  1,
			Weight:      decimal.NewFromFloat(25.5),
		},
		{
			ID:          5,
			Name:        "AirPods Pro 2",
			Description: "主动降噪无线耳机",
			Price:       decimal.NewFromFloat(1899.00),
			Stock:       80,
			Status:      model.ProductStatusActive,
			CategoryID:  1,
			BrandID:     1,
			MerchantID:  1,
			Weight:      decimal.NewFromFloat(0.056),
		},
		{
			ID:          6,
			Name:        "小米13 Pro",
			Description: "小米旗舰手机，徕卡影像",
			Price:       decimal.NewFromFloat(3999.00),
			Stock:       120,
			Status:      model.ProductStatusActive,
			CategoryID:  1,
			BrandID:     4,
			MerchantID:  1,
			Weight:      decimal.NewFromFloat(0.210),
		},
		{
			ID:          7,
			Name:        "Nike Dunk Low",
			Description: "经典板鞋，街头时尚",
			Price:       decimal.NewFromFloat(699.00),
			Stock:       180,
			Status:      model.ProductStatusActive,
			CategoryID:  2,
			BrandID:     2,
			MerchantID:  1,
			Weight:      decimal.NewFromFloat(0.9),
		},
		{
			ID:          8,
			Name:        "IKEA POÄNG扶手椅",
			Description: "舒适休闲椅，北欧设计",
			Price:       decimal.NewFromFloat(599.00),
			Stock:       60,
			Status:      model.ProductStatusActive,
			CategoryID:  3,
			BrandID:     3,
			MerchantID:  1,
			Weight:      decimal.NewFromFloat(12.0),
		},
		{
			ID:          9,
			Name:        "编程珠玑",
			Description: "经典编程思维训练书籍",
			Price:       decimal.NewFromFloat(89.00),
			Stock:       200,
			Status:      model.ProductStatusActive,
			CategoryID:  4,
			BrandID:     4,
			MerchantID:  1,
			Weight:      decimal.NewFromFloat(0.3),
		},
		{
			ID:          10,
			Name:        "无线充电器",
			Description: "支持快充的无线充电板",
			Price:       decimal.NewFromFloat(199.00),
			Stock:       300,
			Status:      model.ProductStatusActive,
			CategoryID:  1,
			BrandID:     4,
			MerchantID:  1,
			Weight:      decimal.NewFromFloat(0.2),
		},
	}

	for _, product := range products {
		if err := db.Create(&product).Error; err != nil {
			fmt.Printf("   ⚠️ 创建商品失败: %s - %v\n", product.Name, err)
		} else {
			fmt.Printf("   ✅ 创建商品: %s (¥%.2f)\n", product.Name, product.Price)
		}
	}

	fmt.Println("✅ 商品数据初始化完成！")
	fmt.Println()
	fmt.Println("📊 数据统计:")

	var categoryCount, brandCount, productCount int64
	db.Model(&model.Category{}).Count(&categoryCount)
	db.Model(&model.Brand{}).Count(&brandCount)
	db.Model(&model.Product{}).Count(&productCount)

	fmt.Printf("   分类数量: %d\n", categoryCount)
	fmt.Printf("   品牌数量: %d\n", brandCount)
	fmt.Printf("   商品数量: %d\n", productCount)

	fmt.Println()
	fmt.Println("🎉 初始化完成！现在可以测试API了。")
}
