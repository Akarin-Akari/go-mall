package user

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"mall-go/internal/model"
	"mall-go/pkg/auth"
	"mall-go/pkg/verification"

	"gorm.io/gorm"
)

// RegisterService 用户注册服务
type RegisterService struct {
	db                  *gorm.DB
	verificationService *verification.VerificationService
}

// NewRegisterService 创建用户注册服务
func NewRegisterService(db *gorm.DB) *RegisterService {
	return &RegisterService{
		db:                  db,
		verificationService: verification.NewVerificationService(db),
	}
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username         string `json:"username" binding:"required,min=3,max=50"`
	Email            string `json:"email" binding:"required,email"`
	Password         string `json:"password" binding:"required,min=6,max=50"`
	ConfirmPassword  string `json:"confirm_password" binding:"required"`
	Phone            string `json:"phone"`
	Nickname         string `json:"nickname"`
	EmailCode        string `json:"email_code"`        // 邮箱验证码
	PhoneCode        string `json:"phone_code"`        // 手机验证码
	AgreeTerms       bool   `json:"agree_terms" binding:"required"` // 同意服务条款
	InviteCode       string `json:"invite_code"`       // 邀请码（可选）
}

// RegisterResponse 注册响应
type RegisterResponse struct {
	UserID      uint   `json:"user_id"`
	Username    string `json:"username"`
	Email       string `json:"email"`
	Nickname    string `json:"nickname"`
	Token       string `json:"token"`        // 自动登录token
	ExpiresAt   int64  `json:"expires_at"`   // token过期时间
	NeedVerify  bool   `json:"need_verify"`  // 是否需要验证
	Message     string `json:"message"`
}

// SendEmailCodeRequest 发送邮箱验证码请求
type SendEmailCodeRequest struct {
	Email string `json:"email" binding:"required,email"`
	Type  string `json:"type" binding:"required"` // register/reset_password等
}

// SendPhoneCodeRequest 发送手机验证码请求
type SendPhoneCodeRequest struct {
	Phone string `json:"phone" binding:"required"`
	Type  string `json:"type" binding:"required"` // register/reset_password等
}

// CheckUsernameRequest 检查用户名请求
type CheckUsernameRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
}

// CheckEmailRequest 检查邮箱请求
type CheckEmailRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// Register 用户注册
func (rs *RegisterService) Register(req *RegisterRequest) (*RegisterResponse, error) {
	// 1. 验证请求数据
	if err := rs.validateRegisterRequest(req); err != nil {
		return nil, err
	}

	// 2. 检查用户名是否已存在
	if exists, err := rs.IsUsernameExists(req.Username); err != nil {
		return nil, fmt.Errorf("检查用户名失败: %v", err)
	} else if exists {
		return nil, fmt.Errorf("用户名已存在")
	}

	// 3. 检查邮箱是否已存在
	if exists, err := rs.IsEmailExists(req.Email); err != nil {
		return nil, fmt.Errorf("检查邮箱失败: %v", err)
	} else if exists {
		return nil, fmt.Errorf("邮箱已被注册")
	}

	// 4. 检查手机号是否已存在（如果提供了手机号）
	if req.Phone != "" {
		if exists, err := rs.IsPhoneExists(req.Phone); err != nil {
			return nil, fmt.Errorf("检查手机号失败: %v", err)
		} else if exists {
			return nil, fmt.Errorf("手机号已被注册")
		}
	}

	// 5. 验证邮箱验证码
	if req.EmailCode != "" {
		_, err := rs.verificationService.VerifyEmailCode(req.Email, req.EmailCode, model.VerificationTypeRegister)
		if err != nil {
			return nil, fmt.Errorf("邮箱验证码验证失败: %v", err)
		}
	}

	// 6. 验证手机验证码（如果提供了手机号和验证码）
	if req.Phone != "" && req.PhoneCode != "" {
		_, err := rs.verificationService.VerifyPhoneCode(req.Phone, req.PhoneCode, model.VerificationTypeRegister)
		if err != nil {
			return nil, fmt.Errorf("手机验证码验证失败: %v", err)
		}
	}

	// 7. 加密密码
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("密码加密失败: %v", err)
	}

	// 8. 创建用户
	user := &model.User{
		Username:      req.Username,
		Email:         req.Email,
		Password:      hashedPassword,
		Nickname:      req.Nickname,
		Phone:         req.Phone,
		Role:          model.RoleUser,
		Status:        model.StatusActive,
		EmailVerified: req.EmailCode != "", // 如果提供了验证码则认为已验证
		PhoneVerified: req.Phone != "" && req.PhoneCode != "",
	}

	// 设置验证时间
	now := time.Now()
	if user.EmailVerified {
		user.EmailVerifiedAt = &now
	}
	if user.PhoneVerified {
		user.PhoneVerifiedAt = &now
	}

	// 如果没有设置昵称，使用用户名
	if user.Nickname == "" {
		user.Nickname = user.Username
	}

	// 9. 保存用户到数据库
	if err := rs.db.Create(user).Error; err != nil {
		return nil, fmt.Errorf("创建用户失败: %v", err)
	}

	// 10. 创建用户资料
	profile := &model.UserProfile{
		UserID: user.ID,
	}
	rs.db.Create(profile)

	// 11. 生成登录token（自动登录）
	token, expiresAt, err := auth.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		// 注册成功但token生成失败，不影响注册结果
		return &RegisterResponse{
			UserID:     user.ID,
			Username:   user.Username,
			Email:      user.Email,
			Nickname:   user.Nickname,
			NeedVerify: !user.EmailVerified,
			Message:    "注册成功，请手动登录",
		}, nil
	}

	return &RegisterResponse{
		UserID:     user.ID,
		Username:   user.Username,
		Email:      user.Email,
		Nickname:   user.Nickname,
		Token:      token,
		ExpiresAt:  expiresAt,
		NeedVerify: !user.EmailVerified,
		Message:    "注册成功",
	}, nil
}

// SendEmailCode 发送邮箱验证码
func (rs *RegisterService) SendEmailCode(req *SendEmailCodeRequest) error {
	// 验证邮箱格式
	if !rs.isValidEmail(req.Email) {
		return fmt.Errorf("邮箱格式不正确")
	}

	// 如果是注册验证码，检查邮箱是否已被注册
	if req.Type == model.VerificationTypeRegister {
		if exists, err := rs.IsEmailExists(req.Email); err != nil {
			return fmt.Errorf("检查邮箱失败: %v", err)
		} else if exists {
			return fmt.Errorf("邮箱已被注册")
		}
	}

	// 发送验证码
	_, err := rs.verificationService.SendEmailVerification(req.Email, req.Type, 0)
	if err != nil {
		return fmt.Errorf("发送邮箱验证码失败: %v", err)
	}

	return nil
}

// SendPhoneCode 发送手机验证码
func (rs *RegisterService) SendPhoneCode(req *SendPhoneCodeRequest) error {
	// 验证手机号格式
	if !rs.isValidPhone(req.Phone) {
		return fmt.Errorf("手机号格式不正确")
	}

	// 如果是注册验证码，检查手机号是否已被注册
	if req.Type == model.VerificationTypeRegister {
		if exists, err := rs.IsPhoneExists(req.Phone); err != nil {
			return fmt.Errorf("检查手机号失败: %v", err)
		} else if exists {
			return fmt.Errorf("手机号已被注册")
		}
	}

	// 发送验证码
	_, err := rs.verificationService.SendSMSVerification(req.Phone, req.Type, 0)
	if err != nil {
		return fmt.Errorf("发送手机验证码失败: %v", err)
	}

	return nil
}

// IsUsernameExists 检查用户名是否存在
func (rs *RegisterService) IsUsernameExists(username string) (bool, error) {
	var count int64
	err := rs.db.Model(&model.User{}).Where("username = ?", username).Count(&count).Error
	return count > 0, err
}

// IsEmailExists 检查邮箱是否存在
func (rs *RegisterService) IsEmailExists(email string) (bool, error) {
	var count int64
	err := rs.db.Model(&model.User{}).Where("email = ?", email).Count(&count).Error
	return count > 0, err
}

// IsPhoneExists 检查手机号是否存在
func (rs *RegisterService) IsPhoneExists(phone string) (bool, error) {
	if phone == "" {
		return false, nil
	}
	var count int64
	err := rs.db.Model(&model.User{}).Where("phone = ?", phone).Count(&count).Error
	return count > 0, err
}

// CheckUsername 检查用户名可用性
func (rs *RegisterService) CheckUsername(req *CheckUsernameRequest) (bool, error) {
	// 验证用户名格式
	if !rs.isValidUsername(req.Username) {
		return false, fmt.Errorf("用户名格式不正确")
	}

	exists, err := rs.IsUsernameExists(req.Username)
	if err != nil {
		return false, err
	}

	return !exists, nil // 返回是否可用（不存在则可用）
}

// CheckEmail 检查邮箱可用性
func (rs *RegisterService) CheckEmail(req *CheckEmailRequest) (bool, error) {
	// 验证邮箱格式
	if !rs.isValidEmail(req.Email) {
		return false, fmt.Errorf("邮箱格式不正确")
	}

	exists, err := rs.IsEmailExists(req.Email)
	if err != nil {
		return false, err
	}

	return !exists, nil // 返回是否可用（不存在则可用）
}

// validateRegisterRequest 验证注册请求
func (rs *RegisterService) validateRegisterRequest(req *RegisterRequest) error {
	// 验证密码确认
	if req.Password != req.ConfirmPassword {
		return fmt.Errorf("两次输入的密码不一致")
	}

	// 验证用户名格式
	if !rs.isValidUsername(req.Username) {
		return fmt.Errorf("用户名格式不正确，只能包含字母、数字和下划线，长度3-50字符")
	}

	// 验证邮箱格式
	if !rs.isValidEmail(req.Email) {
		return fmt.Errorf("邮箱格式不正确")
	}

	// 验证密码强度
	if !rs.isValidPassword(req.Password) {
		return fmt.Errorf("密码强度不够，至少包含6个字符")
	}

	// 验证手机号格式（如果提供了）
	if req.Phone != "" && !rs.isValidPhone(req.Phone) {
		return fmt.Errorf("手机号格式不正确")
	}

	// 验证服务条款同意
	if !req.AgreeTerms {
		return fmt.Errorf("请同意服务条款")
	}

	return nil
}

// isValidUsername 验证用户名格式
func (rs *RegisterService) isValidUsername(username string) bool {
	// 用户名只能包含字母、数字和下划线，长度3-50
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_]{3,50}$`, username)
	return matched
}

// isValidEmail 验证邮箱格式
func (rs *RegisterService) isValidEmail(email string) bool {
	// 简单的邮箱格式验证
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`, email)
	return matched
}

// isValidPassword 验证密码强度
func (rs *RegisterService) isValidPassword(password string) bool {
	// 密码至少6个字符
	return len(strings.TrimSpace(password)) >= 6
}

// isValidPhone 验证手机号格式
func (rs *RegisterService) isValidPhone(phone string) bool {
	// 中国大陆手机号格式验证
	matched, _ := regexp.MatchString(`^1[3-9]\d{9}$`, phone)
	return matched
}

// 全局注册服务实例
var globalRegisterService *RegisterService

// InitGlobalRegisterService 初始化全局注册服务
func InitGlobalRegisterService(db *gorm.DB) {
	globalRegisterService = NewRegisterService(db)
}

// GetGlobalRegisterService 获取全局注册服务
func GetGlobalRegisterService() *RegisterService {
	return globalRegisterService
}
