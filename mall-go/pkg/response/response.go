package response

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Response 统一响应结构
type Response struct {
	Code      int         `json:"code"`                // 状态码
	Message   string      `json:"message"`             // 消息
	Data      interface{} `json:"data"`                // 数据
	TraceID   string      `json:"trace_id,omitempty"`  // 追踪ID
	Timestamp int64       `json:"timestamp,omitempty"` // 时间戳
}

// ErrorDetail 详细错误信息
type ErrorDetail struct {
	Field   string      `json:"field,omitempty"`   // 错误字段
	Value   interface{} `json:"value,omitempty"`   // 错误值
	Message string      `json:"message"`           // 错误消息
	Code    string      `json:"code,omitempty"`    // 错误代码
}

// ErrorResponse 详细错误响应结构
type ErrorResponse struct {
	Code      int           `json:"code"`                // 状态码
	Message   string        `json:"message"`             // 主要错误消息
	Details   []ErrorDetail `json:"details,omitempty"`   // 详细错误列表
	TraceID   string        `json:"trace_id"`            // 追踪ID
	Timestamp int64         `json:"timestamp"`           // 时间戳
	Path      string        `json:"path,omitempty"`      // 请求路径
	Method    string        `json:"method,omitempty"`    // 请求方法
}

// PageResult 分页响应结构
type PageResult struct {
	List     interface{} `json:"list"`      // 数据列表
	Total    int64       `json:"total"`     // 总数
	Page     int         `json:"page"`      // 当前页
	PageSize int         `json:"page_size"` // 每页大小
}

// 状态码常量
const (
	CodeSuccess      = 200  // 成功
	CodeError        = 500  // 服务器错误
	CodeInvalidParam = 400  // 参数错误
	CodeUnauthorized = 401  // 未授权
	CodeForbidden    = 403  // 禁止访问
	CodeNotFound     = 404  // 资源不存在
	CodeConflict     = 409  // 资源冲突
	CodeTooManyReq   = 429  // 请求过多
)

// 业务错误代码常量
const (
	// 用户相关错误
	ErrUserNotFound     = "USER_NOT_FOUND"
	ErrUserExists       = "USER_EXISTS"
	ErrInvalidPassword  = "INVALID_PASSWORD"
	ErrTokenExpired     = "TOKEN_EXPIRED"
	ErrTokenInvalid     = "TOKEN_INVALID"
	
	// 商品相关错误
	ErrProductNotFound  = "PRODUCT_NOT_FOUND"
	ErrInsufficientStock = "INSUFFICIENT_STOCK"
	ErrProductOffline   = "PRODUCT_OFFLINE"
	
	// 订单相关错误
	ErrOrderNotFound    = "ORDER_NOT_FOUND"
	ErrOrderStatusError = "ORDER_STATUS_ERROR"
	ErrPaymentFailed    = "PAYMENT_FAILED"
	
	// 购物车相关错误
	ErrCartItemNotFound = "CART_ITEM_NOT_FOUND"
	ErrCartEmpty        = "CART_EMPTY"
	
	// 通用错误
	ErrValidationFailed = "VALIDATION_FAILED"
	ErrDatabaseError    = "DATABASE_ERROR"
	ErrNetworkError     = "NETWORK_ERROR"
	ErrServiceUnavailable = "SERVICE_UNAVAILABLE"
)

// generateTraceID 生成追踪ID
func generateTraceID() string {
	return uuid.New().String()
}

// Success 成功响应
func Success(c *gin.Context, message string, data interface{}) {
	c.Header("Content-Type", "application/json; charset=utf-8")
	c.JSON(http.StatusOK, Response{
		Code:      CodeSuccess,
		Message:   message,
		Data:      data,
		TraceID:   generateTraceID(),
		Timestamp: time.Now().Unix(),
	})
}

// Error 错误响应
func Error(c *gin.Context, code int, message string) {
	c.Header("Content-Type", "application/json; charset=utf-8")
	c.JSON(code, Response{
		Code:      code,
		Message:   message,
		Data:      nil,
		TraceID:   generateTraceID(),
		Timestamp: time.Now().Unix(),
	})
}

// ErrorWithDetails 详细错误响应
func ErrorWithDetails(c *gin.Context, httpCode int, message string, details []ErrorDetail) {
	c.Header("Content-Type", "application/json; charset=utf-8")
	c.JSON(httpCode, ErrorResponse{
		Code:      httpCode,
		Message:   message,
		Details:   details,
		TraceID:   generateTraceID(),
		Timestamp: time.Now().Unix(),
		Path:      c.Request.URL.Path,
		Method:    c.Request.Method,
	})
}

// ValidationError 参数验证错误响应（支持多个字段错误）
func ValidationError(c *gin.Context, details []ErrorDetail) {
	ErrorWithDetails(c, http.StatusBadRequest, "参数验证失败", details)
}

// BusinessErrorWithCode 业务错误响应（带错误代码）
func BusinessErrorWithCode(c *gin.Context, message, errorCode string) {
	details := []ErrorDetail{
		{
			Message: message,
			Code:    errorCode,
		},
	}
	ErrorWithDetails(c, http.StatusBadRequest, message, details)
}

// ServerError 服务器错误响应
func ServerError(c *gin.Context, message string) {
	c.JSON(http.StatusInternalServerError, Response{
		Code:    CodeError,
		Message: message,
		Data:    nil,
	})
}

// BadRequest 参数错误响应
func BadRequest(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, Response{
		Code:    CodeInvalidParam,
		Message: message,
		Data:    nil,
	})
}

// Unauthorized 未授权响应
func Unauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, Response{
		Code:    CodeUnauthorized,
		Message: message,
		Data:    nil,
	})
}

// Forbidden 禁止访问响应
func Forbidden(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, Response{
		Code:    CodeForbidden,
		Message: message,
		Data:    nil,
	})
}

// NotFound 资源不存在响应
func NotFound(c *gin.Context, message string) {
	c.JSON(http.StatusNotFound, Response{
		Code:    CodeNotFound,
		Message: message,
		Data:    nil,
	})
}

// Conflict 资源冲突响应
func Conflict(c *gin.Context, message string) {
	c.JSON(http.StatusConflict, Response{
		Code:    CodeConflict,
		Message: message,
		Data:    nil,
	})
}

// TooManyRequests 请求过多响应
func TooManyRequests(c *gin.Context, message string) {
	c.JSON(http.StatusTooManyRequests, Response{
		Code:    CodeTooManyReq,
		Message: message,
		Data:    nil,
	})
}

// SuccessWithData 成功响应（带数据）
func SuccessWithData(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: "操作成功",
		Data:    data,
	})
}

// SuccessWithMessage 成功响应（仅消息）
func SuccessWithMessage(c *gin.Context, message string) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: message,
		Data:    nil,
	})
}

// SuccessWithPage 分页成功响应
func SuccessWithPage(c *gin.Context, message string, list interface{}, total int64, page, pageSize int) {
	pageResult := PageResult{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}

	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: message,
		Data:    pageResult,
	})
}

// ErrorWithCode 自定义状态码错误响应
func ErrorWithCode(c *gin.Context, httpCode, businessCode int, message string) {
	c.JSON(httpCode, Response{
		Code:    businessCode,
		Message: message,
		Data:    nil,
	})
}

// JSON 通用JSON响应
func JSON(c *gin.Context, httpCode, businessCode int, message string, data interface{}) {
	c.JSON(httpCode, Response{
		Code:    businessCode,
		Message: message,
		Data:    data,
	})
}

// Abort 中断请求并返回错误
func Abort(c *gin.Context, httpCode, businessCode int, message string) {
	c.JSON(httpCode, Response{
		Code:    businessCode,
		Message: message,
		Data:    nil,
	})
	c.Abort()
}

// AbortWithError 中断请求并返回服务器错误
func AbortWithError(c *gin.Context, message string) {
	c.JSON(http.StatusInternalServerError, Response{
		Code:    CodeError,
		Message: message,
		Data:    nil,
	})
	c.Abort()
}

// AbortWithUnauthorized 中断请求并返回未授权错误
func AbortWithUnauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, Response{
		Code:    CodeUnauthorized,
		Message: message,
		Data:    nil,
	})
	c.Abort()
}

// AbortWithForbidden 中断请求并返回禁止访问错误
func AbortWithForbidden(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, Response{
		Code:    CodeForbidden,
		Message: message,
		Data:    nil,
	})
	c.Abort()
}

// ValidateError 参数验证错误响应
func ValidateError(c *gin.Context, err error) {
	c.JSON(http.StatusBadRequest, Response{
		Code:    CodeInvalidParam,
		Message: "参数验证失败: " + err.Error(),
		Data:    nil,
	})
}

// DatabaseError 数据库错误响应
func DatabaseError(c *gin.Context, err error) {
	c.JSON(http.StatusInternalServerError, Response{
		Code:    CodeError,
		Message: "数据库操作失败: " + err.Error(),
		Data:    nil,
	})
}

// BusinessError 业务逻辑错误响应
func BusinessError(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, Response{
		Code:    CodeError,
		Message: message,
		Data:    nil,
	})
}

// GetResponseData 获取响应数据结构（用于测试）
func GetResponseData(code int, message string, data interface{}) Response {
	return Response{
		Code:    code,
		Message: message,
		Data:    data,
	}
}

// GetPageData 获取分页数据结构（用于测试）
func GetPageData(list interface{}, total int64, page, pageSize int) PageResult {
	return PageResult{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}
}
