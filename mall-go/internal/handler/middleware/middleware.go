package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"mall-go/pkg/auth"
	"mall-go/internal/model"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// 统一的错误响应函数，遵循DRY原则
func respondWithError(c *gin.Context, code int, message string) {
	c.JSON(code, gin.H{
		"code":    code,
		"message": message,
	})
	c.Abort()
}

// CorsMiddleware 跨域中间件
func CorsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// LoggerMiddleware 日志中间件
func LoggerMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
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

// RecoveryMiddleware 恢复中间件
func RecoveryMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		if err, ok := recovered.(string); ok {
			c.String(http.StatusInternalServerError, fmt.Sprintf("error: %s", err))
		}
		c.AbortWithStatus(http.StatusInternalServerError)
	})
}

// AuthMiddleware JWT认证中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取Authorization头
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			respondWithError(c, http.StatusUnauthorized, "未提供认证令牌")
			return
		}

		// 检查Bearer前缀
		const bearerPrefix = "Bearer "
		if !strings.HasPrefix(authHeader, bearerPrefix) {
			respondWithError(c, http.StatusUnauthorized, "认证令牌格式错误")
			return
		}

		// 提取token
		tokenString := authHeader[len(bearerPrefix):]
		if tokenString == "" {
			respondWithError(c, http.StatusUnauthorized, "认证令牌为空")
			return
		}

		// 验证JWT令牌
		claims, err := auth.ParseToken(tokenString)
		if err != nil {
			respondWithError(c, http.StatusUnauthorized, "认证令牌无效")
			return
		}

		// 将用户信息设置到上下文中
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("user_role", claims.Role)
		c.Set("user_claims", claims)

		c.Next()
	}
}

// GetUserFromContext 从上下文中获取用户信息
func GetUserFromContext(c *gin.Context) (userID uint, username string, role string, exists bool) {
	userIDVal, exists1 := c.Get("user_id")
	usernameVal, exists2 := c.Get("username")
	roleVal, exists3 := c.Get("user_role")

	if !exists1 || !exists2 || !exists3 {
		return 0, "", "", false
	}

	userID, ok1 := userIDVal.(uint)
	username, ok2 := usernameVal.(string)
	role, ok3 := roleVal.(string)

	if !ok1 || !ok2 || !ok3 {
		return 0, "", "", false
	}

	return userID, username, role, true
}

// TODO: RequireAuth别名在MVP阶段不需要，遵循YAGNI原则
// 如果将来需要更清晰的语义可以再添加
/*
func RequireAuth() gin.HandlerFunc {
	return AuthMiddleware()
}
*/

// RequireRole 要求特定角色的中间件
func RequireRole(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		_, _, role, exists := GetUserFromContext(c)
		if !exists {
			respondWithError(c, http.StatusUnauthorized, "用户未认证")
			return
		}

		if role != requiredRole {
			respondWithError(c, http.StatusForbidden, "权限不足")
			return
		}

		c.Next()
	}
}

// AdminMiddleware 管理员权限中间件
func AdminMiddleware() gin.HandlerFunc {
	return RequireRole(model.RoleAdmin)
}

// MerchantMiddleware 商家权限中间件
func MerchantMiddleware() gin.HandlerFunc {
	return RequireRole(model.RoleMerchant)
}

// AdminOrMerchantMiddleware 管理员或商家权限中间件
func AdminOrMerchantMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, _, role, exists := GetUserFromContext(c)
		if !exists {
			respondWithError(c, http.StatusUnauthorized, "用户未认证")
			return
		}

		if role != model.RoleAdmin && role != model.RoleMerchant {
			respondWithError(c, http.StatusForbidden, "权限不足，需要管理员或商家权限")
			return
		}

		c.Next()
	}
}
