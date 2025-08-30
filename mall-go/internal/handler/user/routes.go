package user

import (
	"mall-go/internal/handler/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterRoutes 注册用户相关路由
func RegisterRoutes(router *gin.RouterGroup, db *gorm.DB) {
	// 创建处理器
	userHandler := NewHandler(db)
	registerHandler := NewRegisterHandler(db)

	// 用户相关路由组
	userGroup := router.Group("/users")
	{
		// 公开路由（无需认证）
		public := userGroup.Group("")
		{
			// 用户注册相关
			public.POST("/register", registerHandler.Register)                    // 用户注册
			public.POST("/send-email-code", registerHandler.SendEmailCode)       // 发送邮箱验证码
			public.POST("/send-phone-code", registerHandler.SendPhoneCode)       // 发送手机验证码
			public.POST("/check-username", registerHandler.CheckUsername)        // 检查用户名可用性
			public.POST("/check-email", registerHandler.CheckEmail)              // 检查邮箱可用性
			public.GET("/check-username", registerHandler.CheckUsernameByQuery)  // 通过查询参数检查用户名
			public.GET("/check-email", registerHandler.CheckEmailByQuery)        // 通过查询参数检查邮箱
			public.GET("/register-config", registerHandler.GetRegisterConfig)    // 获取注册配置
			public.POST("/resend-email-code", registerHandler.ResendEmailCode)   // 重新发送邮箱验证码
			public.POST("/resend-phone-code", registerHandler.ResendPhoneCode)   // 重新发送手机验证码
			public.POST("/validate-register", registerHandler.ValidateRegisterData) // 验证注册数据

			// 用户登录相关
			public.POST("/login", userHandler.Login)                             // 用户登录
			public.POST("/refresh-token", userHandler.RefreshToken)              // 刷新token
			
			// 密码重置相关
			public.POST("/forgot-password", userHandler.ForgotPassword)          // 忘记密码
			public.POST("/reset-password", userHandler.ResetPassword)            // 重置密码
			public.POST("/verify-reset-code", userHandler.VerifyResetCode)       // 验证重置密码验证码
		}

		// 需要认证的路由
		auth := userGroup.Group("")
		auth.Use(middleware.AuthMiddleware())
		{
			// 用户信息管理
			auth.GET("/profile", userHandler.GetProfile)                         // 获取用户资料
			auth.PUT("/profile", userHandler.UpdateProfile)                      // 更新用户资料
			auth.POST("/upload-avatar", userHandler.UploadAvatar)                // 上传头像
			auth.GET("/me", userHandler.GetCurrentUser)                          // 获取当前用户信息
			
			// 密码管理
			auth.POST("/change-password", userHandler.ChangePassword)            // 修改密码
			
			// 邮箱和手机号管理
			auth.POST("/change-email", userHandler.ChangeEmail)                  // 修改邮箱
			auth.POST("/verify-email", userHandler.VerifyEmail)                  // 验证邮箱
			auth.POST("/change-phone", userHandler.ChangePhone)                  // 修改手机号
			auth.POST("/verify-phone", userHandler.VerifyPhone)                  // 验证手机号
			
			// 安全设置
			auth.GET("/security", userHandler.GetSecuritySettings)               // 获取安全设置
			auth.POST("/enable-2fa", userHandler.Enable2FA)                      // 启用双因子认证
			auth.POST("/disable-2fa", userHandler.Disable2FA)                    // 禁用双因子认证
			auth.GET("/login-logs", userHandler.GetLoginLogs)                    // 获取登录日志
			
			// 社交账号绑定
			auth.GET("/social-accounts", userHandler.GetSocialAccounts)          // 获取社交账号绑定
			auth.POST("/bind-social", userHandler.BindSocialAccount)             // 绑定社交账号
			auth.DELETE("/unbind-social/:provider", userHandler.UnbindSocialAccount) // 解绑社交账号
			
			// 用户注销
			auth.POST("/logout", userHandler.Logout)                             // 用户登出
			auth.DELETE("/account", userHandler.DeleteAccount)                   // 删除账户
		}

		// 管理员路由
		admin := userGroup.Group("/admin")
		admin.Use(middleware.AuthMiddleware(), middleware.AdminMiddleware())
		{
			admin.GET("", userHandler.GetUsers)                                  // 获取用户列表
			admin.GET("/:id", userHandler.GetUserByID)                           // 获取指定用户信息
			admin.PUT("/:id", userHandler.UpdateUser)                            // 更新用户信息
			admin.DELETE("/:id", userHandler.DeleteUser)                         // 删除用户
			admin.POST("/:id/lock", userHandler.LockUser)                        // 锁定用户
			admin.POST("/:id/unlock", userHandler.UnlockUser)                    // 解锁用户
			admin.POST("/:id/ban", userHandler.BanUser)                          // 封禁用户
			admin.POST("/:id/unban", userHandler.UnbanUser)                      // 解封用户
			admin.POST("/:id/reset-password", userHandler.AdminResetPassword)    // 管理员重置用户密码
			admin.GET("/:id/login-logs", userHandler.GetUserLoginLogs)           // 获取用户登录日志
			admin.GET("/statistics", userHandler.GetUserStatistics)              // 获取用户统计信息
		}
	}

	// 用户验证相关路由（独立分组）
	verifyGroup := router.Group("/verify")
	{
		verifyGroup.GET("/email/:token", userHandler.VerifyEmailByToken)        // 通过邮件链接验证邮箱
		verifyGroup.GET("/phone/:token", userHandler.VerifyPhoneByToken)        // 通过短信链接验证手机号
	}

	// OAuth 社交登录路由
	oauthGroup := router.Group("/oauth")
	{
		oauthGroup.GET("/:provider", userHandler.OAuthLogin)                    // 社交登录跳转
		oauthGroup.GET("/:provider/callback", userHandler.OAuthCallback)        // 社交登录回调
	}
}

// RegisterPublicRoutes 注册公开路由（不需要认证）
func RegisterPublicRoutes(router *gin.RouterGroup, db *gorm.DB) {
	registerHandler := NewRegisterHandler(db)
	userHandler := NewHandler(db)

	// 注册相关
	router.POST("/register", registerHandler.Register)
	router.POST("/send-email-code", registerHandler.SendEmailCode)
	router.POST("/send-phone-code", registerHandler.SendPhoneCode)
	router.GET("/check-username", registerHandler.CheckUsernameByQuery)
	router.GET("/check-email", registerHandler.CheckEmailByQuery)
	router.GET("/register-config", registerHandler.GetRegisterConfig)

	// 登录相关
	router.POST("/login", userHandler.Login)
	router.POST("/refresh-token", userHandler.RefreshToken)
	router.POST("/forgot-password", userHandler.ForgotPassword)
	router.POST("/reset-password", userHandler.ResetPassword)

	// 验证相关
	router.GET("/verify/email/:token", userHandler.VerifyEmailByToken)
	router.GET("/verify/phone/:token", userHandler.VerifyPhoneByToken)

	// OAuth
	router.GET("/oauth/:provider", userHandler.OAuthLogin)
	router.GET("/oauth/:provider/callback", userHandler.OAuthCallback)
}

// RegisterAuthRoutes 注册需要认证的路由
func RegisterAuthRoutes(router *gin.RouterGroup, db *gorm.DB) {
	userHandler := NewHandler(db)

	// 应用认证中间件
	router.Use(middleware.AuthMiddleware())

	// 用户信息管理
	router.GET("/profile", userHandler.GetProfile)
	router.PUT("/profile", userHandler.UpdateProfile)
	router.POST("/upload-avatar", userHandler.UploadAvatar)
	router.GET("/me", userHandler.GetCurrentUser)

	// 密码和安全
	router.POST("/change-password", userHandler.ChangePassword)
	router.POST("/change-email", userHandler.ChangeEmail)
	router.POST("/change-phone", userHandler.ChangePhone)
	router.GET("/security", userHandler.GetSecuritySettings)
	router.GET("/login-logs", userHandler.GetLoginLogs)

	// 社交账号
	router.GET("/social-accounts", userHandler.GetSocialAccounts)
	router.POST("/bind-social", userHandler.BindSocialAccount)
	router.DELETE("/unbind-social/:provider", userHandler.UnbindSocialAccount)

	// 登出和删除账户
	router.POST("/logout", userHandler.Logout)
	router.DELETE("/account", userHandler.DeleteAccount)
}

// RegisterAdminRoutes 注册管理员路由
func RegisterAdminRoutes(router *gin.RouterGroup, db *gorm.DB) {
	userHandler := NewHandler(db)

	// 应用认证和管理员中间件
	router.Use(middleware.AuthMiddleware(), middleware.AdminMiddleware())

	// 用户管理
	router.GET("/users", userHandler.GetUsers)
	router.GET("/users/:id", userHandler.GetUserByID)
	router.PUT("/users/:id", userHandler.UpdateUser)
	router.DELETE("/users/:id", userHandler.DeleteUser)
	router.POST("/users/:id/lock", userHandler.LockUser)
	router.POST("/users/:id/unlock", userHandler.UnlockUser)
	router.POST("/users/:id/ban", userHandler.BanUser)
	router.POST("/users/:id/unban", userHandler.UnbanUser)
	router.POST("/users/:id/reset-password", userHandler.AdminResetPassword)
	router.GET("/users/:id/login-logs", userHandler.GetUserLoginLogs)
	router.GET("/users/statistics", userHandler.GetUserStatistics)
}
