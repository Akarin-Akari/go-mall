package integration

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"mall-go/tests/helpers"
)

// HealthTestSuite 健康检查测试套件
type HealthTestSuite struct {
	suite.Suite
	helper *helpers.TestHelper
}

// SetupSuite 测试套件初始化
func (suite *HealthTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
	suite.helper = helpers.NewTestHelper()
	
	// 设置简单的路由用于测试
	router := gin.New()
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"message": "Mall-Go API is running",
		})
	})
	suite.helper.Router = router
}

// TearDownSuite 测试套件清理
func (suite *HealthTestSuite) TearDownSuite() {
	suite.helper.Cleanup()
}

// SetupTest 每个测试前的准备
func (suite *HealthTestSuite) SetupTest() {
	suite.helper.CleanupDatabase()
}

// TestHealthCheck 测试健康检查接口
func (suite *HealthTestSuite) TestHealthCheck() {
	// Given - 准备测试环境
	// (无需特殊准备)

	// When - 执行健康检查请求
	w := suite.helper.MakeRequest("GET", "/health", nil, nil)

	// Then - 验证响应
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	response := suite.helper.AssertJSONResponse(w, http.StatusOK)
	assert.Equal(suite.T(), "ok", response["status"])
	assert.Equal(suite.T(), "Mall-Go API is running", response["message"])
}

// TestDatabaseConnection 测试数据库连接
func (suite *HealthTestSuite) TestDatabaseConnection() {
	// Given - 准备测试数据
	user := suite.helper.CreateTestUser("dbtest", "dbtest@example.com")

	// When - 查询数据库
	var count int64
	err := suite.helper.DB.Table("users").Where("username = ?", "dbtest").Count(&count).Error

	// Then - 验证结果
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(1), count)
	assert.Equal(suite.T(), "dbtest", user.Username)
	assert.Equal(suite.T(), "dbtest@example.com", user.Email)
}

// TestJWTTokenGeneration 测试JWT Token生成
func (suite *HealthTestSuite) TestJWTTokenGeneration() {
	// Given - 准备测试用户
	user := suite.helper.CreateTestUser("jwttest", "jwt@example.com")

	// When - 生成JWT Token
	token := suite.helper.GenerateTestToken(user.ID)

	// Then - 验证Token
	assert.NotEmpty(suite.T(), token)
	assert.Greater(suite.T(), len(token), 50) // JWT Token应该有一定长度
}

// TestTestHelperFunctions 测试辅助函数
func (suite *HealthTestSuite) TestTestHelperFunctions() {
	// 测试创建测试分类
	category := suite.helper.CreateTestCategory("测试分类", "这是一个测试分类")
	assert.NotZero(suite.T(), category.ID)
	assert.Equal(suite.T(), "测试分类", category.Name)

	// 测试创建测试商品
	product := suite.helper.CreateTestProduct("测试商品", "这是一个测试商品", 99.99)
	assert.NotZero(suite.T(), product.ID)
	assert.Equal(suite.T(), "测试商品", product.Name)
	assert.Equal(suite.T(), 99.99, product.Price)

	// 测试创建测试用户
	user := suite.helper.CreateTestUser("helpertest", "helper@example.com")
	assert.NotZero(suite.T(), user.ID)
	assert.Equal(suite.T(), "helpertest", user.Username)

	// 测试创建购物车项
	cartItem := suite.helper.CreateTestCartItem(user.ID, product.ID, 2)
	assert.NotZero(suite.T(), cartItem.ID)
	assert.Equal(suite.T(), user.ID, cartItem.UserID)
	assert.Equal(suite.T(), product.ID, cartItem.ProductID)
	assert.Equal(suite.T(), 2, cartItem.Quantity)

	// 测试创建订单
	order := suite.helper.CreateTestOrder(user.ID, 199.98)
	assert.NotZero(suite.T(), order.ID)
	assert.Equal(suite.T(), user.ID, order.UserID)
	assert.Equal(suite.T(), 199.98, order.TotalAmount)
}

// TestSeedTestData 测试种子数据
func (suite *HealthTestSuite) TestSeedTestData() {
	// When - 执行种子数据创建
	suite.helper.SeedTestData()

	// Then - 验证种子数据
	var userCount, productCount, categoryCount, cartItemCount, orderCount int64

	suite.helper.DB.Table("users").Count(&userCount)
	suite.helper.DB.Table("products").Count(&productCount)
	suite.helper.DB.Table("categories").Count(&categoryCount)
	suite.helper.DB.Table("cart_items").Count(&cartItemCount)
	suite.helper.DB.Table("orders").Count(&orderCount)

	assert.Equal(suite.T(), int64(1), userCount)
	assert.Equal(suite.T(), int64(2), productCount)
	assert.Equal(suite.T(), int64(1), categoryCount)
	assert.Equal(suite.T(), int64(2), cartItemCount)
	assert.Equal(suite.T(), int64(1), orderCount)
}

// TestCleanupDatabase 测试数据库清理
func (suite *HealthTestSuite) TestCleanupDatabase() {
	// Given - 创建一些测试数据
	suite.helper.SeedTestData()

	// 验证数据存在
	var count int64
	suite.helper.DB.Table("users").Count(&count)
	assert.Greater(suite.T(), count, int64(0))

	// When - 执行数据库清理
	suite.helper.CleanupDatabase()

	// Then - 验证数据已清理
	suite.helper.DB.Table("users").Count(&count)
	assert.Equal(suite.T(), int64(0), count)

	suite.helper.DB.Table("products").Count(&count)
	assert.Equal(suite.T(), int64(0), count)

	suite.helper.DB.Table("categories").Count(&count)
	assert.Equal(suite.T(), int64(0), count)
}

// 运行测试套件
func TestHealthTestSuite(t *testing.T) {
	suite.Run(t, new(HealthTestSuite))
}
