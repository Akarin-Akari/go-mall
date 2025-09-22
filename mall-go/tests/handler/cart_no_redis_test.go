package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"mall-go/internal/config"
	"mall-go/internal/model"
	"mall-go/pkg/auth"
	"mall-go/pkg/response"
	testConfig "mall-go/tests/config"
	"mall-go/tests/helpers"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// SimpleCartHandler 简化的购物车处理器（不依赖Redis）
type SimpleCartHandler struct {
	db *gorm.DB
}

// NewSimpleCartHandler 创建简化的购物车处理器
func NewSimpleCartHandler(db *gorm.DB) *SimpleCartHandler {
	return &SimpleCartHandler{db: db}
}

// AddToCart 添加商品到购物车（简化版本）
func (h *SimpleCartHandler) AddToCart(c *gin.Context) {
	var req model.AddToCartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}

	// 获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "用户未认证")
		return
	}

	// 验证商品是否存在
	var product model.Product
	if err := h.db.First(&product, req.ProductID).Error; err != nil {
		response.Error(c, http.StatusNotFound, "商品不存在")
		return
	}

	// 检查库存
	if product.Stock < req.Quantity {
		response.Error(c, http.StatusBadRequest, "库存不足")
		return
	}

	// 查找或创建购物车
	var cart model.Cart
	err := h.db.Where("user_id = ?", userID).First(&cart).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 创建新购物车
			cart = model.Cart{
				UserID: userID.(uint),
			}
			if err := h.db.Create(&cart).Error; err != nil {
				response.Error(c, http.StatusInternalServerError, "创建购物车失败")
				return
			}
		} else {
			response.Error(c, http.StatusInternalServerError, "查询购物车失败")
			return
		}
	}

	// 查找是否已存在该商品
	var cartItem model.CartItem
	err = h.db.Where("cart_id = ? AND product_id = ?", cart.ID, req.ProductID).First(&cartItem).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 创建新的购物车项
			cartItem = model.CartItem{
				CartID:    cart.ID,
				ProductID: req.ProductID,
				Quantity:  req.Quantity,
				Selected:  true,
			}
			if err := h.db.Create(&cartItem).Error; err != nil {
				response.Error(c, http.StatusInternalServerError, "添加商品到购物车失败")
				return
			}
		} else {
			response.Error(c, http.StatusInternalServerError, "查询购物车商品失败")
			return
		}
	} else {
		// 更新数量
		cartItem.Quantity += req.Quantity
		if err := h.db.Save(&cartItem).Error; err != nil {
			response.Error(c, http.StatusInternalServerError, "更新购物车商品失败")
			return
		}
	}

	// 返回成功响应
	response.Success(c, "添加商品到购物车成功", gin.H{
		"cart_item_id": cartItem.ID,
		"quantity":     cartItem.Quantity,
	})
}

// TestSimpleCartAddNoRedis 不依赖Redis的购物车添加测试
func TestSimpleCartAddNoRedis(t *testing.T) {
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

	// 初始化测试辅助工具
	helper := helpers.NewTestHelper(db)

	// 创建测试数据
	testUser := helper.CreateTestUser("cartuser", "cart@example.com", "password123")
	testProduct := helper.CreateTestProduct("测试商品", "100.00", 50)

	// 生成JWT令牌
	token, err := auth.GenerateToken(testUser.ID, testUser.Username, testUser.Role)
	assert.NoError(t, err)

	// 初始化Handler
	handler := NewSimpleCartHandler(db)

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
	assert.NoError(t, err)

	// 检查响应
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "添加商品到购物车成功", response["message"])
	assert.NotNil(t, response["data"])

	// 验证数据库中的数据
	var cartItem model.CartItem
	err = db.Where("product_id = ?", testProduct.ID).First(&cartItem).Error
	assert.NoError(t, err)
	assert.Equal(t, testProduct.ID, cartItem.ProductID)
	assert.Equal(t, 2, cartItem.Quantity)

	// 清理
	helper.CleanupTestData()
}
