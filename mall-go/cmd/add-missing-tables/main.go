package main

import (
	"fmt"
	"log"
	"mall-go/internal/config"
	"mall-go/internal/model"
	"mall-go/pkg/database"
)

func main() {
	fmt.Println("🚀 开始添加缺失的数据库表...")

	// 初始化配置
	config.Load()

	// 初始化数据库连接
	db := database.Init()
	if db == nil {
		log.Fatal("数据库连接失败")
	}

	// 检查并创建缺失的表
	fmt.Println("📋 检查并创建缺失的表...")

	// 订单状态日志表
	if !db.Migrator().HasTable(&model.OrderStatusLog{}) {
		fmt.Println("   创建订单状态日志表...")
		if err := db.AutoMigrate(&model.OrderStatusLog{}); err != nil {
			log.Fatalf("创建订单状态日志表失败: %v", err)
		}
		fmt.Println("   ✅ 订单状态日志表创建成功")
	} else {
		fmt.Println("   ⚠️ 订单状态日志表已存在")
	}

	// 其他可能缺失的订单相关表
	missingTables := []interface{}{
		&model.OrderPayment{},
		&model.OrderShipment{},
		&model.OrderAfterSale{},
	}

	for _, table := range missingTables {
		tableName := fmt.Sprintf("%T", table)
		if !db.Migrator().HasTable(table) {
			fmt.Printf("   创建表: %s...\n", tableName)
			if err := db.AutoMigrate(table); err != nil {
				fmt.Printf("   ⚠️ 创建表 %s 失败: %v\n", tableName, err)
			} else {
				fmt.Printf("   ✅ 表 %s 创建成功\n", tableName)
			}
		} else {
			fmt.Printf("   ⚠️ 表 %s 已存在\n", tableName)
		}
	}

	// 验证表是否创建成功
	fmt.Println("🔍 验证表结构...")
	if db.Migrator().HasTable(&model.OrderStatusLog{}) {
		fmt.Println("   ✅ 订单状态日志表验证成功")

		// 检查表结构
		var count int64
		if err := db.Model(&model.OrderStatusLog{}).Count(&count).Error; err != nil {
			fmt.Printf("   ⚠️ 查询表失败: %v\n", err)
		} else {
			fmt.Printf("   📊 当前表中记录数: %d\n", count)
		}
	} else {
		fmt.Println("   ❌ 订单状态日志表验证失败")
	}

	fmt.Println("✅ 数据库表检查完成！")
}
