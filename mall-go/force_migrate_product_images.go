package main

import (
	"fmt"
	"mall-go/internal/config"
	"mall-go/internal/model"
	"mall-go/pkg/database"
	"mall-go/pkg/logger"
)

func main() {
	fmt.Println("🔧 强制迁移ProductImage表")

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

	// 检查ProductImage表是否存在
	if db.Migrator().HasTable(&model.ProductImage{}) {
		fmt.Println("📋 ProductImage表已存在")
	} else {
		fmt.Println("⚠️  ProductImage表不存在，开始创建...")

		// 强制迁移ProductImage表
		err := db.AutoMigrate(&model.ProductImage{})
		if err != nil {
			fmt.Printf("❌ 创建ProductImage表失败: %v\n", err)
			return
		}

		fmt.Println("✅ ProductImage表创建成功")
	}

	// 验证表结构
	fmt.Println("\n🔍 验证表结构:")

	// 检查所有必要的表
	tables := []interface{}{
		&model.Product{},
		&model.ProductImage{},
		&model.Category{},
		&model.User{},
	}

	tableNames := []string{"products", "product_images", "categories", "users"}
	for i, table := range tables {
		if db.Migrator().HasTable(table) {
			fmt.Printf("  ✅ %s 表存在\n", tableNames[i])
		} else {
			fmt.Printf("  ❌ %s 表不存在\n", tableNames[i])
		}
	}

	// 测试Preload查询
	fmt.Println("\n🧪 测试Preload查询:")

	var product model.Product
	err := db.Preload("Category").Preload("Images").First(&product, 1).Error
	if err != nil {
		fmt.Printf("❌ Preload查询失败: %v\n", err)

		// 尝试不使用Images的Preload
		err2 := db.Preload("Category").First(&product, 1).Error
		if err2 != nil {
			fmt.Printf("❌ 基础查询也失败: %v\n", err2)
		} else {
			fmt.Println("✅ 基础查询成功，问题确实在Images Preload")
		}
	} else {
		fmt.Println("✅ Preload查询成功")
		fmt.Printf("📋 商品信息: ID=%d, Name=%s, Images数量=%d\n",
			product.ID, product.Name, len(product.Images))
	}

	fmt.Println("\n🎉 ProductImage表迁移验证完成！")
}
