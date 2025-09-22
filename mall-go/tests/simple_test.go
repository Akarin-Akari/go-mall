package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"mall-go/tests/config"
	"mall-go/tests/helpers"
)

// TestEnvironmentSetup 测试环境搭建验证
func TestEnvironmentSetup(t *testing.T) {
	// 测试配置加载
	testConfig := config.LoadTestConfig()
	assert.NotNil(t, testConfig)
	assert.Equal(t, "sqlite", testConfig.Database.Driver)
	assert.Equal(t, ":memory:", testConfig.Database.Database)
	assert.NotEmpty(t, testConfig.JWT.Secret)
}

// TestDatabaseConnection 测试数据库连接
func TestDatabaseConnection(t *testing.T) {
	helper := helpers.NewTestHelper()
	defer helper.Cleanup()

	// 测试数据库连接
	assert.NotNil(t, helper.DB)

	// 测试创建用户
	user := helper.CreateTestUser("testuser", "test@example.com")
	assert.NotZero(t, user.ID)
	assert.Equal(t, "testuser", user.Username)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, "active", user.Status)
}

// TestJWTTokenGeneration 测试JWT Token生成
func TestJWTTokenGeneration(t *testing.T) {
	helper := helpers.NewTestHelper()
	defer helper.Cleanup()

	// 创建测试用户
	user := helper.CreateTestUser("jwttest", "jwt@example.com")

	// 生成JWT Token
	token := helper.GenerateTestToken(user.ID)
	assert.NotEmpty(t, token)
	assert.Greater(t, len(token), 50) // JWT Token应该有一定长度
}

// TestModelCreation 测试模型创建
func TestModelCreation(t *testing.T) {
	helper := helpers.NewTestHelper()
	defer helper.Cleanup()

	// 测试创建分类
	category := helper.CreateTestCategory("测试分类", "这是一个测试分类")
	assert.NotZero(t, category.ID)
	assert.Equal(t, "测试分类", category.Name)
	assert.Equal(t, "active", category.Status)

	// 测试创建商品
	product := helper.CreateTestProduct("测试商品", "这是一个测试商品", 99.99)
	assert.NotZero(t, product.ID)
	assert.Equal(t, "测试商品", product.Name)
	assert.Equal(t, "99.99", product.Price.String())
	assert.Equal(t, "on_sale", product.Status)

	// 测试创建用户
	user := helper.CreateTestUser("modeltest", "model@example.com")
	assert.NotZero(t, user.ID)

	// 测试创建购物车项
	cartItem := helper.CreateTestCartItem(user.ID, product.ID, 2)
	assert.NotZero(t, cartItem.ID)
	assert.Equal(t, product.ID, cartItem.ProductID)
	assert.Equal(t, 2, cartItem.Quantity)
	assert.True(t, cartItem.Selected)

	// 测试创建订单
	order := helper.CreateTestOrder(user.ID, 199.98)
	assert.NotZero(t, order.ID)
	assert.Equal(t, user.ID, order.UserID)
	assert.Equal(t, "199.98", order.TotalAmount.String())
	assert.Equal(t, "pending", order.Status)
	assert.Equal(t, "unpaid", order.PaymentStatus)
}

// TestDatabaseCleanup 测试数据库清理
func TestDatabaseCleanup(t *testing.T) {
	helper := helpers.NewTestHelper()
	defer helper.Cleanup()

	// 创建一些测试数据
	helper.SeedTestData()

	// 验证数据存在
	var count int64
	helper.DB.Table("users").Count(&count)
	assert.Greater(t, count, int64(0))

	helper.DB.Table("products").Count(&count)
	assert.Greater(t, count, int64(0))

	// 执行数据库清理
	helper.CleanupDatabase()

	// 验证数据已清理
	helper.DB.Table("users").Count(&count)
	assert.Equal(t, int64(0), count)

	helper.DB.Table("products").Count(&count)
	assert.Equal(t, int64(0), count)

	helper.DB.Table("categories").Count(&count)
	assert.Equal(t, int64(0), count)
}
