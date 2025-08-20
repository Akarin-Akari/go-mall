package model

import (
	"time"
	"mall-go/pkg/auth"

	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID            uint           `gorm:"primarykey" json:"id"`
	Username      string         `gorm:"uniqueIndex;not null;size:50" json:"username"`
	Email         string         `gorm:"uniqueIndex;not null;size:100" json:"email"`
	Password      string         `gorm:"not null;size:255" json:"-"`           // 密码哈希
	Nickname      string         `gorm:"size:50" json:"nickname"`
	Avatar        string         `gorm:"size:255" json:"avatar"`
	Phone         string         `gorm:"size:20" json:"phone"`
	Role          string         `gorm:"default:'user';size:20" json:"role"`
	Status        string         `gorm:"default:'active';size:20" json:"status"`
	LastLoginAt   *time.Time     `gorm:"null" json:"last_login_at"`            // 最后登录时间
	LoginAttempts int            `gorm:"default:0" json:"-"`                   // 登录尝试次数
	LockedUntil   *time.Time     `gorm:"null" json:"-"`                        // 账户锁定到期时间
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// UserRegisterRequest 用户注册请求
type UserRegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=128"`  // 更新密码长度要求
	Nickname string `json:"nickname" binding:"required,min=2,max=50"`
	Role     string `json:"role" binding:"omitempty,oneof=user merchant admin"`  // 可选角色
}

// UserLoginRequest 用户登录请求
type UserLoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	DeviceID string `json:"device_id" binding:"omitempty,max=100"`  // 设备标识
}

// UserResponse 用户响应
type UserResponse struct {
	ID          uint       `json:"id"`
	Username    string     `json:"username"`
	Email       string     `json:"email"`
	Nickname    string     `json:"nickname"`
	Avatar      string     `json:"avatar"`
	Phone       string     `json:"phone"`
	Role        string     `json:"role"`
	Status      string     `json:"status"`
	LastLoginAt *time.Time `json:"last_login_at"`
	CreatedAt   time.Time  `json:"created_at"`
}

// 用户角色常量
const (
	RoleUser     = "user"     // 普通用户
	RoleMerchant = "merchant" // 商家
	RoleAdmin    = "admin"    // 管理员
)

// 用户状态常量
const (
	StatusActive   = "active"   // 活跃
	StatusInactive = "inactive" // 非活跃
	StatusLocked   = "locked"   // 锁定
	StatusBanned   = "banned"   // 封禁
)

// SetPassword 设置用户密码（加密存储）
func (u *User) SetPassword(password string) error {
	hashedPassword, err := auth.HashPassword(password)
	if err != nil {
		return err
	}
	u.Password = hashedPassword
	return nil
}

// CheckPassword 验证用户密码
func (u *User) CheckPassword(password string) bool {
	return auth.IsPasswordValid(u.Password, password)
}

// IsActive 检查用户是否活跃
func (u *User) IsActive() bool {
	return u.Status == StatusActive
}

// IsLocked 检查用户是否被锁定
func (u *User) IsLocked() bool {
	if u.Status == StatusLocked {
		return true
	}
	// 检查临时锁定
	if u.LockedUntil != nil && u.LockedUntil.After(time.Now()) {
		return true
	}
	return false
}

// CanLogin 检查用户是否可以登录
func (u *User) CanLogin() bool {
	return u.IsActive() && !u.IsLocked()
}

// IncrementLoginAttempts 增加登录尝试次数
func (u *User) IncrementLoginAttempts() {
	u.LoginAttempts++
	// 如果尝试次数超过5次，锁定账户30分钟
	if u.LoginAttempts >= 5 {
		lockUntil := time.Now().Add(30 * time.Minute)
		u.LockedUntil = &lockUntil
	}
}

// ResetLoginAttempts 重置登录尝试次数
func (u *User) ResetLoginAttempts() {
	u.LoginAttempts = 0
	u.LockedUntil = nil
}

// UpdateLastLogin 更新最后登录时间
func (u *User) UpdateLastLogin() {
	now := time.Now()
	u.LastLoginAt = &now
}

// HasRole 检查用户是否具有指定角色
func (u *User) HasRole(role string) bool {
	return u.Role == role
}

// IsAdmin 检查用户是否为管理员
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

// IsMerchant 检查用户是否为商家
func (u *User) IsMerchant() bool {
	return u.Role == RoleMerchant
}

// ToResponse 转换为响应格式
func (u *User) ToResponse() *UserResponse {
	return &UserResponse{
		ID:          u.ID,
		Username:    u.Username,
		Email:       u.Email,
		Nickname:    u.Nickname,
		Avatar:      u.Avatar,
		Phone:       u.Phone,
		Role:        u.Role,
		Status:      u.Status,
		LastLoginAt: u.LastLoginAt,
		CreatedAt:   u.CreatedAt,
	}
}
