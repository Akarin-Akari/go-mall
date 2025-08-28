package handler

import (
	"mall-go/internal/handler/file"
	"mall-go/internal/handler/middleware"
	"mall-go/internal/handler/order"
	"mall-go/internal/handler/product"
	"mall-go/internal/handler/user"
	"mall-go/internal/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterRoutes 注册所有路由
func RegisterRoutes(r *gin.Engine, db *gorm.DB) {
	// API版本组
	v1 := r.Group("/api/v1")

	// 健康检查
	v1.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "Mall Go API is running",
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
	orderHandler := order.NewHandler(db)
	orderGroup := v1.Group("/orders")
	orderGroup.Use(middleware.AuthMiddleware())
	{
		orderGroup.GET("", orderHandler.List)
		orderGroup.GET("/:id", orderHandler.Get)
		orderGroup.POST("", orderHandler.Create)
		orderGroup.PUT("/:id/status", orderHandler.UpdateStatus)
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
