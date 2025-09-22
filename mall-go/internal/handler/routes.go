package handler

import (
	"mall-go/internal/handler/address"
	"mall-go/internal/handler/cart"
	"mall-go/internal/handler/file"
	"mall-go/internal/handler/middleware"
	"mall-go/internal/handler/order"
	"mall-go/internal/handler/payment"
	"mall-go/internal/handler/product"
	"mall-go/internal/handler/user"
	"mall-go/internal/model"
	paymentpkg "mall-go/pkg/payment"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// RegisterRoutes 注册所有路由
func RegisterRoutes(r *gin.Engine, db *gorm.DB, rdb *redis.Client, paymentService *paymentpkg.Service) {
	// 根路径健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "Mall Go API is running",
		})
	})

	// 根路径API信息
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Welcome to Mall-Go API",
			"version": "1.0.0",
			"endpoints": []string{
				"GET /health - 健康检查",
				"GET /api/v1/products - 获取商品列表",
				"POST /api/v1/users/register - 用户注册",
				"POST /api/v1/users/login - 用户登录",
				"GET /api/v1/cart - 获取购物车",
				"POST /api/v1/cart/add - 添加商品到购物车",
				"POST /api/v1/payments - 创建支付",
				"GET /api/v1/payments/:id - 查询支付状态",
				"POST /api/v1/payments/callback/alipay - 支付宝回调",
				"POST /api/v1/payments/callback/wechat - 微信支付回调",
			},
		})
	})

	// API版本组
	v1 := r.Group("/api/v1")

	// API版本健康检查
	v1.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "Mall Go API v1 is running",
		})
	})

	// 用户相关路由
	userHandler := user.NewHandler(db)
	userGroup := v1.Group("/users")
	{
		userGroup.POST("/register", userHandler.Register)
		userGroup.POST("/login", userHandler.Login)
		userGroup.GET("/profile", middleware.AuthMiddleware(), userHandler.GetProfile)
		userGroup.PUT("/profile", middleware.AuthMiddleware(), userHandler.UpdateProfile)
	}

	// 商品相关路由
	productHandler := product.NewHandler(db)
	productGroup := v1.Group("/products")
	{
		productGroup.GET("", productHandler.List)
		productGroup.GET("/:id", productHandler.Get)
		productGroup.POST("", middleware.AuthMiddleware(), middleware.AdminMiddleware(), productHandler.Create)
		productGroup.PUT("/:id", middleware.AuthMiddleware(), middleware.AdminMiddleware(), productHandler.Update)
		productGroup.DELETE("/:id", middleware.AuthMiddleware(), middleware.AdminMiddleware(), productHandler.Delete)
	}

	// 订单相关路由
	orderHandler := order.NewOrderHandler(db, rdb) // 使用正确的构造函数，传递Redis客户端
	orderGroup := v1.Group("/orders")
	orderGroup.Use(middleware.AuthMiddleware())
	{
		orderGroup.GET("", orderHandler.GetOrderList)                 // 使用正确的方法名
		orderGroup.GET("/:id", orderHandler.GetOrder)                 // 使用正确的方法名
		orderGroup.POST("", orderHandler.CreateOrder)                 // 使用正确的方法名
		orderGroup.PUT("/:id/status", orderHandler.UpdateOrderStatus) // 使用正确的方法名
		orderGroup.PUT("/:id/cancel", orderHandler.CancelOrder)       // 取消订单
	}

	// 购物车相关路由
	cartHandler := cart.NewCartHandler(db, rdb)
	cartGroup := v1.Group("/cart")
	cartGroup.Use(middleware.AuthMiddleware())
	{
		cartGroup.GET("", cartHandler.GetCart)                    // 获取购物车
		cartGroup.POST("/add", cartHandler.AddToCart)             // 添加商品到购物车
		cartGroup.PUT("/:id", cartHandler.UpdateCartItem)         // 更新购物车商品
		cartGroup.DELETE("/:id", cartHandler.RemoveFromCart)      // 从购物车移除商品
		cartGroup.DELETE("/clear", cartHandler.ClearCart)         // 清空购物车
		cartGroup.POST("/batch", cartHandler.BatchUpdateCart)     // 批量更新购物车
		cartGroup.POST("/select-all", cartHandler.SelectAllItems) // 全选/取消全选
		cartGroup.GET("/count", cartHandler.GetCartItemCount)     // 获取购物车商品数量
		cartGroup.POST("/sync", cartHandler.SyncCartItems)        // 同步购物车商品信息
	}

	// 支付相关路由
	paymentHandler := payment.NewHandler(db, paymentService)
	paymentGroup := v1.Group("/payments")
	paymentGroup.Use(middleware.AuthMiddleware())
	{
		paymentGroup.POST("", paymentHandler.CreatePayment)        // 创建支付
		paymentGroup.GET("", paymentHandler.ListPayments)          // 获取支付列表
		paymentGroup.GET("/:id", paymentHandler.GetPaymentByID)    // 根据ID获取支付详情
		paymentGroup.GET("/query", paymentHandler.QueryPayment)    // 查询支付状态
		paymentGroup.POST("/refund", paymentHandler.RefundPayment) // 申请退款
	}

	// 支付回调路由（无需认证）
	// 创建回调处理器（简化版本，实际应用中应该通过依赖注入配置完整的依赖）
	callbackHandler := payment.NewCallbackHandler(db, paymentService, nil, nil, nil, nil)
	callbackGroup := v1.Group("/payments/callback")
	{
		// 支付宝回调路由
		callbackGroup.POST("/alipay", callbackHandler.AlipayCallback)

		// 微信支付回调路由
		callbackGroup.POST("/wechat", callbackHandler.WechatCallback)
	}

	// 文件管理路由
	fileHandler := file.NewFileHandler(db, "uploads", "http://localhost:8080")
	fileGroup := v1.Group("/files")
	{
		// 公开文件访问（无需认证）
		fileGroup.GET("/public/:uuid", fileHandler.DownloadPublicFile)

		// 需要认证的文件操作
		fileAuth := fileGroup.Group("")
		fileAuth.Use(middleware.AuthMiddleware())
		{
			// 文件上传（需要文件创建权限）
			fileAuth.POST("/upload", middleware.RequirePermission(model.ResourceFile, model.ActionCreate), fileHandler.UploadSingle)
			fileAuth.POST("/upload/multiple", middleware.RequirePermission(model.ResourceFile, model.ActionCreate), fileHandler.UploadMultiple)

			// 文件查看和下载（需要文件读取权限）
			fileAuth.GET("/:uuid", middleware.RequirePermission(model.ResourceFile, model.ActionRead), fileHandler.GetFile)
			fileAuth.GET("/private/:uuid", middleware.RequirePermission(model.ResourceFile, model.ActionRead), fileHandler.DownloadPrivateFile)
			fileAuth.GET("", middleware.RequirePermission(model.ResourceFile, model.ActionRead), fileHandler.ListFiles)

			// 文件删除（需要文件删除权限）
			fileAuth.DELETE("/:uuid", middleware.RequirePermission(model.ResourceFile, model.ActionDelete), fileHandler.DeleteFile)
		}
	}

	// 地址管理路由
	addressHandler := address.NewHandler(db)
	addressGroup := v1.Group("/addresses")
	addressGroup.Use(middleware.AuthMiddleware())
	{
		addressGroup.GET("", addressHandler.GetAddresses)                  // 获取地址列表
		addressGroup.POST("", addressHandler.CreateAddress)                // 创建地址
		addressGroup.PUT("/:id", addressHandler.UpdateAddress)             // 更新地址
		addressGroup.DELETE("/:id", addressHandler.DeleteAddress)          // 删除地址
		addressGroup.PUT("/:id/default", addressHandler.SetDefaultAddress) // 设置默认地址

		// 地区数据（无需认证）
		addressGroup.GET("/regions", addressHandler.GetRegions) // 获取地区数据
	}
}

// RegisterMiddleware 注册中间件
func RegisterMiddleware(r *gin.Engine) {
	// 跨域中间件
	r.Use(middleware.CorsMiddleware())

	// 日志中间件
	r.Use(middleware.LoggerMiddleware())

	// 恢复中间件
	r.Use(middleware.RecoveryMiddleware())
}
