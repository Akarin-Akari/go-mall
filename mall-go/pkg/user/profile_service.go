package user

import (
	"fmt"
	"time"

	"mall-go/internal/model"
	"mall-go/pkg/auth"
	"mall-go/pkg/upload"
	"mall-go/pkg/verification"

	"gorm.io/gorm"
)

// ProfileService 用户资料服务
type ProfileService struct {
	db                  *gorm.DB
	verificationService *verification.VerificationService
	fileManager         *upload.FileManager
}

// NewProfileService 创建用户资料服务
func NewProfileService(db *gorm.DB) *ProfileService {
	return &ProfileService{
		db:                  db,
		verificationService: verification.NewVerificationService(db),
	}
}

// UpdateProfileRequest 更新用户资料请求
type UpdateProfileRequest struct {
	Nickname  string     `json:"nickname"`
	RealName  string     `json:"real_name"`
	Gender    string     `json:"gender"`
	Birthday  *time.Time `json:"birthday"`
	Bio       string     `json:"bio"`
	Location  string     `json:"location"`
	Website   string     `json:"website"`
	Company   string     `json:"company"`
	Position  string     `json:"position"`
	Industry  string     `json:"industry"`
	Education string     `json:"education"`
	Skills    string     `json:"skills"`
	Interests string     `json:"interests"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword     string `json:"old_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=6"`
	ConfirmPassword string `json:"confirm_password" binding:"required"`
}

// ChangeEmailRequest 修改邮箱请求
type ChangeEmailRequest struct {
	NewEmail    string `json:"new_email" binding:"required,email"`
	Password    string `json:"password" binding:"required"`
	Code        string `json:"code" binding:"required"`
}

// ChangePhoneRequest 修改手机号请求
type ChangePhoneRequest struct {
	NewPhone    string `json:"new_phone" binding:"required"`
	Password    string `json:"password" binding:"required"`
	Code        string `json:"code" binding:"required"`
}

// UploadAvatarRequest 上传头像请求
type UploadAvatarRequest struct {
	FileID uint `json:"file_id" binding:"required"`
}

// GetProfile 获取用户资料
func (ps *ProfileService) GetProfile(userID uint) (*model.UserProfile, error) {
	var profile model.UserProfile
	err := ps.db.Where("user_id = ?", userID).First(&profile).Error
	
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 如果资料不存在，创建默认资料
			profile = model.UserProfile{
				UserID: userID,
			}
			if err := ps.db.Create(&profile).Error; err != nil {
				return nil, fmt.Errorf("创建用户资料失败: %v", err)
			}
		} else {
			return nil, fmt.Errorf("查询用户资料失败: %v", err)
		}
	}

	return &profile, nil
}

// UpdateProfile 更新用户资料
func (ps *ProfileService) UpdateProfile(userID uint, req *UpdateProfileRequest) (*model.UserProfile, error) {
	// 获取或创建用户资料
	profile, err := ps.GetProfile(userID)
	if err != nil {
		return nil, err
	}

	// 更新资料字段
	if req.Nickname != "" {
		profile.User.Nickname = req.Nickname
	}
	if req.RealName != "" {
		profile.User.RealName = req.RealName
	}
	if req.Gender != "" && ps.isValidGender(req.Gender) {
		profile.User.Gender = req.Gender
	}
	if req.Birthday != nil {
		profile.User.Birthday = req.Birthday
	}
	if req.Bio != "" {
		profile.User.Bio = req.Bio
	}
	if req.Location != "" {
		profile.User.Location = req.Location
	}
	if req.Website != "" {
		profile.User.Website = req.Website
	}

	// 更新扩展资料
	if req.Company != "" {
		profile.Company = req.Company
	}
	if req.Position != "" {
		profile.Position = req.Position
	}
	if req.Industry != "" {
		profile.Industry = req.Industry
	}
	if req.Education != "" {
		profile.Education = req.Education
	}
	if req.Skills != "" {
		profile.Skills = req.Skills
	}
	if req.Interests != "" {
		profile.Interests = req.Interests
	}

	// 保存更新
	if err := ps.db.Save(profile).Error; err != nil {
		return nil, fmt.Errorf("更新用户资料失败: %v", err)
	}

	// 同时更新用户表中的基本信息
	var user model.User
	if err := ps.db.First(&user, userID).Error; err != nil {
		return nil, fmt.Errorf("查询用户失败: %v", err)
	}

	user.Nickname = profile.User.Nickname
	user.RealName = profile.User.RealName
	user.Gender = profile.User.Gender
	user.Birthday = profile.User.Birthday
	user.Bio = profile.User.Bio
	user.Location = profile.User.Location
	user.Website = profile.User.Website

	if err := ps.db.Save(&user).Error; err != nil {
		return nil, fmt.Errorf("更新用户基本信息失败: %v", err)
	}

	return profile, nil
}

// ChangePassword 修改密码
func (ps *ProfileService) ChangePassword(userID uint, req *ChangePasswordRequest) error {
	// 验证密码确认
	if req.NewPassword != req.ConfirmPassword {
		return fmt.Errorf("两次输入的密码不一致")
	}

	// 查询用户
	var user model.User
	if err := ps.db.First(&user, userID).Error; err != nil {
		return fmt.Errorf("用户不存在")
	}

	// 验证旧密码
	if !auth.CheckPassword(req.OldPassword, user.Password) {
		return fmt.Errorf("原密码错误")
	}

	// 检查新密码是否与旧密码相同
	if auth.CheckPassword(req.NewPassword, user.Password) {
		return fmt.Errorf("新密码不能与原密码相同")
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

	if err := ps.db.Save(&user).Error; err != nil {
		return fmt.Errorf("更新密码失败: %v", err)
	}

	return nil
}

// ChangeEmail 修改邮箱
func (ps *ProfileService) ChangeEmail(userID uint, req *ChangeEmailRequest) error {
	// 查询用户
	var user model.User
	if err := ps.db.First(&user, userID).Error; err != nil {
		return fmt.Errorf("用户不存在")
	}

	// 验证密码
	if !auth.CheckPassword(req.Password, user.Password) {
		return fmt.Errorf("密码错误")
	}

	// 检查新邮箱是否已被使用
	var existingUser model.User
	if err := ps.db.Where("email = ? AND id != ?", req.NewEmail, userID).First(&existingUser).Error; err == nil {
		return fmt.Errorf("邮箱已被其他用户使用")
	}

	// 验证邮箱验证码
	_, err := ps.verificationService.VerifyEmailCode(req.NewEmail, req.Code, model.VerificationTypeChangeEmail)
	if err != nil {
		return fmt.Errorf("验证码验证失败: %v", err)
	}

	// 更新邮箱
	user.Email = req.NewEmail
	user.EmailVerified = true
	now := time.Now()
	user.EmailVerifiedAt = &now

	if err := ps.db.Save(&user).Error; err != nil {
		return fmt.Errorf("更新邮箱失败: %v", err)
	}

	return nil
}

// ChangePhone 修改手机号
func (ps *ProfileService) ChangePhone(userID uint, req *ChangePhoneRequest) error {
	// 查询用户
	var user model.User
	if err := ps.db.First(&user, userID).Error; err != nil {
		return fmt.Errorf("用户不存在")
	}

	// 验证密码
	if !auth.CheckPassword(req.Password, user.Password) {
		return fmt.Errorf("密码错误")
	}

	// 检查新手机号是否已被使用
	var existingUser model.User
	if err := ps.db.Where("phone = ? AND id != ?", req.NewPhone, userID).First(&existingUser).Error; err == nil {
		return fmt.Errorf("手机号已被其他用户使用")
	}

	// 验证手机验证码
	_, err := ps.verificationService.VerifyPhoneCode(req.NewPhone, req.Code, model.VerificationTypeChangePhone)
	if err != nil {
		return fmt.Errorf("验证码验证失败: %v", err)
	}

	// 更新手机号
	user.Phone = req.NewPhone
	user.PhoneVerified = true
	now := time.Now()
	user.PhoneVerifiedAt = &now

	if err := ps.db.Save(&user).Error; err != nil {
		return fmt.Errorf("更新手机号失败: %v", err)
	}

	return nil
}

// UploadAvatar 上传头像
func (ps *ProfileService) UploadAvatar(userID uint, req *UploadAvatarRequest) error {
	// 查询用户
	var user model.User
	if err := ps.db.First(&user, userID).Error; err != nil {
		return fmt.Errorf("用户不存在")
	}

	// 查询文件信息
	var file model.File
	if err := ps.db.Where("id = ? AND user_id = ?", req.FileID, userID).First(&file).Error; err != nil {
		return fmt.Errorf("文件不存在或无权限")
	}

	// 检查文件类型
	if !file.IsImage() {
		return fmt.Errorf("只能上传图片文件")
	}

	// 更新用户头像
	user.Avatar = file.URL

	if err := ps.db.Save(&user).Error; err != nil {
		return fmt.Errorf("更新头像失败: %v", err)
	}

	return nil
}

// GetCurrentUser 获取当前用户信息
func (ps *ProfileService) GetCurrentUser(userID uint) (*model.UserResponse, error) {
	var user model.User
	if err := ps.db.First(&user, userID).Error; err != nil {
		return nil, fmt.Errorf("用户不存在")
	}

	return user.ToResponse(), nil
}

// GetUserByID 根据ID获取用户信息
func (ps *ProfileService) GetUserByID(userID uint) (*model.User, error) {
	var user model.User
	if err := ps.db.Preload("Profile").First(&user, userID).Error; err != nil {
		return nil, fmt.Errorf("用户不存在")
	}

	return &user, nil
}

// GetLoginLogs 获取登录日志
func (ps *ProfileService) GetLoginLogs(userID uint, page, pageSize int) ([]model.UserLoginLog, int64, error) {
	var logs []model.UserLoginLog
	var total int64

	// 获取总数
	if err := ps.db.Model(&model.UserLoginLog{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("查询登录日志总数失败: %v", err)
	}

	// 获取分页数据
	offset := (page - 1) * pageSize
	if err := ps.db.Where("user_id = ?", userID).
		Order("login_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&logs).Error; err != nil {
		return nil, 0, fmt.Errorf("查询登录日志失败: %v", err)
	}

	return logs, total, nil
}

// GetSecuritySettings 获取安全设置
func (ps *ProfileService) GetSecuritySettings(userID uint) (map[string]interface{}, error) {
	var user model.User
	if err := ps.db.First(&user, userID).Error; err != nil {
		return nil, fmt.Errorf("用户不存在")
	}

	settings := map[string]interface{}{
		"email_verified":      user.EmailVerified,
		"phone_verified":      user.PhoneVerified,
		"two_factor_enabled":  user.TwoFactorEnabled,
		"password_changed_at": user.PasswordChangedAt,
		"last_login_at":       user.LastLoginAt,
		"last_login_ip":       user.LastLoginIP,
		"login_count":         user.LoginCount,
	}

	return settings, nil
}

// isValidGender 验证性别值是否有效
func (ps *ProfileService) isValidGender(gender string) bool {
	validGenders := []string{model.GenderMale, model.GenderFemale, model.GenderUnknown}
	for _, validGender := range validGenders {
		if gender == validGender {
			return true
		}
	}
	return false
}

// 全局资料服务实例
var globalProfileService *ProfileService

// InitGlobalProfileService 初始化全局资料服务
func InitGlobalProfileService(db *gorm.DB) {
	globalProfileService = NewProfileService(db)
}

// GetGlobalProfileService 获取全局资料服务
func GetGlobalProfileService() *ProfileService {
	return globalProfileService
}
