package handler

import (
	"testing"

	"mall-go/internal/model"
	"mall-go/tests/config"

	"github.com/stretchr/testify/assert"
)

// TestBasicSetup 测试基础设施设置
func TestBasicSetup(t *testing.T) {
	// 初始化测试数据库
	db := config.SetupTestDB()
	defer config.CleanupTestDB(db)

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

	// 测试创建用户
	user := &model.User{
		Username: "testuser",
		Email:    "test@example.com",
		Nickname: "测试用户",
		Role:     "user",
		Status:   "active",
	}

	// 设置密码
	err = user.SetPassword("password123")
	assert.NoError(t, err)

	// 创建用户
	err = db.Create(user).Error
	assert.NoError(t, err)
	assert.NotZero(t, user.ID)

	// 验证用户创建成功
	var foundUser model.User
	err = db.First(&foundUser, user.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, "testuser", foundUser.Username)
	assert.Equal(t, "test@example.com", foundUser.Email)

	// 验证密码
	assert.True(t, foundUser.CheckPassword("password123"))
	assert.False(t, foundUser.CheckPassword("wrongpassword"))

	t.Log("基础设施测试通过")
}
