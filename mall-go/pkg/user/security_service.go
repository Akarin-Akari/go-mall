package user

import (
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"time"

	"mall-go/internal/model"

	"gorm.io/gorm"
)

// SecurityService 用户安全服务
type SecurityService struct {
	db *gorm.DB
}

// NewSecurityService 创建用户安全服务
func NewSecurityService(db *gorm.DB) *SecurityService {
	return &SecurityService{
		db: db,
	}
}

// LockUserRequest 锁定用户请求
type LockUserRequest struct {
	UserID   uint   `json:"user_id" binding:"required"`
	Duration int    `json:"duration"`                   // 锁定时长（分钟），0表示永久锁定
	Reason   string `json:"reason" binding:"required"`  // 锁定原因
}

// UnlockUserRequest 解锁用户请求
type UnlockUserRequest struct {
	UserID uint   `json:"user_id" binding:"required"`
	Reason string `json:"reason"` // 解锁原因
}

// BanUserRequest 封禁用户请求
type BanUserRequest struct {
	UserID uint   `json:"user_id" binding:"required"`
	Reason string `json:"reason" binding:"required"` // 封禁原因
}

// Enable2FARequest 启用双因子认证请求
type Enable2FARequest struct {
	Code string `json:"code" binding:"required"` // TOTP验证码
}

// Disable2FARequest 禁用双因子认证请求
type Disable2FARequest struct {
	Password string `json:"password" binding:"required"` // 用户密码
	Code     string `json:"code" binding:"required"`     // TOTP验证码
}

// SecurityLog 安全日志模型
type SecurityLog struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	UserID    uint      `gorm:"not null;index" json:"user_id"`
	AdminID   uint      `gorm:"index" json:"admin_id"`              // 操作管理员ID（可为空）
	Action    string    `gorm:"size:50;not null" json:"action"`     // 操作类型
	Target    string    `gorm:"size:100" json:"target"`             // 操作目标
	Reason    string    `gorm:"size:500" json:"reason"`             // 操作原因
	IPAddress string    `gorm:"size:45" json:"ip_address"`          // IP地址
	UserAgent string    `gorm:"size:500" json:"user_agent"`         // 用户代理
	Status    string    `gorm:"size:20" json:"status"`              // 状态：success/failed
	Details   string    `gorm:"type:text" json:"details"`           // 详细信息
	CreatedAt time.Time `json:"created_at"`
	
	// 关联关系
	User  *model.User `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Admin *model.User `gorm:"foreignKey:AdminID" json:"admin,omitempty"`
}

// TableName 指定表名
func (SecurityLog) TableName() string {
	return "security_logs"
}

// LockUser 锁定用户
func (ss *SecurityService) LockUser(adminID uint, req *LockUserRequest) error {
	// 查询用户
	var user model.User
	if err := ss.db.First(&user, req.UserID).Error; err != nil {
		return fmt.Errorf("用户不存在")
	}

	// 检查用户是否已被锁定
	if user.IsLocked() {
		return fmt.Errorf("用户已被锁定")
	}

	// 设置锁定状态
	if req.Duration > 0 {
		// 临时锁定
		lockUntil := time.Now().Add(time.Duration(req.Duration) * time.Minute)
		user.LockedUntil = &lockUntil
	} else {
		// 永久锁定
		user.Status = model.StatusLocked
	}

	// 保存用户状态
	if err := ss.db.Save(&user).Error; err != nil {
		return fmt.Errorf("锁定用户失败: %v", err)
	}

	// 记录安全日志
	ss.recordSecurityLog(req.UserID, adminID, "lock_user", fmt.Sprintf("user:%d", req.UserID), 
		req.Reason, "", "", "success", fmt.Sprintf("锁定时长: %d分钟", req.Duration))

	return nil
}

// UnlockUser 解锁用户
func (ss *SecurityService) UnlockUser(adminID uint, req *UnlockUserRequest) error {
	// 查询用户
	var user model.User
	if err := ss.db.First(&user, req.UserID).Error; err != nil {
		return fmt.Errorf("用户不存在")
	}

	// 检查用户是否被锁定
	if !user.IsLocked() {
		return fmt.Errorf("用户未被锁定")
	}

	// 解锁用户
	user.Status = model.StatusActive
	user.LockedUntil = nil
	user.ResetLoginAttempts()

	// 保存用户状态
	if err := ss.db.Save(&user).Error; err != nil {
		return fmt.Errorf("解锁用户失败: %v", err)
	}

	// 记录安全日志
	ss.recordSecurityLog(req.UserID, adminID, "unlock_user", fmt.Sprintf("user:%d", req.UserID), 
		req.Reason, "", "", "success", "用户已解锁")

	return nil
}

// BanUser 封禁用户
func (ss *SecurityService) BanUser(adminID uint, req *BanUserRequest) error {
	// 查询用户
	var user model.User
	if err := ss.db.First(&user, req.UserID).Error; err != nil {
		return fmt.Errorf("用户不存在")
	}

	// 检查用户是否已被封禁
	if user.Status == model.StatusBanned {
		return fmt.Errorf("用户已被封禁")
	}

	// 封禁用户
	user.Status = model.StatusBanned

	// 保存用户状态
	if err := ss.db.Save(&user).Error; err != nil {
		return fmt.Errorf("封禁用户失败: %v", err)
	}

	// 记录安全日志
	ss.recordSecurityLog(req.UserID, adminID, "ban_user", fmt.Sprintf("user:%d", req.UserID), 
		req.Reason, "", "", "success", "用户已封禁")

	return nil
}

// UnbanUser 解封用户
func (ss *SecurityService) UnbanUser(adminID uint, userID uint, reason string) error {
	// 查询用户
	var user model.User
	if err := ss.db.First(&user, userID).Error; err != nil {
		return fmt.Errorf("用户不存在")
	}

	// 检查用户是否被封禁
	if user.Status != model.StatusBanned {
		return fmt.Errorf("用户未被封禁")
	}

	// 解封用户
	user.Status = model.StatusActive

	// 保存用户状态
	if err := ss.db.Save(&user).Error; err != nil {
		return fmt.Errorf("解封用户失败: %v", err)
	}

	// 记录安全日志
	ss.recordSecurityLog(userID, adminID, "unban_user", fmt.Sprintf("user:%d", userID), 
		reason, "", "", "success", "用户已解封")

	return nil
}

// Enable2FA 启用双因子认证
func (ss *SecurityService) Enable2FA(userID uint, req *Enable2FARequest) (string, error) {
	// 查询用户
	var user model.User
	if err := ss.db.First(&user, userID).Error; err != nil {
		return "", fmt.Errorf("用户不存在")
	}

	// 检查是否已启用
	if user.TwoFactorEnabled {
		return "", fmt.Errorf("双因子认证已启用")
	}

	// 生成密钥
	secret, err := ss.generate2FASecret()
	if err != nil {
		return "", fmt.Errorf("生成密钥失败: %v", err)
	}

	// 验证TOTP码
	if !ss.verifyTOTP(secret, req.Code) {
		return "", fmt.Errorf("验证码错误")
	}

	// 启用双因子认证
	user.TwoFactorEnabled = true
	user.TwoFactorSecret = secret

	// 保存用户状态
	if err := ss.db.Save(&user).Error; err != nil {
		return "", fmt.Errorf("启用双因子认证失败: %v", err)
	}

	// 记录安全日志
	ss.recordSecurityLog(userID, 0, "enable_2fa", fmt.Sprintf("user:%d", userID), 
		"用户启用双因子认证", "", "", "success", "双因子认证已启用")

	return secret, nil
}

// Disable2FA 禁用双因子认证
func (ss *SecurityService) Disable2FA(userID uint, req *Disable2FARequest) error {
	// 查询用户
	var user model.User
	if err := ss.db.First(&user, userID).Error; err != nil {
		return fmt.Errorf("用户不存在")
	}

	// 检查是否已启用
	if !user.TwoFactorEnabled {
		return fmt.Errorf("双因子认证未启用")
	}

	// 验证密码
	if !user.CheckPassword(req.Password) {
		return fmt.Errorf("密码错误")
	}

	// 验证TOTP码
	if !ss.verifyTOTP(user.TwoFactorSecret, req.Code) {
		return fmt.Errorf("验证码错误")
	}

	// 禁用双因子认证
	user.TwoFactorEnabled = false
	user.TwoFactorSecret = ""

	// 保存用户状态
	if err := ss.db.Save(&user).Error; err != nil {
		return fmt.Errorf("禁用双因子认证失败: %v", err)
	}

	// 记录安全日志
	ss.recordSecurityLog(userID, 0, "disable_2fa", fmt.Sprintf("user:%d", userID), 
		"用户禁用双因子认证", "", "", "success", "双因子认证已禁用")

	return nil
}

// GetSecurityLogs 获取安全日志
func (ss *SecurityService) GetSecurityLogs(userID uint, page, pageSize int) ([]SecurityLog, int64, error) {
	var logs []SecurityLog
	var total int64

	query := ss.db.Model(&SecurityLog{})
	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("查询安全日志总数失败: %v", err)
	}

	// 获取分页数据
	offset := (page - 1) * pageSize
	if err := query.Preload("User").Preload("Admin").
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&logs).Error; err != nil {
		return nil, 0, fmt.Errorf("查询安全日志失败: %v", err)
	}

	return logs, total, nil
}

// CheckAccountSecurity 检查账户安全状态
func (ss *SecurityService) CheckAccountSecurity(userID uint) (map[string]interface{}, error) {
	var user model.User
	if err := ss.db.First(&user, userID).Error; err != nil {
		return nil, fmt.Errorf("用户不存在")
	}

	// 获取最近登录日志
	var recentLogins []model.UserLoginLog
	ss.db.Where("user_id = ?", userID).
		Order("login_at DESC").
		Limit(5).
		Find(&recentLogins)

	// 检查异常登录
	suspiciousLogins := ss.detectSuspiciousLogins(recentLogins)

	// 计算安全评分
	securityScore := ss.calculateSecurityScore(&user, recentLogins)

	result := map[string]interface{}{
		"user_id":            user.ID,
		"account_status":     user.Status,
		"is_locked":          user.IsLocked(),
		"email_verified":     user.EmailVerified,
		"phone_verified":     user.PhoneVerified,
		"two_factor_enabled": user.TwoFactorEnabled,
		"password_changed_at": user.PasswordChangedAt,
		"last_login_at":      user.LastLoginAt,
		"last_login_ip":      user.LastLoginIP,
		"login_attempts":     user.LoginAttempts,
		"locked_until":       user.LockedUntil,
		"recent_logins":      recentLogins,
		"suspicious_logins":  suspiciousLogins,
		"security_score":     securityScore,
		"recommendations":    ss.getSecurityRecommendations(&user, securityScore),
	}

	return result, nil
}

// recordSecurityLog 记录安全日志
func (ss *SecurityService) recordSecurityLog(userID, adminID uint, action, target, reason, ip, userAgent, status, details string) {
	log := &SecurityLog{
		UserID:    userID,
		AdminID:   adminID,
		Action:    action,
		Target:    target,
		Reason:    reason,
		IPAddress: ip,
		UserAgent: userAgent,
		Status:    status,
		Details:   details,
	}

	ss.db.Create(log)
}

// generate2FASecret 生成双因子认证密钥
func (ss *SecurityService) generate2FASecret() (string, error) {
	secret := make([]byte, 20)
	_, err := rand.Read(secret)
	if err != nil {
		return "", err
	}
	
	return base32.StdEncoding.EncodeToString(secret), nil
}

// verifyTOTP 验证TOTP码（简化实现）
func (ss *SecurityService) verifyTOTP(secret, code string) bool {
	// TODO: 实现真正的TOTP验证
	// 这里只是简单的模拟验证
	return len(code) == 6
}

// detectSuspiciousLogins 检测可疑登录
func (ss *SecurityService) detectSuspiciousLogins(logins []model.UserLoginLog) []model.UserLoginLog {
	var suspicious []model.UserLoginLog
	
	if len(logins) < 2 {
		return suspicious
	}

	// 检查IP地址变化
	lastIP := ""
	for _, login := range logins {
		if lastIP != "" && lastIP != login.LoginIP {
			// IP地址发生变化，可能是可疑登录
			suspicious = append(suspicious, login)
		}
		lastIP = login.LoginIP
	}

	return suspicious
}

// calculateSecurityScore 计算安全评分
func (ss *SecurityService) calculateSecurityScore(user *model.User, recentLogins []model.UserLoginLog) int {
	score := 0

	// 基础分数
	score += 20

	// 邮箱验证
	if user.EmailVerified {
		score += 20
	}

	// 手机验证
	if user.PhoneVerified {
		score += 20
	}

	// 双因子认证
	if user.TwoFactorEnabled {
		score += 30
	}

	// 密码强度（简化检查）
	if user.PasswordChangedAt != nil && time.Since(*user.PasswordChangedAt) < 90*24*time.Hour {
		score += 10 // 90天内修改过密码
	}

	// 登录安全性
	if len(recentLogins) > 0 {
		// 检查最近是否有失败登录
		hasFailedLogin := false
		for _, login := range recentLogins {
			if login.Status == model.LoginStatusFailed {
				hasFailedLogin = true
				break
			}
		}
		if !hasFailedLogin {
			score += 10
		}
	}

	// 确保分数在0-100之间
	if score > 100 {
		score = 100
	}
	if score < 0 {
		score = 0
	}

	return score
}

// getSecurityRecommendations 获取安全建议
func (ss *SecurityService) getSecurityRecommendations(user *model.User, score int) []string {
	var recommendations []string

	if !user.EmailVerified {
		recommendations = append(recommendations, "请验证您的邮箱地址")
	}

	if !user.PhoneVerified {
		recommendations = append(recommendations, "请验证您的手机号码")
	}

	if !user.TwoFactorEnabled {
		recommendations = append(recommendations, "建议启用双因子认证以提高账户安全性")
	}

	if user.PasswordChangedAt == nil || time.Since(*user.PasswordChangedAt) > 90*24*time.Hour {
		recommendations = append(recommendations, "建议定期更换密码，上次修改密码已超过90天")
	}

	if score < 60 {
		recommendations = append(recommendations, "您的账户安全评分较低，建议完善安全设置")
	}

	return recommendations
}

// 全局安全服务实例
var globalSecurityService *SecurityService

// InitGlobalSecurityService 初始化全局安全服务
func InitGlobalSecurityService(db *gorm.DB) {
	globalSecurityService = NewSecurityService(db)
}

// GetGlobalSecurityService 获取全局安全服务
func GetGlobalSecurityService() *SecurityService {
	return globalSecurityService
}
