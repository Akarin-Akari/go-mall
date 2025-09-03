package verification

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"mall-go/internal/model"

	"gorm.io/gorm"
)

// VerificationService 验证码服务
type VerificationService struct {
	db *gorm.DB
}

// NewVerificationService 创建验证码服务
func NewVerificationService(db *gorm.DB) *VerificationService {
	return &VerificationService{
		db: db,
	}
}

// SendEmailVerification 发送邮箱验证码
func (vs *VerificationService) SendEmailVerification(email, codeType string, userID uint) (*model.UserVerificationCode, error) {
	// 生成验证码
	code, err := vs.generateCode(6)
	if err != nil {
		return nil, fmt.Errorf("生成验证码失败: %v", err)
	}

	// 检查是否存在未过期的验证码
	var existingCode model.UserVerificationCode
	err = vs.db.Where("email = ? AND type = ? AND used = false AND expires_at > ?", 
		email, codeType, time.Now()).First(&existingCode).Error
	
	if err == nil {
		// 如果存在未过期的验证码，更新它
		existingCode.Code = code
		existingCode.ExpiresAt = time.Now().Add(10 * time.Minute) // 10分钟有效期
		err = vs.db.Save(&existingCode).Error
		if err != nil {
			return nil, fmt.Errorf("更新验证码失败: %v", err)
		}
		
		// 发送邮件
		err = vs.sendEmail(email, code, codeType)
		if err != nil {
			return nil, fmt.Errorf("发送邮件失败: %v", err)
		}
		
		return &existingCode, nil
	}

	// 创建新的验证码记录
	verificationCode := &model.UserVerificationCode{
		UserID:    userID,
		Email:     email,
		Code:      code,
		Type:      codeType,
		Used:      false,
		ExpiresAt: time.Now().Add(10 * time.Minute), // 10分钟有效期
	}

	err = vs.db.Create(verificationCode).Error
	if err != nil {
		return nil, fmt.Errorf("保存验证码失败: %v", err)
	}

	// 发送邮件
	err = vs.sendEmail(email, code, codeType)
	if err != nil {
		return nil, fmt.Errorf("发送邮件失败: %v", err)
	}

	return verificationCode, nil
}

// SendSMSVerification 发送短信验证码
func (vs *VerificationService) SendSMSVerification(phone, codeType string, userID uint) (*model.UserVerificationCode, error) {
	// 生成验证码
	code, err := vs.generateCode(6)
	if err != nil {
		return nil, fmt.Errorf("生成验证码失败: %v", err)
	}

	// 检查是否存在未过期的验证码
	var existingCode model.UserVerificationCode
	err = vs.db.Where("phone = ? AND type = ? AND used = false AND expires_at > ?", 
		phone, codeType, time.Now()).First(&existingCode).Error
	
	if err == nil {
		// 如果存在未过期的验证码，更新它
		existingCode.Code = code
		existingCode.ExpiresAt = time.Now().Add(5 * time.Minute) // 5分钟有效期
		err = vs.db.Save(&existingCode).Error
		if err != nil {
			return nil, fmt.Errorf("更新验证码失败: %v", err)
		}
		
		// 发送短信
		err = vs.sendSMS(phone, code, codeType)
		if err != nil {
			return nil, fmt.Errorf("发送短信失败: %v", err)
		}
		
		return &existingCode, nil
	}

	// 创建新的验证码记录
	verificationCode := &model.UserVerificationCode{
		UserID:    userID,
		Phone:     phone,
		Code:      code,
		Type:      codeType,
		Used:      false,
		ExpiresAt: time.Now().Add(5 * time.Minute), // 5分钟有效期
	}

	err = vs.db.Create(verificationCode).Error
	if err != nil {
		return nil, fmt.Errorf("保存验证码失败: %v", err)
	}

	// 发送短信
	err = vs.sendSMS(phone, code, codeType)
	if err != nil {
		return nil, fmt.Errorf("发送短信失败: %v", err)
	}

	return verificationCode, nil
}

// VerifyEmailCode 验证邮箱验证码
func (vs *VerificationService) VerifyEmailCode(email, code, codeType string) (*model.UserVerificationCode, error) {
	var verificationCode model.UserVerificationCode
	err := vs.db.Where("email = ? AND code = ? AND type = ? AND used = false AND expires_at > ?", 
		email, code, codeType, time.Now()).First(&verificationCode).Error
	
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("验证码无效或已过期")
		}
		return nil, fmt.Errorf("验证失败: %v", err)
	}

	// 标记验证码为已使用
	verificationCode.Used = true
	err = vs.db.Save(&verificationCode).Error
	if err != nil {
		return nil, fmt.Errorf("更新验证码状态失败: %v", err)
	}

	return &verificationCode, nil
}

// VerifyPhoneCode 验证手机验证码
func (vs *VerificationService) VerifyPhoneCode(phone, code, codeType string) (*model.UserVerificationCode, error) {
	var verificationCode model.UserVerificationCode
	err := vs.db.Where("phone = ? AND code = ? AND type = ? AND used = false AND expires_at > ?", 
		phone, code, codeType, time.Now()).First(&verificationCode).Error
	
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("验证码无效或已过期")
		}
		return nil, fmt.Errorf("验证失败: %v", err)
	}

	// 标记验证码为已使用
	verificationCode.Used = true
	err = vs.db.Save(&verificationCode).Error
	if err != nil {
		return nil, fmt.Errorf("更新验证码状态失败: %v", err)
	}

	return &verificationCode, nil
}

// CleanupExpiredCodes 清理过期的验证码
func (vs *VerificationService) CleanupExpiredCodes() error {
	return vs.db.Where("expires_at < ?", time.Now()).Delete(&model.UserVerificationCode{}).Error
}

// generateCode 生成指定长度的数字验证码
func (vs *VerificationService) generateCode(length int) (string, error) {
	const digits = "0123456789"
	code := make([]byte, length)
	
	for i := range code {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			return "", err
		}
		code[i] = digits[num.Int64()]
	}
	
	return string(code), nil
}

// sendEmail 发送邮件（简化实现，实际应该使用邮件服务）
func (vs *VerificationService) sendEmail(email, code, codeType string) error {
	// TODO: 集成真实的邮件服务（如阿里云邮件推送、SendGrid等）
	// 这里只是模拟发送
	
	var subject, content string
	switch codeType {
	case model.VerificationTypeRegister:
		subject = "注册验证码"
		content = fmt.Sprintf("您的注册验证码是：%s，有效期10分钟。", code)
	case model.VerificationTypeResetPassword:
		subject = "重置密码验证码"
		content = fmt.Sprintf("您的重置密码验证码是：%s，有效期10分钟。", code)
	case model.VerificationTypeChangeEmail:
		subject = "修改邮箱验证码"
		content = fmt.Sprintf("您的修改邮箱验证码是：%s，有效期10分钟。", code)
	default:
		subject = "验证码"
		content = fmt.Sprintf("您的验证码是：%s，有效期10分钟。", code)
	}
	
	// 模拟发送邮件
	fmt.Printf("发送邮件到 %s: %s - %s\n", email, subject, content)
	
	return nil
}

// sendSMS 发送短信（简化实现，实际应该使用短信服务）
func (vs *VerificationService) sendSMS(phone, code, codeType string) error {
	// TODO: 集成真实的短信服务（如阿里云短信、腾讯云短信等）
	// 这里只是模拟发送
	
	var content string
	switch codeType {
	case model.VerificationTypeRegister:
		content = fmt.Sprintf("【商城】您的注册验证码是：%s，有效期5分钟。", code)
	case model.VerificationTypeResetPassword:
		content = fmt.Sprintf("【商城】您的重置密码验证码是：%s，有效期5分钟。", code)
	case model.VerificationTypeChangePhone:
		content = fmt.Sprintf("【商城】您的修改手机号验证码是：%s，有效期5分钟。", code)
	default:
		content = fmt.Sprintf("【商城】您的验证码是：%s，有效期5分钟。", code)
	}
	
	// 模拟发送短信
	fmt.Printf("发送短信到 %s: %s\n", phone, content)
	
	return nil
}

// GetVerificationCodesByUser 获取用户的验证码记录
func (vs *VerificationService) GetVerificationCodesByUser(userID uint, limit int) ([]model.UserVerificationCode, error) {
	var codes []model.UserVerificationCode
	query := vs.db.Where("user_id = ?", userID).Order("created_at DESC")
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	err := query.Find(&codes).Error
	return codes, err
}

// 全局验证码服务实例
var globalVerificationService *VerificationService

// InitGlobalVerificationService 初始化全局验证码服务
func InitGlobalVerificationService(db *gorm.DB) {
	globalVerificationService = NewVerificationService(db)
}

// GetGlobalVerificationService 获取全局验证码服务
func GetGlobalVerificationService() *VerificationService {
	return globalVerificationService
}

// 便捷函数

// SendEmailVerificationGlobal 发送邮箱验证码（全局函数）
func SendEmailVerificationGlobal(email, codeType string, userID uint) (*model.UserVerificationCode, error) {
	if globalVerificationService == nil {
		return nil, fmt.Errorf("验证码服务未初始化")
	}
	return globalVerificationService.SendEmailVerification(email, codeType, userID)
}

// VerifyEmailCodeGlobal 验证邮箱验证码（全局函数）
func VerifyEmailCodeGlobal(email, code, codeType string) (*model.UserVerificationCode, error) {
	if globalVerificationService == nil {
		return nil, fmt.Errorf("验证码服务未初始化")
	}
	return globalVerificationService.VerifyEmailCode(email, code, codeType)
}
