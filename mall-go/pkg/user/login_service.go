package user

import (
	"fmt"
	"net"
	"regexp"
	"time"

	"mall-go/internal/model"
	"mall-go/pkg/auth"
	"mall-go/pkg/verification"

	"gorm.io/gorm"
)

// LoginService 用户登录服务
type LoginService struct {
	db *gorm.DB
}

// NewLoginService 创建用户登录服务
func NewLoginService(db *gorm.DB) *LoginService {
	return &LoginService{
		db: db,
	}
}

// LoginRequest 登录请求
type LoginRequest struct {
	Account    string `json:"account" binding:"required"`  // 账号（用户名/邮箱/手机号）
	Password   string `json:"password" binding:"required"` // 密码
	LoginType  string `json:"login_type"`                  // 登录类型：username/email/phone
	DeviceID   string `json:"device_id"`                   // 设备ID
	Platform   string `json:"platform"`                    // 平台：web/ios/android/mini
	RememberMe bool   `json:"remember_me"`                 // 记住我
	ClientIP   string `json:"-"`                           // 客户端IP（由中间件设置）
	UserAgent  string `json:"-"`                           // 用户代理（由中间件设置）
}

// LoginResponse 登录响应
type LoginResponse struct {
	UserID       uint                `json:"user_id"`
	Username     string              `json:"username"`
	Email        string              `json:"email"`
	Nickname     string              `json:"nickname"`
	Avatar       string              `json:"avatar"`
	Role         string              `json:"role"`
	Token        string              `json:"token"`
	RefreshToken string              `json:"refresh_token"`
	ExpiresAt    int64               `json:"expires_at"`
	User         *model.UserResponse `json:"user"`
	FirstLogin   bool                `json:"first_login"` // 是否首次登录
	NeedVerify   bool                `json:"need_verify"` // 是否需要验证
	Message      string              `json:"message"`
}

// RefreshTokenRequest 刷新token请求
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// ForgotPasswordRequest 忘记密码请求
type ForgotPasswordRequest struct {
	Account string `json:"account" binding:"required"` // 邮箱或手机号
	Type    string `json:"type" binding:"required"`    // email/phone
}

// ResetPasswordRequest 重置密码请求
type ResetPasswordRequest struct {
	Account         string `json:"account" binding:"required"`
	Code            string `json:"code" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=6"`
	ConfirmPassword string `json:"confirm_password" binding:"required"`
	Type            string `json:"type" binding:"required"` // email/phone
}

// Login 用户登录
func (ls *LoginService) Login(req *LoginRequest) (*LoginResponse, error) {
	// 1. 确定登录类型
	loginType := ls.determineLoginType(req.Account, req.LoginType)

	// 2. 查找用户
	user, err := ls.findUserByAccount(req.Account, loginType)
	if err != nil {
		// 记录登录失败日志
		ls.recordLoginLog(0, req, model.LoginStatusFailed, "用户不存在")
		return nil, fmt.Errorf("用户名或密码错误")
	}

	// 3. 检查用户状态
	if !user.CanLogin() {
		reason := "账户异常"
		if user.IsLocked() {
			reason = "账户已锁定"
		} else if user.Status == model.StatusBanned {
			reason = "账户已封禁"
		} else if user.Status == model.StatusInactive {
			reason = "账户未激活"
		}

		ls.recordLoginLog(user.ID, req, model.LoginStatusFailed, reason)
		return nil, fmt.Errorf(reason)
	}

	// 4. 验证密码
	if !auth.CheckPassword(req.Password, user.Password) {
		// 增加登录尝试次数
		user.IncrementLoginAttempts()
		ls.db.Save(user)

		ls.recordLoginLog(user.ID, req, model.LoginStatusFailed, "密码错误")
		return nil, fmt.Errorf("用户名或密码错误")
	}

	// 5. 登录成功，重置登录尝试次数
	user.ResetLoginAttempts()
	user.UpdateLastLogin()
	user.LastLoginIP = req.ClientIP
	user.LoginCount++

	// 检查是否首次登录
	firstLogin := user.LoginCount == 1

	// 6. 保存用户更新
	if err := ls.db.Save(user).Error; err != nil {
		return nil, fmt.Errorf("更新用户信息失败: %v", err)
	}

	// 7. 生成token
	tokenExpiry := 24 * time.Hour // 默认24小时
	if req.RememberMe {
		tokenExpiry = 30 * 24 * time.Hour // 记住我30天
	}

	token, expiresAt, err := auth.GenerateTokenWithExpiry(user.ID, user.Username, user.Role, tokenExpiry)
	if err != nil {
		ls.recordLoginLog(user.ID, req, model.LoginStatusFailed, "生成token失败")
		return nil, fmt.Errorf("登录失败: %v", err)
	}

	// 8. 生成刷新token
	refreshToken, _, err := auth.GenerateRefreshToken(user.ID, user.Username, user.Role)
	if err != nil {
		// 刷新token生成失败不影响登录
		refreshToken = ""
	}

	// 9. 记录登录成功日志
	ls.recordLoginLog(user.ID, req, model.LoginStatusSuccess, "登录成功")

	// 10. 构造响应
	response := &LoginResponse{
		UserID:       user.ID,
		Username:     user.Username,
		Email:        user.Email,
		Nickname:     user.Nickname,
		Avatar:       user.Avatar,
		Role:         user.Role,
		Token:        token,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt.Unix(),
		User:         user.ToResponse(),
		FirstLogin:   firstLogin,
		NeedVerify:   !user.EmailVerified,
		Message:      "登录成功",
	}

	return response, nil
}

// RefreshToken 刷新token
func (ls *LoginService) RefreshToken(req *RefreshTokenRequest) (*LoginResponse, error) {
	// 验证刷新token
	claims, err := auth.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("刷新token无效: %v", err)
	}

	// 查找用户
	var user model.User
	if err := ls.db.First(&user, claims.UserID).Error; err != nil {
		return nil, fmt.Errorf("用户不存在")
	}

	// 检查用户状态
	if !user.CanLogin() {
		return nil, fmt.Errorf("账户状态异常")
	}

	// 生成新的token
	token, err := auth.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		return nil, fmt.Errorf("生成token失败: %v", err)
	}

	// 计算过期时间（默认24小时）
	expiresAt := time.Now().Add(24 * time.Hour)

	// 生成新的刷新token
	refreshToken, _, err := auth.GenerateRefreshToken(user.ID, user.Username, user.Role)
	if err != nil {
		refreshToken = req.RefreshToken // 使用原刷新token
	}

	return &LoginResponse{
		UserID:       user.ID,
		Username:     user.Username,
		Email:        user.Email,
		Nickname:     user.Nickname,
		Avatar:       user.Avatar,
		Role:         user.Role,
		Token:        token,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt.Unix(),
		User:         user.ToResponse(),
		Message:      "token刷新成功",
	}, nil
}

// ForgotPassword 忘记密码
func (ls *LoginService) ForgotPassword(req *ForgotPasswordRequest) error {
	// 查找用户
	var user model.User
	var err error

	if req.Type == "email" {
		err = ls.db.Where("email = ?", req.Account).First(&user).Error
	} else if req.Type == "phone" {
		err = ls.db.Where("phone = ?", req.Account).First(&user).Error
	} else {
		return fmt.Errorf("不支持的重置类型")
	}

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 为了安全，即使用户不存在也返回成功
			return nil
		}
		return fmt.Errorf("查询用户失败: %v", err)
	}

	// 发送重置密码验证码
	verificationService := verification.NewVerificationService(ls.db)
	if req.Type == "email" {
		_, err = verificationService.SendEmailVerification(user.Email, model.VerificationTypeResetPassword, user.ID)
	} else {
		_, err = verificationService.SendSMSVerification(user.Phone, model.VerificationTypeResetPassword, user.ID)
	}

	if err != nil {
		return fmt.Errorf("发送验证码失败: %v", err)
	}

	return nil
}

// ResetPassword 重置密码
func (ls *LoginService) ResetPassword(req *ResetPasswordRequest) error {
	// 验证密码确认
	if req.NewPassword != req.ConfirmPassword {
		return fmt.Errorf("两次输入的密码不一致")
	}

	// 验证验证码
	verificationService := verification.NewVerificationService(ls.db)
	var verificationCode *model.UserVerificationCode
	var err error

	if req.Type == "email" {
		verificationCode, err = verificationService.VerifyEmailCode(req.Account, req.Code, model.VerificationTypeResetPassword)
	} else if req.Type == "phone" {
		verificationCode, err = verificationService.VerifyPhoneCode(req.Account, req.Code, model.VerificationTypeResetPassword)
	} else {
		return fmt.Errorf("不支持的重置类型")
	}

	if err != nil {
		return fmt.Errorf("验证码验证失败: %v", err)
	}

	// 查找用户
	var user model.User
	if req.Type == "email" {
		err = ls.db.Where("email = ?", req.Account).First(&user).Error
	} else {
		err = ls.db.Where("phone = ?", req.Account).First(&user).Error
	}

	if err != nil {
		return fmt.Errorf("用户不存在")
	}

	// 验证验证码是否属于该用户
	if verificationCode.UserID != 0 && verificationCode.UserID != user.ID {
		return fmt.Errorf("验证码无效")
	}

	// 加密新密码
	hashedPassword, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		return fmt.Errorf("密码加密失败: %v", err)
	}

	// 更新密码
	now := time.Now()
	user.Password = hashedPassword
	user.PasswordChangedAt = &now
	user.ResetLoginAttempts() // 重置登录尝试次数

	if err := ls.db.Save(&user).Error; err != nil {
		return fmt.Errorf("更新密码失败: %v", err)
	}

	return nil
}

// determineLoginType 确定登录类型
func (ls *LoginService) determineLoginType(account, loginType string) string {
	if loginType != "" {
		return loginType
	}

	// 自动判断登录类型
	if ls.isEmail(account) {
		return "email"
	} else if ls.isPhone(account) {
		return "phone"
	} else {
		return "username"
	}
}

// findUserByAccount 根据账号查找用户
func (ls *LoginService) findUserByAccount(account, loginType string) (*model.User, error) {
	var user model.User
	var err error

	switch loginType {
	case "email":
		err = ls.db.Where("email = ?", account).First(&user).Error
	case "phone":
		err = ls.db.Where("phone = ?", account).First(&user).Error
	case "username":
		err = ls.db.Where("username = ?", account).First(&user).Error
	default:
		// 尝试多种方式查找
		err = ls.db.Where("username = ? OR email = ? OR phone = ?", account, account, account).First(&user).Error
	}

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// recordLoginLog 记录登录日志
func (ls *LoginService) recordLoginLog(userID uint, req *LoginRequest, status, reason string) {
	loginLog := &model.UserLoginLog{
		UserID:     userID,
		LoginIP:    req.ClientIP,
		UserAgent:  req.UserAgent,
		DeviceID:   req.DeviceID,
		Platform:   req.Platform,
		Status:     status,
		FailReason: reason,
		LoginAt:    time.Now(),
	}

	// 尝试解析IP地理位置（简化实现）
	if req.ClientIP != "" {
		loginLog.Location = ls.getLocationByIP(req.ClientIP)
	}

	ls.db.Create(loginLog)
}

// getLocationByIP 根据IP获取地理位置（简化实现）
func (ls *LoginService) getLocationByIP(ip string) string {
	// TODO: 集成IP地理位置服务
	// 这里只是简单判断是否为内网IP
	if ls.isPrivateIP(ip) {
		return "内网"
	}
	return "未知"
}

// isPrivateIP 判断是否为内网IP
func (ls *LoginService) isPrivateIP(ip string) bool {
	privateRanges := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"127.0.0.0/8",
	}

	for _, cidr := range privateRanges {
		_, network, _ := net.ParseCIDR(cidr)
		if network.Contains(net.ParseIP(ip)) {
			return true
		}
	}
	return false
}

// isEmail 判断是否为邮箱格式
func (ls *LoginService) isEmail(account string) bool {
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`, account)
	return matched
}

// isPhone 判断是否为手机号格式
func (ls *LoginService) isPhone(account string) bool {
	matched, _ := regexp.MatchString(`^1[3-9]\d{9}$`, account)
	return matched
}

// 全局登录服务实例
var globalLoginService *LoginService

// InitGlobalLoginService 初始化全局登录服务
func InitGlobalLoginService(db *gorm.DB) {
	globalLoginService = NewLoginService(db)
}

// GetGlobalLoginService 获取全局登录服务
func GetGlobalLoginService() *LoginService {
	return globalLoginService
}
