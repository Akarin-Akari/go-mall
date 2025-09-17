package main

import (
	"log"
	"mall-go/internal/config"
	"mall-go/internal/handler"
	"mall-go/pkg/cache"
	"mall-go/pkg/database"
	"mall-go/pkg/logger"
	"mall-go/pkg/payment"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
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

	// 初始化Redis客户端（可选，如果Redis服务不可用则使用nil）
	var rdb *redis.Client
	redisConfig := config.RedisConfig{
		Host:         "localhost",
		Port:         6379,
		Password:     "",
		DB:           0,
		PoolSize:     10,
		MinIdleConns: 5,
		MaxRetries:   3,
		DialTimeout:  5,
		ReadTimeout:  3,
		WriteTimeout: 3,
		IdleTimeout:  300,
		MaxConnAge:   3600,
	}

	_, err := cache.NewRedisClient(redisConfig)
	if err != nil {
		logger.Warn("Redis服务连接失败，将在没有缓存的情况下运行", zap.Error(err))
		rdb = nil
	} else {
		// 将封装的RedisClient转换为标准的redis.Client
		// 这里需要根据实际的RedisClient实现来获取底层的redis.Client
		// 暂时使用简单的redis.NewClient
		rdb = redis.NewClient(&redis.Options{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		})
	}

	// 初始化支付服务
	paymentConfig := &payment.PaymentConfig{
		// 使用默认配置或从环境变量读取
	}
	paymentService, err := payment.NewService(db, paymentConfig)
	if err != nil {
		logger.Warn("支付服务初始化失败，将在没有支付功能的情况下运行", zap.Error(err))
		paymentService = nil
	}

	// 设置Gin模式
	gin.SetMode(gin.ReleaseMode)

	// 创建Gin实例
	r := gin.Default()

	// 注册中间件
	handler.RegisterMiddleware(r)

	// 注册路由
	handler.RegisterRoutes(r, db, rdb, paymentService)

	// 启动服务器
	logger.Info("服务器启动在端口: 8080")
	log.Fatal(r.Run(":8080"))
}
