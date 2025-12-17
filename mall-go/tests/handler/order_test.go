package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"mall-go/internal/config"
	"mall-go/internal/handler/order"
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

// OrderHandlerTestSuite 订单Handler测试套件
type OrderHandlerTestSuite struct {
	suite.Suite
	db          *gorm.DB
	rdb         *redis.Client
	handler     *order.OrderHandler
	router      *gin.Engine
	helper      *helpers.TestHelper
	testUser    *model.User
	testProduct *model.Product
	testOrder   *model.Order
	token       string
}

// SetupSuite 测试套件初始化
func (suite *OrderHandlerTestSuite) SetupSuite() {
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
		Addr: "localhost:6379",
		DB:   2, // 使用测试数据库
	})

	// 初始化测试辅助工具
	suite.helper = helpers.NewTestHelper(suite.db)

	// 初始化Handler
	suite.handler = order.NewOrderHandler(suite.db, suite.rdb)

	// 设置路由
	suite.router = gin.New()
	suite.setupRoutes()

	// 创建测试用户和商品
	suite.testUser = suite.helper.CreateTestUser("orderuser", "order@example.com", "password123")
	suite.testProduct = suite.helper.CreateTestProduct("测试商品", "100.00", 50)

	// 生成JWT令牌
	suite.token, err = auth.GenerateToken(suite.testUser.ID, suite.testUser.Username, suite.testUser.Role)
	suite.Require().NoError(err)
}

// TearDownSuite 测试套件清理
func (suite *OrderHandlerTestSuite) TearDownSuite() {
	if suite.rdb != nil {
		suite.rdb.Close()
	}
	if suite.db != nil {
		testConfig.CleanupTestDB(suite.db)
	}
}

// SetupTest 每个测试前的准备
func (suite *OrderHandlerTestSuite) SetupTest() {
	// 清理订单数据
	suite.helper.CleanupOrderData()
	// 清理Redis缓存
	suite.rdb.FlushDB(suite.rdb.Context())
}

// setupRoutes 设置测试路由
func (suite *OrderHandlerTestSuite) setupRoutes() {
	v1 := suite.router.Group("/api/v1")
	{
		orders := v1.Group("/orders")
		{
			orders.GET("/", suite.authMiddleware(), suite.handler.GetOrderList)
			orders.POST("/", suite.authMiddleware(), suite.handler.Create)
			orders.GET("/:id", suite.authMiddleware(), suite.handler.Get)
			orders.PUT("/:id/status", suite.authMiddleware(), suite.handler.UpdateStatus)
			orders.POST("/:id/cancel", suite.authMiddleware(), suite.handler.CancelOrder)
		}
	}
}

// authMiddleware 简单的认证中间件用于测试
func (suite *OrderHandlerTestSuite) authMiddleware() gin.HandlerFunc {
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
func (suite *OrderHandlerTestSuite) makeRequest(method, url string, body interface{}, headers map[string]string) *httptest.ResponseRecorder {
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
func (suite *OrderHandlerTestSuite) getAuthHeaders() map[string]string {
	return map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", suite.token),
	}
}

// TestCreateOrder 测试创建订单
func (suite *OrderHandlerTestSuite) TestCreateOrder() {
	tests := []struct {
		name           string
		requestBody    model.OrderCreateRequest
		expectedStatus int
		expectedMsg    string
	}{
		{
			name: "成功创建订单",
			requestBody: model.OrderCreateRequest{
				CartItemIDs:     []uint{1}, // 假设购物车项ID为1
				ReceiverName:    "张三",
				ReceiverPhone:   "13800138000",
				ReceiverAddress: "某某街道123号",
				Province:        "北京市",
				City:            "北京市",
				District:        "朝阳区",
				ShippingMethod:  "express",
				BuyerMessage:    "测试订单",
			},
			expectedStatus: http.StatusOK,
			expectedMsg:    "订单创建成功",
		},
		{
			name: "购物车项不存在",
			requestBody: model.OrderCreateRequest{
				CartItemIDs:     []uint{99999}, // 不存在的购物车项ID
				ReceiverName:    "张三",
				ReceiverPhone:   "13800138000",
				ReceiverAddress: "某某街道123号",
				Province:        "北京市",
				City:            "北京市",
				District:        "朝阳区",
			},
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "购物车项不存在",
		},
		{
			name: "缺少收货地址",
			requestBody: model.OrderCreateRequest{
				CartItemIDs:  []uint{1},
				ReceiverName: "张三",
				// 缺少必填的收货地址字段
			},
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "请求参数错误",
		},
		{
			name: "购物车项为空",
			requestBody: model.OrderCreateRequest{
				CartItemIDs:     []uint{}, // 空的购物车项
				ReceiverName:    "张三",
				ReceiverPhone:   "13800138000",
				ReceiverAddress: "某某街道123号",
				Province:        "北京市",
				City:            "北京市",
				District:        "朝阳区",
			},
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "请求参数错误",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			w := suite.makeRequest("POST", "/api/v1/orders/", tt.requestBody, suite.getAuthHeaders())

			assert.Equal(suite.T(), tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(suite.T(), err)

			if tt.expectedStatus == http.StatusOK {
				assert.Equal(suite.T(), tt.expectedMsg, response["message"])
				assert.NotNil(suite.T(), response["data"])

				// 验证返回的订单数据
				orderData := response["data"].(map[string]interface{})
				assert.NotEmpty(suite.T(), orderData["order_id"])
				assert.NotEmpty(suite.T(), orderData["order_no"])
			} else {
				assert.Contains(suite.T(), response["message"].(string), tt.expectedMsg)
			}
		})
	}
}

// TestGetOrderList 测试获取订单列表
func (suite *OrderHandlerTestSuite) TestGetOrderList() {
	// 先创建一个测试订单
	suite.testOrder = suite.helper.CreateTestOrder(suite.testUser.ID, suite.testProduct.ID)

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
		expectedMsg    string
	}{
		{
			name:           "成功获取订单列表",
			queryParams:    "?page=1&page_size=10",
			expectedStatus: http.StatusOK,
			expectedMsg:    "获取订单列表成功",
		},
		{
			name:           "按状态筛选订单",
			queryParams:    "?status=pending&page=1&page_size=10",
			expectedStatus: http.StatusOK,
			expectedMsg:    "获取订单列表成功",
		},
		{
			name:           "默认分页参数",
			queryParams:    "",
			expectedStatus: http.StatusOK,
			expectedMsg:    "获取订单列表成功",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			url := "/api/v1/orders/" + tt.queryParams
			w := suite.makeRequest("GET", url, nil, suite.getAuthHeaders())

			assert.Equal(suite.T(), tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(suite.T(), err)

			assert.Equal(suite.T(), tt.expectedMsg, response["message"])
			assert.NotNil(suite.T(), response["data"])

			// 验证订单列表数据结构
			listData := response["data"].(map[string]interface{})
			assert.NotNil(suite.T(), listData["orders"])
			assert.NotNil(suite.T(), listData["total"])
			assert.NotNil(suite.T(), listData["page"])
			assert.NotNil(suite.T(), listData["page_size"])
		})
	}
}

// TestGetOrderDetail 测试获取订单详情
func (suite *OrderHandlerTestSuite) TestGetOrderDetail() {
	// 先创建一个测试订单
	suite.testOrder = suite.helper.CreateTestOrder(suite.testUser.ID, suite.testProduct.ID)

	tests := []struct {
		name           string
		orderID        string
		expectedStatus int
		expectedMsg    string
	}{
		{
			name:           "成功获取订单详情",
			orderID:        fmt.Sprintf("%d", suite.testOrder.ID),
			expectedStatus: http.StatusOK,
			expectedMsg:    "",
		},
		{
			name:           "订单ID无效",
			orderID:        "invalid",
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "无效的订单ID",
		},
		{
			name:           "订单不存在",
			orderID:        "99999",
			expectedStatus: http.StatusNotFound,
			expectedMsg:    "订单不存在",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			url := fmt.Sprintf("/api/v1/orders/%s", tt.orderID)
			w := suite.makeRequest("GET", url, nil, suite.getAuthHeaders())

			assert.Equal(suite.T(), tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(suite.T(), err)

			if tt.expectedStatus == http.StatusOK {
				assert.NotNil(suite.T(), response["data"])

				// 验证订单详情数据
				orderData := response["data"].(map[string]interface{})
				assert.NotEmpty(suite.T(), orderData["id"])
				assert.NotEmpty(suite.T(), orderData["order_no"])
				assert.NotEmpty(suite.T(), orderData["status"])
			} else {
				assert.Contains(suite.T(), response["message"].(string), tt.expectedMsg)
			}
		})
	}
}

// TestOrderHandlerSuite 运行测试套件
func TestOrderHandlerSuite(t *testing.T) {
	suite.Run(t, new(OrderHandlerTestSuite))
}
