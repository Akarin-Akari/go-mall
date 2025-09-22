package main

import (
	"fmt"
	"mall-go/internal/config"
	"mall-go/internal/model"
	"mall-go/pkg/database"
	"mall-go/pkg/logger"
)

func main() {
	fmt.Println("🔍 检查数据库中的商品数据")
	
	// 初始化配置
	config.Load()
	
	// 初始化日志
	logger.Init()
	
	// 初始化数据库
	db := database.Init()
	if db == nil {
		fmt.Println("❌ 数据库初始化失败")
		return
	}
	
	fmt.Println("✅ 数据库连接成功")
	
	// 检查商品数据
	fmt.Println("\n📊 商品数据统计:")
	
	var productCount int64
	db.Model(&model.Product{}).Count(&productCount)
	fmt.Printf("  商品总数: %d\n", productCount)
	
	// 检查分类数据
	var categoryCount int64
	db.Model(&model.Category{}).Count(&categoryCount)
	fmt.Printf("  分类总数: %d\n", categoryCount)
	
	// 检查用户数据
	var userCount int64
	db.Model(&model.User{}).Count(&userCount)
	fmt.Printf("  用户总数: %d\n", userCount)
	
	// 显示前5个商品的详细信息
	if productCount > 0 {
		fmt.Println("\n📋 前5个商品详情:")
		var products []model.Product
		db.Limit(5).Find(&products)
		
		for i, product := range products {
			fmt.Printf("  %d. ID=%d, Name=%s, Price=%s, Stock=%d, Status=%s\n", 
				i+1, product.ID, product.Name, product.Price.String(), product.Stock, product.Status)
		}
	} else {
		fmt.Println("\n⚠️  数据库中没有商品数据")
		fmt.Println("建议运行: go run scripts/init_test_data.go")
	}
	
	// 显示前3个分类的详细信息
	if categoryCount > 0 {
		fmt.Println("\n📂 前3个分类详情:")
		var categories []model.Category
		db.Limit(3).Find(&categories)
		
		for i, category := range categories {
			fmt.Printf("  %d. ID=%d, Name=%s, Description=%s\n", 
				i+1, category.ID, category.Name, category.Description)
		}
	} else {
		fmt.Println("\n⚠️  数据库中没有分类数据")
	}
	
	// 显示用户信息
	if userCount > 0 {
		fmt.Println("\n👤 用户信息:")
		var users []model.User
		db.Find(&users)
		
		for i, user := range users {
			fmt.Printf("  %d. ID=%d, Username=%s, Email=%s, Role=%s, Status=%s\n", 
				i+1, user.ID, user.Username, user.Email, user.Role, user.Status)
		}
	}
	
	fmt.Println("\n🎉 数据检查完成！")
}
