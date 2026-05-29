package middleware

import (
	"net/http"
	"strings"

	"mall-go/internal/config"
	"mall-go/pkg/logger"
	"mall-go/pkg/security"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// SecurityHeadersMiddleware 安全响应头中间件
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// X-Content-Type-Options: 防止MIME类型嗅探
		c.Header("X-Content-Type-Options", "nosniff")

		// X-Frame-Options: 防止点击劫持
		c.Header("X-Frame-Options", "DENY")

		// X-XSS-Protection: 启用浏览器XSS防护
		c.Header("X-XSS-Protection", "1; mode=block")

		// Content-Security-Policy: 内容安全策略
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self'; connect-src 'self'")

		// Strict-Transport-Security: 强制HTTPS
		if config.GlobalConfig.Server.Mode == "release" {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		// Referrer-Policy: 控制Referrer信息
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Permissions-Policy: 控制浏览器功能
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		c.Next()
	}
}

// CORSSecureMiddleware 安全的CORS中间件
func CORSSecureMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取允许的源列表
		allowedOrigins := config.GlobalConfig.CORS.AllowOrigins

		// 生产环境禁止使用通配符
		if config.GlobalConfig.Server.Mode == "release" {
			// 检查是否包含通配符
			for _, origin := range allowedOrigins {
				if origin == "*" {
					logger.Warn("生产环境检测到CORS通配符配置，已禁用")
					c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
						"code": 403,
						"msg":  "生产环境禁止使用CORS通配符",
					})
					return
				}
			}
		}

		// 获取请求的Origin
		origin := c.GetHeader("Origin")

		// 检查Origin是否在白名单中
		allowed := false
		for _, allowedOrigin := range allowedOrigins {
			if allowedOrigin == "*" || allowedOrigin == origin {
				allowed = true
				break
			}
		}

		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Methods", strings.Join(config.GlobalConfig.CORS.AllowMethods, ", "))
			c.Header("Access-Control-Allow-Headers", strings.Join(config.GlobalConfig.CORS.AllowHeaders, ", "))
			c.Header("Access-Control-Max-Age", "86400")
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		// 处理预检请求
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// RateLimitMiddleware 速率限制中间件
func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取客户端IP
		clientIP := c.ClientIP()

		// 获取速率限制器
		limiter := security.GetRateLimiter()

		// 检查是否超过限制
		allowed, remaining, err := limiter.CheckLimit(c.Request.Context(), clientIP)

		// 设置限流信息头
		c.Header("X-RateLimit-Limit", "100")
		c.Header("X-RateLimit-Remaining", string(rune(remaining)))

		if !allowed {
			logger.Warn("请求被速率限制", zap.String("ip", clientIP))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code": 429,
				"msg":  "请求过于频繁，请稍后再试",
			})
			return
		}

		if err != nil {
			logger.Error("速率限制检查失败", zap.Error(err))
		}

		c.Next()
	}
}

// InputValidationMiddleware 输入验证中间件
func InputValidationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取输入验证器
		validator := security.GetInputValidator()

		// 验证查询参数
		for key, values := range c.Request.URL.Query() {
			for _, value := range values {
				// 检查SQL注入
				if !validator.ValidateSQL(value) {
					logger.Warn("检测到可能的SQL注入", zap.String("param", key), zap.String("value", value))
					c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
						"code": 400,
						"msg":  "非法的输入参数",
					})
					return
				}

				// 检查XSS
				if !validator.ValidateXSS(value) {
					logger.Warn("检测到可能的XSS攻击", zap.String("param", key))
					c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
						"code": 400,
						"msg":  "非法的输入参数",
					})
					return
				}
			}
		}

		c.Next()
	}
}

// RequestSizeLimitMiddleware 请求体大小限制中间件
func RequestSizeLimitMiddleware(maxSize int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 设置请求体大小限制（默认10MB）
		if maxSize == 0 {
			maxSize = 10 * 1024 * 1024
		}

		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSize)

		c.Next()

		// 检查是否超过大小限制
		if c.Request.ContentLength > maxSize {
			logger.Warn("请求体过大", zap.Int64("size", c.Request.ContentLength), zap.Int64("limit", maxSize))
			c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, gin.H{
				"code": 413,
				"msg":  "请求体过大",
			})
		}
	}
}

// JWTBlacklistMiddleware JWT黑名单检查中间件
func JWTBlacklistMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取Authorization头
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		// 提取token
		const bearerPrefix = "Bearer "
		if !strings.HasPrefix(authHeader, bearerPrefix) {
			c.Next()
			return
		}

		token := authHeader[len(bearerPrefix):]

		// 检查token是否在黑名单中
		blacklist := security.GetJWTBlacklist()
		if blacklist.IsBlacklisted(c.Request.Context(), token) {
			logger.Warn("检测到黑名单中的JWT", zap.String("token", token[:20]+"..."))
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code": 401,
				"msg":  "认证令牌已失效",
			})
			return
		}

		c.Next()
	}
}

