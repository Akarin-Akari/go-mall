package main

import (
	"fmt"
	"log"
	"mall-go/internal/config"
	"mall-go/pkg/database"
	"mall-go/pkg/logger"
)

func main() {
	fmt.Println("🔍 Mall-Go 数据库连接测试")
	fmt.Println("==========================")

	// 初始化日志
	logger.Init()

	// 加载配置
	fmt.Println("📋 加载配置文件...")
	config.Load()

	fmt.Printf("✅ 配置加载成功\n")
	fmt.Printf("   数据库主机: %s:%d\n", config.GlobalConfig.Database.Host, config.GlobalConfig.Database.Port)
	fmt.Printf("   数据库名称: %s\n", config.GlobalConfig.Database.DBName)
	fmt.Printf("   用户名: %s\n", config.GlobalConfig.Database.Username)

	// 测试数据库连接
	fmt.Println("\n🔌 测试数据库连接...")

	db := database.Init()
	if db == nil {
		log.Fatal("❌ 数据库初始化失败")
	}

	fmt.Println("✅ 数据库连接成功!")

	// 测试数据库操作
	fmt.Println("\n🧪 测试数据库操作...")

	// 测试查询用户表
	var userCount int64
	if err := db.Table("users").Count(&userCount).Error; err != nil {
		fmt.Printf("❌ 查询用户表失败: %v\n", err)
	} else {
		fmt.Printf("✅ 用户表查询成功，共有 %d 个用户\n", userCount)
	}

	// 测试查询商品表
	var productCount int64
	if err := db.Table("products").Count(&productCount).Error; err != nil {
		fmt.Printf("❌ 查询商品表失败: %v\n", err)
	} else {
		fmt.Printf("✅ 商品表查询成功，共有 %d 个商品\n", productCount)
	}

	// 测试查询分类表
	var categoryCount int64
	if err := db.Table("categories").Count(&categoryCount).Error; err != nil {
		fmt.Printf("❌ 查询分类表失败: %v\n", err)
	} else {
		fmt.Printf("✅ 分类表查询成功，共有 %d 个分类\n", categoryCount)
	}

	// 测试查询订单表
	var orderCount int64
	if err := db.Table("orders").Count(&orderCount).Error; err != nil {
		fmt.Printf("❌ 查询订单表失败: %v\n", err)
	} else {
		fmt.Printf("✅ 订单表查询成功，共有 %d 个订单\n", orderCount)
	}

	fmt.Println("\n🎉 数据库测试完成!")
}
