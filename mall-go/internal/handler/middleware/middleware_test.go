package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"mall-go/internal/config"
	"mall-go/internal/model"
	"mall-go/pkg/auth"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// 初始化测试配置
func init() {
	config.GlobalConfig = config.Config{
		JWT: config.JWTConfig{
			Secret: "test-secret-key-for-middleware-testing",
			Expire: "24h",
		},
	}
	gin.SetMode(gin.TestMode)
}

func TestAuthMiddleware_Success(t *testing.T) {
	// 生成有效的JWT token
	token, err := auth.GenerateToken(1, "testuser", model.RoleUser)
	assert.NoError(t, err)

	// 创建测试路由
	router := gin.New()
	router.Use(AuthMiddleware())
	router.GET("/test", func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		username, _ := c.Get("username")
		role, _ := c.Get("user_role")
		
		c.JSON(http.StatusOK, gin.H{
			"user_id":  userID,
			"username": username,
			"role":     role,
		})
	})

	// 创建请求
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	
	// 执行请求
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证结果
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "testuser")
}

func TestAuthMiddleware_NoToken(t *testing.T) {
	// 创建测试路由
	router := gin.New()
	router.Use(AuthMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// 创建没有token的请求
	req, _ := http.NewRequest("GET", "/test", nil)
	
	// 执行请求
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证结果
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "未提供认证令牌")
}

func TestAuthMiddleware_InvalidTokenFormat(t *testing.T) {
	// 创建测试路由
	router := gin.New()
	router.Use(AuthMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// 创建格式错误的token请求
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "InvalidToken")
	
	// 执行请求
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证结果
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "认证令牌格式错误")
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	// 创建测试路由
	router := gin.New()
	router.Use(AuthMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// 创建无效token的请求
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	
	// 执行请求
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证结果
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "认证令牌无效")
}

func TestGetUserFromContext(t *testing.T) {
	// 创建Gin上下文
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	
	// 设置用户信息
	c.Set("user_id", uint(1))
	c.Set("username", "testuser")
	c.Set("user_role", model.RoleUser)

	// 测试获取用户信息
	userID, username, role, exists := GetUserFromContext(c)
	
	assert.True(t, exists)
	assert.Equal(t, uint(1), userID)
	assert.Equal(t, "testuser", username)
	assert.Equal(t, model.RoleUser, role)
}

func TestGetUserFromContext_NotExists(t *testing.T) {
	// 创建空的Gin上下文
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// 测试获取不存在的用户信息
	_, _, _, exists := GetUserFromContext(c)
	
	assert.False(t, exists)
}

func TestAdminMiddleware_Success(t *testing.T) {
	// 生成管理员token
	token, err := auth.GenerateToken(1, "admin", model.RoleAdmin)
	assert.NoError(t, err)

	// 创建测试路由
	router := gin.New()
	router.Use(AuthMiddleware())
	router.Use(AdminMiddleware())
	router.GET("/admin", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "admin access granted"})
	})

	// 创建请求
	req, _ := http.NewRequest("GET", "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	
	// 执行请求
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证结果
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "admin access granted")
}

func TestAdminMiddleware_Forbidden(t *testing.T) {
	// 生成普通用户token
	token, err := auth.GenerateToken(1, "user", model.RoleUser)
	assert.NoError(t, err)

	// 创建测试路由
	router := gin.New()
	router.Use(AuthMiddleware())
	router.Use(AdminMiddleware())
	router.GET("/admin", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "admin access granted"})
	})

	// 创建请求
	req, _ := http.NewRequest("GET", "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	
	// 执行请求
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证结果
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "权限不足")
}

func TestMerchantMiddleware_Success(t *testing.T) {
	// 生成商家token
	token, err := auth.GenerateToken(1, "merchant", model.RoleMerchant)
	assert.NoError(t, err)

	// 创建测试路由
	router := gin.New()
	router.Use(AuthMiddleware())
	router.Use(MerchantMiddleware())
	router.GET("/merchant", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "merchant access granted"})
	})

	// 创建请求
	req, _ := http.NewRequest("GET", "/merchant", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	
	// 执行请求
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证结果
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "merchant access granted")
}

func TestAdminOrMerchantMiddleware_AdminSuccess(t *testing.T) {
	// 生成管理员token
	token, err := auth.GenerateToken(1, "admin", model.RoleAdmin)
	assert.NoError(t, err)

	// 创建测试路由
	router := gin.New()
	router.Use(AuthMiddleware())
	router.Use(AdminOrMerchantMiddleware())
	router.GET("/admin-or-merchant", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "access granted"})
	})

	// 创建请求
	req, _ := http.NewRequest("GET", "/admin-or-merchant", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	
	// 执行请求
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证结果
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAdminOrMerchantMiddleware_MerchantSuccess(t *testing.T) {
	// 生成商家token
	token, err := auth.GenerateToken(1, "merchant", model.RoleMerchant)
	assert.NoError(t, err)

	// 创建测试路由
	router := gin.New()
	router.Use(AuthMiddleware())
	router.Use(AdminOrMerchantMiddleware())
	router.GET("/admin-or-merchant", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "access granted"})
	})

	// 创建请求
	req, _ := http.NewRequest("GET", "/admin-or-merchant", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	
	// 执行请求
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证结果
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAdminOrMerchantMiddleware_UserForbidden(t *testing.T) {
	// 生成普通用户token
	token, err := auth.GenerateToken(1, "user", model.RoleUser)
	assert.NoError(t, err)

	// 创建测试路由
	router := gin.New()
	router.Use(AuthMiddleware())
	router.Use(AdminOrMerchantMiddleware())
	router.GET("/admin-or-merchant", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "access granted"})
	})

	// 创建请求
	req, _ := http.NewRequest("GET", "/admin-or-merchant", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	
	// 执行请求
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证结果
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "需要管理员或商家权限")
}
