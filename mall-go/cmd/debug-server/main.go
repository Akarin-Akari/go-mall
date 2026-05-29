package main

import (
	"context"
	"fmt"
	"log"
	"mall-go/internal/config"
	"mall-go/internal/handler"
	"mall-go/pkg/cache"
	"mall-go/pkg/database"
	"mall-go/pkg/logger"
	"mall-go/pkg/payment"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func main() {
	// 初始化日志
	logger.Init()
	logger.Info("🚀 启动调试模式服务器...")

	// 加载配置
	config.Load()
	logger.Info("✅ 配置加载完成")

	// 初始化数据库
	db := database.Init()
	logger.Info("✅ 数据库连接成功")

	// 初始化Redis客户端（可选，如果Redis服务不可用则使用nil）
	var rdb *redis.Client
	redisConfig := config.RedisConfig{
		Host:         "localhost",
		Port:         6379,
		Password:     "123456", // 使用正确的Redis密码
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
		rdb = redis.NewClient(&redis.Options{
			Addr:     "localhost:6379",
			Password: "123456", // 使用正确的Redis密码
			DB:       0,
		})
		// 测试连接
		ctx := context.Background()
		if err := rdb.Ping(ctx).Err(); err != nil {
			logger.Warn("Redis ping失败，将在没有缓存的情况下运行", zap.Error(err))
			rdb = nil
		} else {
			logger.Info("✅ Redis连接成功")
		}
	}

	// 初始化支付服务
	paymentConfig := &payment.PaymentConfig{
		// 使用默认配置或从环境变量读取
	}
	paymentService, err := payment.NewService(db, paymentConfig)
	if err != nil {
		logger.Warn("支付服务初始化失败，将在没有支付功能的情况下运行", zap.Error(err))
		paymentService = nil
	} else {
		logger.Info("✅ 支付服务初始化成功")
	}

	// 设置Gin为调试模式
	gin.SetMode(gin.DebugMode)
	logger.Info("🔧 Gin设置为调试模式")

	// 创建Gin实例
	r := gin.Default()

	// 添加详细的请求日志中间件
	r.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("[%s] %s %s %d %s %s\n",
			param.TimeStamp.Format("2006-01-02 15:04:05"),
			param.Method,
			param.Path,
			param.StatusCode,
			param.Latency,
			param.ErrorMessage,
		)
	}))

	// 添加错误恢复中间件
	r.Use(gin.Recovery())

	// 注册中间件
	handler.RegisterMiddleware(r)
	logger.Info("✅ 中间件注册完成")

	// 注册路由
	handler.RegisterRoutes(r, db, rdb, paymentService)
	logger.Info("✅ 路由注册完成")

	// 启动服务器
	logger.Info("🌟 调试服务器启动在端口: 8081")
	log.Fatal(r.Run(":8081"))
}
