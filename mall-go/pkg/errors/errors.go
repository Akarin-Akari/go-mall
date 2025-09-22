package errors

import (
	"encoding/json"
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// ErrorCode 错误码类型
type ErrorCode string

// ErrorLevel 错误级别
type ErrorLevel string

// ErrorCategory 错误分类
type ErrorCategory string

// 错误级别定义
const (
	ErrorLevelInfo     ErrorLevel = "INFO"     // 信息级别
	ErrorLevelWarn     ErrorLevel = "WARN"     // 警告级别
	ErrorLevelError    ErrorLevel = "ERROR"    // 错误级别
	ErrorLevelCritical ErrorLevel = "CRITICAL" // 严重错误级别
	ErrorLevelFatal    ErrorLevel = "FATAL"    // 致命错误级别
)

// 错误分类定义
const (
	CategorySystem     ErrorCategory = "SYSTEM"     // 系统错误
	CategoryBusiness   ErrorCategory = "BUSINESS"   // 业务错误
	CategoryValidation ErrorCategory = "VALIDATION" // 验证错误
	CategoryAuth       ErrorCategory = "AUTH"       // 认证错误
	CategoryPermission ErrorCategory = "PERMISSION" // 权限错误
	CategoryNetwork    ErrorCategory = "NETWORK"    // 网络错误
	CategoryDatabase   ErrorCategory = "DATABASE"   // 数据库错误
	CategoryThirdParty ErrorCategory = "THIRDPARTY" // 第三方服务错误
	CategoryPayment    ErrorCategory = "PAYMENT"    // 支付错误
	CategoryUpload     ErrorCategory = "UPLOAD"     // 上传错误
)

// 系统错误码定义
const (
	// 系统错误 (10000-19999)
	ErrCodeSystemInternal      ErrorCode = "10001" // 系统内部错误
	ErrCodeSystemTimeout       ErrorCode = "10002" // 系统超时
	ErrCodeSystemOverload      ErrorCode = "10003" // 系统过载
	ErrCodeSystemMaintenance   ErrorCode = "10004" // 系统维护
	ErrCodeSystemConfigError   ErrorCode = "10005" // 系统配置错误
	ErrCodeSystemResourceLimit ErrorCode = "10006" // 系统资源限制

	// 业务错误 (20000-29999)
	ErrCodeBusinessLogic       ErrorCode = "20001" // 业务逻辑错误
	ErrCodeBusinessRuleViolation ErrorCode = "20002" // 业务规则违反
	ErrCodeBusinessStateError  ErrorCode = "20003" // 业务状态错误
	ErrCodeBusinessQuotaExceed ErrorCode = "20004" // 业务配额超限
	ErrCodeBusinessOperation   ErrorCode = "20005" // 业务操作错误

	// 验证错误 (30000-39999)
	ErrCodeValidationRequired  ErrorCode = "30001" // 必填字段验证失败
	ErrCodeValidationFormat    ErrorCode = "30002" // 格式验证失败
	ErrCodeValidationLength    ErrorCode = "30003" // 长度验证失败
	ErrCodeValidationRange     ErrorCode = "30004" // 范围验证失败
	ErrCodeValidationUnique    ErrorCode = "30005" // 唯一性验证失败
	ErrCodeValidationRegex     ErrorCode = "30006" // 正则验证失败

	// 认证错误 (40000-49999)
	ErrCodeAuthRequired        ErrorCode = "40001" // 需要认证
	ErrCodeAuthInvalidToken    ErrorCode = "40002" // 无效的令牌
	ErrCodeAuthTokenExpired    ErrorCode = "40003" // 令牌已过期
	ErrCodeAuthInvalidCredentials ErrorCode = "40004" // 无效的凭据
	ErrCodeAuthAccountLocked   ErrorCode = "40005" // 账户被锁定
	ErrCodeAuthAccountDisabled ErrorCode = "40006" // 账户被禁用

	// 权限错误 (50000-59999)
	ErrCodePermissionDenied    ErrorCode = "50001" // 权限被拒绝
	ErrCodePermissionInsufficientRights ErrorCode = "50002" // 权限不足
	ErrCodePermissionResourceAccess ErrorCode = "50003" // 资源访问权限错误
	ErrCodePermissionOperationForbidden ErrorCode = "50004" // 操作被禁止

	// 数据库错误 (60000-69999)
	ErrCodeDatabaseConnection  ErrorCode = "60001" // 数据库连接错误
	ErrCodeDatabaseQuery       ErrorCode = "60002" // 数据库查询错误
	ErrCodeDatabaseTransaction ErrorCode = "60003" // 数据库事务错误
	ErrCodeDatabaseConstraint  ErrorCode = "60004" // 数据库约束错误
	ErrCodeDatabaseNotFound    ErrorCode = "60005" // 记录不存在
	ErrCodeDatabaseDuplicate   ErrorCode = "60006" // 重复记录

	// 网络错误 (70000-79999)
	ErrCodeNetworkTimeout      ErrorCode = "70001" // 网络超时
	ErrCodeNetworkConnection   ErrorCode = "70002" // 网络连接错误
	ErrCodeNetworkUnavailable  ErrorCode = "70003" // 网络不可用
	ErrCodeNetworkDNS          ErrorCode = "70004" // DNS解析错误

	// 第三方服务错误 (80000-89999)
	ErrCodeThirdPartyUnavailable ErrorCode = "80001" // 第三方服务不可用
	ErrCodeThirdPartyTimeout     ErrorCode = "80002" // 第三方服务超时
	ErrCodeThirdPartyInvalidResponse ErrorCode = "80003" // 第三方服务响应无效
	ErrCodeThirdPartyQuotaExceed ErrorCode = "80004" // 第三方服务配额超限

	// 支付错误 (90000-99999)
	ErrCodePaymentInvalidMethod   ErrorCode = "90001" // 无效的支付方式
	ErrCodePaymentInvalidAmount   ErrorCode = "90002" // 无效的支付金额
	ErrCodePaymentInsufficientFunds ErrorCode = "90003" // 余额不足
	ErrCodePaymentExpired         ErrorCode = "90004" // 支付已过期
	ErrCodePaymentAlreadyPaid     ErrorCode = "90005" // 支付已完成
	ErrCodePaymentRefundFailed    ErrorCode = "90006" // 退款失败
	ErrCodePaymentSignatureInvalid ErrorCode = "90007" // 支付签名无效
	ErrCodePaymentChannelUnavailable ErrorCode = "90008" // 支付渠道不可用
)

// BusinessError 业务错误结构
type BusinessError struct {
	Code        ErrorCode              `json:"code"`                  // 错误码
	Message     string                 `json:"message"`               // 错误消息
	Details     string                 `json:"details,omitempty"`     // 详细信息
	Level       ErrorLevel             `json:"level"`                 // 错误级别
	Category    ErrorCategory          `json:"category"`              // 错误分类
	Timestamp   time.Time              `json:"timestamp"`             // 错误时间
	RequestID   string                 `json:"request_id,omitempty"`  // 请求ID
	UserID      uint                   `json:"user_id,omitempty"`     // 用户ID
	TraceID     string                 `json:"trace_id,omitempty"`    // 链路追踪ID
	StackTrace  []StackFrame           `json:"stack_trace,omitempty"` // 调用栈
	Context     map[string]interface{} `json:"context,omitempty"`     // 错误上下文
	Cause       error                  `json:"-"`                     // 原始错误
	Suggestion  string                 `json:"suggestion,omitempty"`  // 解决建议
	Retryable   bool                   `json:"retryable"`             // 是否可重试
	RetryAfter  *time.Duration         `json:"retry_after,omitempty"` // 重试间隔
}

// StackFrame 调用栈帧
type StackFrame struct {
	Function string `json:"function"` // 函数名
	File     string `json:"file"`     // 文件名
	Line     int    `json:"line"`     // 行号
}

// Error 实现error接口
func (e *BusinessError) Error() string {
	return fmt.Sprintf("[%s] %s: %s", e.Code, e.Category, e.Message)
}

// Is 实现errors.Is接口
func (e *BusinessError) Is(target error) bool {
	if be, ok := target.(*BusinessError); ok {
		return e.Code == be.Code
	}
	return false
}

// Unwrap 实现errors.Unwrap接口
func (e *BusinessError) Unwrap() error {
	return e.Cause
}

// WithContext 添加上下文信息
func (e *BusinessError) WithContext(key string, value interface{}) *BusinessError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// WithCause 添加原始错误
func (e *BusinessError) WithCause(cause error) *BusinessError {
	e.Cause = cause
	return e
}

// WithDetails 添加详细信息
func (e *BusinessError) WithDetails(details string) *BusinessError {
	e.Details = details
	return e
}

// WithSuggestion 添加解决建议
func (e *BusinessError) WithSuggestion(suggestion string) *BusinessError {
	e.Suggestion = suggestion
	return e
}

// WithRetry 设置重试信息
func (e *BusinessError) WithRetry(retryable bool, retryAfter *time.Duration) *BusinessError {
	e.Retryable = retryable
	e.RetryAfter = retryAfter
	return e
}

// ToJSON 转换为JSON
func (e *BusinessError) ToJSON() string {
	data, _ := json.Marshal(e)
	return string(data)
}

// NewError 创建新的业务错误
func NewError(code ErrorCode, message string, level ErrorLevel, category ErrorCategory) *BusinessError {
	return &BusinessError{
		Code:       code,
		Message:    message,
		Level:      level,
		Category:   category,
		Timestamp:  time.Now(),
		StackTrace: captureStackTrace(2), // 跳过当前函数和调用方
		Context:    make(map[string]interface{}),
		Retryable:  false,
	}
}

// NewSystemError 创建系统错误
func NewSystemError(code ErrorCode, message string) *BusinessError {
	return NewError(code, message, ErrorLevelError, CategorySystem)
}

// NewBusinessError 创建业务错误
func NewBusinessError(code ErrorCode, message string) *BusinessError {
	return NewError(code, message, ErrorLevelWarn, CategoryBusiness)
}

// NewValidationError 创建验证错误
func NewValidationError(code ErrorCode, message string) *BusinessError {
	return NewError(code, message, ErrorLevelWarn, CategoryValidation)
}

// NewAuthError 创建认证错误
func NewAuthError(code ErrorCode, message string) *BusinessError {
	return NewError(code, message, ErrorLevelWarn, CategoryAuth)
}

// NewPermissionError 创建权限错误
func NewPermissionError(code ErrorCode, message string) *BusinessError {
	return NewError(code, message, ErrorLevelWarn, CategoryPermission)
}

// NewDatabaseError 创建数据库错误
func NewDatabaseError(code ErrorCode, message string) *BusinessError {
	return NewError(code, message, ErrorLevelError, CategoryDatabase)
}

// NewNetworkError 创建网络错误
func NewNetworkError(code ErrorCode, message string) *BusinessError {
	return NewError(code, message, ErrorLevelError, CategoryNetwork)
}

// NewThirdPartyError 创建第三方错误
func NewThirdPartyError(code ErrorCode, message string) *BusinessError {
	return NewError(code, message, ErrorLevelError, CategoryThirdParty)
}

// NewPaymentError 创建支付错误
func NewPaymentError(code ErrorCode, message string) *BusinessError {
	return NewError(code, message, ErrorLevelError, CategoryPayment)
}

// captureStackTrace 捕获调用栈
func captureStackTrace(skip int) []StackFrame {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(skip+1, pcs[:])
	frames := runtime.CallersFrames(pcs[:n])
	
	var stackTrace []StackFrame
	for {
		frame, more := frames.Next()
		stackTrace = append(stackTrace, StackFrame{
			Function: frame.Function,
			File:     frame.File,
			Line:     frame.Line,
		})
		if !more {
			break
		}
		// 限制栈帧数量
		if len(stackTrace) >= 10 {
			break
		}
	}
	
	return stackTrace
}

// WrapError 包装标准错误为业务错误
func WrapError(err error, code ErrorCode, message string, category ErrorCategory) *BusinessError {
	if err == nil {
		return nil
	}
	
	// 如果已经是BusinessError，直接返回
	if be, ok := err.(*BusinessError); ok {
		return be
	}
	
	return NewError(code, message, ErrorLevelError, category).WithCause(err).WithDetails(err.Error())
}

// IsRetryable 检查错误是否可重试
func IsRetryable(err error) bool {
	if be, ok := err.(*BusinessError); ok {
		return be.Retryable
	}
	return false
}

// GetErrorCode 获取错误码
func GetErrorCode(err error) ErrorCode {
	if be, ok := err.(*BusinessError); ok {
		return be.Code
	}
	return ErrCodeSystemInternal
}

// GetErrorLevel 获取错误级别
func GetErrorLevel(err error) ErrorLevel {
	if be, ok := err.(*BusinessError); ok {
		return be.Level
	}
	return ErrorLevelError
}

// GetErrorCategory 获取错误分类
func GetErrorCategory(err error) ErrorCategory {
	if be, ok := err.(*BusinessError); ok {
		return be.Category
	}
	return CategorySystem
}

// FormatErrorCode 格式化错误码显示
func FormatErrorCode(code ErrorCode) string {
	codeStr := string(code)
	if len(codeStr) != 5 {
		return codeStr
	}
	
	// 根据错误码前缀确定分类
	prefix := codeStr[:2]
	var category string
	switch prefix {
	case "10":
		category = "SYSTEM"
	case "20":
		category = "BUSINESS"
	case "30":
		category = "VALIDATION"
	case "40":
		category = "AUTH"
	case "50":
		category = "PERMISSION"
	case "60":
		category = "DATABASE"
	case "70":
		category = "NETWORK"
	case "80":
		category = "THIRDPARTY"
	case "90":
		category = "PAYMENT"
	default:
		category = "UNKNOWN"
	}
	
	return fmt.Sprintf("%s-%s", category, codeStr)
}

// ParseErrorCode 解析错误码字符串
func ParseErrorCode(codeStr string) (ErrorCode, bool) {
	// 支持 "SYSTEM-10001" 或 "10001" 格式
	parts := strings.Split(codeStr, "-")
	if len(parts) == 2 {
		codeStr = parts[1]
	}
	
	// 验证错误码格式
	if len(codeStr) == 5 {
		if _, err := strconv.Atoi(codeStr); err == nil {
			return ErrorCode(codeStr), true
		}
	}
	
	return "", false
}