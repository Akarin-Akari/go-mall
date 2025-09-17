//go:build ignore

package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"mall-go/internal/config"
	"mall-go/internal/model"
	"mall-go/pkg/database"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"gorm.io/gorm"
)

// 用户模拟数据
type User struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"-"`
}

// 商品模拟数据
type Product struct {
	ID          uint     `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Price       string   `json:"price"`
	Stock       int      `json:"stock"`
	CategoryID  uint     `json:"category_id"`
	Images      []string `json:"images"`
	Rating      float64  `json:"rating"`
	SalesCount  int      `json:"sales_count"`
	CreatedAt   string   `json:"created_at"`
	UpdatedAt   string   `json:"updated_at"`
}

// 购物车项模拟数据
type CartItem struct {
	ID          uint   `json:"id"`
	ProductID   uint   `json:"product_id"`
	ProductName string `json:"product_name"`
	Price       string `json:"price"`
	Quantity    int    `json:"quantity"`
	Image       string `json:"image"`
	Selected    bool   `json:"selected"`
}

// JWT密钥
var jwtSecret = []byte("your-secret-key")

// 模拟数据
var users = []User{
	{ID: 1, Username: "admin", Email: "admin@example.com", Password: "123456"},
	{ID: 2, Username: "user", Email: "user@example.com", Password: "123456"},
}

var products = []Product{
	{
		ID: 1, Name: "iPhone 15 Pro Max", Description: "最新款iPhone", Price: "8999.00", Stock: 50,
		CategoryID: 1, Images: []string{"https://images.unsplash.com/photo-1592750475338-74b7b21085ab?w=400"},
		Rating: 4.8, SalesCount: 1250, CreatedAt: "2024-01-15T10:00:00Z", UpdatedAt: "2024-01-15T10:00:00Z",
	},
	{
		ID: 2, Name: "MacBook Pro 14英寸", Description: "M3芯片MacBook", Price: "13999.00", Stock: 30,
		CategoryID: 1, Images: []string{"https://images.unsplash.com/photo-1517336714731-489689fd1ca8?w=400"},
		Rating: 4.9, SalesCount: 856, CreatedAt: "2024-01-10T10:00:00Z", UpdatedAt: "2024-01-10T10:00:00Z",
	},
}

var cartItems = []CartItem{
	{ID: 1, ProductID: 1, ProductName: "iPhone 15 Pro Max", Price: "8999.00", Quantity: 1, Image: "https://images.unsplash.com/photo-1592750475338-74b7b21085ab?w=400", Selected: true},
	{ID: 2, ProductID: 2, ProductName: "MacBook Pro 14英寸", Price: "13999.00", Quantity: 1, Image: "https://images.unsplash.com/photo-1517336714731-489689fd1ca8?w=400", Selected: true},
}

// 生成JWT Token
func generateToken(userID uint) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// JWT中间件
func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未提供认证token"})
			c.Abort()
			return
		}

		// 移除Bearer前缀
		if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
			tokenString = tokenString[7:]
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "无效的token"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "无效的token格式"})
			c.Abort()
			return
		}

		userID := uint(claims["user_id"].(float64))
		c.Set("user_id", userID)
		c.Next()
	}
}

// 全局数据库连接
var db *gorm.DB

// 自动迁移数据库表
func autoMigrate() error {
	// 检查表是否存在，如果存在则跳过迁移
	if db.Migrator().HasTable(&model.User{}) {
		log.Println("数据库表已存在，跳过迁移")
		return nil
	}

	return db.AutoMigrate(
		&model.User{},
		&model.Product{},
		&model.Cart{},
		&model.CartItem{},
		&model.Order{},
		&model.OrderItem{},
		&model.Payment{},
		&model.File{},
	)
}

func main() {
	fmt.Println("🚀 启动Mall Go API服务器...")

	// 加载配置
	config.Load()
	fmt.Println("✅ 配置加载成功")

	// 初始化数据库连接
	db = database.Init()
	fmt.Println("✅ 数据库连接成功")

	// 自动迁移数据库表
	if err := autoMigrate(); err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}
	fmt.Println("✅ 数据库表结构同步完成")

	r := gin.Default()

	// CORS中间件
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	// 基础路由
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "Mall Go API Server is running!",
			"data": gin.H{
				"version": "1.0.0",
				"status":  "ok",
			},
		})
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "success",
			"data": gin.H{
				"status": "healthy",
				"time":   time.Now().Format("2006-01-02 15:04:05"),
			},
		})
	})

	// API路由组
	api := r.Group("/api/v1")
	{
		// 认证相关
		auth := api.Group("/auth")
		{
			auth.POST("/login", loginHandler)
			auth.POST("/register", registerHandler)
		}

		// 商品相关
		products := api.Group("/products")
		{
			products.GET("", getProductsHandler)
			products.GET("/:id", getProductDetailHandler)
		}

		// 购物车相关（需要认证）
		cart := api.Group("/cart")
		cart.Use(authMiddleware())
		{
			cart.GET("", getCartHandler)
			cart.POST("/items", addToCartHandler)
			cart.PUT("/items/:id", updateCartItemHandler)
			cart.DELETE("/items/:id", removeCartItemHandler)
			cart.DELETE("/clear", clearCartHandler)
		}
	}

	fmt.Println("🚀 Mall Go API Server starting on port 8080")
	r.Run(":8080")
}

// 登录处理器
func loginHandler(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	// 从数据库查找用户
	var user model.User
	if err := db.Where("username = ?", req.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "用户名或密码错误"})
		return
	}

	// 验证密码（这里简化处理，实际应该使用bcrypt等加密）
	if user.Password != req.Password {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "用户名或密码错误"})
		return
	}

	// 生成token
	token, err := generateToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "生成token失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "登录成功",
		"data": gin.H{
			"token": token,
			"user": gin.H{
				"id":       user.ID,
				"username": user.Username,
				"email":    user.Email,
			},
		},
	})
}

// 注册处理器
func registerHandler(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	// 检查用户是否已存在
	for _, u := range users {
		if u.Username == req.Username || u.Email == req.Email {
			c.JSON(http.StatusConflict, gin.H{"code": 409, "message": "用户名或邮箱已存在"})
			return
		}
	}

	// 创建新用户
	newUser := User{
		ID:       uint(len(users) + 1),
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
	}
	users = append(users, newUser)

	c.JSON(http.StatusCreated, gin.H{
		"code":    0,
		"message": "注册成功",
		"data": gin.H{
			"id":       newUser.ID,
			"username": newUser.Username,
			"email":    newUser.Email,
		},
	})
}

// 获取商品列表
func getProductsHandler(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	start := (page - 1) * pageSize
	end := start + pageSize

	if start >= len(products) {
		start = len(products)
	}
	if end > len(products) {
		end = len(products)
	}

	result := products[start:end]

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"list":      result,
			"total":     len(products),
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// 获取商品详情
func getProductDetailHandler(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的商品ID"})
		return
	}

	for _, product := range products {
		if product.ID == uint(id) {
			c.JSON(http.StatusOK, gin.H{
				"code":    0,
				"message": "success",
				"data":    product,
			})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "商品不存在"})
}

// 获取购物车
func getCartHandler(c *gin.Context) {
	totalAmount := 0.0
	totalQuantity := 0

	for _, item := range cartItems {
		if item.Selected {
			price, _ := strconv.ParseFloat(item.Price, 64)
			totalAmount += price * float64(item.Quantity)
			totalQuantity += item.Quantity
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"items":          cartItems,
			"total_amount":   fmt.Sprintf("%.2f", totalAmount),
			"total_quantity": totalQuantity,
		},
	})
}

// 添加到购物车
func addToCartHandler(c *gin.Context) {
	var req struct {
		ProductID uint `json:"product_id" binding:"required"`
		Quantity  int  `json:"quantity" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	// 查找商品
	var product *Product
	for _, p := range products {
		if p.ID == req.ProductID {
			product = &p
			break
		}
	}

	if product == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "商品不存在"})
		return
	}

	// 检查是否已在购物车中
	for i, item := range cartItems {
		if item.ProductID == req.ProductID {
			cartItems[i].Quantity += req.Quantity
			c.JSON(http.StatusOK, gin.H{
				"code":    0,
				"message": "添加成功",
				"data":    cartItems[i],
			})
			return
		}
	}

	// 添加新项目
	newItem := CartItem{
		ID:          uint(len(cartItems) + 1),
		ProductID:   product.ID,
		ProductName: product.Name,
		Price:       product.Price,
		Quantity:    req.Quantity,
		Image:       product.Images[0],
		Selected:    true,
	}
	cartItems = append(cartItems, newItem)

	c.JSON(http.StatusCreated, gin.H{
		"code":    0,
		"message": "添加成功",
		"data":    newItem,
	})
}

// 更新购物车项
func updateCartItemHandler(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的ID"})
		return
	}

	var req struct {
		Quantity int   `json:"quantity"`
		Selected *bool `json:"selected"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	for i, item := range cartItems {
		if item.ID == uint(id) {
			if req.Quantity > 0 {
				cartItems[i].Quantity = req.Quantity
			}
			if req.Selected != nil {
				cartItems[i].Selected = *req.Selected
			}
			c.JSON(http.StatusOK, gin.H{
				"code":    0,
				"message": "更新成功",
				"data":    cartItems[i],
			})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "购物车项不存在"})
}

// 删除购物车项
func removeCartItemHandler(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的ID"})
		return
	}

	for i, item := range cartItems {
		if item.ID == uint(id) {
			cartItems = append(cartItems[:i], cartItems[i+1:]...)
			c.JSON(http.StatusOK, gin.H{
				"code":    0,
				"message": "删除成功",
			})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "购物车项不存在"})
}

// 清空购物车
func clearCartHandler(c *gin.Context) {
	cartItems = []CartItem{}
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "清空成功",
	})
}
