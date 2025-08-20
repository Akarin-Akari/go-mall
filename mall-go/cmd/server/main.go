package main

import (
	"log"
	"mall-go/internal/config"
	"mall-go/internal/handler"
	"mall-go/pkg/database"
	"mall-go/pkg/logger"

	"github.com/gin-gonic/gin"
)

// @title Mall Go API
// @version 1.0
// @description Go语言商城后端API
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api/v1

func main() {
	// 初始化日志
	logger.Init()

	// 加载配置
	config.Load()

	// 初始化数据库
	db := database.Init()

	// 设置Gin模式
	gin.SetMode(gin.ReleaseMode)

	// 创建Gin实例
	r := gin.Default()

	// 注册中间件
	handler.RegisterMiddleware(r)

	// 注册路由
	handler.RegisterRoutes(r, db)

	// 启动服务器
	logger.Info("服务器启动在端口: 8080")
	log.Fatal(r.Run(":8080"))
}
