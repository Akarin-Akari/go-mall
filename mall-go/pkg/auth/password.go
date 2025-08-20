package auth

import (
	"errors"
	"regexp"

	"golang.org/x/crypto/bcrypt"
)

// 密码相关常量
const (
	// bcrypt 加密成本，推荐值为 12
	DefaultCost = 12
	// 密码最小长度
	MinPasswordLength = 8
	// 密码最大长度
	MaxPasswordLength = 128
)

// 密码强度验证规则
var (
	// 至少包含一个小写字母
	hasLowerCase = regexp.MustCompile(`[a-z]`)
	// 至少包含一个大写字母
	hasUpperCase = regexp.MustCompile(`[A-Z]`)
	// 至少包含一个数字
	hasNumber = regexp.MustCompile(`\d`)
	// 至少包含一个特殊字符
	hasSpecialChar = regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?~` + "`" + `]`)
)

// PasswordError 密码相关错误
type PasswordError struct {
	Message string
}

func (e *PasswordError) Error() string {
	return e.Message
}

// 预定义错误
var (
	ErrPasswordTooShort    = &PasswordError{"密码长度不能少于8位"}
	ErrPasswordTooLong     = &PasswordError{"密码长度不能超过128位"}
	ErrPasswordTooWeak     = &PasswordError{"密码强度不够，需要包含大小写字母、数字和特殊字符"}
	ErrPasswordHashFailed  = &PasswordError{"密码加密失败"}
	ErrPasswordVerifyFailed = &PasswordError{"密码验证失败"}
)

// ValidatePasswordStrength 验证密码强度
// 要求：8-128位，包含大小写字母、数字和特殊字符
func ValidatePasswordStrength(password string) error {
	// 检查长度
	if len(password) < MinPasswordLength {
		return ErrPasswordTooShort
	}
	if len(password) > MaxPasswordLength {
		return ErrPasswordTooLong
	}

	// 检查复杂度
	checks := []struct {
		regex *regexp.Regexp
		desc  string
	}{
		{hasLowerCase, "小写字母"},
		{hasUpperCase, "大写字母"},
		{hasNumber, "数字"},
		{hasSpecialChar, "特殊字符"},
	}

	for _, check := range checks {
		if !check.regex.MatchString(password) {
			return ErrPasswordTooWeak
		}
	}

	return nil
}

// HashPassword 使用 bcrypt 加密密码
func HashPassword(password string) (string, error) {
	// 先验证密码强度
	if err := ValidatePasswordStrength(password); err != nil {
		return "", err
	}

	// 生成密码哈希
	hash, err := bcrypt.GenerateFromPassword([]byte(password), DefaultCost)
	if err != nil {
		return "", ErrPasswordHashFailed
	}

	return string(hash), nil
}

// VerifyPassword 验证密码是否正确
func VerifyPassword(hashedPassword, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return ErrPasswordVerifyFailed
		}
		return err
	}
	return nil
}

// IsPasswordValid 检查密码是否有效（不验证强度，仅用于登录）
func IsPasswordValid(hashedPassword, password string) bool {
	return VerifyPassword(hashedPassword, password) == nil
}

// TODO: 以下功能在MVP阶段暂不需要，遵循YAGNI原则
// 如果将来需要密码强度评分或随机密码生成，可以再添加

/*
// GenerateRandomPassword 生成随机密码（用于重置密码等场景）
// 当前MVP阶段不需要此功能
func GenerateRandomPassword(length int) (string, error) {
	// 实现留待将来需要时添加
	return "", errors.New("功能暂未实现")
}

// PasswordStrengthScore 计算密码强度分数 (0-100)
// 当前MVP阶段不需要此功能
func PasswordStrengthScore(password string) int {
	// 实现留待将来需要时添加
	return 0
}

// GetPasswordStrengthLevel 获取密码强度等级
// 当前MVP阶段不需要此功能
func GetPasswordStrengthLevel(password string) string {
	// 实现留待将来需要时添加
	return "未知"
}
*/
