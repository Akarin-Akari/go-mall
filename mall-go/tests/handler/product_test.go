package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"mall-go/internal/config"
	"mall-go/internal/handler/product"
	"mall-go/internal/model"
	"mall-go/pkg/auth"
	testConfig "mall-go/tests/config"
	"mall-go/tests/helpers"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

// ProductHandlerTestSuite 商品Handler测试套件
type ProductHandlerTestSuite struct {
	suite.Suite
	db           *gorm.DB
	handler      *product.Handler
	router       *gin.Engine
	helper       *helpers.TestHelper
	testUser     *model.User
	testCategory *model.Category
	testProduct  *model.Product
	token        string
}

// SetupSuite 测试套件初始化
func (suite *ProductHandlerTestSuite) SetupSuite() {
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
		&model.ProductImage{},
		&model.Category{},
		&model.Order{},
		&model.OrderItem{},
		&model.CartItem{},
		&model.Cart{},
	)
	if err != nil {
		panic("Failed to migrate test database: " + err.Error())
	}

	// 初始化测试辅助工具
	suite.helper = helpers.NewTestHelper(suite.db)

	// 初始化Handler
	suite.handler = product.NewHandler(suite.db)

	// 设置路由
	suite.router = gin.New()
	suite.setupRoutes()

	// 创建测试用户和分类
	suite.testUser = suite.helper.CreateTestUser("productuser", "product@example.com", "password123")
	suite.testCategory = suite.helper.CreateTestCategory("测试分类", "test-category")

	// 生成JWT令牌
	suite.token, err = auth.GenerateToken(suite.testUser.ID, suite.testUser.Username, suite.testUser.Role)
	suite.Require().NoError(err)
}

// TearDownSuite 测试套件清理
func (suite *ProductHandlerTestSuite) TearDownSuite() {
	if suite.db != nil {
		testConfig.CleanupTestDB(suite.db)
	}
}

// SetupTest 每个测试前的准备
func (suite *ProductHandlerTestSuite) SetupTest() {
	// 清理商品数据
	suite.helper.CleanupProductData()
}

// setupRoutes 设置测试路由
func (suite *ProductHandlerTestSuite) setupRoutes() {
	v1 := suite.router.Group("/api/v1")
	{
		products := v1.Group("/products")
		{
			products.GET("/", suite.handler.List)
			products.GET("/:id", suite.handler.Get)
			products.POST("/", suite.authMiddleware(), suite.handler.Create)
			products.PUT("/:id", suite.authMiddleware(), suite.handler.Update)
			products.DELETE("/:id", suite.authMiddleware(), suite.handler.Delete)
		}
	}
}

// authMiddleware 简单的认证中间件用于测试
func (suite *ProductHandlerTestSuite) authMiddleware() gin.HandlerFunc {
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
func (suite *ProductHandlerTestSuite) makeRequest(method, url string, body interface{}, headers map[string]string) *httptest.ResponseRecorder {
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
func (suite *ProductHandlerTestSuite) getAuthHeaders() map[string]string {
	return map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", suite.token),
	}
}

// TestProductList 测试获取商品列表
func (suite *ProductHandlerTestSuite) TestProductList() {
	// 先创建一些测试商品
	suite.testProduct = suite.helper.CreateTestProduct("测试商品1", "100.00", 50)
	suite.helper.CreateTestProduct("测试商品2", "200.00", 30)

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
		expectedMsg    string
	}{
		{
			name:           "成功获取商品列表",
			queryParams:    "?page=1&page_size=10",
			expectedStatus: http.StatusOK,
			expectedMsg:    "查询成功",
		},
		{
			name:           "按分类筛选商品",
			queryParams:    fmt.Sprintf("?category_id=%d&page=1&page_size=10", suite.testCategory.ID),
			expectedStatus: http.StatusOK,
			expectedMsg:    "查询成功",
		},
		{
			name:           "关键词搜索商品",
			queryParams:    "?keyword=测试&page=1&page_size=10",
			expectedStatus: http.StatusOK,
			expectedMsg:    "查询成功",
		},
		{
			name:           "按状态筛选商品",
			queryParams:    "?status=active&page=1&page_size=10",
			expectedStatus: http.StatusOK,
			expectedMsg:    "查询成功",
		},
		{
			name:           "默认分页参数",
			queryParams:    "",
			expectedStatus: http.StatusOK,
			expectedMsg:    "查询成功",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			url := "/api/v1/products/" + tt.queryParams
			w := suite.makeRequest("GET", url, nil, nil)

			assert.Equal(suite.T(), tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(suite.T(), err)

			assert.Equal(suite.T(), tt.expectedMsg, response["msg"])
			assert.NotNil(suite.T(), response["data"])

			// 验证分页数据结构
			assert.NotNil(suite.T(), response["total"])
			assert.NotNil(suite.T(), response["page"])
			assert.NotNil(suite.T(), response["page_size"])
		})
	}
}

// TestProductGet 测试获取商品详情
func (suite *ProductHandlerTestSuite) TestProductGet() {
	// 先创建一个测试商品
	suite.testProduct = suite.helper.CreateTestProduct("测试商品", "100.00", 50)
	// 确保商品关联到测试分类
	suite.testProduct.CategoryID = suite.testCategory.ID
	suite.db.Save(suite.testProduct)

	tests := []struct {
		name           string
		productID      string
		expectedStatus int
		expectedMsg    string
	}{
		{
			name:           "成功获取商品详情",
			productID:      fmt.Sprintf("%d", suite.testProduct.ID),
			expectedStatus: http.StatusOK,
			expectedMsg:    "",
		},
		{
			name:           "商品ID无效",
			productID:      "invalid",
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "无效的商品ID",
		},
		{
			name:           "商品不存在",
			productID:      "99999",
			expectedStatus: http.StatusNotFound,
			expectedMsg:    "商品不存在",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			url := fmt.Sprintf("/api/v1/products/%s", tt.productID)
			w := suite.makeRequest("GET", url, nil, nil)

			// 添加调试信息
			suite.T().Logf("Response Status: %d", w.Code)
			suite.T().Logf("Response Body: %s", w.Body.String())

			assert.Equal(suite.T(), tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(suite.T(), err)

			if tt.expectedStatus == http.StatusOK {
				assert.NotNil(suite.T(), response["data"])

				// 验证商品详情数据
				productData := response["data"].(map[string]interface{})
				assert.NotEmpty(suite.T(), productData["id"])
				assert.NotEmpty(suite.T(), productData["name"])
				assert.NotEmpty(suite.T(), productData["price"])
			} else {
				assert.Contains(suite.T(), response["message"].(string), tt.expectedMsg)
			}
		})
	}
}

// TestProductCreate 测试创建商品
func (suite *ProductHandlerTestSuite) TestProductCreate() {
	tests := []struct {
		name           string
		requestBody    model.ProductCreateRequest
		expectedStatus int
		expectedMsg    string
	}{
		{
			name: "成功创建商品",
			requestBody: model.ProductCreateRequest{
				Name:        "新商品",
				Description: "这是一个新商品",
				Price:       150.00,
				Stock:       100,
				CategoryID:  suite.testCategory.ID,
				Images:      []string{"https://example.com/image1.jpg", "https://example.com/image2.jpg"},
			},
			expectedStatus: http.StatusOK,
			expectedMsg:    "商品创建成功",
		},
		{
			name: "分类不存在",
			requestBody: model.ProductCreateRequest{
				Name:        "新商品",
				Description: "这是一个新商品",
				Price:       150.00,
				Stock:       100,
				CategoryID:  99999, // 不存在的分类ID
				Images:      []string{"https://example.com/image1.jpg"},
			},
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "商品分类不存在",
		},
		{
			name: "缺少必填字段",
			requestBody: model.ProductCreateRequest{
				// 缺少Name
				Description: "这是一个新商品",
				Price:       150.00,
				Stock:       100,
				CategoryID:  suite.testCategory.ID,
			},
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "请求参数错误",
		},
		{
			name: "价格无效",
			requestBody: model.ProductCreateRequest{
				Name:        "新商品",
				Description: "这是一个新商品",
				Price:       -10.00, // 负价格
				Stock:       100,
				CategoryID:  suite.testCategory.ID,
			},
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "请求参数错误",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			w := suite.makeRequest("POST", "/api/v1/products/", tt.requestBody, suite.getAuthHeaders())

			assert.Equal(suite.T(), tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(suite.T(), err)

			if tt.expectedStatus == http.StatusOK {
				assert.Equal(suite.T(), tt.expectedMsg, response["message"])
				assert.NotNil(suite.T(), response["data"])

				// 验证返回的商品数据
				productData := response["data"].(map[string]interface{})
				assert.Equal(suite.T(), tt.requestBody.Name, productData["name"])
				assert.Equal(suite.T(), tt.requestBody.Description, productData["description"])
				assert.NotEmpty(suite.T(), productData["id"])
			} else {
				assert.Contains(suite.T(), response["message"].(string), tt.expectedMsg)
			}
		})
	}
}

// TestProductHandlerSuite 运行测试套件
func TestProductHandlerSuite(t *testing.T) {
	suite.Run(t, new(ProductHandlerTestSuite))
}
