//go:build ignore

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
	log.Println("🚀 开始初始化商品测试数据...")

	// 初始化配置和数据库
	config.Load()
	db := database.Init()
	if db == nil {
		log.Fatal("❌ 数据库初始化失败")
	}

	// 创建分类测试数据
	if err := createCategories(db); err != nil {
		log.Fatalf("❌ 创建分类数据失败: %v", err)
	}

	// 创建商品测试数据
	if err := createProducts(db); err != nil {
		log.Fatalf("❌ 创建商品数据失败: %v", err)
	}

	log.Println("✅ 商品测试数据初始化完成!")
}

// createCategories 创建分类测试数据
func createCategories(db *gorm.DB) error {
	log.Println("📋 创建商品分类数据...")

	categories := []model.Category{
		{
			Name:        "电子产品",
			Description: "各种电子设备和数码产品",
			Level:       1,
			Path:        "电子产品",
			Sort:        1,
			Status:      "active",
		},
		{
			Name:        "服装鞋帽",
			Description: "时尚服装、鞋子和帽子",
			Level:       1,
			Path:        "服装鞋帽",
			Sort:        2,
			Status:      "active",
		},
		{
			Name:        "家居用品",
			Description: "家庭生活用品和装饰",
			Level:       1,
			Path:        "家居用品",
			Sort:        3,
			Status:      "active",
		},
	}

	for _, category := range categories {
		var existing model.Category
		if err := db.Where("name = ?", category.Name).First(&existing).Error; err != nil {
			// 分类不存在，创建新分类
			if err := db.Create(&category).Error; err != nil {
				return fmt.Errorf("创建分类 %s 失败: %v", category.Name, err)
			}
			log.Printf("✅ 创建分类: %s", category.Name)
		} else {
			log.Printf("⚠️ 分类已存在: %s", category.Name)
		}
	}

	return nil
}

// createProducts 创建商品测试数据
func createProducts(db *gorm.DB) error {
	log.Println("📦 创建商品数据...")

	// 获取分类ID
	var categories []model.Category
	if err := db.Find(&categories).Error; err != nil {
		return fmt.Errorf("获取分类失败: %v", err)
	}

	if len(categories) == 0 {
		return fmt.Errorf("没有找到分类数据")
	}

	// 创建分类ID映射
	categoryMap := make(map[string]uint)
	for _, cat := range categories {
		categoryMap[cat.Name] = cat.ID
	}

	// 电子产品
	electronicProducts := []model.Product{
		{
			Name:        "iPhone 15 Pro",
			Description: "全新iPhone 15 Pro，采用钛金属设计，搭载A17 Pro芯片，拍照更清晰，性能更强劲",
			CategoryID:  categoryMap["电子产品"],
			Price:       decimal.NewFromFloat(8999.00),
			Stock:       100,
			Status:      "active",
		},
		{
			Name:        "MacBook Pro 14英寸",
			Description: "全新MacBook Pro 14英寸，搭载M3芯片，适合专业创作和开发工作",
			CategoryID:  categoryMap["电子产品"],
			Price:       decimal.NewFromFloat(14999.00),
			Stock:       50,
			Status:      "active",
		},
		{
			Name:        "iPad Air",
			Description: "iPad Air，轻薄设计，强劲性能，适合学习和娱乐",
			CategoryID:  categoryMap["电子产品"],
			Price:       decimal.NewFromFloat(4399.00),
			Stock:       80,
			Status:      "active",
		},
		{
			Name:        "AirPods Pro",
			Description: "AirPods Pro，主动降噪技术，空间音频体验，音质更清晰",
			CategoryID:  categoryMap["电子产品"],
			Price:       decimal.NewFromFloat(1899.00),
			Stock:       200,
			Status:      "active",
		},
		{
			Name:        "Apple Watch Series 9",
			Description: "Apple Watch Series 9，全面健康监测，智能生活助手",
			CategoryID:  categoryMap["电子产品"],
			Price:       decimal.NewFromFloat(2999.00),
			Stock:       120,
			Status:      "active",
		},
	}

	// 服装鞋帽
	clothingProducts := []model.Product{
		{
			Name:        "Nike Air Max 270",
			Description: "Nike Air Max 270运动鞋，舒适透气，适合日常运动和休闲穿着",
			CategoryID:  categoryMap["服装鞋帽"],
			Price:       decimal.NewFromFloat(899.00),
			Stock:       150,
			Status:      "active",
		},
		{
			Name:        "Adidas三叶草卫衣",
			Description: "Adidas三叶草经典卫衣，舒适面料，时尚设计，适合春秋季穿着",
			CategoryID:  categoryMap["服装鞋帽"],
			Price:       decimal.NewFromFloat(599.00),
			Stock:       200,
			Status:      "active",
		},
		{
			Name:        "Levi's 501牛仔裤",
			Description: "Levi's 501经典牛仔裤，优质面料，经典版型，百搭单品",
			CategoryID:  categoryMap["服装鞋帽"],
			Price:       decimal.NewFromFloat(799.00),
			Stock:       100,
			Status:      "active",
		},
		{
			Name:        "优衣库羽绒服",
			Description: "优衣库轻薄羽绒服，保暖性能优异，时尚设计，适合冬季穿着",
			CategoryID:  categoryMap["服装鞋帽"],
			Price:       decimal.NewFromFloat(499.00),
			Stock:       80,
			Status:      "active",
		},
		{
			Name:        "New Balance 990v5",
			Description: "New Balance 990v5复古跑鞋，经典设计，舒适脚感，适合日常穿着",
			CategoryID:  categoryMap["服装鞋帽"],
			Price:       decimal.NewFromFloat(1299.00),
			Stock:       60,
			Status:      "active",
		},
	}

	// 家居用品
	homeProducts := []model.Product{
		{
			Name:        "无印良品香薰机",
			Description: "无印良品超声波香薰机，简约设计，静音运行，营造舒适家居环境",
			CategoryID:  categoryMap["家居用品"],
			Price:       decimal.NewFromFloat(299.00),
			Stock:       100,
			Status:      "active",
		},
		{
			Name:        "宜家北欧风台灯",
			Description: "宜家北欧风格台灯，简约设计，护眼LED光源，适合阅读和工作",
			CategoryID:  categoryMap["家居用品"],
			Price:       decimal.NewFromFloat(199.00),
			Stock:       150,
			Status:      "active",
		},
		{
			Name:        "小米空气净化器",
			Description: "小米空气净化器，高效HEPA过滤，智能APP控制，守护家人健康",
			CategoryID:  categoryMap["家居用品"],
			Price:       decimal.NewFromFloat(899.00),
			Stock:       80,
			Status:      "active",
		},
		{
			Name:        "网易严选四件套",
			Description: "网易严选纯棉四件套，舒适亲肤，透气性好，提升睡眠质量",
			CategoryID:  categoryMap["家居用品"],
			Price:       decimal.NewFromFloat(399.00),
			Stock:       120,
			Status:      "active",
		},
		{
			Name:        "戴森吸尘器V15",
			Description: "戴森V15无线吸尘器，强劲吸力，智能灰尘检测，深度清洁家居",
			CategoryID:  categoryMap["家居用品"],
			Price:       decimal.NewFromFloat(3999.00),
			Stock:       40,
			Status:      "active",
		},
	}

	// 合并所有商品
	allProducts := append(electronicProducts, clothingProducts...)
	allProducts = append(allProducts, homeProducts...)

	// 创建商品
	for _, product := range allProducts {
		var existing model.Product
		if err := db.Where("name = ?", product.Name).First(&existing).Error; err != nil {
			// 商品不存在，创建新商品
			if err := db.Create(&product).Error; err != nil {
				return fmt.Errorf("创建商品 %s 失败: %v", product.Name, err)
			}
			log.Printf("✅ 创建商品: %s", product.Name)
		} else {
			log.Printf("⚠️ 商品已存在: %s", product.Name)
		}
	}

	return nil
}
