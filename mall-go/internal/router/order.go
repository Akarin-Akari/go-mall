package router

import (
	"context"
	"fmt"
	"time"

	"mall-go/internal/handler/middleware"
	"mall-go/internal/handler/order"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

// SetupOrderRoutes 设置订单相关路由
func SetupOrderRoutes(r *gin.Engine, db *gorm.DB, rdb *redis.Client) {
	// 创建订单处理器
	orderHandler := order.NewOrderHandler(db, rdb)

	// API版本分组
	v1 := r.Group("/api/v1")
	{
		// 用户端订单路由（需要登录）
		userOrders := v1.Group("/orders")
		userOrders.Use(middleware.AuthMiddleware())
		{
			// 订单基础操作
			userOrders.GET("", orderHandler.GetOrderList)             // 获取订单列表
			userOrders.POST("", orderHandler.CreateOrder)             // 创建订单
			userOrders.GET("/:id", orderHandler.GetOrder)             // 获取订单详情
			userOrders.GET("/no/:orderNo", orderHandler.GetOrderByNo) // 根据订单号获取订单
			userOrders.PUT("/:id/cancel", orderHandler.CancelOrder)   // 取消订单

			// 订单支付相关
			userOrders.POST("/payment", orderHandler.CreatePayment)        // 创建支付
			userOrders.POST("/payment/notify", orderHandler.PaymentNotify) // 支付回调（无需认证）

			// 订单物流相关
			userOrders.GET("/:id/shipping", orderHandler.GetShippingInfo) // 获取物流信息
			userOrders.PUT("/:id/confirm", orderHandler.ConfirmReceipt)   // 确认收货

			// 订单售后相关
			userOrders.POST("/aftersale", orderHandler.CreateAfterSale) // 创建售后申请
		}

		// 管理端订单路由（需要管理员权限）
		adminOrders := v1.Group("/admin/orders")
		adminOrders.Use(middleware.AuthMiddleware())
		adminOrders.Use(middleware.AdminMiddleware())
		{
			// 订单管理
			adminOrders.GET("", orderHandler.GetOrderList)                  // 获取所有订单列表
			adminOrders.GET("/:id", orderHandler.GetOrder)                  // 获取订单详情
			adminOrders.PUT("/:id/status", orderHandler.UpdateOrderStatus)  // 更新订单状态
			adminOrders.GET("/statistics", orderHandler.GetOrderStatistics) // 获取订单统计

			// 发货管理
			adminOrders.POST("/:id/shipment", orderHandler.CreateShipment) // 创建发货
			adminOrders.GET("/:id/shipping", orderHandler.GetShippingInfo) // 获取物流信息

			// 售后管理
			adminOrders.PUT("/aftersale/:id", orderHandler.HandleAfterSale)    // 处理售后申请
			adminOrders.GET("/aftersale", orderHandler.GetAfterSaleList)       // 获取售后申请列表
			adminOrders.GET("/aftersale/:id", orderHandler.GetAfterSaleDetail) // 获取售后申请详情
		}

		// 商家端订单路由（需要商家权限）
		merchantOrders := v1.Group("/merchant/orders")
		merchantOrders.Use(middleware.AuthMiddleware())
		merchantOrders.Use(middleware.MerchantMiddleware())
		{
			// 商家订单管理
			merchantOrders.GET("", orderHandler.GetMerchantOrderList)         // 获取商家订单列表
			merchantOrders.GET("/:id", orderHandler.GetOrder)                 // 获取订单详情
			merchantOrders.PUT("/:id/status", orderHandler.UpdateOrderStatus) // 更新订单状态

			// 商家发货管理
			merchantOrders.POST("/:id/shipment", orderHandler.CreateShipment) // 创建发货
			merchantOrders.GET("/:id/shipping", orderHandler.GetShippingInfo) // 获取物流信息

			// 商家售后管理
			merchantOrders.PUT("/aftersale/:id", orderHandler.HandleAfterSale)      // 处理售后申请
			merchantOrders.GET("/aftersale", orderHandler.GetMerchantAfterSaleList) // 获取商家售后申请列表
		}

		// 公开路由（无需认证）
		public := v1.Group("/public")
		{
			// 支付回调（第三方支付平台回调）
			public.POST("/payment/alipay/notify", orderHandler.AlipayNotify) // 支付宝回调
			public.POST("/payment/wechat/notify", orderHandler.WechatNotify) // 微信支付回调

			// 物流查询（公开接口）
			public.GET("/shipping/track/:trackingNumber", orderHandler.TrackShipment) // 物流跟踪
			public.GET("/shipping/companies", orderHandler.GetShippingCompanies)      // 获取物流公司列表
		}
	}
}

// SetupOrderWebhooks 设置订单相关的Webhook路由
func SetupOrderWebhooks(r *gin.Engine, db *gorm.DB, rdb *redis.Client) {
	orderHandler := order.NewOrderHandler(db, rdb)

	// Webhook路由组
	webhooks := r.Group("/webhooks")
	{
		// 支付相关Webhook
		payment := webhooks.Group("/payment")
		{
			payment.POST("/alipay", orderHandler.AlipayWebhook) // 支付宝Webhook
			payment.POST("/wechat", orderHandler.WechatWebhook) // 微信支付Webhook
			payment.POST("/stripe", orderHandler.StripeWebhook) // Stripe Webhook
			payment.POST("/paypal", orderHandler.PaypalWebhook) // PayPal Webhook
		}

		// 物流相关Webhook
		shipping := webhooks.Group("/shipping")
		{
			shipping.POST("/sf", orderHandler.SFWebhook) // 顺丰Webhook
			shipping.POST("/yt", orderHandler.YTWebhook) // 圆通Webhook
			shipping.POST("/zt", orderHandler.ZTWebhook) // 中通Webhook
			shipping.POST("/st", orderHandler.STWebhook) // 申通Webhook
			shipping.POST("/yd", orderHandler.YDWebhook) // 韵达Webhook
		}

		// 第三方服务Webhook
		services := webhooks.Group("/services")
		{
			services.POST("/sms", orderHandler.SMSWebhook)     // 短信服务Webhook
			services.POST("/email", orderHandler.EmailWebhook) // 邮件服务Webhook
			services.POST("/push", orderHandler.PushWebhook)   // 推送服务Webhook
		}
	}
}

// SetupOrderMiddleware 设置订单相关中间件
func SetupOrderMiddleware(r *gin.Engine, db *gorm.DB, rdb *redis.Client) {
	// 订单操作日志中间件
	r.Use(OrderLogMiddleware(db))

	// 订单限流中间件
	r.Use(OrderRateLimitMiddleware(rdb))

	// 订单缓存中间件
	r.Use(OrderCacheMiddleware(rdb))
}

// OrderLogMiddleware 订单操作日志中间件
func OrderLogMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 记录订单相关操作日志
		if isOrderOperation(c.Request.URL.Path) {
			// 记录操作前状态
			beforeLog := captureOrderState(c, db)

			c.Next()

			// 记录操作后状态
			afterLog := captureOrderState(c, db)

			// 保存操作日志
			saveOrderOperationLog(db, beforeLog, afterLog, c)
		} else {
			c.Next()
		}
	}
}

// OrderRateLimitMiddleware 订单限流中间件
func OrderRateLimitMiddleware(rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 对订单创建等敏感操作进行限流
		if isOrderSensitiveOperation(c.Request.URL.Path, c.Request.Method) {
			userID := getUserIDFromContext(c)
			if userID > 0 {
				// 检查用户订单创建频率
				if isOrderRateLimited(rdb, userID) {
					c.JSON(429, gin.H{
						"code":    429,
						"message": "操作过于频繁，请稍后再试",
					})
					c.Abort()
					return
				}
			}
		}
		c.Next()
	}
}

// OrderCacheMiddleware 订单缓存中间件
func OrderCacheMiddleware(rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 对订单查询操作进行缓存
		if isOrderQueryOperation(c.Request.URL.Path, c.Request.Method) {
			cacheKey := generateOrderCacheKey(c)

			// 尝试从缓存获取
			if cachedResponse := getFromCache(rdb, cacheKey); cachedResponse != nil {
				c.Header("X-Cache", "HIT")
				c.Data(200, "application/json", cachedResponse)
				c.Abort()
				return
			}

			// 缓存未命中，继续处理请求
			c.Header("X-Cache", "MISS")
			c.Next()

			// 缓存响应结果
			if c.Writer.Status() == 200 {
				cacheResponse(rdb, cacheKey, c)
			}
		} else {
			c.Next()
		}
	}
}

// 辅助函数
func isOrderOperation(path string) bool {
	orderPaths := []string{"/api/v1/orders", "/api/v1/admin/orders", "/api/v1/merchant/orders"}
	for _, orderPath := range orderPaths {
		if len(path) >= len(orderPath) && path[:len(orderPath)] == orderPath {
			return true
		}
	}
	return false
}

func isOrderSensitiveOperation(path, method string) bool {
	// 订单创建、支付、取消等敏感操作
	sensitiveOps := map[string][]string{
		"POST": {"/api/v1/orders", "/api/v1/orders/payment"},
		"PUT":  {"/api/v1/orders/*/cancel", "/api/v1/orders/*/status"},
	}

	if methods, exists := sensitiveOps[method]; exists {
		for _, pattern := range methods {
			if matchPath(path, pattern) {
				return true
			}
		}
	}
	return false
}

func isOrderQueryOperation(path, method string) bool {
	return method == "GET" && isOrderOperation(path)
}

func matchPath(path, pattern string) bool {
	// 简单的路径匹配，支持*通配符
	// 实际实现中可以使用更复杂的路径匹配算法
	return path == pattern || (len(pattern) > 0 && pattern[len(pattern)-1] == '*' &&
		len(path) >= len(pattern)-1 && path[:len(pattern)-1] == pattern[:len(pattern)-1])
}

func captureOrderState(c *gin.Context, db *gorm.DB) map[string]interface{} {
	// 捕获订单状态快照
	return map[string]interface{}{
		"timestamp": time.Now(),
		"user_id":   getUserIDFromContext(c),
		"path":      c.Request.URL.Path,
		"method":    c.Request.Method,
		"ip":        c.ClientIP(),
	}
}

func saveOrderOperationLog(db *gorm.DB, before, after map[string]interface{}, c *gin.Context) {
	// 保存订单操作日志到数据库
	// 实际实现中需要定义操作日志表结构
}

func isOrderRateLimited(rdb *redis.Client, userID uint) bool {
	// 检查用户是否被限流
	// 例如：每分钟最多创建5个订单
	key := fmt.Sprintf("order_rate_limit:user:%d", userID)
	count, err := rdb.Incr(context.Background(), key).Result()
	if err != nil {
		return false
	}

	if count == 1 {
		rdb.Expire(context.Background(), key, time.Minute)
	}

	return count > 5
}

func generateOrderCacheKey(c *gin.Context) string {
	// 生成订单缓存键
	userID := getUserIDFromContext(c)
	return fmt.Sprintf("order_cache:user:%d:path:%s:query:%s",
		userID, c.Request.URL.Path, c.Request.URL.RawQuery)
}

func getFromCache(rdb *redis.Client, key string) []byte {
	// 从Redis获取缓存数据
	data, err := rdb.Get(context.Background(), key).Bytes()
	if err != nil {
		return nil
	}
	return data
}

func cacheResponse(rdb *redis.Client, key string, c *gin.Context) {
	// 缓存响应数据到Redis
	// 实际实现中需要获取响应体数据
}

func getUserIDFromContext(c *gin.Context) uint {
	if uid, exists := c.Get("user_id"); exists {
		return uid.(uint)
	}
	return 0
}
