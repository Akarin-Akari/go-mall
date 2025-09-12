package cache

import (
	"encoding/json"
	"fmt"
	"mall-go/internal/model"
	"mall-go/pkg/logger"
	"time"
)

// SessionCacheService 用户会话缓存服务
type SessionCacheService struct {
	cacheManager CacheManager
	keyManager   *CacheKeyManager
}

// NewSessionCacheService 创建用户会话缓存服务
func NewSessionCacheService(cacheManager CacheManager, keyManager *CacheKeyManager) *SessionCacheService {
	return &SessionCacheService{
		cacheManager: cacheManager,
		keyManager:   keyManager,
	}
}

// UserSessionData 用户会话缓存数据结构
type UserSessionData struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Phone    string `json:"phone"`
	Role     string `json:"role"`
	Status   string `json:"status"`

	// 登录信息
	LoginTime    time.Time `json:"login_time"`
	LastActiveAt time.Time `json:"last_active_at"`
	LoginIP      string    `json:"login_ip"`
	DeviceInfo   string    `json:"device_info"`
	UserAgent    string    `json:"user_agent"`

	// 权限信息
	Permissions []string `json:"permissions"`

	// 偏好设置
	Language string `json:"language"`
	Timezone string `json:"timezone"`
	Theme    string `json:"theme"`

	// 统计信息
	LoginCount int `json:"login_count"`

	// 缓存元数据
	CachedAt  time.Time `json:"cached_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Version   int       `json:"version"` // 乐观锁版本号
}

// TokenCacheData JWT Token缓存数据结构
type TokenCacheData struct {
	Token            string    `json:"token"`
	RefreshToken     string    `json:"refresh_token"`
	UserID           uint      `json:"user_id"`
	Username         string    `json:"username"`
	Role             string    `json:"role"`
	ExpiresAt        time.Time `json:"expires_at"`
	RefreshExpiresAt time.Time `json:"refresh_expires_at"`
	IssuedAt         time.Time `json:"issued_at"`
	DeviceID         string    `json:"device_id"`
	LoginIP          string    `json:"login_ip"`

	// 缓存元数据
	CachedAt time.Time `json:"cached_at"`
	Version  int       `json:"version"`
}

// LoginSessionRequest 登录会话请求
type LoginSessionRequest struct {
	UserID     uint   `json:"user_id"`
	Token      string `json:"token"`
	DeviceID   string `json:"device_id"`
	LoginIP    string `json:"login_ip"`
	UserAgent  string `json:"user_agent"`
	RememberMe bool   `json:"remember_me"`
}

// SessionStatsData 会话统计数据
type SessionStatsData struct {
	OnlineUsers    int       `json:"online_users"`
	TotalSessions  int       `json:"total_sessions"`
	ActiveSessions int       `json:"active_sessions"`
	PeakSessions   int       `json:"peak_sessions"`
	LastUpdated    time.Time `json:"last_updated"`
}

// ConvertToUserSessionData 转换用户模型为会话缓存数据
func ConvertToUserSessionData(user *model.User, loginInfo *LoginSessionRequest) *UserSessionData {
	now := time.Now()

	// 获取用户偏好设置
	language := "zh-CN"
	timezone := "Asia/Shanghai"
	theme := "light"

	if user.Profile != nil {
		language = user.Profile.Language
		timezone = user.Profile.Timezone
		theme = user.Profile.Theme
	}

	return &UserSessionData{
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
		Nickname: user.Nickname,
		Avatar:   user.Avatar,
		Phone:    user.Phone,
		Role:     user.Role,
		Status:   user.Status,

		LoginTime:    now,
		LastActiveAt: now,
		LoginIP:      loginInfo.LoginIP,
		DeviceInfo:   loginInfo.DeviceID,
		UserAgent:    loginInfo.UserAgent,

		Permissions: []string{}, // 权限列表，可根据角色动态生成

		Language: language,
		Timezone: timezone,
		Theme:    theme,

		LoginCount: user.LoginCount,

		CachedAt:  now,
		UpdatedAt: user.UpdatedAt,
		Version:   1, // 默认版本号
	}
}

// ConvertToTokenCacheData 转换为Token缓存数据
func ConvertToTokenCacheData(token, refreshToken string, user *model.User, expiresAt, refreshExpiresAt time.Time, loginInfo *LoginSessionRequest) *TokenCacheData {
	now := time.Now()

	return &TokenCacheData{
		Token:            token,
		RefreshToken:     refreshToken,
		UserID:           user.ID,
		Username:         user.Username,
		Role:             user.Role,
		ExpiresAt:        expiresAt,
		RefreshExpiresAt: refreshExpiresAt,
		IssuedAt:         now,
		DeviceID:         loginInfo.DeviceID,
		LoginIP:          loginInfo.LoginIP,

		CachedAt: now,
		Version:  1, // 默认版本号
	}
}

// GetUserSession 获取用户会话缓存
func (scs *SessionCacheService) GetUserSession(userID uint) (*UserSessionData, error) {
	key := scs.keyManager.GenerateUserSessionKey(userID)

	result, err := scs.cacheManager.Get(key)
	if err != nil {
		logger.Error(fmt.Sprintf("获取用户会话缓存失败: UserID=%d, Error=%v", userID, err))
		return nil, fmt.Errorf("获取用户会话缓存失败: %w", err)
	}

	if result == nil {
		return nil, nil // 缓存未命中
	}

	// 反序列化
	var sessionData UserSessionData
	if err := json.Unmarshal([]byte(result.(string)), &sessionData); err != nil {
		logger.Error(fmt.Sprintf("用户会话数据反序列化失败: UserID=%d, Error=%v", userID, err))
		return nil, fmt.Errorf("用户会话数据反序列化失败: %w", err)
	}

	logger.Info(fmt.Sprintf("用户会话缓存命中: UserID=%d, Username=%s", userID, sessionData.Username))
	return &sessionData, nil
}

// SetUserSession 设置用户会话缓存
func (scs *SessionCacheService) SetUserSession(user *model.User, loginInfo *LoginSessionRequest) error {
	key := scs.keyManager.GenerateUserSessionKey(user.ID)

	// 转换为缓存数据
	sessionData := ConvertToUserSessionData(user, loginInfo)

	// 序列化
	jsonData, err := json.Marshal(sessionData)
	if err != nil {
		logger.Error(fmt.Sprintf("用户会话数据序列化失败: UserID=%d, Error=%v", user.ID, err))
		return fmt.Errorf("用户会话数据序列化失败: %w", err)
	}

	// 获取TTL - 根据是否记住我设置不同的过期时间
	var ttl time.Duration
	if loginInfo.RememberMe {
		ttl = 30 * 24 * time.Hour // 记住我30天
	} else {
		ttl = GetTTL("session") // 默认2小时
	}

	// 存储到缓存
	if err := scs.cacheManager.Set(key, string(jsonData), ttl); err != nil {
		logger.Error(fmt.Sprintf("设置用户会话缓存失败: UserID=%d, Error=%v", user.ID, err))
		return fmt.Errorf("设置用户会话缓存失败: %w", err)
	}

	logger.Info(fmt.Sprintf("用户会话缓存设置成功: UserID=%d, Username=%s, TTL=%v",
		user.ID, user.Username, ttl))
	return nil
}

// GetTokenSession 获取Token会话缓存
func (scs *SessionCacheService) GetTokenSession(token string) (*TokenCacheData, error) {
	key := scs.keyManager.GenerateTokenKey(token)

	result, err := scs.cacheManager.Get(key)
	if err != nil {
		logger.Error(fmt.Sprintf("获取Token会话缓存失败: Error=%v", err))
		return nil, fmt.Errorf("获取Token会话缓存失败: %w", err)
	}

	if result == nil {
		return nil, nil // 缓存未命中
	}

	// 反序列化
	var tokenData TokenCacheData
	if err := json.Unmarshal([]byte(result.(string)), &tokenData); err != nil {
		logger.Error(fmt.Sprintf("Token会话数据反序列化失败: Error=%v", err))
		return nil, fmt.Errorf("Token会话数据反序列化失败: %w", err)
	}

	// 检查Token是否过期
	if time.Now().After(tokenData.ExpiresAt) {
		// Token已过期，删除缓存
		scs.DeleteTokenSession(token)
		return nil, nil
	}

	logger.Info(fmt.Sprintf("Token会话缓存命中: UserID=%d, Username=%s",
		tokenData.UserID, tokenData.Username))
	return &tokenData, nil
}

// SetTokenSession 设置Token会话缓存
func (scs *SessionCacheService) SetTokenSession(token, refreshToken string, user *model.User,
	expiresAt, refreshExpiresAt time.Time, loginInfo *LoginSessionRequest) error {

	key := scs.keyManager.GenerateTokenKey(token)

	// 转换为缓存数据
	tokenData := ConvertToTokenCacheData(token, refreshToken, user, expiresAt, refreshExpiresAt, loginInfo)

	// 序列化
	jsonData, err := json.Marshal(tokenData)
	if err != nil {
		logger.Error(fmt.Sprintf("Token会话数据序列化失败: UserID=%d, Error=%v", user.ID, err))
		return fmt.Errorf("Token会话数据序列化失败: %w", err)
	}

	// 获取TTL - 使用Token的过期时间
	ttl := GetTTL("token")
	if time.Until(expiresAt) < ttl {
		ttl = time.Until(expiresAt)
	}

	// 存储到缓存
	if err := scs.cacheManager.Set(key, string(jsonData), ttl); err != nil {
		logger.Error(fmt.Sprintf("设置Token会话缓存失败: UserID=%d, Error=%v", user.ID, err))
		return fmt.Errorf("设置Token会话缓存失败: %w", err)
	}

	logger.Info(fmt.Sprintf("Token会话缓存设置成功: UserID=%d, Username=%s, TTL=%v",
		user.ID, user.Username, ttl))
	return nil
}

// UpdateLastActive 更新用户最后活跃时间
func (scs *SessionCacheService) UpdateLastActive(userID uint) error {
	// 获取当前会话数据
	sessionData, err := scs.GetUserSession(userID)
	if err != nil {
		return err
	}

	if sessionData == nil {
		return fmt.Errorf("用户会话不存在: UserID=%d", userID)
	}

	// 更新最后活跃时间
	sessionData.LastActiveAt = time.Now()
	sessionData.Version++

	// 重新设置缓存
	key := scs.keyManager.GenerateUserSessionKey(userID)
	jsonData, err := json.Marshal(sessionData)
	if err != nil {
		return fmt.Errorf("序列化会话数据失败: %w", err)
	}

	// 保持原有TTL
	ttl, err := scs.cacheManager.TTL(key)
	if err != nil || ttl <= 0 {
		ttl = GetTTL("session")
	}

	if err := scs.cacheManager.Set(key, string(jsonData), ttl); err != nil {
		return fmt.Errorf("更新用户活跃时间失败: %w", err)
	}

	logger.Info(fmt.Sprintf("用户活跃时间更新成功: UserID=%d", userID))
	return nil
}

// DeleteUserSession 删除用户会话缓存
func (scs *SessionCacheService) DeleteUserSession(userID uint) error {
	key := scs.keyManager.GenerateUserSessionKey(userID)

	if err := scs.cacheManager.Delete(key); err != nil {
		logger.Error(fmt.Sprintf("删除用户会话缓存失败: UserID=%d, Error=%v", userID, err))
		return fmt.Errorf("删除用户会话缓存失败: %w", err)
	}

	logger.Info(fmt.Sprintf("用户会话缓存删除成功: UserID=%d", userID))
	return nil
}

// DeleteTokenSession 删除Token会话缓存
func (scs *SessionCacheService) DeleteTokenSession(token string) error {
	key := scs.keyManager.GenerateTokenKey(token)

	if err := scs.cacheManager.Delete(key); err != nil {
		logger.Error(fmt.Sprintf("删除Token会话缓存失败: Error=%v", err))
		return fmt.Errorf("删除Token会话缓存失败: %w", err)
	}

	logger.Info("Token会话缓存删除成功")
	return nil
}

// ExistsUserSession 检查用户会话缓存是否存在
func (scs *SessionCacheService) ExistsUserSession(userID uint) bool {
	key := scs.keyManager.GenerateUserSessionKey(userID)
	return scs.cacheManager.Exists(key)
}

// ExistsTokenSession 检查Token会话缓存是否存在
func (scs *SessionCacheService) ExistsTokenSession(token string) bool {
	key := scs.keyManager.GenerateTokenKey(token)
	return scs.cacheManager.Exists(key)
}

// GetUserSessionTTL 获取用户会话缓存TTL
func (scs *SessionCacheService) GetUserSessionTTL(userID uint) (time.Duration, error) {
	key := scs.keyManager.GenerateUserSessionKey(userID)
	return scs.cacheManager.TTL(key)
}

// RefreshUserSessionTTL 刷新用户会话缓存TTL
func (scs *SessionCacheService) RefreshUserSessionTTL(userID uint) error {
	key := scs.keyManager.GenerateUserSessionKey(userID)
	ttl := GetTTL("session")

	if err := scs.cacheManager.Expire(key, ttl); err != nil {
		logger.Error(fmt.Sprintf("刷新用户会话TTL失败: UserID=%d, Error=%v", userID, err))
		return fmt.Errorf("刷新用户会话TTL失败: %w", err)
	}

	logger.Info(fmt.Sprintf("用户会话TTL刷新成功: UserID=%d, TTL=%v", userID, ttl))
	return nil
}

// GetOnlineUsers 获取在线用户列表
func (scs *SessionCacheService) GetOnlineUsers() ([]uint, error) {
	// 这里需要实现键模式搜索，暂时返回空列表
	// 在实际实现中，可以使用Redis的SCAN命令
	// pattern := scs.keyManager.prefix + ":user:session:*"
	return []uint{}, nil
}

// GetSessionStats 获取会话统计信息
func (scs *SessionCacheService) GetSessionStats() (*SessionStatsData, error) {
	// 获取在线用户数量
	onlineUsers, err := scs.GetOnlineUsers()
	if err != nil {
		return nil, fmt.Errorf("获取在线用户失败: %w", err)
	}

	stats := &SessionStatsData{
		OnlineUsers:    len(onlineUsers),
		TotalSessions:  len(onlineUsers), // 简化实现
		ActiveSessions: len(onlineUsers),
		PeakSessions:   len(onlineUsers),
		LastUpdated:    time.Now(),
	}

	return stats, nil
}

// BatchDeleteUserSessions 批量删除用户会话缓存
func (scs *SessionCacheService) BatchDeleteUserSessions(userIDs []uint) error {
	if len(userIDs) == 0 {
		return nil
	}

	// 生成批量键
	keys := make([]string, len(userIDs))
	for i, userID := range userIDs {
		keys[i] = scs.keyManager.GenerateUserSessionKey(userID)
	}

	// 批量删除
	if err := scs.cacheManager.MDelete(keys); err != nil {
		logger.Error(fmt.Sprintf("批量删除用户会话缓存失败: %v", err))
		return fmt.Errorf("批量删除用户会话缓存失败: %w", err)
	}

	logger.Info(fmt.Sprintf("批量删除用户会话缓存成功: 数量=%d", len(keys)))
	return nil
}

// ValidateSession 验证会话有效性
func (scs *SessionCacheService) ValidateSession(userID uint) (bool, *UserSessionData, error) {
	sessionData, err := scs.GetUserSession(userID)
	if err != nil {
		return false, nil, err
	}

	if sessionData == nil {
		return false, nil, nil // 会话不存在
	}

	// 检查会话是否过期（通过最后活跃时间）
	sessionTimeout := 30 * time.Minute // 30分钟无活动则认为过期
	if time.Since(sessionData.LastActiveAt) > sessionTimeout {
		// 会话已过期，删除缓存
		scs.DeleteUserSession(userID)
		return false, nil, nil
	}

	// 检查用户状态
	if sessionData.Status != "active" {
		return false, sessionData, nil
	}

	return true, sessionData, nil
}

// RefreshToken 刷新Token
func (scs *SessionCacheService) RefreshToken(refreshToken string) (*TokenCacheData, error) {
	// 获取刷新Token对应的数据
	key := scs.keyManager.GenerateRefreshTokenKey(refreshToken)

	result, err := scs.cacheManager.Get(key)
	if err != nil {
		return nil, fmt.Errorf("获取刷新Token失败: %w", err)
	}

	if result == nil {
		return nil, fmt.Errorf("刷新Token不存在或已过期")
	}

	// 反序列化
	var tokenData TokenCacheData
	if err := json.Unmarshal([]byte(result.(string)), &tokenData); err != nil {
		return nil, fmt.Errorf("刷新Token数据反序列化失败: %w", err)
	}

	// 检查刷新Token是否过期
	if time.Now().After(tokenData.RefreshExpiresAt) {
		// 刷新Token已过期，删除缓存
		scs.cacheManager.Delete(key)
		return nil, fmt.Errorf("刷新Token已过期")
	}

	return &tokenData, nil
}

// SetRefreshToken 设置刷新Token缓存
func (scs *SessionCacheService) SetRefreshToken(refreshToken string, tokenData *TokenCacheData) error {
	key := scs.keyManager.GenerateRefreshTokenKey(refreshToken)

	// 序列化
	jsonData, err := json.Marshal(tokenData)
	if err != nil {
		return fmt.Errorf("刷新Token数据序列化失败: %w", err)
	}

	// 获取TTL - 使用刷新Token的过期时间
	ttl := time.Until(tokenData.RefreshExpiresAt)
	if ttl <= 0 {
		ttl = 7 * 24 * time.Hour // 默认7天
	}

	// 存储到缓存
	if err := scs.cacheManager.Set(key, string(jsonData), ttl); err != nil {
		return fmt.Errorf("设置刷新Token缓存失败: %w", err)
	}

	logger.Info(fmt.Sprintf("刷新Token缓存设置成功: UserID=%d, TTL=%v", tokenData.UserID, ttl))
	return nil
}

// GetSessionCacheStats 获取会话缓存统计信息
func (scs *SessionCacheService) GetSessionCacheStats() map[string]interface{} {
	stats := make(map[string]interface{})

	// 获取缓存管理器统计
	metrics := scs.cacheManager.GetMetrics()
	if metrics != nil {
		stats["total_ops"] = metrics.TotalOps
		stats["hit_count"] = metrics.HitCount
		stats["miss_count"] = metrics.MissCount
		stats["hit_rate"] = metrics.HitRate
		stats["error_count"] = metrics.ErrorCount
		stats["last_updated"] = metrics.LastUpdated
	}

	// 获取连接池统计
	connStats := scs.cacheManager.GetConnectionStats()
	if connStats != nil {
		stats["total_conns"] = connStats.TotalConns
		stats["idle_conns"] = connStats.IdleConns
		stats["stale_conns"] = connStats.StaleConns
		stats["hits"] = connStats.Hits
		stats["misses"] = connStats.Misses
		stats["timeouts"] = connStats.Timeouts
	}

	return stats
}
