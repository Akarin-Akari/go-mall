package main

import (
	"fmt"
	"log"
	"mall-go/internal/config"
	"mall-go/pkg/database"
	"mall-go/pkg/logger"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	fmt.Println("=== Mall-Go 服务器启动调试 ===")
	
	// 步骤1: 初始化日志
	fmt.Println("步骤1: 初始化日志...")
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("日志初始化失败: %v\n", r)
			os.Exit(1)
		}
	}()
	logger.Init()
	fmt.Println("✅ 日志初始化成功")

	// 步骤2: 加载配置
	fmt.Println("步骤2: 加载配置...")
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("配置加载失败: %v\n", r)
			os.Exit(1)
		}
	}()
	config.Load()
	fmt.Println("✅ 配置加载成功")

	// 步骤3: 初始化数据库
	fmt.Println("步骤3: 初始化数据库...")
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("数据库初始化失败: %v\n", r)
			os.Exit(1)
		}
	}()
	db := database.Init()
	if db == nil {
		fmt.Println("❌ 数据库连接失败")
		os.Exit(1)
	}
	fmt.Println("✅ 数据库初始化成功")

	// 步骤4: 创建Gin实例
	fmt.Println("步骤4: 创建Gin实例...")
	gin.SetMode(gin.DebugMode) // 使用调试模式
	r := gin.Default()
	fmt.Println("✅ Gin实例创建成功")

	// 步骤5: 添加基本路由
	fmt.Println("步骤5: 添加基本路由...")
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"message": "Mall-Go API is running",
		})
	})
	
	r.GET("/api/v1/products", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"message": "Products endpoint is working",
			"data": []interface{}{},
		})
	})
	fmt.Println("✅ 基本路由添加成功")

	// 步骤6: 启动服务器
	fmt.Println("步骤6: 启动服务器...")
	fmt.Println("🚀 服务器启动在端口: 8080")
	fmt.Println("🔗 健康检查: http://localhost:8080/health")
	fmt.Println("🔗 产品接口: http://localhost:8080/api/v1/products")
	
	log.Fatal(r.Run(":8080"))
}
