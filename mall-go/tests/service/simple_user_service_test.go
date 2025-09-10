package service

import (
	"testing"

	"mall-go/internal/config"
	"mall-go/internal/model"
	"mall-go/pkg/auth"
	"mall-go/pkg/user"
	testConfig "mall-go/tests/config"

	"github.com/stretchr/testify/assert"
)

// TestSimpleUserRegisterService 简单的用户注册服务测试
func TestSimpleUserRegisterService(t *testing.T) {
	// 初始化配置
	config.GlobalConfig = config.Config{
		JWT: config.JWTConfig{
			Secret: "test-secret-key-for-service-layer-testing",
			Expire: "24h",
		},
	}

	// 初始化测试数据库
	db := testConfig.SetupTestDB()
	defer testConfig.CleanupTestDB(db)

	// 自动迁移
	err := db.AutoMigrate(&model.User{}, &model.Product{}, &model.Category{})
	assert.NoError(t, err, "数据库迁移失败")

	// 创建注册服务
	registerService := user.NewRegisterService(db)

	t.Run("成功注册用户", func(t *testing.T) {
		req := &user.RegisterRequest{
			Username:        "testuser",
			Email:           "test@example.com",
			Password:        "password123",
			ConfirmPassword: "password123",
			Phone:           "13800138000",
			Nickname:        "测试用户",
			AgreeTerms:      true,
		}

		response, err := registerService.Register(req)

		// 验证注册结果
		if err != nil {
			t.Logf("注册错误: %v", err)
			t.FailNow()
		}

		assert.NotNil(t, response, "注册响应不应为空")
		assert.NotZero(t, response.UserID, "用户ID应该不为0")
		assert.Equal(t, "testuser", response.Username, "用户名应该匹配")
		assert.Equal(t, "test@example.com", response.Email, "邮箱应该匹配")
		assert.NotEmpty(t, response.Token, "应该返回登录Token")

		// 验证数据库中的用户数据
		var dbUser model.User
		err = db.Where("username = ?", "testuser").First(&dbUser).Error
		assert.NoError(t, err, "应该能在数据库中找到用户")
		assert.Equal(t, "testuser", dbUser.Username, "数据库中用户名应该匹配")
		assert.Equal(t, "test@example.com", dbUser.Email, "数据库中邮箱应该匹配")

		// 验证密码已加密
		assert.NotEqual(t, "password123", dbUser.Password, "密码应该已加密")

		// 验证密码加密正确性
		isValid := auth.CheckPassword(dbUser.Password, "password123")
		assert.True(t, isValid, "密码加密应该正确")

		t.Logf("✅ 用户注册测试通过 - 用户ID: %d", response.UserID)
	})

	t.Run("用户名重复注册失败", func(t *testing.T) {
		// 先创建一个用户
		firstReq := &user.RegisterRequest{
			Username:        "existuser",
			Email:           "exist@example.com",
			Password:        "password123",
			ConfirmPassword: "password123",
			AgreeTerms:      true,
		}
		_, err := registerService.Register(firstReq)
		assert.NoError(t, err, "第一次注册应该成功")

		// 尝试用相同用户名注册
		secondReq := &user.RegisterRequest{
			Username:        "existuser", // 使用已存在的用户名
			Email:           "new@example.com",
			Password:        "password123",
			ConfirmPassword: "password123",
			AgreeTerms:      true,
		}

		response, err := registerService.Register(secondReq)
		assert.Error(t, err, "重复用户名注册应该失败")
		assert.Contains(t, err.Error(), "用户名", "错误信息应该包含用户名相关信息")
		assert.Nil(t, response, "注册失败时不应返回结果")

		t.Logf("✅ 用户名重复注册测试通过")
	})
}

// TestSimpleUserLoginService 简单的用户登录服务测试
func TestSimpleUserLoginService(t *testing.T) {
	// 初始化配置
	config.GlobalConfig = config.Config{
		JWT: config.JWTConfig{
			Secret: "test-secret-key-for-service-layer-testing",
			Expire: "24h",
		},
	}

	// 初始化测试数据库
	db := testConfig.SetupTestDB()
	defer testConfig.CleanupTestDB(db)

	// 自动迁移
	err := db.AutoMigrate(&model.User{})
	assert.NoError(t, err, "数据库迁移失败")

	// 创建服务
	registerService := user.NewRegisterService(db)
	loginService := user.NewLoginService(db)

	t.Run("成功登录", func(t *testing.T) {
		// 先注册一个用户
		registerReq := &user.RegisterRequest{
			Username:        "loginuser",
			Email:           "login@example.com",
			Password:        "password123",
			ConfirmPassword: "password123",
			Phone:           "13800138001", // 使用不同的手机号
			AgreeTerms:      true,
		}
		registerResp, err := registerService.Register(registerReq)
		assert.NoError(t, err, "用户注册应该成功")
		assert.NotNil(t, registerResp, "注册响应不应为空")

		t.Logf("注册成功，用户ID: %d", registerResp.UserID)

		// 尝试登录
		loginReq := &user.LoginRequest{
			Account:  "loginuser",
			Password: "password123",
		}

		loginResp, err := loginService.Login(loginReq)

		if err != nil {
			t.Logf("登录错误: %v", err)

			// 检查数据库中的用户数据
			var dbUser model.User
			dbErr := db.Where("username = ?", "loginuser").First(&dbUser).Error
			if dbErr != nil {
				t.Logf("数据库查询错误: %v", dbErr)
			} else {
				t.Logf("数据库中的用户: ID=%d, Username=%s, Password=%s",
					dbUser.ID, dbUser.Username, dbUser.Password)

				// 测试密码验证
				isValid := auth.CheckPassword("password123", dbUser.Password)
				t.Logf("密码验证结果: %v", isValid)
			}

			t.FailNow()
		}

		assert.NotNil(t, loginResp, "登录响应不应为空")
		assert.NotEmpty(t, loginResp.Token, "应该返回Token")
		assert.Equal(t, "loginuser", loginResp.Username, "用户名应该匹配")
		assert.Equal(t, "login@example.com", loginResp.Email, "邮箱应该匹配")

		// 验证Token有效性
		claims, err := auth.ParseToken(loginResp.Token)
		assert.NoError(t, err, "Token应该有效")
		assert.Equal(t, "loginuser", claims.Username, "Token中用户名应该匹配")

		t.Logf("✅ 用户登录测试通过 - Token: %s", loginResp.Token[:20]+"...")
	})

	t.Run("用户名不存在", func(t *testing.T) {
		loginReq := &user.LoginRequest{
			Account:  "nonexistuser",
			Password: "password123",
		}

		loginResp, err := loginService.Login(loginReq)
		assert.Error(t, err, "不存在用户登录应该失败")
		assert.Contains(t, err.Error(), "用户名或密码", "错误信息应该包含用户名或密码相关信息")
		assert.Nil(t, loginResp, "登录失败时不应返回结果")

		t.Logf("✅ 不存在用户登录测试通过")
	})

	t.Run("密码错误", func(t *testing.T) {
		// 先注册一个用户
		registerReq := &user.RegisterRequest{
			Username:        "pwdtestuser",
			Email:           "pwdtest@example.com",
			Password:        "password123",
			ConfirmPassword: "password123",
			Phone:           "13800138002", // 使用不同的手机号
			AgreeTerms:      true,
		}
		_, err := registerService.Register(registerReq)
		assert.NoError(t, err, "用户注册应该成功")

		// 尝试用错误密码登录
		loginReq := &user.LoginRequest{
			Account:  "pwdtestuser",
			Password: "wrongpassword",
		}

		loginResp, err := loginService.Login(loginReq)
		assert.Error(t, err, "错误密码登录应该失败")
		assert.Contains(t, err.Error(), "用户名或密码", "错误信息应该包含用户名或密码相关信息")
		assert.Nil(t, loginResp, "登录失败时不应返回结果")

		t.Logf("✅ 错误密码登录测试通过")
	})
}
