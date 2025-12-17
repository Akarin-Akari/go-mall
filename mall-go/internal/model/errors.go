package model

import "errors"

// Common errors for model package
var (
	ErrInvalidParam     = errors.New("无效的参数")
	ErrRecordNotFound   = errors.New("记录不存在")
	ErrDuplicateEntry   = errors.New("重复的条目")
	ErrInvalidOperation = errors.New("无效的操作")
	ErrPermissionDenied = errors.New("权限被拒绝")
	ErrValidationFailed = errors.New("验证失败")
)
