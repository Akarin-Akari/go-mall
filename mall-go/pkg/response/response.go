package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 统一响应结构
type Response struct {
	Code    int         `json:"code"`    // 状态码
	Message string      `json:"message"` // 消息
	Data    interface{} `json:"data"`    // 数据
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

// Success 成功响应
func Success(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: message,
		Data:    data,
	})
}

// Error 错误响应
func Error(c *gin.Context, code int, message string) {
	c.JSON(code, Response{
		Code:    code,
		Message: message,
		Data:    nil,
	})
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
