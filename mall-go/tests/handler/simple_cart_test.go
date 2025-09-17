package handler

import (
	"bytes"
	"encoding/json"
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
)

// TestSimpleCartAdd 简单的购物车添加测试
func TestSimpleCartAdd(t *testing.T) {
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
		&model.Category{},
		&model.Order{},
		&model.OrderItem{},
		&model.CartItem{},
		&model.Cart{},
	)
	assert.NoError(t, err)

	// 初始化Redis客户端（使用内存模拟）
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // 无密码
		DB:       1,  // 使用测试数据库
	})
	defer rdb.Close()

	// 测试Redis连接，如果失败则跳过
	_, err = rdb.Ping(rdb.Context()).Result()
	if err != nil {
		t.Skipf("Redis not available, skipping test: %v", err)
	}

	// 初始化测试辅助工具
	helper := helpers.NewTestHelper(db)

	// 创建测试数据
	testUser := helper.CreateTestUser("cartuser", "cart@example.com", "password123")
	testProduct := helper.CreateTestProduct("测试商品", "100.00", 50)

	// 生成JWT令牌
	token, err := auth.GenerateToken(testUser.ID, testUser.Username, testUser.Role)
	assert.NoError(t, err)

	// 初始化Handler
	handler := cart.NewCartHandler(db, rdb)

	// 设置路由
	router := gin.New()
	v1 := router.Group("/api/v1")
	{
		cartGroup := v1.Group("/cart")
		{
			cartGroup.POST("/add", handler.AddToCart)
		}
	}

	// 准备测试数据
	requestBody := model.AddToCartRequest{
		ProductID: testProduct.ID,
		Quantity:  2,
		SKUID:     0, // 可选
	}

	// 发送请求
	reqBodyBytes, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/api/v1/cart/add", bytes.NewBuffer(reqBodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 打印响应内容用于调试
	t.Logf("Response Status: %d", w.Code)
	t.Logf("Response Body: %s", w.Body.String())

	// 解析响应
	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// 打印响应结构
	t.Logf("Response Structure: %+v", response)

	// 检查响应
	if w.Code == http.StatusOK {
		assert.Equal(t, "添加商品到购物车成功", response["message"])
		assert.NotNil(t, response["data"])
	} else {
		t.Logf("Request failed with status %d", w.Code)
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
