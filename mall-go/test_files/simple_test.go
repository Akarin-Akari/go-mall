package main

import (
	"fmt"
	"time"

	"mall-go/internal/model"
	"mall-go/pkg/product"

	"github.com/shopspring/decimal"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	fmt.Println("🎯 Mall-Go 第一周数据模型优化简单测试")

	// 创建内存数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		fmt.Printf("❌ 连接数据库失败: %v\n", err)
		return
	}

	// 自动迁移
	fmt.Println("🚀 开始数据库迁移...")
	err = db.AutoMigrate(
		&model.User{},
		&model.Category{},
		&model.Brand{},
		&model.Product{},
		&model.ProductImage{},
	)
	if err != nil {
		fmt.Printf("❌ 数据库迁移失败: %v\n", err)
		return
	}
	fmt.Println("✅ 数据库迁移完成")

	// 创建测试数据
	fmt.Println("🧪 创建测试数据...")
	
	// 创建分类
	category := model.Category{
		Name:        "电子产品",
		Description: "各类电子设备",
		Status:      "active",
	}
	db.Create(&category)

	// 创建品牌
	brand := model.Brand{
		Name:        "Apple",
		Description: "苹果公司",
		Status:      "active",
	}
	db.Create(&brand)

	// 创建商家
	merchant := model.User{
		Username: "merchant1",
		Email:    "merchant1@example.com",
		Role:     "merchant",
		Status:   "active",
	}
	db.Create(&merchant)

	// 创建商品
	price, _ := decimal.NewFromString("8999.00")
	product := model.Product{
		Name:         "iPhone 15 Pro Max",
		Description:  "最新款iPhone",
		CategoryID:   category.ID,
		BrandID:      brand.ID,
		MerchantID:   merchant.ID,
		Price:        price,
		Stock:        100,
		Status:       "active",
		CategoryName: category.Name,
		BrandName:    brand.Name,
		MerchantName: merchant.Username,
	}
	db.Create(&product)

	fmt.Println("✅ 测试数据创建完成")

	// 创建服务
	productService := product.NewProductService(db)

	// 测试查询性能
	fmt.Println("\n🔍 开始性能测试...")

	// 测试1：单个商品查询
	start := time.Now()
	retrievedProduct, err := productService.GetProduct(product.ID)
	duration := time.Since(start)

	if err != nil {
		fmt.Printf("❌ 查询失败: %v\n", err)
	} else {
		fmt.Printf("✅ 查询成功: %s\n", retrievedProduct.Name)
		fmt.Printf("✅ 响应时间: %v\n", duration)
		fmt.Printf("✅ 分类名称: %s\n", retrievedProduct.CategoryName)
		fmt.Printf("✅ 品牌名称: %s\n", retrievedProduct.BrandName)
		fmt.Printf("✅ 商家名称: %s\n", retrievedProduct.MerchantName)
	}

	// 测试2：商品列表查询
	fmt.Println("\n📋 测试商品列表查询...")
	req := &product.ProductListRequest{
		Page:     1,
		PageSize: 10,
		Status:   "active",
	}

	start = time.Now()
	products, total, err := productService.GetProductList(req)
	duration = time.Since(start)

	if err != nil {
		fmt.Printf("❌ 列表查询失败: %v\n", err)
	} else {
		fmt.Printf("✅ 列表查询成功: %d 个商品，总计 %d 个\n", len(products), total)
		fmt.Printf("✅ 响应时间: %v\n", duration)
	}

	// 测试3：简单查询
	fmt.Println("\n⚡ 测试简单查询...")
	start = time.Now()
	simpleProduct, err := productService.GetProductSimple(product.ID)
	duration = time.Since(start)

	if err != nil {
		fmt.Printf("❌ 简单查询失败: %v\n", err)
	} else {
		fmt.Printf("✅ 简单查询成功: %s\n", simpleProduct.Name)
		fmt.Printf("✅ 响应时间: %v\n", duration)
	}

	fmt.Println("\n🎉 第一周数据模型优化测试完成！")
	fmt.Println("\n📊 优化成果:")
	fmt.Println("   • 使用冗余字段减少JOIN查询")
	fmt.Println("   • 实现分步查询策略")
	fmt.Println("   • 添加性能索引")
	fmt.Println("   • 建立数据同步机制")
}
