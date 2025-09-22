package errors

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// HTTPErrorResponse HTTP错误响应结构
type HTTPErrorResponse struct {
	Success   bool                   `json:"success"`              // 是否成功
	Code      string                 `json:"code"`                 // 错误码
	Message   string                 `json:"message"`              // 错误消息
	Details   string                 `json:"details,omitempty"`    // 详细信息
	Timestamp int64                  `json:"timestamp"`            // 时间戳
	RequestID string                 `json:"request_id,omitempty"` // 请求ID
	TraceID   string                 `json:"trace_id,omitempty"`   // 链路追踪ID
	Path      string                 `json:"path,omitempty"`       // 请求路径
	Method    string                 `json:"method,omitempty"`     // 请求方法
	Retryable bool                   `json:"retryable,omitempty"`  // 是否可重试
	Context   map[string]interface{} `json:"context,omitempty"`    // 错误上下文
}

// ErrorHandlerMiddleware 错误处理中间件
func ErrorHandlerMiddleware() gin.HandlerFunc {
	return gin.CustomRecoveryWithWriter(gin.DefaultWriter, func(c *gin.Context, recovered interface{}) {
		if err, ok := recovered.(error); ok {
			HandleError(c, err)
		} else {
			// 非错误类型的panic
			HandleError(c, NewSystemError(ErrCodeSystemInternal, "系统内部错误"))
		}
		c.Abort()
	})
}

// HandleError 统一错误处理
func HandleError(c *gin.Context, err error) {
	if err == nil {
		return
	}

	var businessErr *BusinessError
	var httpStatus int
	var errorCode string
	var message string
	var details string
	var retryable bool
	var context map[string]interface{}

	// 检查是否为BusinessError
	if be, ok := err.(*BusinessError); ok {
		businessErr = be
		httpStatus = getHTTPStatusFromErrorCode(be.Code)
		errorCode = string(be.Code)
		message = be.Message
		details = be.Details
		retryable = be.Retryable
		context = be.Context
	} else {
		// 普通错误转换为系统错误
		businessErr = WrapError(err, ErrCodeSystemInternal, "系统内部错误", CategorySystem)
		httpStatus = http.StatusInternalServerError
		errorCode = string(ErrCodeSystemInternal)
		message = "系统内部错误"
		details = err.Error()
		retryable = false
	}

	// 添加请求上下文信息
	if businessErr.RequestID == "" {
		if requestID := c.GetString("request_id"); requestID != "" {
			businessErr.RequestID = requestID
		}
	}

	if businessErr.TraceID == "" {
		if traceID := c.GetString("trace_id"); traceID != "" {
			businessErr.TraceID = traceID
		}
	}

	if businessErr.UserID == 0 {
		if userID, exists := c.Get("user_id"); exists {
			if uid, ok := userID.(uint); ok {
				businessErr.UserID = uid
			}
		}
	}

	// 构建HTTP响应
	response := HTTPErrorResponse{
		Success:   false,
		Code:      errorCode,
		Message:   message,
		Details:   details,
		Timestamp: businessErr.Timestamp.Unix(),
		RequestID: businessErr.RequestID,
		TraceID:   businessErr.TraceID,
		Path:      c.Request.URL.Path,
		Method:    c.Request.Method,
		Retryable: retryable,
		Context:   context,
	}

	// 记录错误日志
	LogError(c, businessErr)

	// 根据错误级别设置不同的响应头
	setResponseHeaders(c, businessErr)

	// 返回JSON响应
	c.JSON(httpStatus, response)
}

// getHTTPStatusFromErrorCode 根据错误码获取HTTP状态码
func getHTTPStatusFromErrorCode(code ErrorCode) int {
	codeStr := string(code)
	if len(codeStr) < 2 {
		return http.StatusInternalServerError
	}

	prefix := codeStr[:2]
	switch prefix {
	case "10": // 系统错误
		return http.StatusInternalServerError
	case "20": // 业务错误
		return http.StatusBadRequest
	case "30": // 验证错误
		return http.StatusBadRequest
	case "40": // 认证错误
		switch code {
		case ErrCodeAuthRequired:
			return http.StatusUnauthorized
		case ErrCodeAuthInvalidToken, ErrCodeAuthTokenExpired:
			return http.StatusUnauthorized
		case ErrCodeAuthInvalidCredentials:
			return http.StatusUnauthorized
		case ErrCodeAuthAccountLocked, ErrCodeAuthAccountDisabled:
			return http.StatusForbidden
		default:
			return http.StatusUnauthorized
		}
	case "50": // 权限错误
		return http.StatusForbidden
	case "60": // 数据库错误
		if code == ErrCodeDatabaseNotFound {
			return http.StatusNotFound
		}
		return http.StatusInternalServerError
	case "70": // 网络错误
		return http.StatusServiceUnavailable
	case "80": // 第三方服务错误
		return http.StatusServiceUnavailable
	case "90", "91", "92": // 支付错误
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

// setResponseHeaders 设置响应头
func setResponseHeaders(c *gin.Context, err *BusinessError) {
	// 设置错误级别头
	c.Header("X-Error-Level", string(err.Level))
	c.Header("X-Error-Category", string(err.Category))
	c.Header("X-Error-Code", string(err.Code))

	// 如果可重试，设置Retry-After头
	if err.Retryable && err.RetryAfter != nil {
		c.Header("Retry-After", strconv.Itoa(int(err.RetryAfter.Seconds())))
	}

	// 设置请求ID和链路追踪ID
	if err.RequestID != "" {
		c.Header("X-Request-ID", err.RequestID)
	}
	if err.TraceID != "" {
		c.Header("X-Trace-ID", err.TraceID)
	}

	// 根据错误级别设置缓存控制
	switch err.Level {
	case ErrorLevelCritical, ErrorLevelFatal:
		c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	case ErrorLevelError:
		c.Header("Cache-Control", "no-cache")
	}
}

// LogError 记录错误日志
func LogError(c *gin.Context, err *BusinessError) {
	// 构建日志上下文
	logContext := map[string]interface{}{
		"error_code":     err.Code,
		"error_message":  err.Message,
		"error_level":    err.Level,
		"error_category": err.Category,
		"request_id":     err.RequestID,
		"trace_id":       err.TraceID,
		"user_id":        err.UserID,
		"path":           c.Request.URL.Path,
		"method":         c.Request.Method,
		"user_agent":     c.Request.UserAgent(),
		"client_ip":      c.ClientIP(),
	}

	if err.Context != nil {
		for k, v := range err.Context {
			logContext["ctx_"+k] = v
		}
	}

	// 这里可以集成实际的日志库，如zap、logrus等
	logData, _ := json.Marshal(logContext)

	switch err.Level {
	case ErrorLevelFatal, ErrorLevelCritical:
		// 使用实际的日志库记录错误
		// logger.Error(string(logData))
		println("[CRITICAL ERROR]", string(logData))
	case ErrorLevelError:
		// logger.Error(string(logData))
		println("[ERROR]", string(logData))
	case ErrorLevelWarn:
		// logger.Warn(string(logData))
		println("[WARN]", string(logData))
	case ErrorLevelInfo:
		// logger.Info(string(logData))
		println("[INFO]", string(logData))
	}
}

// AbortWithError 中断请求并返回错误
func AbortWithError(c *gin.Context, err error) {
	HandleError(c, err)
	c.Abort()
}

// AbortWithBusinessError 中断请求并返回业务错误
func AbortWithBusinessError(c *gin.Context, code ErrorCode, message string) {
	err := GetBusinessErrorByCode(code)
	if err.Message == "" || err.Message == "未知错误" {
		err.Message = message
	}
	AbortWithError(c, err)
}

// AbortWithValidationError 中断请求并返回验证错误
func AbortWithValidationError(c *gin.Context, field, message string) {
	err := NewValidationError(ErrCodeValidationRequired, message).
		WithContext("field", field).
		WithSuggestion("请检查输入参数格式")
	AbortWithError(c, err)
}

// AbortWithAuthError 中断请求并返回认证错误
func AbortWithAuthError(c *gin.Context, message string) {
	err := NewAuthError(ErrCodeAuthRequired, message).
		WithSuggestion("请先登录")
	AbortWithError(c, err)
}

// AbortWithPermissionError 中断请求并返回权限错误
func AbortWithPermissionError(c *gin.Context, resource, action string) {
	err := NewPermissionError(ErrCodePermissionDenied, "权限不足").
		WithContext("resource", resource).
		WithContext("action", action).
		WithSuggestion("请联系管理员申请权限")
	AbortWithError(c, err)
}

// CheckError 检查错误并处理
func CheckError(c *gin.Context, err error) bool {
	if err != nil {
		HandleError(c, err)
		return true
	}
	return false
}

// MustNotError 确保没有错误，否则panic
func MustNotError(err error) {
	if err != nil {
		panic(err)
	}
}

// ValidateAndAbort 验证条件，失败则中断请求
func ValidateAndAbort(c *gin.Context, condition bool, code ErrorCode, message string) bool {
	if !condition {
		AbortWithBusinessError(c, code, message)
		return false
	}
	return true
}

// RequireAuth 要求认证，失败则中断请求
func RequireAuth(c *gin.Context) bool {
	if userID, exists := c.Get("user_id"); !exists || userID == nil {
		AbortWithAuthError(c, "请先登录")
		return false
	}
	return true
}

// RequirePermission 要求权限，失败则中断请求
func RequirePermission(c *gin.Context, permission string) bool {
	if !RequireAuth(c) {
		return false
	}

	// 这里应该集成实际的权限检查逻辑
	// 现在简化为检查是否有权限标识
	if perms, exists := c.Get("permissions"); exists {
		if permList, ok := perms.([]string); ok {
			for _, perm := range permList {
				if perm == permission {
					return true
				}
			}
		}
	}

	AbortWithPermissionError(c, permission, "access")
	return false
}
