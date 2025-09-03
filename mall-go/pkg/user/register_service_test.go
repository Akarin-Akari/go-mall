package user

import (
	"testing"

	"mall-go/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// RegisterServiceTestSuite 注册服务测试套件
type RegisterServiceTestSuite struct {
	suite.Suite
	db              *gorm.DB
	registerService *RegisterService
}

// SetupSuite 设置测试套件
func (suite *RegisterServiceTestSuite) SetupSuite() {
	// 创建内存数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	suite.Require().NoError(err)
	suite.db = db

	// 自动迁移
	err = db.AutoMigrate(&model.User{}, &model.UserProfile{}, &model.UserVerificationCode{})
	suite.Require().NoError(err)

	// 创建注册服务
	suite.registerService = NewRegisterService(db)
}

// TestRegisterService_Register 测试用户注册
func (suite *RegisterServiceTestSuite) TestRegisterService_Register() {
	req := &RegisterRequest{
		Username:        "testuser",
		Email:           "test@example.com",
		Password:        "password123",
		ConfirmPassword: "password123",
		Nickname:        "Test User",
		AgreeTerms:      true,
	}

	resp, err := suite.registerService.Register(req)
	suite.NoError(err)
	suite.NotNil(resp)

	// 验证响应
	suite.Equal("testuser", resp.Username)
	suite.Equal("test@example.com", resp.Email)
	suite.Equal("Test User", resp.Nickname)
	suite.NotEmpty(resp.Token)
	suite.Equal("注册成功", resp.Message)

	// 验证数据库记录
	var user model.User
	err = suite.db.Where("username = ?", "testuser").First(&user).Error
	suite.NoError(err)
	suite.Equal("testuser", user.Username)
	suite.Equal("test@example.com", user.Email)
	suite.Equal("Test User", user.Nickname)
	suite.Equal(model.RoleUser, user.Role)
	suite.Equal(model.StatusActive, user.Status)
}

// TestRegisterService_Register_DuplicateUsername 测试重复用户名注册
func (suite *RegisterServiceTestSuite) TestRegisterService_Register_DuplicateUsername() {
	// 先创建一个用户
	user := &model.User{
		Username: "duplicate",
		Email:    "first@example.com",
		Password: "hashedpassword",
	}
	suite.db.Create(user)

	// 尝试用相同用户名注册
	req := &RegisterRequest{
		Username:        "duplicate",
		Email:           "second@example.com",
		Password:        "password123",
		ConfirmPassword: "password123",
		AgreeTerms:      true,
	}

	resp, err := suite.registerService.Register(req)
	suite.Error(err)
	suite.Nil(resp)
	suite.Contains(err.Error(), "用户名已存在")
}

// TestRegisterService_Register_DuplicateEmail 测试重复邮箱注册
func (suite *RegisterServiceTestSuite) TestRegisterService_Register_DuplicateEmail() {
	// 先创建一个用户
	user := &model.User{
		Username: "user1",
		Email:    "duplicate@example.com",
		Password: "hashedpassword",
	}
	suite.db.Create(user)

	// 尝试用相同邮箱注册
	req := &RegisterRequest{
		Username:        "user2",
		Email:           "duplicate@example.com",
		Password:        "password123",
		ConfirmPassword: "password123",
		AgreeTerms:      true,
	}

	resp, err := suite.registerService.Register(req)
	suite.Error(err)
	suite.Nil(resp)
	suite.Contains(err.Error(), "邮箱已被注册")
}

// TestRegisterService_Register_PasswordMismatch 测试密码不匹配
func (suite *RegisterServiceTestSuite) TestRegisterService_Register_PasswordMismatch() {
	req := &RegisterRequest{
		Username:        "testuser2",
		Email:           "test2@example.com",
		Password:        "password123",
		ConfirmPassword: "password456",
		AgreeTerms:      true,
	}

	resp, err := suite.registerService.Register(req)
	suite.Error(err)
	suite.Nil(resp)
	suite.Contains(err.Error(), "两次输入的密码不一致")
}

// TestRegisterService_Register_InvalidData 测试无效数据
func (suite *RegisterServiceTestSuite) TestRegisterService_Register_InvalidData() {
	tests := []struct {
		name string
		req  *RegisterRequest
		err  string
	}{
		{
			name: "用户名太短",
			req: &RegisterRequest{
				Username:        "ab",
				Email:           "test@example.com",
				Password:        "password123",
				ConfirmPassword: "password123",
				AgreeTerms:      true,
			},
			err: "用户名格式不正确",
		},
		{
			name: "邮箱格式错误",
			req: &RegisterRequest{
				Username:        "testuser",
				Email:           "invalid-email",
				Password:        "password123",
				ConfirmPassword: "password123",
				AgreeTerms:      true,
			},
			err: "邮箱格式不正确",
		},
		{
			name: "密码太短",
			req: &RegisterRequest{
				Username:        "testuser",
				Email:           "test@example.com",
				Password:        "123",
				ConfirmPassword: "123",
				AgreeTerms:      true,
			},
			err: "密码强度不够",
		},
		{
			name: "未同意服务条款",
			req: &RegisterRequest{
				Username:        "testuser",
				Email:           "test@example.com",
				Password:        "password123",
				ConfirmPassword: "password123",
				AgreeTerms:      false,
			},
			err: "请同意服务条款",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			resp, err := suite.registerService.Register(tt.req)
			suite.Error(err)
			suite.Nil(resp)
			suite.Contains(err.Error(), tt.err)
		})
	}
}

// TestRegisterService_IsUsernameExists 测试检查用户名是否存在
func (suite *RegisterServiceTestSuite) TestRegisterService_IsUsernameExists() {
	// 创建测试用户
	user := &model.User{
		Username: "existinguser",
		Email:    "existing@example.com",
		Password: "hashedpassword",
	}
	suite.db.Create(user)

	// 测试存在的用户名
	exists, err := suite.registerService.IsUsernameExists("existinguser")
	suite.NoError(err)
	suite.True(exists)

	// 测试不存在的用户名
	exists, err = suite.registerService.IsUsernameExists("nonexistentuser")
	suite.NoError(err)
	suite.False(exists)
}

// TestRegisterService_IsEmailExists 测试检查邮箱是否存在
func (suite *RegisterServiceTestSuite) TestRegisterService_IsEmailExists() {
	// 创建测试用户
	user := &model.User{
		Username: "testuser",
		Email:    "existing@example.com",
		Password: "hashedpassword",
	}
	suite.db.Create(user)

	// 测试存在的邮箱
	exists, err := suite.registerService.IsEmailExists("existing@example.com")
	suite.NoError(err)
	suite.True(exists)

	// 测试不存在的邮箱
	exists, err = suite.registerService.IsEmailExists("nonexistent@example.com")
	suite.NoError(err)
	suite.False(exists)
}

// TestRegisterService_CheckUsername 测试检查用户名可用性
func (suite *RegisterServiceTestSuite) TestRegisterService_CheckUsername() {
	// 创建测试用户
	user := &model.User{
		Username: "takenuser",
		Email:    "taken@example.com",
		Password: "hashedpassword",
	}
	suite.db.Create(user)

	// 测试已被占用的用户名
	req := &CheckUsernameRequest{Username: "takenuser"}
	available, err := suite.registerService.CheckUsername(req)
	suite.NoError(err)
	suite.False(available)

	// 测试可用的用户名
	req = &CheckUsernameRequest{Username: "availableuser"}
	available, err = suite.registerService.CheckUsername(req)
	suite.NoError(err)
	suite.True(available)

	// 测试无效的用户名格式
	req = &CheckUsernameRequest{Username: "ab"}
	available, err = suite.registerService.CheckUsername(req)
	suite.Error(err)
	suite.False(available)
	suite.Contains(err.Error(), "用户名格式不正确")
}

// TestRegisterService_CheckEmail 测试检查邮箱可用性
func (suite *RegisterServiceTestSuite) TestRegisterService_CheckEmail() {
	// 创建测试用户
	user := &model.User{
		Username: "testuser",
		Email:    "taken@example.com",
		Password: "hashedpassword",
	}
	suite.db.Create(user)

	// 测试已被占用的邮箱
	req := &CheckEmailRequest{Email: "taken@example.com"}
	available, err := suite.registerService.CheckEmail(req)
	suite.NoError(err)
	suite.False(available)

	// 测试可用的邮箱
	req = &CheckEmailRequest{Email: "available@example.com"}
	available, err = suite.registerService.CheckEmail(req)
	suite.NoError(err)
	suite.True(available)

	// 测试无效的邮箱格式
	req = &CheckEmailRequest{Email: "invalid-email"}
	available, err = suite.registerService.CheckEmail(req)
	suite.Error(err)
	suite.False(available)
	suite.Contains(err.Error(), "邮箱格式不正确")
}

// 运行注册服务测试套件
func TestRegisterServiceSuite(t *testing.T) {
	suite.Run(t, new(RegisterServiceTestSuite))
}

// TestNewRegisterService 测试创建注册服务
func TestNewRegisterService(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	service := NewRegisterService(db)
	assert.NotNil(t, service)
	assert.Equal(t, db, service.db)
	assert.NotNil(t, service.verificationService)
}

// TestRegisterService_ValidationMethods 测试验证方法
func TestRegisterService_ValidationMethods(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	service := NewRegisterService(db)

	// 测试用户名验证
	assert.True(t, service.isValidUsername("validuser123"))
	assert.True(t, service.isValidUsername("user_name"))
	assert.False(t, service.isValidUsername("ab"))           // 太短
	assert.False(t, service.isValidUsername("user@name"))    // 包含非法字符
	assert.False(t, service.isValidUsername("verylongusernamethatexceedsfiftycharacterslimit")) // 太长

	// 测试邮箱验证
	assert.True(t, service.isValidEmail("test@example.com"))
	assert.True(t, service.isValidEmail("user.name+tag@domain.co.uk"))
	assert.False(t, service.isValidEmail("invalid-email"))
	assert.False(t, service.isValidEmail("@example.com"))
	assert.False(t, service.isValidEmail("test@"))

	// 测试密码验证
	assert.True(t, service.isValidPassword("password123"))
	assert.True(t, service.isValidPassword("123456"))
	assert.False(t, service.isValidPassword("12345"))  // 太短
	assert.False(t, service.isValidPassword(""))       // 空密码
	assert.False(t, service.isValidPassword("   "))    // 只有空格

	// 测试手机号验证
	assert.True(t, service.isValidPhone("13812345678"))
	assert.True(t, service.isValidPhone("18888888888"))
	assert.False(t, service.isValidPhone("12812345678"))  // 不是1开头的有效号段
	assert.False(t, service.isValidPhone("1381234567"))   // 太短
	assert.False(t, service.isValidPhone("138123456789")) // 太长
	assert.False(t, service.isValidPhone("abcdefghijk"))  // 非数字
}
