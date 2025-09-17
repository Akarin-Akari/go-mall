//go:build ignore

package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	fmt.Println("🚀 启动测试服务器...")
	
	// 创建Gin实例
	r := gin.Default()
	
	// 添加健康检查路由
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"message": "Mall-Go API is running",
		})
	})
	
	// 添加产品路由
	r.GET("/api/v1/products", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"message": "Products endpoint is working",
			"data": []gin.H{
				{"id": 1, "name": "测试商品1", "price": 99.99},
				{"id": 2, "name": "测试商品2", "price": 199.99},
			},
		})
	})
	
	// 添加根路由
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Welcome to Mall-Go API",
			"version": "1.0.0",
			"endpoints": []string{
				"GET /health - 健康检查",
				"GET /api/v1/products - 获取商品列表",
			},
		})
	})
	
	fmt.Println("✅ 路由配置完成")
	fmt.Println("🔗 服务器地址: http://localhost:8080")
	fmt.Println("🔗 健康检查: http://localhost:8080/health")
	fmt.Println("🔗 产品接口: http://localhost:8080/api/v1/products")
	fmt.Println("🚀 服务器启动中...")
	
	// 启动服务器
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal("服务器启动失败:", err)
	}
}
