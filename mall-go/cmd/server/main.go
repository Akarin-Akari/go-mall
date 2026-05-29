package main

import (
	"net/http"
	"time"

	"mall-go/internal/config"
	"mall-go/internal/handler"
	"mall-go/pkg/audit"
	"mall-go/pkg/cache"
	"mall-go/pkg/database"
	"mall-go/pkg/logger"
	"mall-go/pkg/metrics"
	"mall-go/pkg/order"
	"mall-go/pkg/payment"
	"mall-go/pkg/security"
	"mall-go/pkg/shutdown"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
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

	// 初始化审计日志服务（P1-6）
	audit.InitGlobalAuditService(db)

	// 初始化Redis客户端（从配置文件读取配置）
	var rdb *redis.Client
	redisConfig := config.GlobalConfig.Redis

	redisClient, err := cache.NewRedisClient(redisConfig)
	if err != nil {
		logger.Warn("Redis服务连接失败，将在没有缓存的情况下运行", zap.Error(err))
		rdb = nil
	} else {
		// 使用封装的RedisClient获取底层的redis.Client
		rdb = redisClient.GetClient()
		logger.Info("Redis连接成功，缓存功能已启用")
	}

	// 初始化支付服务
	// 使用默认配置（开发环境 Debug=true，支持无数据库配置的回退逻辑）
	paymentConfig := payment.DefaultPaymentConfig()
	paymentService, err := payment.NewService(db, paymentConfig)
	if err != nil {
		logger.Warn("支付服务初始化失败，将在没有支付功能的情况下运行", zap.Error(err))
		paymentService = nil
	}

	// 初始化Prometheus指标
	metrics.InitPrometheusMetrics()

	// 初始化系统指标收集器（15秒采样间隔）
	metrics.InitSystemMetricsCollector(db, 15*time.Second)

	// 初始化安全组件
	security.InitInputValidator()
	security.InitJWTBlacklist(rdb)
	security.InitRateLimiter(rdb, security.RateLimitConfig{
		RequestsPerMinute: 100,
		BurstSize:         10,
		WindowDuration:    time.Minute,
	})

	// 设置Gin模式
	gin.SetMode(gin.ReleaseMode)

	// 创建Gin实例
	r := gin.Default()

	// 注册中间件
	handler.RegisterMiddleware(r)

	// 注册路由
	handler.RegisterRoutes(r, db, rdb, paymentService)

	// 启动订单调度器（P1-4：自动关闭超时未支付订单 + 自动确认收货，每分钟扫描一次）
	orderScheduler := order.NewScheduler(db, time.Minute)
	orderScheduler.Start()

	// 创建HTTP服务器
	server := &http.Server{
		Addr:         ":8081",
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 创建优雅关闭管理器
	gracefulShutdown := shutdown.NewGracefulShutdown(server, db, rdb, 30*time.Second)

	// 启动优雅关闭监听
	gracefulShutdown.Start()

	// 在goroutine中启动HTTP服务器
	go func() {
		logger.Info("服务器启动在端口: 8081")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("服务器启动失败", zap.Error(err))
		}
	}()

	// 等待优雅关闭完成
	gracefulShutdown.Wait()
	logger.Info("服务器已完全关闭")
}
