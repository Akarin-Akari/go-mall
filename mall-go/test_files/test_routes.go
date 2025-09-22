//go:build ignore

package main

import (
	"fmt"
	"log"
	"net/http"

	"mall-go/internal/handler"
	"mall-go/pkg/payment"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// 简单的路由测试程序
func main() {
	// 创建内存数据库用于测试
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// 创建Redis客户端（可以为nil，购物车handler会处理）
	var rdb *redis.Client = nil

	// 创建支付服务（简单实例化）
	paymentService := payment.NewService(db)

	// 创建Gin引擎
	r := gin.New()

	// 注册中间件
	handler.RegisterMiddleware(r)

	// 注册路由
	handler.RegisterRoutes(r, db, rdb, paymentService)

	// 打印所有注册的路由
	fmt.Println("=== 已注册的路由 ===")
	routes := r.Routes()
	for _, route := range routes {
		fmt.Printf("%-8s %s\n", route.Method, route.Path)
	}

	fmt.Println("\n=== 购物车相关路由 ===")
	fmt.Println("GET      /api/v1/cart")
	fmt.Println("POST     /api/v1/cart/add")
	fmt.Println("PUT      /api/v1/cart/:id")
	fmt.Println("DELETE   /api/v1/cart/:id")
	fmt.Println("DELETE   /api/v1/cart/clear")
	fmt.Println("POST     /api/v1/cart/batch")
	fmt.Println("POST     /api/v1/cart/select-all")
	fmt.Println("GET      /api/v1/cart/count")
	fmt.Println("POST     /api/v1/cart/sync")

	fmt.Println("\n=== 支付相关路由 ===")
	fmt.Println("POST     /api/v1/payments")
	fmt.Println("GET      /api/v1/payments")
	fmt.Println("GET      /api/v1/payments/:id")
	fmt.Println("GET      /api/v1/payments/query")

	fmt.Println("\n=== 支付回调路由 ===")
	fmt.Println("POST     /api/v1/payments/callback/alipay")
	fmt.Println("POST     /api/v1/payments/callback/wechat")

	fmt.Println("\n=== 测试服务器启动 ===")
	fmt.Println("服务器将在 :8080 端口启动")
	fmt.Println("访问 http://localhost:8080 查看API信息")
	fmt.Println("访问 http://localhost:8080/health 进行健康检查")

	// 启动服务器
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
