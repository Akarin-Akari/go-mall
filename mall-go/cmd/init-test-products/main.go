package main

import (
	"fmt"
	"log"
	"mall-go/internal/config"
	"mall-go/internal/model"
	"mall-go/pkg/database"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

func main() {
	fmt.Println("🚀 开始初始化Mall-Go测试商品数据...")

	// 初始化配置
	cfg := config.Config{
		Database: config.DatabaseConfig{
			Driver: "sqlite",
			DBName: "mall.db",
		},
	}
	config.GlobalConfig = cfg

	// 初始化数据库
	db := database.Init()
	if db == nil {
		log.Fatal("数据库初始化失败")
	}

	// 自动迁移表结构
	err := db.AutoMigrate(
		&model.Category{},
		&model.Product{},
		&model.ProductImage{},
	)
	if err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}

	// 创建测试数据
	if err := createTestData(db); err != nil {
		log.Fatalf("创建测试数据失败: %v", err)
	}

	fmt.Println("✅ 测试商品数据初始化完成！")
}

func createTestData(db *gorm.DB) error {
	// 1. 创建商品分类
	categories := []model.Category{
		{
			Name:        "电子产品",
			Description: "手机、电脑、数码产品等",
			Sort:        1,
			Status:      "active",
		},
		{
			Name:        "服装鞋帽",
			Description: "男装、女装、鞋子、配饰等",
			Sort:        2,
			Status:      "active",
		},
		{
			Name:        "家居用品",
			Description: "家具、装饰、生活用品等",
			Sort:        3,
			Status:      "active",
		},
		{
			Name:        "图书文具",
			Description: "图书、文具、办公用品等",
			Sort:        4,
			Status:      "active",
		},
		{
			Name:        "食品饮料",
			Description: "零食、饮料、生鲜食品等",
			Sort:        5,
			Status:      "active",
		},
	}

	for _, category := range categories {
		var existingCategory model.Category
		if err := db.Where("name = ?", category.Name).First(&existingCategory).Error; err == gorm.ErrRecordNotFound {
			if err := db.Create(&category).Error; err != nil {
				return fmt.Errorf("创建分类失败: %v", err)
			}
			fmt.Printf("✅ 创建分类: %s\n", category.Name)
		} else {
			fmt.Printf("⚠️ 分类已存在: %s\n", category.Name)
		}
	}

	// 2. 获取分类ID
	var categoryMap = make(map[string]uint)
	for _, category := range categories {
		var cat model.Category
		db.Where("name = ?", category.Name).First(&cat)
		categoryMap[category.Name] = cat.ID
	}

	// 3. 创建测试商品
	products := []struct {
		Name        string
		Description string
		Price       float64
		Stock       int
		CategoryKey string
		Images      []string
	}{
		{
			Name:        "iPhone 15 Pro",
			Description: "苹果最新旗舰手机，搭载A17 Pro芯片",
			Price:       8999.00,
			Stock:       50,
			CategoryKey: "电子产品",
			Images:      []string{"https://example.com/iphone15pro.jpg"},
		},
		{
			Name:        "MacBook Pro 14英寸",
			Description: "苹果笔记本电脑，M3芯片，16GB内存",
			Price:       15999.00,
			Stock:       30,
			CategoryKey: "电子产品",
			Images:      []string{"https://example.com/macbook-pro.jpg"},
		},
		{
			Name:        "iPad Air",
			Description: "轻薄便携的平板电脑，适合办公和娱乐",
			Price:       4399.00,
			Stock:       40,
			CategoryKey: "电子产品",
			Images:      []string{"https://example.com/ipad-air.jpg"},
		},
		{
			Name:        "AirPods Pro",
			Description: "主动降噪无线耳机，音质出色",
			Price:       1899.00,
			Stock:       100,
			CategoryKey: "电子产品",
			Images:      []string{"https://example.com/airpods-pro.jpg"},
		},
		{
			Name:        "Nike Air Max 270",
			Description: "时尚运动鞋，舒适透气",
			Price:       899.00,
			Stock:       80,
			CategoryKey: "服装鞋帽",
			Images:      []string{"https://example.com/nike-air-max.jpg"},
		},
		{
			Name:        "Adidas三叶草卫衣",
			Description: "经典款卫衣，舒适保暖",
			Price:       399.00,
			Stock:       60,
			CategoryKey: "服装鞋帽",
			Images:      []string{"https://example.com/adidas-hoodie.jpg"},
		},
		{
			Name:        "宜家书桌",
			Description: "简约现代书桌，适合家庭办公",
			Price:       599.00,
			Stock:       25,
			CategoryKey: "家居用品",
			Images:      []string{"https://example.com/ikea-desk.jpg"},
		},
		{
			Name:        "小米台灯",
			Description: "护眼台灯，可调节亮度和色温",
			Price:       199.00,
			Stock:       120,
			CategoryKey: "家居用品",
			Images:      []string{"https://example.com/xiaomi-lamp.jpg"},
		},
		{
			Name:        "《深入理解计算机系统》",
			Description: "计算机科学经典教材",
			Price:       139.00,
			Stock:       200,
			CategoryKey: "图书文具",
			Images:      []string{"https://example.com/csapp-book.jpg"},
		},
		{
			Name:        "晨光文具套装",
			Description: "学生文具套装，包含笔、橡皮、尺子等",
			Price:       29.90,
			Stock:       300,
			CategoryKey: "图书文具",
			Images:      []string{"https://example.com/stationery-set.jpg"},
		},
		{
			Name:        "三只松鼠坚果礼盒",
			Description: "精选坚果组合，健康美味",
			Price:       88.00,
			Stock:       150,
			CategoryKey: "食品饮料",
			Images:      []string{"https://example.com/nuts-gift-box.jpg"},
		},
		{
			Name:        "农夫山泉天然水",
			Description: "天然弱碱性水，24瓶装",
			Price:       45.00,
			Stock:       500,
			CategoryKey: "食品饮料",
			Images:      []string{"https://example.com/nongfu-water.jpg"},
		},
	}

	for _, productData := range products {
		var existingProduct model.Product
		if err := db.Where("name = ?", productData.Name).First(&existingProduct).Error; err == gorm.ErrRecordNotFound {
			// 创建商品
			product := model.Product{
				Name:        productData.Name,
				Description: productData.Description,
				Price:       decimal.NewFromFloat(productData.Price),
				Stock:       productData.Stock,
				CategoryID:  categoryMap[productData.CategoryKey],
				Status:      "active",
				MerchantID:  1, // 默认商家ID
			}

			if err := db.Create(&product).Error; err != nil {
				return fmt.Errorf("创建商品失败: %v", err)
			}

			// 创建商品图片
			for i, imageURL := range productData.Images {
				productImage := model.ProductImage{
					ProductID: product.ID,
					URL:       imageURL,
					Sort:      i,
					IsMain:    i == 0, // 第一张图片设为主图
				}
				if err := db.Create(&productImage).Error; err != nil {
					fmt.Printf("⚠️ 创建商品图片失败: %v\n", err)
				}
			}

			fmt.Printf("✅ 创建商品: %s (ID: %d)\n", product.Name, product.ID)
		} else {
			fmt.Printf("⚠️ 商品已存在: %s\n", productData.Name)
		}
	}

	return nil
}
