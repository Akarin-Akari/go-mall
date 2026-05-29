package security

import (
	"regexp"
	"strings"
	"unicode"
)

// InputValidator 输入验证器
type InputValidator struct {
	// SQL注入检测模式
	sqlInjectionPattern *regexp.Regexp
	// XSS检测模式
	xssPattern *regexp.Regexp
}

// NewInputValidator 创建输入验证器
func NewInputValidator() *InputValidator {
	return &InputValidator{
		// SQL注入检测：检查常见的SQL关键字
		sqlInjectionPattern: regexp.MustCompile(`(?i)(union|select|insert|update|delete|drop|create|alter|exec|execute|script|javascript|onerror|onload|onclick)`),
		// XSS检测：检查HTML标签和事件处理器
		xssPattern: regexp.MustCompile(`(?i)(<script|<iframe|<object|<embed|javascript:|onerror=|onload=|onclick=|onmouseover=)`),
	}
}

// ValidateSQL 验证SQL注入
func (iv *InputValidator) ValidateSQL(input string) bool {
	if input == "" {
		return true
	}

	// 检查是否包含SQL注入模式
	if iv.sqlInjectionPattern.MatchString(input) {
		// 进一步检查是否是合法的输入
		// 例如，"select"可能出现在正常的文本中
		// 这里我们只检查明显的SQL注入模式
		return !containsSuspiciousSQLPattern(input)
	}

	return true
}

// ValidateXSS 验证XSS攻击
func (iv *InputValidator) ValidateXSS(input string) bool {
	if input == "" {
		return true
	}

	// 检查是否包含XSS模式
	return !iv.xssPattern.MatchString(input)
}

// SanitizeInput 清理输入
func (iv *InputValidator) SanitizeInput(input string) string {
	// 移除危险字符
	input = strings.TrimSpace(input)

	// 移除控制字符
	input = removeControlCharacters(input)

	// 转义HTML特殊字符
	input = escapeHTML(input)

	return input
}

// ValidateEmail 验证邮箱格式
func (iv *InputValidator) ValidateEmail(email string) bool {
	// 简单的邮箱验证
	emailPattern := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailPattern.MatchString(email)
}

// ValidateUsername 验证用户名
func (iv *InputValidator) ValidateUsername(username string) bool {
	if len(username) < 3 || len(username) > 32 {
		return false
	}

	// 用户名只能包含字母、数字、下划线和连字符
	usernamePattern := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	return usernamePattern.MatchString(username)
}

// ValidatePassword 验证密码强度
func (iv *InputValidator) ValidatePassword(password string) (bool, string) {
	if len(password) < 8 {
		return false, "密码长度至少8个字符"
	}

	if len(password) > 128 {
		return false, "密码长度不能超过128个字符"
	}

	// 检查是否包含大写字母
	hasUpper := false
	// 检查是否包含小写字母
	hasLower := false
	// 检查是否包含数字
	hasDigit := false
	// 检查是否包含特殊字符
	hasSpecial := false

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case !unicode.IsLetter(char) && !unicode.IsDigit(char):
			hasSpecial = true
		}
	}

	// 至少需要3种字符类型
	count := 0
	if hasUpper {
		count++
	}
	if hasLower {
		count++
	}
	if hasDigit {
		count++
	}
	if hasSpecial {
		count++
	}

	if count < 3 {
		return false, "密码必须包含大写字母、小写字母、数字和特殊字符中的至少3种"
	}

	return true, ""
}

// 辅助函数

// containsSuspiciousSQLPattern 检查是否包含可疑的SQL模式
func containsSuspiciousSQLPattern(input string) bool {
	suspiciousPatterns := []string{
		"' or '1'='1",
		"' or 1=1",
		"'; drop table",
		"union select",
		"exec(",
		"execute(",
	}

	lowerInput := strings.ToLower(input)
	for _, pattern := range suspiciousPatterns {
		if strings.Contains(lowerInput, pattern) {
			return true
		}
	}

	return false
}

// removeControlCharacters 移除控制字符
func removeControlCharacters(input string) string {
	return strings.Map(func(r rune) rune {
		if r < 32 && r != '\n' && r != '\r' && r != '\t' {
			return -1
		}
		return r
	}, input)
}

// escapeHTML 转义HTML特殊字符
func escapeHTML(input string) string {
	replacer := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		"\"", "&quot;",
		"'", "&#x27;",
	)
	return replacer.Replace(input)
}

// 全局输入验证器实例
var GlobalInputValidator *InputValidator

// InitInputValidator 初始化全局输入验证器
func InitInputValidator() {
	GlobalInputValidator = NewInputValidator()
}

// GetInputValidator 获取全局输入验证器
func GetInputValidator() *InputValidator {
	if GlobalInputValidator == nil {
		InitInputValidator()
	}
	return GlobalInputValidator
}

