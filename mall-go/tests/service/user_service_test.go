package service

import (
	"testing"

	"mall-go/internal/model"
	"mall-go/pkg/auth"
	"mall-go/pkg/user"
)

// UserServiceTestSuite 用户服务测试套件
type UserServiceTestSuite struct {
	ServiceTestSuite
	registerService *user.RegisterService
	loginService    *user.LoginService
	profileService  *user.ProfileService
}

// SetupSuite 测试套件初始化
func (suite *UserServiceTestSuite) SetupSuite() {
	// 调用基础套件初始化
	suite.ServiceTestSuite.SetupSuite()

	// 初始化用户服务
	suite.registerService = user.NewRegisterService(suite.GetDB())
	suite.loginService = user.NewLoginService(suite.GetDB())
	suite.profileService = user.NewProfileService(suite.GetDB())
}

// TestUserRegisterService 测试用户注册服务
func (suite *UserServiceTestSuite) TestUserRegisterService() {
	testCases := []ServiceTestCase{
		{
			Name:        "成功注册用户",
			Description: "验证用户注册的完整业务流程",
			SetupData: func(s *ServiceTestSuite) {
				// 无需特殊设置
			},
			ExecuteAction: func(s *ServiceTestSuite) (interface{}, error) {
				req := &user.RegisterRequest{
					Username:        "testuser",
					Email:           "test@example.com",
					Password:        "password123",
					ConfirmPassword: "password123",
					Phone:           "13800138000",
					Nickname:        "测试用户",
					AgreeTerms:      true,
				}
				return suite.registerService.Register(req)
			},
			ValidateResult: func(s *ServiceTestSuite, result interface{}, err error) {
				// 验证注册成功
				s.AssertNoError(err, "用户注册应该成功")

				response, ok := result.(*user.RegisterResponse)
				s.True(ok, "返回结果应该是RegisterResponse类型")
				s.NotZero(response.UserID, "用户ID应该不为0")
				s.Equal("testuser", response.Username, "用户名应该匹配")
				s.Equal("test@example.com", response.Email, "邮箱应该匹配")
				s.NotEmpty(response.Token, "应该返回登录Token")

				// 验证数据库中的用户数据
				var dbUser model.User
				err = s.GetDB().Where("username = ?", "testuser").First(&dbUser).Error
				s.AssertNoError(err, "应该能在数据库中找到用户")
				s.Equal("testuser", dbUser.Username, "数据库中用户名应该匹配")
				s.Equal("test@example.com", dbUser.Email, "数据库中邮箱应该匹配")

				// 验证密码已加密
				s.NotEqual("password123", dbUser.Password, "密码应该已加密")

				// 验证密码加密正确性
				isValid := auth.CheckPassword("password123", dbUser.Password)
				s.True(isValid, "密码加密应该正确")
			},
		},
		{
			Name:        "用户名重复注册失败",
			Description: "验证用户名唯一性约束",
			SetupData: func(s *ServiceTestSuite) {
				// 先创建一个用户
				s.CreateTestUser("existuser", "exist@example.com", "password123")
			},
			ExecuteAction: func(s *ServiceTestSuite) (interface{}, error) {
				req := &user.RegisterRequest{
					Username:        "existuser", // 使用已存在的用户名
					Email:           "new@example.com",
					Password:        "password123",
					ConfirmPassword: "password123",
					AgreeTerms:      true,
				}
				return suite.registerService.Register(req)
			},
			ValidateResult: func(s *ServiceTestSuite, result interface{}, err error) {
				s.AssertErrorContains(err, "用户名已存在")
				s.Nil(result, "注册失败时不应返回结果")
			},
		},
		{
			Name:        "邮箱重复注册失败",
			Description: "验证邮箱唯一性约束",
			SetupData: func(s *ServiceTestSuite) {
				// 先创建一个用户
				s.CreateTestUser("existuser", "exist@example.com", "password123")
			},
			ExecuteAction: func(s *ServiceTestSuite) (interface{}, error) {
				req := &user.RegisterRequest{
					Username:        "newuser",
					Email:           "exist@example.com", // 使用已存在的邮箱
					Password:        "password123",
					ConfirmPassword: "password123",
					AgreeTerms:      true,
				}
				return suite.registerService.Register(req)
			},
			ValidateResult: func(s *ServiceTestSuite, result interface{}, err error) {
				s.AssertErrorContains(err, "邮箱已存在")
				s.Nil(result, "注册失败时不应返回结果")
			},
		},
		{
			Name:        "密码确认不匹配",
			Description: "验证密码确认逻辑",
			ExecuteAction: func(s *ServiceTestSuite) (interface{}, error) {
				req := &user.RegisterRequest{
					Username:        "testuser",
					Email:           "test@example.com",
					Password:        "password123",
					ConfirmPassword: "password456", // 密码不匹配
					AgreeTerms:      true,
				}
				return suite.registerService.Register(req)
			},
			ValidateResult: func(s *ServiceTestSuite, result interface{}, err error) {
				s.AssertErrorContains(err, "密码确认不匹配")
				s.Nil(result, "注册失败时不应返回结果")
			},
		},
		{
			Name:        "未同意服务条款",
			Description: "验证服务条款同意检查",
			ExecuteAction: func(s *ServiceTestSuite) (interface{}, error) {
				req := &user.RegisterRequest{
					Username:        "testuser",
					Email:           "test@example.com",
					Password:        "password123",
					ConfirmPassword: "password123",
					AgreeTerms:      false, // 未同意条款
				}
				return suite.registerService.Register(req)
			},
			ValidateResult: func(s *ServiceTestSuite, result interface{}, err error) {
				s.AssertErrorContains(err, "必须同意服务条款")
				s.Nil(result, "注册失败时不应返回结果")
			},
		},
	}

	suite.RunServiceTestCases(testCases)
}

// TestUserLoginService 测试用户登录服务
func (suite *UserServiceTestSuite) TestUserLoginService() {
	testCases := []ServiceTestCase{
		{
			Name:        "成功登录",
			Description: "验证用户登录的完整业务流程",
			SetupData: func(s *ServiceTestSuite) {
				// 创建测试用户
				s.CreateTestUser("loginuser", "login@example.com", "password123")
			},
			ExecuteAction: func(s *ServiceTestSuite) (interface{}, error) {
				req := &user.LoginRequest{
					Account:  "loginuser",
					Password: "password123",
				}
				return suite.loginService.Login(req)
			},
			ValidateResult: func(s *ServiceTestSuite, result interface{}, err error) {
				s.AssertNoError(err, "用户登录应该成功")

				response, ok := result.(*user.LoginResponse)
				s.True(ok, "返回结果应该是LoginResponse类型")
				s.NotEmpty(response.Token, "应该返回Token")
				s.Equal("loginuser", response.Username, "用户名应该匹配")
				s.Equal("login@example.com", response.Email, "邮箱应该匹配")

				// 验证Token有效性
				claims, err := auth.ParseToken(response.Token)
				s.AssertNoError(err, "Token应该有效")
				s.Equal("loginuser", claims.Username, "Token中用户名应该匹配")
			},
		},
		{
			Name:        "用户名不存在",
			Description: "验证不存在用户的登录处理",
			ExecuteAction: func(s *ServiceTestSuite) (interface{}, error) {
				req := &user.LoginRequest{
					Account:  "nonexistuser",
					Password: "password123",
				}
				return suite.loginService.Login(req)
			},
			ValidateResult: func(s *ServiceTestSuite, result interface{}, err error) {
				s.AssertErrorContains(err, "用户名或密码错误")
				s.Nil(result, "登录失败时不应返回结果")
			},
		},
		{
			Name:        "密码错误",
			Description: "验证错误密码的登录处理",
			SetupData: func(s *ServiceTestSuite) {
				s.CreateTestUser("loginuser", "login@example.com", "password123")
			},
			ExecuteAction: func(s *ServiceTestSuite) (interface{}, error) {
				req := &user.LoginRequest{
					Account:  "loginuser",
					Password: "wrongpassword",
				}
				return suite.loginService.Login(req)
			},
			ValidateResult: func(s *ServiceTestSuite, result interface{}, err error) {
				s.AssertErrorContains(err, "用户名或密码错误")
				s.Nil(result, "登录失败时不应返回结果")
			},
		},
	}

	suite.RunServiceTestCases(testCases)
}

// TestUserServiceSuite 运行用户服务测试套件
func TestUserServiceSuite(t *testing.T) {
	RunServiceTest(t, new(UserServiceTestSuite))
}
