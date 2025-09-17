package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"mall-go/internal/config"
	"mall-go/internal/handler/user"
	"mall-go/internal/model"
	"mall-go/pkg/auth"
	testConfig "mall-go/tests/config"
	"mall-go/tests/helpers"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

// UserHandlerTestSuite 用户Handler测试套件
type UserHandlerTestSuite struct {
	suite.Suite
	db      *gorm.DB
	handler *user.Handler
	router  *gin.Engine
	helper  *helpers.TestHelper
}

// SetupSuite 测试套件初始化
func (suite *UserHandlerTestSuite) SetupSuite() {
	// 设置Gin为测试模式
	gin.SetMode(gin.TestMode)

	// 初始化全局配置
	config.GlobalConfig = config.Config{
		JWT: config.JWTConfig{
			Secret: "test-secret-key-for-jwt-token-generation",
			Expire: "24h",
		},
	}

	// 初始化测试数据库
	suite.db = testConfig.SetupTestDB()

	// 自动迁移测试表
	err := suite.db.AutoMigrate(
		&model.User{},
		&model.Product{},
		&model.Category{},
		&model.Order{},
		&model.OrderItem{},
		&model.CartItem{},
		&model.Cart{},
	)
	suite.Require().NoError(err)

	// 初始化测试辅助工具
	suite.helper = helpers.NewTestHelper(suite.db)

	// 初始化Handler
	suite.handler = user.NewHandler(suite.db)

	// 设置路由
	suite.router = gin.New()
	suite.setupRoutes()
}

// TearDownSuite 测试套件清理
func (suite *UserHandlerTestSuite) TearDownSuite() {
	if suite.db != nil {
		testConfig.CleanupTestDB(suite.db)
	}
}

// SetupTest 每个测试前的准备
func (suite *UserHandlerTestSuite) SetupTest() {
	// 清理测试数据
	suite.helper.CleanupTestData()
}

// setupRoutes 设置测试路由
func (suite *UserHandlerTestSuite) setupRoutes() {
	v1 := suite.router.Group("/api/v1")
	{
		users := v1.Group("/users")
		{
			users.POST("/register", suite.handler.Register)
			users.POST("/login", suite.handler.Login)
			users.GET("/profile", suite.authMiddleware(), suite.handler.GetProfile)
			users.PUT("/profile", suite.authMiddleware(), suite.handler.UpdateProfile)
			users.GET("/current", suite.authMiddleware(), suite.handler.GetCurrentUser)
		}
	}
}

// authMiddleware 简单的认证中间件用于测试
func (suite *UserHandlerTestSuite) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未提供认证令牌"})
			c.Abort()
			return
		}

		// 移除 "Bearer " 前缀
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}

		// 验证JWT令牌
		claims, err := auth.ParseToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的认证令牌"})
			c.Abort()
			return
		}

		// 设置用户信息到上下文
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Next()
	}
}

// makeRequest 发送HTTP请求的辅助方法
func (suite *UserHandlerTestSuite) makeRequest(method, url string, body interface{}, headers map[string]string) *httptest.ResponseRecorder {
	var reqBody []byte
	if body != nil {
		reqBody, _ = json.Marshal(body)
	}

	req, _ := http.NewRequest(method, url, bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	// 设置额外的请求头
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	return w
}

// TestUserRegister 测试用户注册
func (suite *UserHandlerTestSuite) TestUserRegister() {
	tests := []struct {
		name           string
		requestBody    model.UserRegisterRequest
		expectedStatus int
		expectedMsg    string
	}{
		{
			name: "成功注册",
			requestBody: model.UserRegisterRequest{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "password123",
				Nickname: "测试用户",
				Role:     "user",
			},
			expectedStatus: http.StatusOK,
			expectedMsg:    "用户注册成功",
		},
		{
			name: "用户名已存在",
			requestBody: model.UserRegisterRequest{
				Username: "testuser", // 重复用户名
				Email:    "test2@example.com",
				Password: "password123",
				Nickname: "测试用户2",
				Role:     "user",
			},
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "用户名已存在",
		},
		{
			name: "邮箱已存在",
			requestBody: model.UserRegisterRequest{
				Username: "testuser2",
				Email:    "test@example.com", // 重复邮箱
				Password: "password123",
				Nickname: "测试用户2",
				Role:     "user",
			},
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "邮箱已存在",
		},
		{
			name: "密码过短",
			requestBody: model.UserRegisterRequest{
				Username: "testuser3",
				Email:    "test3@example.com",
				Password: "123", // 密码过短
				Nickname: "测试用户3",
				Role:     "user",
			},
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "请求参数错误",
		},
		{
			name: "缺少必填字段",
			requestBody: model.UserRegisterRequest{
				Username: "testuser4",
				// 缺少Email
				Password: "password123",
				Nickname: "测试用户4",
				Role:     "user",
			},
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "请求参数错误",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			w := suite.makeRequest("POST", "/api/v1/users/register", tt.requestBody, nil)

			assert.Equal(suite.T(), tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(suite.T(), err)

			if tt.expectedStatus == http.StatusOK {
				assert.Equal(suite.T(), tt.expectedMsg, response["message"])
				assert.NotNil(suite.T(), response["data"])

				// 验证返回的用户数据
				userData := response["data"].(map[string]interface{})
				assert.Equal(suite.T(), tt.requestBody.Username, userData["username"])
				assert.Equal(suite.T(), tt.requestBody.Email, userData["email"])
				assert.Equal(suite.T(), tt.requestBody.Nickname, userData["nickname"])
				assert.NotEmpty(suite.T(), userData["id"])
			} else {
				assert.Contains(suite.T(), response["message"].(string), tt.expectedMsg)
			}
		})
	}
}

// TestUserLogin 测试用户登录
func (suite *UserHandlerTestSuite) TestUserLogin() {
	// 先创建一个测试用户
	testUser := suite.helper.CreateTestUser("loginuser", "login@example.com", "password123")

	tests := []struct {
		name           string
		requestBody    model.UserLoginRequest
		expectedStatus int
		expectedMsg    string
	}{
		{
			name: "成功登录",
			requestBody: model.UserLoginRequest{
				Username: testUser.Username,
				Password: "password123",
			},
			expectedStatus: http.StatusOK,
			expectedMsg:    "登录成功",
		},
		{
			name: "用户名不存在",
			requestBody: model.UserLoginRequest{
				Username: "nonexistentuser",
				Password: "password123",
			},
			expectedStatus: http.StatusUnauthorized,
			expectedMsg:    "用户名或密码错误",
		},
		{
			name: "密码错误",
			requestBody: model.UserLoginRequest{
				Username: testUser.Username,
				Password: "wrongpassword",
			},
			expectedStatus: http.StatusUnauthorized,
			expectedMsg:    "用户名或密码错误",
		},
		{
			name: "缺少用户名",
			requestBody: model.UserLoginRequest{
				Password: "password123",
			},
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "请求参数错误",
		},
		{
			name: "缺少密码",
			requestBody: model.UserLoginRequest{
				Username: testUser.Username,
			},
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "请求参数错误",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			w := suite.makeRequest("POST", "/api/v1/users/login", tt.requestBody, nil)

			assert.Equal(suite.T(), tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(suite.T(), err)

			if tt.expectedStatus == http.StatusOK {
				assert.Equal(suite.T(), tt.expectedMsg, response["message"])
				assert.NotNil(suite.T(), response["data"])

				// 验证返回的登录数据
				loginData := response["data"].(map[string]interface{})
				assert.NotEmpty(suite.T(), loginData["token"])
				assert.NotNil(suite.T(), loginData["user"])
				assert.Equal(suite.T(), float64(86400), loginData["expires_in"]) // 24小时
			} else {
				assert.Contains(suite.T(), response["message"].(string), tt.expectedMsg)
			}
		})
	}
}

// TestUserProfile 测试获取用户信息
func (suite *UserHandlerTestSuite) TestUserProfile() {
	// 创建测试用户并生成JWT令牌
	testUser := suite.helper.CreateTestUser("profileuser", "profile@example.com", "password123")
	token, err := auth.GenerateToken(testUser.ID, testUser.Username, testUser.Role)
	suite.Require().NoError(err)

	tests := []struct {
		name           string
		token          string
		expectedStatus int
		expectedMsg    string
	}{
		{
			name:           "成功获取用户信息",
			token:          token,
			expectedStatus: http.StatusOK,
			expectedMsg:    "",
		},
		{
			name:           "未提供认证令牌",
			token:          "",
			expectedStatus: http.StatusUnauthorized,
			expectedMsg:    "未提供认证令牌",
		},
		{
			name:           "无效的认证令牌",
			token:          "invalid_token",
			expectedStatus: http.StatusUnauthorized,
			expectedMsg:    "无效的认证令牌",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			headers := make(map[string]string)
			if tt.token != "" {
				headers["Authorization"] = fmt.Sprintf("Bearer %s", tt.token)
			}

			w := suite.makeRequest("GET", "/api/v1/users/profile", nil, headers)

			assert.Equal(suite.T(), tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(suite.T(), err)

			if tt.expectedStatus == http.StatusOK {
				assert.NotNil(suite.T(), response["data"])

				// 验证返回的用户数据
				userData := response["data"].(map[string]interface{})
				assert.Equal(suite.T(), testUser.Username, userData["username"])
				assert.Equal(suite.T(), testUser.Email, userData["email"])
				assert.Equal(suite.T(), testUser.Nickname, userData["nickname"])
			} else {
				assert.Contains(suite.T(), response["error"].(string), tt.expectedMsg)
			}
		})
	}
}

// TestUpdateProfile 测试更新用户信息
func (suite *UserHandlerTestSuite) TestUpdateProfile() {
	// 创建测试用户并生成JWT令牌
	testUser := suite.helper.CreateTestUser("updateuser", "update@example.com", "password123")
	token, err := auth.GenerateToken(testUser.ID, testUser.Username, testUser.Role)
	suite.Require().NoError(err)

	updateRequest := model.UserResponse{
		Nickname: "更新后的昵称",
		Avatar:   "https://example.com/avatar.jpg",
		Phone:    "13800138000",
	}

	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", token),
	}

	w := suite.makeRequest("PUT", "/api/v1/users/profile", updateRequest, headers)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), "用户信息更新成功", response["message"])
	assert.NotNil(suite.T(), response["data"])

	// 验证更新后的用户数据
	userData := response["data"].(map[string]interface{})
	assert.Equal(suite.T(), updateRequest.Nickname, userData["nickname"])
	assert.Equal(suite.T(), updateRequest.Avatar, userData["avatar"])
	assert.Equal(suite.T(), updateRequest.Phone, userData["phone"])
}

// TestSuite 运行测试套件
func TestUserHandlerSuite(t *testing.T) {
	suite.Run(t, new(UserHandlerTestSuite))
}
