package errors

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"mall-go/pkg/logger"

	"go.uber.org/zap"
)

// Language 语言类型
type Language string

const (
	LangEN   Language = "en"    // 英语
	LangZHCN Language = "zh-cn" // 简体中文
	LangZHTW Language = "zh-tw" // 繁体中文
	LangJA   Language = "ja"    // 日语
	LangKO   Language = "ko"    // 韩语
	LangES   Language = "es"    // 西班牙语
	LangFR   Language = "fr"    // 法语
	LangDE   Language = "de"    // 德语
)

// ErrorMessage 错误消息
type ErrorMessage struct {
	Code       ErrorCode `json:"code"`       // 错误码
	Message    string    `json:"message"`    // 错误消息
	Details    string    `json:"details"`    // 详细信息
	Suggestion string    `json:"suggestion"` // 解决建议
}

// I18nManager 国际化管理器
type I18nManager struct {
	mu           sync.RWMutex
	messages     map[Language]map[ErrorCode]*ErrorMessage
	defaultLang  Language
	fallbackLang Language
	loaded       bool
}

// NewI18nManager 创建国际化管理器
func NewI18nManager() *I18nManager {
	return &I18nManager{
		messages:     make(map[Language]map[ErrorCode]*ErrorMessage),
		defaultLang:  LangZHCN,
		fallbackLang: LangEN,
		loaded:       false,
	}
}

// SetDefaultLanguage 设置默认语言
func (i18n *I18nManager) SetDefaultLanguage(lang Language) {
	i18n.mu.Lock()
	defer i18n.mu.Unlock()
	i18n.defaultLang = lang
}

// SetFallbackLanguage 设置回退语言
func (i18n *I18nManager) SetFallbackLanguage(lang Language) {
	i18n.mu.Lock()
	defer i18n.mu.Unlock()
	i18n.fallbackLang = lang
}

// LoadMessages 加载错误消息
func (i18n *I18nManager) LoadMessages(lang Language, messages map[ErrorCode]*ErrorMessage) {
	i18n.mu.Lock()
	defer i18n.mu.Unlock()

	if i18n.messages[lang] == nil {
		i18n.messages[lang] = make(map[ErrorCode]*ErrorMessage)
	}

	for code, msg := range messages {
		i18n.messages[lang][code] = msg
	}

	logger.Info("加载错误消息",
		zap.String("language", string(lang)),
		zap.Int("count", len(messages)))
}

// LoadFromFile 从文件加载错误消息
func (i18n *I18nManager) LoadFromFile(lang Language, filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("读取错误消息文件失败: %v", err)
	}

	var messages map[ErrorCode]*ErrorMessage
	if err := json.Unmarshal(data, &messages); err != nil {
		return fmt.Errorf("解析错误消息文件失败: %v", err)
	}

	i18n.LoadMessages(lang, messages)
	return nil
}

// LoadFromDirectory 从目录加载所有语言的错误消息
func (i18n *I18nManager) LoadFromDirectory(dirPath string) error {
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("读取错误消息目录失败: %v", err)
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		// 从文件名提取语言代码
		fileName := strings.TrimSuffix(file.Name(), ".json")
		parts := strings.Split(fileName, ".")
		if len(parts) < 2 || parts[0] != "errors" {
			continue
		}

		lang := Language(parts[1])
		filePath := filepath.Join(dirPath, file.Name())

		if err := i18n.LoadFromFile(lang, filePath); err != nil {
			logger.Error("加载错误消息文件失败",
				zap.String("file", filePath),
				zap.Error(err))
			continue
		}
	}

	i18n.mu.Lock()
	i18n.loaded = true
	i18n.mu.Unlock()

	return nil
}

// GetMessage 获取错误消息
func (i18n *I18nManager) GetMessage(code ErrorCode, lang Language) *ErrorMessage {
	i18n.mu.RLock()
	defer i18n.mu.RUnlock()

	// 尝试获取指定语言的消息
	if langMessages, exists := i18n.messages[lang]; exists {
		if msg, found := langMessages[code]; found {
			return msg
		}
	}

	// 回退到默认语言
	if lang != i18n.defaultLang {
		if langMessages, exists := i18n.messages[i18n.defaultLang]; exists {
			if msg, found := langMessages[code]; found {
				return msg
			}
		}
	}

	// 回退到fallback语言
	if lang != i18n.fallbackLang && i18n.defaultLang != i18n.fallbackLang {
		if langMessages, exists := i18n.messages[i18n.fallbackLang]; exists {
			if msg, found := langMessages[code]; found {
				return msg
			}
		}
	}

	// 返回默认消息
	return &ErrorMessage{
		Code:       code,
		Message:    fmt.Sprintf("Error code: %s", code),
		Details:    "Error message not found",
		Suggestion: "Please contact support",
	}
}

// LocalizeError 本地化错误
func (i18n *I18nManager) LocalizeError(err *BusinessError, lang Language) *BusinessError {
	msg := i18n.GetMessage(err.Code, lang)

	// 创建错误副本
	localizedErr := *err
	localizedErr.Message = msg.Message
	if msg.Details != "" && localizedErr.Details == "" {
		localizedErr.Details = msg.Details
	}
	if msg.Suggestion != "" && localizedErr.Suggestion == "" {
		localizedErr.Suggestion = msg.Suggestion
	}

	return &localizedErr
}

// GetSupportedLanguages 获取支持的语言列表
func (i18n *I18nManager) GetSupportedLanguages() []Language {
	i18n.mu.RLock()
	defer i18n.mu.RUnlock()

	languages := make([]Language, 0, len(i18n.messages))
	for lang := range i18n.messages {
		languages = append(languages, lang)
	}

	return languages
}

// IsLoaded 检查是否已加载消息
func (i18n *I18nManager) IsLoaded() bool {
	i18n.mu.RLock()
	defer i18n.mu.RUnlock()
	return i18n.loaded
}

// GetStatistics 获取统计信息
func (i18n *I18nManager) GetStatistics() map[string]interface{} {
	i18n.mu.RLock()
	defer i18n.mu.RUnlock()

	stats := map[string]interface{}{
		"loaded":         i18n.loaded,
		"default_lang":   i18n.defaultLang,
		"fallback_lang":  i18n.fallbackLang,
		"languages":      make(map[Language]int),
		"total_messages": 0,
	}

	totalMessages := 0
	for lang, messages := range i18n.messages {
		count := len(messages)
		stats["languages"].(map[Language]int)[lang] = count
		totalMessages += count
	}
	stats["total_messages"] = totalMessages

	return stats
}

// 预定义的错误消息
var (
	// 中文错误消息
	ChineseMessages = map[ErrorCode]*ErrorMessage{
		// 系统错误
		ErrCodeSystemInternal: {
			Code:       ErrCodeSystemInternal,
			Message:    "系统内部错误",
			Details:    "服务器遇到内部错误，无法完成请求",
			Suggestion: "请稍后重试，如问题持续存在请联系技术支持",
		},
		ErrCodeSystemTimeout: {
			Code:       ErrCodeSystemTimeout,
			Message:    "系统响应超时",
			Details:    "系统处理请求时间过长",
			Suggestion: "请稍后重试",
		},
		ErrCodeSystemOverload: {
			Code:       ErrCodeSystemOverload,
			Message:    "系统负载过高",
			Details:    "当前系统负载较高，暂时无法处理请求",
			Suggestion: "请稍后重试",
		},

		// 业务错误
		ErrCodeBusinessLogic: {
			Code:       ErrCodeBusinessLogic,
			Message:    "业务逻辑错误",
			Details:    "请求不符合业务规则",
			Suggestion: "请检查输入参数是否正确",
		},

		// 验证错误
		ErrCodeValidationRequired: {
			Code:       ErrCodeValidationRequired,
			Message:    "必填字段缺失",
			Details:    "请求中缺少必填字段",
			Suggestion: "请填写所有必填字段",
		},
		ErrCodeValidationFormat: {
			Code:       ErrCodeValidationFormat,
			Message:    "字段格式错误",
			Details:    "输入字段格式不正确",
			Suggestion: "请检查输入格式是否正确",
		},

		// 认证错误
		ErrCodeAuthRequired: {
			Code:       ErrCodeAuthRequired,
			Message:    "需要登录",
			Details:    "此操作需要用户登录",
			Suggestion: "请先登录再进行操作",
		},
		ErrCodeAuthInvalidToken: {
			Code:       ErrCodeAuthInvalidToken,
			Message:    "登录令牌无效",
			Details:    "提供的登录令牌无效或已过期",
			Suggestion: "请重新登录",
		},

		// 权限错误
		ErrCodePermissionDenied: {
			Code:       ErrCodePermissionDenied,
			Message:    "权限不足",
			Details:    "您没有执行此操作的权限",
			Suggestion: "请联系管理员申请相应权限",
		},

		// 支付错误
		ErrCodePaymentInvalidMethod: {
			Code:       ErrCodePaymentInvalidMethod,
			Message:    "支付方式无效",
			Details:    "选择的支付方式不可用",
			Suggestion: "请选择其他支付方式",
		},
		ErrCodePaymentInsufficientFunds: {
			Code:       ErrCodePaymentInsufficientFunds,
			Message:    "余额不足",
			Details:    "账户余额不足以完成此次支付",
			Suggestion: "请充值后再试",
		},
	}

	// 英文错误消息
	EnglishMessages = map[ErrorCode]*ErrorMessage{
		// System errors
		ErrCodeSystemInternal: {
			Code:       ErrCodeSystemInternal,
			Message:    "Internal server error",
			Details:    "The server encountered an internal error and could not complete the request",
			Suggestion: "Please try again later. If the problem persists, contact technical support",
		},
		ErrCodeSystemTimeout: {
			Code:       ErrCodeSystemTimeout,
			Message:    "System timeout",
			Details:    "The system took too long to process the request",
			Suggestion: "Please try again later",
		},
		ErrCodeSystemOverload: {
			Code:       ErrCodeSystemOverload,
			Message:    "System overloaded",
			Details:    "The system is currently under high load and cannot process the request",
			Suggestion: "Please try again later",
		},

		// Business errors
		ErrCodeBusinessLogic: {
			Code:       ErrCodeBusinessLogic,
			Message:    "Business logic error",
			Details:    "The request does not comply with business rules",
			Suggestion: "Please check if the input parameters are correct",
		},

		// Validation errors
		ErrCodeValidationRequired: {
			Code:       ErrCodeValidationRequired,
			Message:    "Required field missing",
			Details:    "Required field is missing in the request",
			Suggestion: "Please fill in all required fields",
		},
		ErrCodeValidationFormat: {
			Code:       ErrCodeValidationFormat,
			Message:    "Field format error",
			Details:    "The input field format is incorrect",
			Suggestion: "Please check if the input format is correct",
		},

		// Authentication errors
		ErrCodeAuthRequired: {
			Code:       ErrCodeAuthRequired,
			Message:    "Authentication required",
			Details:    "This operation requires user authentication",
			Suggestion: "Please log in first",
		},
		ErrCodeAuthInvalidToken: {
			Code:       ErrCodeAuthInvalidToken,
			Message:    "Invalid authentication token",
			Details:    "The provided authentication token is invalid or expired",
			Suggestion: "Please log in again",
		},

		// Permission errors
		ErrCodePermissionDenied: {
			Code:       ErrCodePermissionDenied,
			Message:    "Permission denied",
			Details:    "You do not have permission to perform this operation",
			Suggestion: "Please contact the administrator to request appropriate permissions",
		},

		// Payment errors
		ErrCodePaymentInvalidMethod: {
			Code:       ErrCodePaymentInvalidMethod,
			Message:    "Invalid payment method",
			Details:    "The selected payment method is not available",
			Suggestion: "Please choose another payment method",
		},
		ErrCodePaymentInsufficientFunds: {
			Code:       ErrCodePaymentInsufficientFunds,
			Message:    "Insufficient funds",
			Details:    "Account balance is insufficient to complete this payment",
			Suggestion: "Please top up and try again",
		},
	}
)

// 全局国际化管理器实例
var GlobalI18nManager = NewI18nManager()

// init 初始化默认错误消息
func init() {
	GlobalI18nManager.LoadMessages(LangZHCN, ChineseMessages)
	GlobalI18nManager.LoadMessages(LangEN, EnglishMessages)
	GlobalI18nManager.SetDefaultLanguage(LangZHCN)
	GlobalI18nManager.SetFallbackLanguage(LangEN)

	logger.Info("错误消息国际化系统已初始化",
		zap.Int("zh_messages", len(ChineseMessages)),
		zap.Int("en_messages", len(EnglishMessages)))
}

// LocalizeBusinessError 本地化业务错误
func LocalizeBusinessError(err *BusinessError, lang Language) *BusinessError {
	return GlobalI18nManager.LocalizeError(err, lang)
}

// GetLocalizedMessage 获取本地化消息
func GetLocalizedMessage(code ErrorCode, lang Language) *ErrorMessage {
	return GlobalI18nManager.GetMessage(code, lang)
}
