package helpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"mall-go/internal/model"
	"mall-go/tests/config"
)

// TestHelper 测试辅助工具结构
type TestHelper struct {
	DB     *gorm.DB
	Router *gin.Engine
}

// NewTestHelper 创建测试辅助工具
func NewTestHelper(db *gorm.DB) *TestHelper {
	return &TestHelper{
		DB: db,
	}
}

// Cleanup 清理测试环境
func (h *TestHelper) Cleanup() {
	config.CleanupTestDB(h.DB)
}

// MakeRequest 创建HTTP测试请求
func (h *TestHelper) MakeRequest(method, url string, body interface{}, headers map[string]string) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		json.NewEncoder(&buf).Encode(body)
	}

	req := httptest.NewRequest(method, url, &buf)
	req.Header.Set("Content-Type", "application/json")

	// 添加自定义头部
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	w := httptest.NewRecorder()
	if h.Router != nil {
		h.Router.ServeHTTP(w, req)
	}

	return w
}

// MakeAuthenticatedRequest 创建带认证的HTTP测试请求
func (h *TestHelper) MakeAuthenticatedRequest(method, url string, body interface{}, userID uint) *httptest.ResponseRecorder {
	token := h.GenerateTestToken(userID)
	headers := map[string]string{
		"Authorization": "Bearer " + token,
	}
	return h.MakeRequest(method, url, body, headers)
}

// GenerateTestToken 生成测试JWT Token
func (h *TestHelper) GenerateTestToken(userID uint) string {
	testConfig := config.GetTestConfig()

	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(testConfig.JWT.Secret))
	if err != nil {
		panic("failed to generate test token: " + err.Error())
	}

	return tokenString
}

// CreateTestUser 创建测试用户
func (h *TestHelper) CreateTestUser(username, email, password string) *model.User {
	user := &model.User{
		Username: username,
		Email:    email,
		Status:   "active",
		Role:     "user",
		Nickname: username,
	}

	// 设置密码
	if err := user.SetPassword(password); err != nil {
		panic("failed to set password: " + err.Error())
	}

	if err := h.DB.Create(user).Error; err != nil {
		panic("failed to create test user: " + err.Error())
	}

	return user
}

// CreateTestProduct 创建测试商品
func (h *TestHelper) CreateTestProduct(name, price string, stock int) *model.Product {
	// 先创建一个默认分类（如果不存在）
	var category model.Category
	err := h.DB.Where("name = ?", "默认分类").First(&category).Error
	if err != nil {
		category = model.Category{
			Name:        "默认分类",
			Description: "测试用默认分类",
			Status:      "active",
		}
		h.DB.Create(&category)
	}

	product := &model.Product{
		Name:        name,
		Description: "测试商品描述",
		Price:       decimal.RequireFromString(price),
		Stock:       stock,
		Status:      "active",
		CategoryID:  category.ID,
	}

	if err := h.DB.Create(product).Error; err != nil {
		panic("failed to create test product: " + err.Error())
	}

	return product
}

// CreateTestCategory 创建测试分类
func (h *TestHelper) CreateTestCategory(name, description string) *model.Category {
	category := &model.Category{
		Name:        name,
		Description: description,
		Status:      "active",
	}

	if err := h.DB.Create(category).Error; err != nil {
		panic("failed to create test category: " + err.Error())
	}

	return category
}

// CreateTestOrder 创建测试订单
func (h *TestHelper) CreateTestOrder(userID, productID uint) *model.Order {
	order := &model.Order{
		UserID:          userID,
		OrderNo:         fmt.Sprintf("TEST%d", time.Now().Unix()),
		TotalAmount:     decimal.NewFromFloat(100.00),
		PayableAmount:   decimal.NewFromFloat(100.00),
		Status:          "pending",
		PaymentStatus:   "unpaid",
		ReceiverName:    "测试收货人",
		ReceiverPhone:   "13800138000",
		ReceiverAddress: "测试地址",
	}

	if err := h.DB.Create(order).Error; err != nil {
		panic("failed to create test order: " + err.Error())
	}

	// 创建订单项
	orderItem := &model.OrderItem{
		OrderID:   order.ID,
		ProductID: productID,
		Quantity:  1,
		Price:     decimal.NewFromFloat(100.00),
	}

	if err := h.DB.Create(orderItem).Error; err != nil {
		panic("failed to create test order item: " + err.Error())
	}

	return order
}

// CreateTestCartItem 创建测试购物车项
func (h *TestHelper) CreateTestCartItem(userID, productID uint, quantity int) *model.CartItem {
	// 首先创建或获取购物车
	var cart model.Cart
	err := h.DB.Where("user_id = ?", userID).First(&cart).Error
	if err != nil {
		// 如果购物车不存在，创建一个
		cart = model.Cart{
			UserID: userID,
			Status: "active",
		}
		h.DB.Create(&cart)
	}

	// 获取商品信息用于快照
	var product model.Product
	h.DB.First(&product, productID)

	cartItem := &model.CartItem{
		CartID:       cart.ID,
		ProductID:    productID,
		Quantity:     quantity,
		Price:        product.Price,
		ProductName:  product.Name,
		ProductImage: "",
		Selected:     true,
	}

	if err := h.DB.Create(cartItem).Error; err != nil {
		panic("failed to create test cart item: " + err.Error())
	}

	return cartItem
}

// CleanupDatabase 清理数据库数据
func (h *TestHelper) CleanupDatabase() {
	// 按照外键依赖顺序删除数据
	tables := []string{
		"order_items",
		"orders",
		"cart_items",
		"product_images",
		"products",
		"categories",
		"users",
	}

	for _, table := range tables {
		h.DB.Exec("DELETE FROM " + table)
	}
}

// CleanupTestData 清理测试数据
func (h *TestHelper) CleanupTestData() {
	h.CleanupDatabase()
}

// CleanupCartData 清理购物车数据
func (h *TestHelper) CleanupCartData() {
	h.DB.Exec("DELETE FROM cart_items")
}

// CleanupOrderData 清理订单数据
func (h *TestHelper) CleanupOrderData() {
	h.DB.Exec("DELETE FROM order_items")
	h.DB.Exec("DELETE FROM orders")
}

// CleanupProductData 清理商品数据
func (h *TestHelper) CleanupProductData() {
	h.DB.Exec("DELETE FROM product_images")
	h.DB.Exec("DELETE FROM products")
}

// AssertJSONResponse 断言JSON响应
func (h *TestHelper) AssertJSONResponse(w *httptest.ResponseRecorder, expectedCode int) map[string]interface{} {
	if w.Code != expectedCode {
		panic(fmt.Sprintf("Expected status code %d, got %d. Response: %s", expectedCode, w.Code, w.Body.String()))
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		panic("Failed to unmarshal JSON response: " + err.Error())
	}

	return response
}

// SeedTestData 种子测试数据
func (h *TestHelper) SeedTestData() {
	// 创建测试分类
	category := h.CreateTestCategory("电子产品", "各种电子产品")

	// 创建测试用户
	user := h.CreateTestUser("testuser", "test@example.com", "password123")

	// 创建测试商品
	product1 := h.CreateTestProduct("iPhone 15", "8999.00", 50)
	product1.CategoryID = category.ID
	h.DB.Save(product1)

	product2 := h.CreateTestProduct("MacBook Pro", "15999.00", 30)
	product2.CategoryID = category.ID
	h.DB.Save(product2)

	// 创建测试购物车项
	h.CreateTestCartItem(user.ID, product1.ID, 1)
	h.CreateTestCartItem(user.ID, product2.ID, 2)

	// 创建测试订单
	h.CreateTestOrder(user.ID, product1.ID)

	// 订单项已在CreateTestOrder中创建
}
