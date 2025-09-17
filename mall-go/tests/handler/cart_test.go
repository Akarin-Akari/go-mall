package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"mall-go/internal/config"
	"mall-go/internal/handler/cart"
	"mall-go/internal/model"
	"mall-go/pkg/auth"
	testConfig "mall-go/tests/config"
	"mall-go/tests/helpers"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

// CartHandlerTestSuite 购物车Handler测试套件
type CartHandlerTestSuite struct {
	suite.Suite
	db          *gorm.DB
	rdb         *redis.Client
	handler     *cart.CartHandler
	router      *gin.Engine
	helper      *helpers.TestHelper
	testUser    *model.User
	testProduct *model.Product
	token       string
}

// SetupSuite 测试套件初始化
func (suite *CartHandlerTestSuite) SetupSuite() {
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

	// 初始化Redis客户端（使用内存模拟）
	suite.rdb = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // 无密码
		DB:       1,  // 使用测试数据库
	})

	// 测试Redis连接，如果失败则跳过Redis相关功能
	_, err = suite.rdb.Ping(suite.rdb.Context()).Result()
	if err != nil {
		// Redis不可用时，使用nil客户端，购物车Handler会降级到仅数据库模式
		suite.rdb = nil
	}

	// 初始化测试辅助工具
	suite.helper = helpers.NewTestHelper(suite.db)

	// 初始化Handler
	suite.handler = cart.NewCartHandler(suite.db, suite.rdb)

	// 设置路由
	suite.router = gin.New()
	suite.setupRoutes()

	// 创建测试用户和商品
	suite.testUser = suite.helper.CreateTestUser("cartuser", "cart@example.com", "password123")
	suite.testProduct = suite.helper.CreateTestProduct("测试商品", "100.00", 50)

	// 生成JWT令牌
	suite.token, err = auth.GenerateToken(suite.testUser.ID, suite.testUser.Username, suite.testUser.Role)
	suite.Require().NoError(err)
}

// TearDownSuite 测试套件清理
func (suite *CartHandlerTestSuite) TearDownSuite() {
	if suite.rdb != nil {
		suite.rdb.Close()
	}
	if suite.db != nil {
		testConfig.CleanupTestDB(suite.db)
	}
}

// SetupTest 每个测试前的准备
func (suite *CartHandlerTestSuite) SetupTest() {
	// 清理购物车数据
	suite.helper.CleanupCartData()
	// 清理Redis缓存（如果Redis可用）
	if suite.rdb != nil {
		suite.rdb.FlushDB(suite.rdb.Context())
	}
}

// setupRoutes 设置测试路由
func (suite *CartHandlerTestSuite) setupRoutes() {
	v1 := suite.router.Group("/api/v1")
	{
		carts := v1.Group("/cart")
		{
			carts.POST("/add", suite.authMiddleware(), suite.handler.AddToCart)
			carts.GET("/", suite.authMiddleware(), suite.handler.GetCart)
			carts.PUT("/items/:id", suite.authMiddleware(), suite.handler.UpdateCartItem)
			carts.DELETE("/items/:id", suite.authMiddleware(), suite.handler.RemoveFromCart)
			carts.DELETE("/clear", suite.authMiddleware(), suite.handler.ClearCart)
			carts.POST("/batch-update", suite.authMiddleware(), suite.handler.BatchUpdateCart)
			carts.POST("/select-all", suite.authMiddleware(), suite.handler.SelectAllItems)
			carts.GET("/count", suite.authMiddleware(), suite.handler.GetCartItemCount)
		}
	}
}

// authMiddleware 简单的认证中间件用于测试
func (suite *CartHandlerTestSuite) authMiddleware() gin.HandlerFunc {
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
func (suite *CartHandlerTestSuite) makeRequest(method, url string, body interface{}, headers map[string]string) *httptest.ResponseRecorder {
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

// getAuthHeaders 获取认证请求头
func (suite *CartHandlerTestSuite) getAuthHeaders() map[string]string {
	return map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", suite.token),
	}
}

// TestAddToCart 测试添加商品到购物车
func (suite *CartHandlerTestSuite) TestAddToCart() {
	tests := []struct {
		name           string
		requestBody    model.AddToCartRequest
		expectedStatus int
		expectedMsg    string
	}{
		{
			name: "成功添加商品到购物车",
			requestBody: model.AddToCartRequest{
				ProductID: suite.testProduct.ID,
				Quantity:  2,
				SKUID:     0, // 可选
			},
			expectedStatus: http.StatusOK,
			expectedMsg:    "添加商品到购物车成功",
		},
		{
			name: "商品ID无效",
			requestBody: model.AddToCartRequest{
				ProductID: 99999, // 不存在的商品ID
				Quantity:  1,
			},
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "商品不存在",
		},
		{
			name: "数量无效",
			requestBody: model.AddToCartRequest{
				ProductID: suite.testProduct.ID,
				Quantity:  0, // 数量为0
			},
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "商品数量必须大于0",
		},
		{
			name: "数量超过库存",
			requestBody: model.AddToCartRequest{
				ProductID: suite.testProduct.ID,
				Quantity:  1000, // 超过库存
			},
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "库存不足",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			w := suite.makeRequest("POST", "/api/v1/cart/add", tt.requestBody, suite.getAuthHeaders())

			assert.Equal(suite.T(), tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(suite.T(), err)

			if tt.expectedStatus == http.StatusOK {
				assert.Equal(suite.T(), tt.expectedMsg, response["message"])
				assert.NotNil(suite.T(), response["data"])
			} else {
				assert.Contains(suite.T(), response["message"].(string), tt.expectedMsg)
			}
		})
	}
}

// TestGetCart 测试获取购物车
func (suite *CartHandlerTestSuite) TestGetCart() {
	// 先添加一个商品到购物车
	addRequest := model.AddToCartRequest{
		ProductID: suite.testProduct.ID,
		Quantity:  2,
	}
	suite.makeRequest("POST", "/api/v1/cart/add", addRequest, suite.getAuthHeaders())

	// 测试获取购物车
	w := suite.makeRequest("GET", "/api/v1/cart/", nil, suite.getAuthHeaders())

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), "获取购物车成功", response["message"])
	assert.NotNil(suite.T(), response["data"])

	// 验证购物车数据结构
	cartData := response["data"].(map[string]interface{})
	assert.NotNil(suite.T(), cartData["cart"])
	assert.NotNil(suite.T(), cartData["summary"])
}

// TestUpdateCartItem 测试更新购物车商品
func (suite *CartHandlerTestSuite) TestUpdateCartItem() {
	// 先添加一个商品到购物车
	addRequest := model.AddToCartRequest{
		ProductID: suite.testProduct.ID,
		Quantity:  2,
	}
	addResponse := suite.makeRequest("POST", "/api/v1/cart/add", addRequest, suite.getAuthHeaders())

	// 解析添加响应获取商品ID
	var addResult map[string]interface{}
	json.Unmarshal(addResponse.Body.Bytes(), &addResult)
	cartItem := addResult["data"].(map[string]interface{})
	itemID := uint(cartItem["id"].(float64))

	// 测试更新商品数量
	updateRequest := model.UpdateCartItemRequest{
		Quantity: 5,
	}

	w := suite.makeRequest("PUT", fmt.Sprintf("/api/v1/cart/items/%d", itemID), updateRequest, suite.getAuthHeaders())

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), "更新购物车商品成功", response["message"])
	assert.NotNil(suite.T(), response["data"])
}

// TestRemoveFromCart 测试从购物车移除商品
func (suite *CartHandlerTestSuite) TestRemoveFromCart() {
	// 先添加一个商品到购物车
	addRequest := model.AddToCartRequest{
		ProductID: suite.testProduct.ID,
		Quantity:  2,
	}
	addResponse := suite.makeRequest("POST", "/api/v1/cart/add", addRequest, suite.getAuthHeaders())

	// 解析添加响应获取商品ID
	var addResult map[string]interface{}
	json.Unmarshal(addResponse.Body.Bytes(), &addResult)
	cartItem := addResult["data"].(map[string]interface{})
	itemID := uint(cartItem["id"].(float64))

	// 测试移除商品
	w := suite.makeRequest("DELETE", fmt.Sprintf("/api/v1/cart/items/%d", itemID), nil, suite.getAuthHeaders())

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), "移除商品成功", response["message"])
}

// TestClearCart 测试清空购物车
func (suite *CartHandlerTestSuite) TestClearCart() {
	// 先添加一个商品到购物车
	addRequest := model.AddToCartRequest{
		ProductID: suite.testProduct.ID,
		Quantity:  2,
	}
	suite.makeRequest("POST", "/api/v1/cart/add", addRequest, suite.getAuthHeaders())

	// 测试清空购物车
	w := suite.makeRequest("DELETE", "/api/v1/cart/clear", nil, suite.getAuthHeaders())

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), "清空购物车成功", response["message"])
}

// TestGetCartItemCount 测试获取购物车商品数量
func (suite *CartHandlerTestSuite) TestGetCartItemCount() {
	// 先添加一个商品到购物车
	addRequest := model.AddToCartRequest{
		ProductID: suite.testProduct.ID,
		Quantity:  3,
	}
	suite.makeRequest("POST", "/api/v1/cart/add", addRequest, suite.getAuthHeaders())

	// 测试获取商品数量
	w := suite.makeRequest("GET", "/api/v1/cart/count", nil, suite.getAuthHeaders())

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.NotNil(suite.T(), response["data"])

	// 验证返回的数量数据
	countData := response["data"].(map[string]interface{})
	assert.NotNil(suite.T(), countData["total_count"])
	assert.NotNil(suite.T(), countData["selected_count"])
}

// TestCartHandlerSuite 运行测试套件
func TestCartHandlerSuite(t *testing.T) {
	suite.Run(t, new(CartHandlerTestSuite))
}
