package model

import (
	"time"

	"github.com/shopspring/decimal"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID       uint   `gorm:"primarykey" json:"id"`
	Username string `gorm:"uniqueIndex;not null;size:50" json:"username"`
	Email    string `gorm:"uniqueIndex;not null;size:100" json:"email"`
	Password string `gorm:"not null;size:255" json:"-"` // 密码哈希
	Nickname string `gorm:"size:50" json:"nickname"`
	Avatar   string `gorm:"size:255" json:"avatar"`
	Phone    string `gorm:"uniqueIndex;size:20" json:"phone"`
	Role     string `gorm:"default:'user';size:20;index" json:"role"`
	Status   string `gorm:"default:'active';size:20;index" json:"status"`

	// 扩展信息
	RealName string     `gorm:"size:50" json:"real_name"`                // 真实姓名
	Gender   string     `gorm:"size:10;default:'unknown'" json:"gender"` // 性别: male/female/unknown
	Birthday *time.Time `gorm:"null" json:"birthday"`                    // 生日
	Bio      string     `gorm:"size:500" json:"bio"`                     // 个人简介
	Location string     `gorm:"size:100" json:"location"`                // 所在地
	Website  string     `gorm:"size:255" json:"website"`                 // 个人网站

	// 验证状态
	EmailVerified   bool       `gorm:"default:false" json:"email_verified"` // 邮箱验证状态
	PhoneVerified   bool       `gorm:"default:false" json:"phone_verified"` // 手机验证状态
	EmailVerifiedAt *time.Time `gorm:"null" json:"email_verified_at"`       // 邮箱验证时间
	PhoneVerifiedAt *time.Time `gorm:"null" json:"phone_verified_at"`       // 手机验证时间

	// 安全信息
	LastLoginAt       *time.Time `gorm:"null" json:"last_login_at"`               // 最后登录时间
	LastLoginIP       string     `gorm:"size:45" json:"last_login_ip"`            // 最后登录IP
	LoginAttempts     int        `gorm:"default:0" json:"-"`                      // 登录尝试次数
	LockedUntil       *time.Time `gorm:"null" json:"-"`                           // 账户锁定到期时间
	PasswordChangedAt *time.Time `gorm:"null" json:"-"`                           // 密码修改时间
	TwoFactorEnabled  bool       `gorm:"default:false" json:"two_factor_enabled"` // 双因子认证
	TwoFactorSecret   string     `gorm:"size:32" json:"-"`                        // 双因子认证密钥

	// 财务信息
	Balance decimal.Decimal `gorm:"type:decimal(10,2);default:0.00" json:"balance"` // 账户余额

	// 统计信息
	LoginCount     int `gorm:"default:0" json:"login_count"`     // 登录次数
	PostCount      int `gorm:"default:0" json:"post_count"`      // 发帖数量
	FollowerCount  int `gorm:"default:0" json:"follower_count"`  // 粉丝数量
	FollowingCount int `gorm:"default:0" json:"following_count"` // 关注数量

	// 时间戳
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联关系
	Files     []File         `gorm:"foreignKey:UploadUserID" json:"files,omitempty"` // 用户文件
	LoginLogs []UserLoginLog `gorm:"foreignKey:UserID" json:"login_logs,omitempty"`  // 登录日志
	Profile   *UserProfile   `gorm:"foreignKey:UserID" json:"profile,omitempty"`     // 用户资料
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// UserRegisterRequest 用户注册请求
type UserRegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=128"` // 更新密码长度要求
	Nickname string `json:"nickname" binding:"required,min=2,max=50"`
	Role     string `json:"role" binding:"omitempty,oneof=user merchant admin"` // 可选角色
}

// UserLoginRequest 用户登录请求
type UserLoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	DeviceID string `json:"device_id" binding:"omitempty,max=100"` // 设备标识
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

// 用户角色常量已在permission.go中定义，这里移除重复定义

// 用户状态常量
const (
	StatusActive   = "active"   // 活跃
	StatusInactive = "inactive" // 非活跃
	StatusLocked   = "locked"   // 锁定
	StatusBanned   = "banned"   // 封禁
)

// SetPassword 设置用户密码（加密存储）
func (u *User) SetPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

// CheckPassword 验证用户密码
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
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

// UserProfile 用户详细资料模型
type UserProfile struct {
	ID        uint   `gorm:"primarykey" json:"id"`
	UserID    uint   `gorm:"uniqueIndex;not null" json:"user_id"`
	Company   string `gorm:"size:100" json:"company"`   // 公司
	Position  string `gorm:"size:50" json:"position"`   // 职位
	Industry  string `gorm:"size:50" json:"industry"`   // 行业
	Education string `gorm:"size:100" json:"education"` // 教育背景
	Skills    string `gorm:"size:500" json:"skills"`    // 技能标签
	Interests string `gorm:"size:500" json:"interests"` // 兴趣爱好

	// 社交媒体
	WechatID   string `gorm:"size:50" json:"wechat_id"`    // 微信号
	QQNumber   string `gorm:"size:20" json:"qq_number"`    // QQ号
	WeiboID    string `gorm:"size:50" json:"weibo_id"`     // 微博ID
	GithubID   string `gorm:"size:50" json:"github_id"`    // GitHub ID
	LinkedinID string `gorm:"size:100" json:"linkedin_id"` // LinkedIn ID

	// 偏好设置
	Language    string `gorm:"size:10;default:'zh-CN'" json:"language"`         // 语言偏好
	Timezone    string `gorm:"size:50;default:'Asia/Shanghai'" json:"timezone"` // 时区
	Theme       string `gorm:"size:20;default:'light'" json:"theme"`            // 主题偏好
	EmailNotify bool   `gorm:"default:true" json:"email_notify"`                // 邮件通知
	SmsNotify   bool   `gorm:"default:false" json:"sms_notify"`                 // 短信通知
	PushNotify  bool   `gorm:"default:true" json:"push_notify"`                 // 推送通知

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联关系
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName 指定表名
func (UserProfile) TableName() string {
	return "user_profiles"
}

// UserLoginLog 用户登录日志模型
type UserLoginLog struct {
	ID         uint      `gorm:"primarykey" json:"id"`
	UserID     uint      `gorm:"not null;index" json:"user_id"`
	LoginIP    string    `gorm:"size:45" json:"login_ip"`     // 登录IP
	UserAgent  string    `gorm:"size:500" json:"user_agent"`  // 用户代理
	DeviceID   string    `gorm:"size:100" json:"device_id"`   // 设备ID
	Platform   string    `gorm:"size:20" json:"platform"`     // 平台 (web/ios/android)
	Location   string    `gorm:"size:100" json:"location"`    // 登录地点
	Status     string    `gorm:"size:20" json:"status"`       // 登录状态 (success/failed)
	FailReason string    `gorm:"size:100" json:"fail_reason"` // 失败原因
	LoginAt    time.Time `gorm:"not null" json:"login_at"`    // 登录时间

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联关系
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName 指定表名
func (UserLoginLog) TableName() string {
	return "user_login_logs"
}

// UserVerificationCode 用户验证码模型
type UserVerificationCode struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	UserID    uint      `gorm:"index" json:"user_id"`         // 用户ID，可为空（注册时）
	Email     string    `gorm:"size:100" json:"email"`        // 邮箱
	Phone     string    `gorm:"size:20" json:"phone"`         // 手机号
	Code      string    `gorm:"size:10;not null" json:"code"` // 验证码
	Type      string    `gorm:"size:20;not null" json:"type"` // 类型：register/login/reset_password/change_email/change_phone
	Used      bool      `gorm:"default:false" json:"used"`    // 是否已使用
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`   // 过期时间

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (UserVerificationCode) TableName() string {
	return "user_verification_codes"
}

// IsExpired 检查验证码是否过期
func (c *UserVerificationCode) IsExpired() bool {
	return time.Now().After(c.ExpiresAt)
}

// IsValid 检查验证码是否有效
func (c *UserVerificationCode) IsValid() bool {
	return !c.Used && !c.IsExpired()
}

// UserSocialAccount 用户社交账号绑定模型
type UserSocialAccount struct {
	ID           uint       `gorm:"primarykey" json:"id"`
	UserID       uint       `gorm:"not null;index" json:"user_id"`
	Provider     string     `gorm:"size:20;not null" json:"provider"`     // 提供商：wechat/qq/weibo/github
	ProviderID   string     `gorm:"size:100;not null" json:"provider_id"` // 第三方用户ID
	ProviderName string     `gorm:"size:100" json:"provider_name"`        // 第三方用户名
	Avatar       string     `gorm:"size:255" json:"avatar"`               // 第三方头像
	AccessToken  string     `gorm:"size:500" json:"access_token"`         // 访问令牌
	RefreshToken string     `gorm:"size:500" json:"refresh_token"`        // 刷新令牌
	ExpiresAt    *time.Time `gorm:"null" json:"expires_at"`               // 令牌过期时间

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联关系
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName 指定表名
func (UserSocialAccount) TableName() string {
	return "user_social_accounts"
}

// 用户角色常量
const (
	RoleSuperAdmin = "super_admin" // 超级管理员
	RoleAdmin      = "admin"       // 管理员
	RoleUser       = "user"        // 普通用户
	RoleMerchant   = "merchant"    // 商家
	RoleCustomer   = "customer"    // 客户（别名，用于兼容）
)

// 用户状态常量
const (
	UserStatusActive   = "active"   // 激活
	UserStatusInactive = "inactive" // 未激活
	UserStatusLocked   = "locked"   // 锁定
	UserStatusBanned   = "banned"   // 禁用
)

// 验证码类型常量
const (
	VerificationTypeRegister      = "register"       // 注册验证
	VerificationTypeLogin         = "login"          // 登录验证
	VerificationTypeResetPassword = "reset_password" // 重置密码
	VerificationTypeChangeEmail   = "change_email"   // 修改邮箱
	VerificationTypeChangePhone   = "change_phone"   // 修改手机号
)

// 性别常量
const (
	GenderMale    = "male"    // 男性
	GenderFemale  = "female"  // 女性
	GenderUnknown = "unknown" // 未知
)

// 社交账号提供商常量
const (
	ProviderWechat = "wechat" // 微信
	ProviderQQ     = "qq"     // QQ
	ProviderWeibo  = "weibo"  // 微博
	ProviderGithub = "github" // GitHub
)

// 登录状态常量
const (
	LoginStatusSuccess = "success" // 登录成功
	LoginStatusFailed  = "failed"  // 登录失败
)

// 平台常量
const (
	PlatformWeb     = "web"     // 网页
	PlatformIOS     = "ios"     // iOS应用
	PlatformAndroid = "android" // Android应用
	PlatformMini    = "mini"    // 小程序
)
