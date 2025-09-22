package payment

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"mall-go/pkg/logger"
	"mall-go/pkg/payment"
	"mall-go/pkg/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// SecurityMiddleware 支付安全中间件
func SecurityMiddleware(securityManager *payment.SecurityManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger.Info("支付安全检查开始",
			zap.String("path", c.Request.URL.Path),
			zap.String("method", c.Request.Method),
			zap.String("client_ip", c.ClientIP()))

		// 1. 检查请求大小
		if err := securityManager.ValidateRequestSize(c.Request.ContentLength); err != nil {
			logger.Error("请求大小验证失败", zap.Error(err))
			response.Error(c, http.StatusBadRequest, "请求大小超过限制")
			c.Abort()
			return
		}

		// 2. 检查IP白名单（仅对回调接口）
		if strings.Contains(c.Request.URL.Path, "/callback/") {
			if err := securityManager.CheckIPWhitelist(c.ClientIP()); err != nil {
				logger.Error("IP白名单检查失败", zap.Error(err), zap.String("client_ip", c.ClientIP()))
				response.Error(c, http.StatusForbidden, "IP地址不被允许")
				c.Abort()
				return
			}
		}

		// 3. 检查限流
		rateLimitKey := c.ClientIP() + ":" + c.Request.URL.Path
		if err := securityManager.CheckRateLimit(rateLimitKey); err != nil {
			logger.Error("限流检查失败", zap.Error(err), zap.String("key", rateLimitKey))
			response.Error(c, http.StatusTooManyRequests, "请求频率过高")
			c.Abort()
			return
		}

		// 4. 对于需要签名验证的接口进行签名检查
		if shouldVerifySignature(c.Request.URL.Path) {
			if err := verifyRequestSignature(c, securityManager); err != nil {
				logger.Error("签名验证失败", zap.Error(err))
				response.Error(c, http.StatusUnauthorized, "签名验证失败")
				c.Abort()
				return
			}
		}

		// 5. 防重放攻击检查
		if err := checkReplayAttack(c, securityManager); err != nil {
			logger.Error("防重放检查失败", zap.Error(err))
			response.Error(c, http.StatusBadRequest, "请求已过期或重复")
			c.Abort()
			return
		}

		logger.Info("支付安全检查通过")
		c.Next()
	}
}

// shouldVerifySignature 判断是否需要验证签名
func shouldVerifySignature(path string) bool {
	// 回调接口需要验证签名
	if strings.Contains(path, "/callback/") {
		return true
	}

	// 敏感操作接口需要验证签名
	sensitiveAPIs := []string{
		"/payments/refund",
		"/payments/statistics",
	}

	for _, api := range sensitiveAPIs {
		if strings.Contains(path, api) {
			return true
		}
	}

	return false
}

// verifyRequestSignature 验证请求签名
func verifyRequestSignature(c *gin.Context, securityManager *payment.SecurityManager) error {
	var params map[string]string

	// 根据请求方法获取参数
	if c.Request.Method == "POST" {
		// POST请求从表单或JSON中获取参数
		if strings.Contains(c.GetHeader("Content-Type"), "application/x-www-form-urlencoded") {
			// 表单参数
			if err := c.Request.ParseForm(); err != nil {
				return err
			}
			params = make(map[string]string)
			for key, values := range c.Request.PostForm {
				if len(values) > 0 {
					params[key] = values[0]
				}
			}
		} else {
			// JSON参数需要特殊处理，这里简化处理
			params = make(map[string]string)
		}
	} else {
		// GET请求从查询参数获取
		params = make(map[string]string)
		for key, values := range c.Request.URL.Query() {
			if len(values) > 0 {
				params[key] = values[0]
			}
		}
	}

	// 获取签名
	signature := params["sign"]
	if signature == "" {
		signature = c.GetHeader("X-Signature")
	}

	if signature == "" {
		return nil // 如果没有签名，跳过验证
	}

	// 获取签名类型
	signType := params["sign_type"]
	if signType == "" {
		signType = "MD5" // 默认使用MD5
	}

	// 验证签名
	return securityManager.VerifySignature(params, signature, signType)
}

// checkReplayAttack 检查重放攻击
func checkReplayAttack(c *gin.Context, securityManager *payment.SecurityManager) error {
	// 检查时间戳
	timestamp := c.GetHeader("X-Timestamp")
	if timestamp == "" {
		timestamp = c.Query("timestamp")
	}

	if timestamp != "" {
		if err := securityManager.CheckTimestamp(timestamp, 5*time.Minute); err != nil {
			return err
		}
	}

	// 检查随机数
	nonce := c.GetHeader("X-Nonce")
	if nonce == "" {
		nonce = c.Query("nonce")
	}

	if nonce != "" {
		if err := securityManager.CheckNonce(nonce); err != nil {
			return err
		}
	}

	return nil
}

// PaymentAuthMiddleware 支付权限中间件
func PaymentAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从JWT中获取用户信息
		userID, exists := c.Get("user_id")
		if !exists {
			response.Error(c, http.StatusUnauthorized, "用户未登录")
			c.Abort()
			return
		}

		// 检查用户权限
		if !hasPaymentPermission(userID) {
			response.Error(c, http.StatusForbidden, "无支付权限")
			c.Abort()
			return
		}

		c.Next()
	}
}

// hasPaymentPermission 检查用户是否有支付权限
func hasPaymentPermission(userID interface{}) bool {
	// 这里应该根据实际的权限系统来检查
	// 简化处理，假设所有登录用户都有支付权限
	return userID != nil
}

// PaymentLogMiddleware 支付日志中间件
func PaymentLogMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// 简单的日志格式化函数
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format("02/Jan/2006:15:04:05 -0700"),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	})
}

// CORSMiddleware 跨域中间件（用于支付页面）
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// 允许的域名列表
		allowedOrigins := []string{
			"https://your-frontend-domain.com",
			"https://localhost:3000", // 开发环境
		}

		// 检查是否为允许的域名
		allowed := false
		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin {
				allowed = true
				break
			}
		}

		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Signature, X-Timestamp, X-Nonce")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
		c.Header("Access-Control-Allow-Credentials", "true")

		// 处理预检请求
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// RequestIDMiddleware 请求ID中间件
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}

		c.Header("X-Request-ID", requestID)
		c.Set("request_id", requestID)

		c.Next()
	}
}

// generateRequestID 生成请求ID
func generateRequestID() string {
	return strconv.FormatInt(time.Now().UnixNano(), 36)
}

// RecoveryMiddleware 恢复中间件
func RecoveryMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		logger.Error("支付接口发生panic",
			zap.Any("recovered", recovered),
			zap.String("path", c.Request.URL.Path),
			zap.String("method", c.Request.Method))

		response.Error(c, http.StatusInternalServerError, "服务器内部错误")
	})
}

// SetupPaymentMiddlewares 设置支付中间件
func SetupPaymentMiddlewares(router *gin.RouterGroup, securityManager *payment.SecurityManager) {
	// 基础中间件
	router.Use(RecoveryMiddleware())
	router.Use(RequestIDMiddleware())
	router.Use(PaymentLogMiddleware())
	router.Use(CORSMiddleware())

	// 安全中间件
	router.Use(SecurityMiddleware(securityManager))

	// 认证中间件（除了回调接口）
	router.Use(func(c *gin.Context) {
		// 回调接口不需要用户认证
		if strings.Contains(c.Request.URL.Path, "/callback/") {
			c.Next()
			return
		}

		// 其他接口需要用户认证
		PaymentAuthMiddleware()(c)
	})
}
