package cache

import (
	"fmt"
	"mall-go/internal/model"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// 创建测试用的用户数据
func createTestUser(id uint, username string) *model.User {
	return &model.User{
		ID:       id,
		Username: username,
		Email:    fmt.Sprintf("%s@example.com", username),
		Nickname: fmt.Sprintf("昵称_%s", username),
		Avatar:   fmt.Sprintf("https://example.com/avatar/%s.jpg", username),
		Phone:    fmt.Sprintf("1380000%04d", id),
		Role:     "user",
		Status:   "active",

		RealName: fmt.Sprintf("真实姓名_%s", username),
		Gender:   "male",
		Birthday: nil,
		Bio:      fmt.Sprintf("个人简介_%s", username),
		Website:  fmt.Sprintf("https://%s.example.com", username),
		Location: fmt.Sprintf("位置_%s", username),

		LastLoginAt:       nil,
		LastLoginIP:       "127.0.0.1",
		LoginAttempts:     0,
		LockedUntil:       nil,
		PasswordChangedAt: nil,
		TwoFactorEnabled:  false,
		TwoFactorSecret:   "",

		Balance: decimal.NewFromFloat(100.00),

		LoginCount:     1,
		PostCount:      0,
		FollowerCount:  0,
		FollowingCount: 0,

		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),

		Profile: &model.UserProfile{
			UserID:      id,
			Language:    "zh-CN",
			Timezone:    "Asia/Shanghai",
			Theme:       "light",
			EmailNotify: true,
			SmsNotify:   false,
			PushNotify:  true,
		},
	}
}

func setupSessionCacheService() (*SessionCacheService, *MockCacheManager) {
	mockCache := new(MockCacheManager)
	keyManager := NewCacheKeyManager("test")
	service := NewSessionCacheService(mockCache, keyManager)
	return service, mockCache
}

func TestSessionCacheService_GetUserSession(t *testing.T) {
	service, mockCache := setupSessionCacheService()

	// 测试缓存命中
	t.Run("缓存命中", func(t *testing.T) {
		userID := uint(1)

		// 模拟缓存数据
		cacheData := `{"user_id":1,"username":"testuser","email":"testuser@example.com","nickname":"测试用户","role":"user","status":"active","login_time":"2025-01-10T10:00:00Z","last_active_at":"2025-01-10T10:30:00Z","login_ip":"127.0.0.1","device_info":"device123","user_agent":"Mozilla/5.0","permissions":[],"language":"zh-CN","timezone":"Asia/Shanghai","theme":"light","login_count":1,"cached_at":"2025-01-10T10:00:00Z","updated_at":"2025-01-10T09:00:00Z","version":1}`

		mockCache.On("Get", mock.AnythingOfType("string")).Return(cacheData, nil)

		result, err := service.GetUserSession(userID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, uint(1), result.UserID)
		assert.Equal(t, "testuser", result.Username)
		assert.Equal(t, "testuser@example.com", result.Email)
		assert.Equal(t, "user", result.Role)
		assert.Equal(t, "active", result.Status)

		mockCache.AssertExpectations(t)
	})

	// 测试缓存未命中
	t.Run("缓存未命中", func(t *testing.T) {
		// 创建新的mock实例避免冲突
		newMockCache := new(MockCacheManager)
		newKeyManager := NewCacheKeyManager("test")
		newService := NewSessionCacheService(newMockCache, newKeyManager)

		userID := uint(2)

		newMockCache.On("Get", mock.AnythingOfType("string")).Return(nil, nil)

		result, err := newService.GetUserSession(userID)

		assert.NoError(t, err)
		assert.Nil(t, result)

		newMockCache.AssertExpectations(t)
	})
}

func TestSessionCacheService_SetUserSession(t *testing.T) {
	service, mockCache := setupSessionCacheService()

	user := createTestUser(1, "testuser")
	loginInfo := &LoginSessionRequest{
		UserID:     1,
		Token:      "test_token_123",
		DeviceID:   "device123",
		LoginIP:    "127.0.0.1",
		UserAgent:  "Mozilla/5.0",
		RememberMe: false,
	}

	mockCache.On("Set", mock.AnythingOfType("string"), mock.AnythingOfType("string"), 2*time.Hour).Return(nil)

	err := service.SetUserSession(user, loginInfo)

	assert.NoError(t, err)
	mockCache.AssertExpectations(t)
}

func TestSessionCacheService_SetUserSession_RememberMe(t *testing.T) {
	service, mockCache := setupSessionCacheService()

	user := createTestUser(1, "testuser")
	loginInfo := &LoginSessionRequest{
		UserID:     1,
		Token:      "test_token_123",
		DeviceID:   "device123",
		LoginIP:    "127.0.0.1",
		UserAgent:  "Mozilla/5.0",
		RememberMe: true, // 记住我
	}

	mockCache.On("Set", mock.AnythingOfType("string"), mock.AnythingOfType("string"), 30*24*time.Hour).Return(nil)

	err := service.SetUserSession(user, loginInfo)

	assert.NoError(t, err)
	mockCache.AssertExpectations(t)
}

func TestSessionCacheService_GetTokenSession(t *testing.T) {
	service, mockCache := setupSessionCacheService()

	// 测试获取有效Token
	t.Run("获取有效Token", func(t *testing.T) {
		token := "test_token_123"

		// 模拟缓存数据 - 未过期的Token
		futureTime := time.Now().Add(1 * time.Hour)
		cacheData := fmt.Sprintf(`{"token":"%s","refresh_token":"refresh_123","user_id":1,"username":"testuser","role":"user","expires_at":"%s","refresh_expires_at":"%s","issued_at":"2025-01-10T10:00:00Z","device_id":"device123","login_ip":"127.0.0.1","cached_at":"2025-01-10T10:00:00Z","version":1}`,
			token, futureTime.Format(time.RFC3339), futureTime.Add(7*24*time.Hour).Format(time.RFC3339))

		mockCache.On("Get", mock.AnythingOfType("string")).Return(cacheData, nil)

		result, err := service.GetTokenSession(token)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, token, result.Token)
		assert.Equal(t, uint(1), result.UserID)
		assert.Equal(t, "testuser", result.Username)

		mockCache.AssertExpectations(t)
	})

	// 测试过期Token
	t.Run("过期Token", func(t *testing.T) {
		token := "expired_token_123"

		// 模拟缓存数据 - 已过期的Token
		pastTime := time.Now().Add(-1 * time.Hour)
		cacheData := fmt.Sprintf(`{"token":"%s","refresh_token":"refresh_123","user_id":1,"username":"testuser","role":"user","expires_at":"%s","refresh_expires_at":"%s","issued_at":"2025-01-10T09:00:00Z","device_id":"device123","login_ip":"127.0.0.1","cached_at":"2025-01-10T09:00:00Z","version":1}`,
			token, pastTime.Format(time.RFC3339), pastTime.Add(7*24*time.Hour).Format(time.RFC3339))

		// 创建新的mock实例避免冲突
		newMockCache := new(MockCacheManager)
		newKeyManager := NewCacheKeyManager("test")
		newService := NewSessionCacheService(newMockCache, newKeyManager)

		newMockCache.On("Get", mock.AnythingOfType("string")).Return(cacheData, nil)
		newMockCache.On("Delete", mock.AnythingOfType("string")).Return(nil) // 删除过期Token

		result, err := newService.GetTokenSession(token)

		assert.NoError(t, err)
		assert.Nil(t, result) // 过期Token应该返回nil

		newMockCache.AssertExpectations(t)
	})
}

func TestSessionCacheService_SetTokenSession(t *testing.T) {
	service, mockCache := setupSessionCacheService()

	user := createTestUser(1, "testuser")
	token := "test_token_123"
	refreshToken := "refresh_token_456"
	expiresAt := time.Now().Add(1 * time.Hour)
	refreshExpiresAt := time.Now().Add(7 * 24 * time.Hour)

	loginInfo := &LoginSessionRequest{
		UserID:     1,
		Token:      token,
		DeviceID:   "device123",
		LoginIP:    "127.0.0.1",
		UserAgent:  "Mozilla/5.0",
		RememberMe: false,
	}

	mockCache.On("Set", mock.AnythingOfType("string"), mock.AnythingOfType("string"), 30*time.Minute).Return(nil)

	err := service.SetTokenSession(token, refreshToken, user, expiresAt, refreshExpiresAt, loginInfo)

	assert.NoError(t, err)
	mockCache.AssertExpectations(t)
}

func TestSessionCacheService_UpdateLastActive(t *testing.T) {
	service, mockCache := setupSessionCacheService()

	userID := uint(1)

	// 模拟获取现有会话数据
	cacheData := `{"user_id":1,"username":"testuser","email":"testuser@example.com","nickname":"测试用户","role":"user","status":"active","login_time":"2025-01-10T10:00:00Z","last_active_at":"2025-01-10T10:00:00Z","login_ip":"127.0.0.1","device_info":"device123","user_agent":"Mozilla/5.0","permissions":[],"language":"zh-CN","timezone":"Asia/Shanghai","theme":"light","login_count":1,"cached_at":"2025-01-10T10:00:00Z","updated_at":"2025-01-10T09:00:00Z","version":1}`

	mockCache.On("Get", mock.AnythingOfType("string")).Return(cacheData, nil)
	mockCache.On("TTL", mock.AnythingOfType("string")).Return(1*time.Hour, nil)
	mockCache.On("Set", mock.AnythingOfType("string"), mock.AnythingOfType("string"), 1*time.Hour).Return(nil)

	err := service.UpdateLastActive(userID)

	assert.NoError(t, err)
	mockCache.AssertExpectations(t)
}

func TestSessionCacheService_DeleteOperations(t *testing.T) {
	service, mockCache := setupSessionCacheService()

	// 测试删除用户会话
	t.Run("删除用户会话", func(t *testing.T) {
		userID := uint(1)

		mockCache.On("Delete", mock.AnythingOfType("string")).Return(nil)

		err := service.DeleteUserSession(userID)

		assert.NoError(t, err)
		mockCache.AssertExpectations(t)
	})

	// 测试删除Token会话
	t.Run("删除Token会话", func(t *testing.T) {
		token := "test_token_123"

		mockCache.On("Delete", mock.AnythingOfType("string")).Return(nil)

		err := service.DeleteTokenSession(token)

		assert.NoError(t, err)
		mockCache.AssertExpectations(t)
	})

	// 测试批量删除用户会话
	t.Run("批量删除用户会话", func(t *testing.T) {
		userIDs := []uint{1, 2, 3}

		mockCache.On("MDelete", mock.AnythingOfType("[]string")).Return(nil)

		err := service.BatchDeleteUserSessions(userIDs)

		assert.NoError(t, err)
		mockCache.AssertExpectations(t)
	})
}

func TestSessionCacheService_ExistsOperations(t *testing.T) {
	service, mockCache := setupSessionCacheService()

	// 测试检查用户会话存在
	t.Run("检查用户会话存在", func(t *testing.T) {
		userID := uint(1)

		mockCache.On("Exists", mock.AnythingOfType("string")).Return(true)

		exists := service.ExistsUserSession(userID)

		assert.True(t, exists)
		mockCache.AssertExpectations(t)
	})

	// 测试检查Token会话存在
	t.Run("检查Token会话存在", func(t *testing.T) {
		token := "test_token_123"

		mockCache.On("Exists", mock.AnythingOfType("string")).Return(true)

		exists := service.ExistsTokenSession(token)

		assert.True(t, exists)
		mockCache.AssertExpectations(t)
	})
}

func TestSessionCacheService_TTLOperations(t *testing.T) {
	service, mockCache := setupSessionCacheService()

	// 测试获取TTL
	t.Run("获取TTL", func(t *testing.T) {
		userID := uint(1)
		expectedTTL := 1 * time.Hour

		mockCache.On("TTL", mock.AnythingOfType("string")).Return(expectedTTL, nil)

		ttl, err := service.GetUserSessionTTL(userID)

		assert.NoError(t, err)
		assert.Equal(t, expectedTTL, ttl)
		mockCache.AssertExpectations(t)
	})

	// 测试刷新TTL
	t.Run("刷新TTL", func(t *testing.T) {
		userID := uint(1)

		mockCache.On("Expire", mock.AnythingOfType("string"), 2*time.Hour).Return(nil)

		err := service.RefreshUserSessionTTL(userID)

		assert.NoError(t, err)
		mockCache.AssertExpectations(t)
	})
}

func TestSessionCacheService_ValidateSession(t *testing.T) {
	service, mockCache := setupSessionCacheService()

	// 测试有效会话
	t.Run("有效会话", func(t *testing.T) {
		userID := uint(1)

		// 模拟最近活跃的会话数据
		recentTime := time.Now().Add(-10 * time.Minute) // 10分钟前活跃
		cacheData := fmt.Sprintf(`{"user_id":1,"username":"testuser","email":"testuser@example.com","nickname":"测试用户","role":"user","status":"active","login_time":"2025-01-10T10:00:00Z","last_active_at":"%s","login_ip":"127.0.0.1","device_info":"device123","user_agent":"Mozilla/5.0","permissions":[],"language":"zh-CN","timezone":"Asia/Shanghai","theme":"light","login_count":1,"cached_at":"2025-01-10T10:00:00Z","updated_at":"2025-01-10T09:00:00Z","version":1}`,
			recentTime.Format(time.RFC3339))

		mockCache.On("Get", mock.AnythingOfType("string")).Return(cacheData, nil)

		valid, sessionData, err := service.ValidateSession(userID)

		assert.NoError(t, err)
		assert.True(t, valid)
		assert.NotNil(t, sessionData)
		assert.Equal(t, uint(1), sessionData.UserID)

		mockCache.AssertExpectations(t)
	})

	// 测试过期会话
	t.Run("过期会话", func(t *testing.T) {
		// 创建新的mock实例避免冲突
		newMockCache := new(MockCacheManager)
		newKeyManager := NewCacheKeyManager("test")
		newService := NewSessionCacheService(newMockCache, newKeyManager)

		userID := uint(2)

		// 模拟过期的会话数据
		oldTime := time.Now().Add(-2 * time.Hour) // 2小时前活跃
		cacheData := fmt.Sprintf(`{"user_id":2,"username":"testuser2","email":"testuser2@example.com","nickname":"测试用户2","role":"user","status":"active","login_time":"2025-01-10T08:00:00Z","last_active_at":"%s","login_ip":"127.0.0.1","device_info":"device123","user_agent":"Mozilla/5.0","permissions":[],"language":"zh-CN","timezone":"Asia/Shanghai","theme":"light","login_count":1,"cached_at":"2025-01-10T08:00:00Z","updated_at":"2025-01-10T07:00:00Z","version":1}`,
			oldTime.Format(time.RFC3339))

		newMockCache.On("Get", mock.AnythingOfType("string")).Return(cacheData, nil)
		newMockCache.On("Delete", mock.AnythingOfType("string")).Return(nil) // 删除过期会话

		valid, sessionData, err := newService.ValidateSession(userID)

		assert.NoError(t, err)
		assert.False(t, valid)
		assert.Nil(t, sessionData)

		newMockCache.AssertExpectations(t)
	})
}
