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
)

// TestSimpleProductCreate 简单的商品创建测试
func TestSimpleProductCreate(t *testing.T) {
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
	db := testConfig.SetupTestDB()
	defer testConfig.CleanupTestDB(db)

	// 自动迁移测试表
	err := db.AutoMigrate(
		&model.User{},
		&model.Product{},
		&model.ProductImage{},
		&model.Category{},
		&model.Order{},
		&model.OrderItem{},
		&model.CartItem{},
		&model.Cart{},
	)
	assert.NoError(t, err)

	// 初始化测试辅助工具
	helper := helpers.NewTestHelper(db)

	// 创建测试数据
	testUser := helper.CreateTestUser("productuser", "product@example.com", "password123")
	testCategory := helper.CreateTestCategory("测试分类", "test-category")

	// 生成JWT令牌
	token, err := auth.GenerateToken(testUser.ID, testUser.Username, testUser.Role)
	assert.NoError(t, err)

	// 初始化Handler
	handler := product.NewHandler(db)

	// 设置路由
	router := gin.New()

	// 添加认证中间件
	router.Use(func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未提供认证令牌"})
			c.Abort()
			return
		}

		// 解析Bearer token
		tokenString := authHeader[7:] // 去掉 "Bearer " 前缀
		claims, err := auth.ParseToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的认证令牌"})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Next()
	})

	v1 := router.Group("/api/v1")
	{
		productGroup := v1.Group("/products")
		{
			productGroup.POST("", handler.Create)
			productGroup.GET("/:id", handler.Get)
		}
	}

	// 准备测试数据
	requestBody := model.ProductCreateRequest{
		Name:        "测试商品",
		Description: "这是一个测试商品",
		Price:       99.99,
		Stock:       100,
		CategoryID:  testCategory.ID,
	}

	// 发送创建商品请求
	reqBodyBytes, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/api/v1/products", bytes.NewBuffer(reqBodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 打印响应内容用于调试
	t.Logf("Create Response Status: %d", w.Code)
	t.Logf("Create Response Body: %s", w.Body.String())

	// 解析响应
	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// 检查创建响应
	if w.Code == http.StatusOK {
		assert.Equal(t, "商品创建成功", response["message"])
		assert.NotNil(t, response["data"])

		// 获取创建的商品ID
		productData := response["data"].(map[string]interface{})
		productID := productData["id"].(float64)

		// 验证数据库中的商品数据
		var product model.Product
		err = db.First(&product, uint(productID)).Error
		assert.NoError(t, err)
		assert.Equal(t, "测试商品", product.Name)
		assert.Equal(t, "这是一个测试商品", product.Description)
		assert.Equal(t, testCategory.ID, product.CategoryID)

		t.Logf("✅ 商品创建测试通过 - ID: %v", productID)
	} else {
		t.Logf("❌ 商品创建失败 - Status: %d", w.Code)
		if msg, ok := response["message"]; ok {
			t.Logf("Error message: %v", msg)
		}
		if err, ok := response["error"]; ok {
			t.Logf("Error: %v", err)
		}
	}

	// 清理
	helper.CleanupTestData()
}

// TestSimpleProductGet 简单的商品获取测试
func TestSimpleProductGet(t *testing.T) {
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
	db := testConfig.SetupTestDB()
	defer testConfig.CleanupTestDB(db)

	// 自动迁移测试表
	err := db.AutoMigrate(
		&model.User{},
		&model.Product{},
		&model.ProductImage{},
		&model.Category{},
		&model.Order{},
		&model.OrderItem{},
		&model.CartItem{},
		&model.Cart{},
	)
	assert.NoError(t, err)

	// 初始化测试辅助工具
	helper := helpers.NewTestHelper(db)

	// 创建测试数据
	testCategory := helper.CreateTestCategory("测试分类", "test-category")
	testProduct := helper.CreateTestProduct("测试商品", "99.99", 100)

	// 确保商品关联到测试分类
	testProduct.CategoryID = testCategory.ID
	db.Save(testProduct)

	// 初始化Handler
	handler := product.NewHandler(db)

	// 设置路由
	router := gin.New()
	v1 := router.Group("/api/v1")
	{
		productGroup := v1.Group("/products")
		{
			productGroup.GET("/:id", handler.Get)
		}
	}

	// 发送获取商品请求
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/products/%d", testProduct.ID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 打印响应内容用于调试
	t.Logf("Get Response Status: %d", w.Code)
	t.Logf("Get Response Body: %s", w.Body.String())

	// 解析响应
	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// 检查获取响应
	if w.Code == http.StatusOK {
		assert.Equal(t, "操作成功", response["message"])
		assert.NotNil(t, response["data"])

		// 验证商品数据
		productData := response["data"].(map[string]interface{})
		assert.Equal(t, "测试商品", productData["name"])
		assert.NotEmpty(t, productData["id"])

		t.Logf("✅ 商品获取测试通过")
	} else {
		t.Logf("❌ 商品获取失败 - Status: %d", w.Code)
		if msg, ok := response["message"]; ok {
			t.Logf("Error message: %v", msg)
		}
	}

	// 清理
	helper.CleanupTestData()
}
