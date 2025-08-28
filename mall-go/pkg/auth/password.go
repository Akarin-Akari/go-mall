package auth

import (
	"errors"
	"regexp"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

// 密码复杂度要求
const (
	MinPasswordLength = 8
	MaxPasswordLength = 128
	DefaultCost       = 12 // bcrypt cost
)

// PasswordStrength 密码强度枚举
type PasswordStrength int

const (
	PasswordWeak PasswordStrength = iota
	PasswordMedium
	PasswordStrong
	PasswordVeryStrong
)

// HashPassword 加密密码
func HashPassword(password string) (string, error) {
	if password == "" {
		return "", errors.New("密码不能为空")
	}

	// 使用bcrypt加密密码
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hashedBytes), nil
}

// IsPasswordValid 验证密码是否正确
func IsPasswordValid(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// ValidatePasswordStrength 验证密码强度
func ValidatePasswordStrength(password string) error {
	if len(password) < MinPasswordLength {
		return errors.New("密码长度至少8位")
	}

	if len(password) > MaxPasswordLength {
		return errors.New("密码长度不能超过128位")
	}

	// 检查密码复杂度
	strength := GetPasswordStrength(password)
	if strength == PasswordWeak {
		return errors.New("密码强度太弱，请包含大小写字母、数字和特殊字符")
	}

	return nil
}

// GetPasswordStrength 获取密码强度
func GetPasswordStrength(password string) PasswordStrength {
	var (
		hasLower   = false
		hasUpper   = false
		hasNumber  = false
		hasSpecial = false
		length     = len(password)
	)

	// 检查字符类型
	for _, char := range password {
		switch {
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	// 计算强度分数
	score := 0
	if hasLower {
		score++
	}
	if hasUpper {
		score++
	}
	if hasNumber {
		score++
	}
	if hasSpecial {
		score++
	}

	// 长度加分
	if length >= 12 {
		score++
	}

	// 根据分数返回强度
	switch {
	case score <= 2:
		return PasswordWeak
	case score == 3:
		return PasswordMedium
	case score == 4:
		return PasswordStrong
	default:
		return PasswordVeryStrong
	}
}

// IsCommonPassword 检查是否为常见密码
func IsCommonPassword(password string) bool {
	// 常见弱密码列表
	commonPasswords := []string{
		"password", "123456", "123456789", "12345678", "12345",
		"1234567", "admin", "qwerty", "abc123", "password123",
		"admin123", "root", "user", "test", "guest",
	}

	for _, common := range commonPasswords {
		if password == common {
			return true
		}
	}

	return false
}

// ValidatePasswordPolicy 验证密码策略
func ValidatePasswordPolicy(password string) error {
	// 基本长度检查
	if len(password) < MinPasswordLength {
		return errors.New("密码长度至少8位")
	}

	if len(password) > MaxPasswordLength {
		return errors.New("密码长度不能超过128位")
	}

	// 检查是否为常见密码
	if IsCommonPassword(password) {
		return errors.New("不能使用常见密码")
	}

	// 检查是否包含空格
	if regexp.MustCompile(`\s`).MatchString(password) {
		return errors.New("密码不能包含空格")
	}

	// 检查字符类型要求
	var (
		hasLower   = regexp.MustCompile(`[a-z]`).MatchString(password)
		hasUpper   = regexp.MustCompile(`[A-Z]`).MatchString(password)
		hasNumber  = regexp.MustCompile(`[0-9]`).MatchString(password)
		hasSpecial = regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]`).MatchString(password)
	)

	if !hasLower {
		return errors.New("密码必须包含小写字母")
	}

	if !hasUpper {
		return errors.New("密码必须包含大写字母")
	}

	if !hasNumber {
		return errors.New("密码必须包含数字")
	}

	if !hasSpecial {
		return errors.New("密码必须包含特殊字符")
	}

	return nil
}

// GenerateRandomPassword 生成随机密码
func GenerateRandomPassword(length int) string {
	if length < MinPasswordLength {
		length = MinPasswordLength
	}

	// 字符集
	const (
		lowerChars   = "abcdefghijklmnopqrstuvwxyz"
		upperChars   = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		numberChars  = "0123456789"
		specialChars = "!@#$%^&*()_+-=[]{}|;:,.<>?"
	)

	// 确保包含各种字符类型
	password := ""
	password += string(lowerChars[0])   // 至少一个小写字母
	password += string(upperChars[0])   // 至少一个大写字母
	password += string(numberChars[0])  // 至少一个数字
	password += string(specialChars[0]) // 至少一个特殊字符

	// 填充剩余长度
	allChars := lowerChars + upperChars + numberChars + specialChars
	for i := 4; i < length; i++ {
		password += string(allChars[i%len(allChars)])
	}

	return password
}

// PasswordStrengthText 获取密码强度文本描述
func PasswordStrengthText(strength PasswordStrength) string {
	switch strength {
	case PasswordWeak:
		return "弱"
	case PasswordMedium:
		return "中等"
	case PasswordStrong:
		return "强"
	case PasswordVeryStrong:
		return "很强"
	default:
		return "未知"
	}
}
