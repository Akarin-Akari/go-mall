package service

import (
	"errors"
	"net/http"
)

// 地址管理相关错误定义
var (
	// 参数验证错误
	ErrInvalidUserID    = errors.New("用户ID无效")
	ErrInvalidAddressID = errors.New("地址ID无效")
	ErrInvalidRequest   = errors.New("请求参数无效")
	ErrInvalidContext   = errors.New("上下文参数无效")

	// 业务逻辑错误
	ErrAddressNotFound     = errors.New("地址不存在")
	ErrAddressLimitReached = errors.New("地址数量已达上限")
	ErrNoDefaultAddress    = errors.New("未找到默认地址")
	ErrPermissionDenied    = errors.New("地址不存在或无权限访问")

	// 数据验证错误
	ErrInvalidPhone       = errors.New("手机号格式不正确")
	ErrInvalidPostalCode  = errors.New("邮政编码格式不正确")
	ErrInvalidAddressType = errors.New("地址类型无效")
	ErrEmptyReceiverName  = errors.New("收货人姓名不能为空")
	ErrEmptyReceiverPhone = errors.New("收货人电话不能为空")
	ErrEmptyProvince      = errors.New("省份不能为空")
	ErrEmptyCity          = errors.New("城市不能为空")
	ErrEmptyDistrict      = errors.New("区县不能为空")
	ErrEmptyDetailAddress = errors.New("详细地址不能为空")

	// 系统错误
	ErrDatabaseOperation = errors.New("数据库操作失败")
	ErrTransactionFailed = errors.New("事务执行失败")
)

// ErrorCode 错误码类型
type ErrorCode int

// 错误码常量定义
const (
	// 成功
	CodeSuccess ErrorCode = 0

	// 参数错误 (1000-1999)
	CodeInvalidUserID    ErrorCode = 1001
	CodeInvalidAddressID ErrorCode = 1002
	CodeInvalidRequest   ErrorCode = 1003
	CodeInvalidContext   ErrorCode = 1004

	// 业务逻辑错误 (2000-2999)
	CodeAddressNotFound     ErrorCode = 2001
	CodeAddressLimitReached ErrorCode = 2002
	CodeNoDefaultAddress    ErrorCode = 2003
	CodePermissionDenied    ErrorCode = 2004

	// 数据验证错误 (3000-3999)
	CodeInvalidPhone       ErrorCode = 3001
	CodeInvalidPostalCode  ErrorCode = 3002
	CodeInvalidAddressType ErrorCode = 3003
	CodeEmptyReceiverName  ErrorCode = 3004
	CodeEmptyReceiverPhone ErrorCode = 3005
	CodeEmptyProvince      ErrorCode = 3006
	CodeEmptyCity          ErrorCode = 3007
	CodeEmptyDistrict      ErrorCode = 3008
	CodeEmptyDetailAddress ErrorCode = 3009

	// 系统错误 (5000-5999)
	CodeDatabaseOperation ErrorCode = 5001
	CodeTransactionFailed ErrorCode = 5002
)

// ServiceError 业务错误结构
type ServiceError struct {
	Code    ErrorCode
	Message string
	Err     error
}

// Error 实现error接口
func (e *ServiceError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

// Unwrap 支持errors.Unwrap
func (e *ServiceError) Unwrap() error {
	return e.Err
}

// NewServiceError 创建业务错误
func NewServiceError(code ErrorCode, message string, err error) *ServiceError {
	return &ServiceError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// 错误到HTTP状态码的映射
var errorToHTTPStatus = map[error]int{
	// 参数错误 -> 400 Bad Request
	ErrInvalidUserID:    http.StatusBadRequest,
	ErrInvalidAddressID: http.StatusBadRequest,
	ErrInvalidRequest:   http.StatusBadRequest,
	ErrInvalidContext:   http.StatusBadRequest,

	// 业务逻辑错误
	ErrAddressNotFound:     http.StatusNotFound,     // 404
	ErrAddressLimitReached: http.StatusBadRequest,   // 400
	ErrNoDefaultAddress:    http.StatusNotFound,     // 404
	ErrPermissionDenied:    http.StatusForbidden,    // 403

	// 数据验证错误 -> 400 Bad Request
	ErrInvalidPhone:       http.StatusBadRequest,
	ErrInvalidPostalCode:  http.StatusBadRequest,
	ErrInvalidAddressType: http.StatusBadRequest,
	ErrEmptyReceiverName:  http.StatusBadRequest,
	ErrEmptyReceiverPhone: http.StatusBadRequest,
	ErrEmptyProvince:      http.StatusBadRequest,
	ErrEmptyCity:          http.StatusBadRequest,
	ErrEmptyDistrict:      http.StatusBadRequest,
	ErrEmptyDetailAddress: http.StatusBadRequest,

	// 系统错误 -> 500 Internal Server Error
	ErrDatabaseOperation: http.StatusInternalServerError,
	ErrTransactionFailed: http.StatusInternalServerError,
}

// MapServiceErrorToHTTP 将Service错误映射为HTTP状态码和消息
func MapServiceErrorToHTTP(err error) (int, string) {
	if err == nil {
		return http.StatusOK, "成功"
	}

	// 检查是否是ServiceError
	var serviceErr *ServiceError
	if errors.As(err, &serviceErr) {
		if status, exists := errorToHTTPStatus[serviceErr.Unwrap()]; exists {
			return status, serviceErr.Message
		}
		// 如果没有映射，默认返回500
		return http.StatusInternalServerError, serviceErr.Message
	}

	// 检查是否是预定义的错误
	if status, exists := errorToHTTPStatus[err]; exists {
		return status, err.Error()
	}

	// 默认处理
	return http.StatusInternalServerError, "服务器内部错误"
}

// IsNotFoundError 判断是否是资源不存在错误
func IsNotFoundError(err error) bool {
	return errors.Is(err, ErrAddressNotFound) || errors.Is(err, ErrNoDefaultAddress)
}

// IsBadRequestError 判断是否是请求参数错误
func IsBadRequestError(err error) bool {
	badRequestErrors := []error{
		ErrInvalidUserID, ErrInvalidAddressID, ErrInvalidRequest,
		ErrInvalidPhone, ErrInvalidPostalCode, ErrInvalidAddressType,
		ErrEmptyReceiverName, ErrEmptyReceiverPhone, ErrEmptyProvince,
		ErrEmptyCity, ErrEmptyDistrict, ErrEmptyDetailAddress,
		ErrAddressLimitReached,
	}

	for _, badErr := range badRequestErrors {
		if errors.Is(err, badErr) {
			return true
		}
	}
	return false
}

// IsPermissionError 判断是否是权限错误
func IsPermissionError(err error) bool {
	return errors.Is(err, ErrPermissionDenied)
}

// IsSystemError 判断是否是系统错误
func IsSystemError(err error) bool {
	systemErrors := []error{
		ErrDatabaseOperation, ErrTransactionFailed,
	}

	for _, sysErr := range systemErrors {
		if errors.Is(err, sysErr) {
			return true
		}
	}
	return false
}
