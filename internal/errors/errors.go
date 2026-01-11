package errors

import (
	"errors"
	"fmt"
)

// ErrorCode 错误码
type ErrorCode int

const (
	// ErrNotFound 资源未找到
	ErrNotFound ErrorCode = iota + 1000
	// ErrValidation 验证错误
	ErrValidation
	// ErrDatabase 数据库错误
	ErrDatabase
	// ErrScheduler 调度器错误
	ErrScheduler
	// ErrCircularDependency DAG循环依赖
	ErrCircularDependency
	// ErrUnauthorized 未授权
	ErrUnauthorized
	// ErrInternal 内部错误
	ErrInternal
)

// AppError 应用错误
type AppError struct {
	Code    ErrorCode
	Message string
	Err     error
}

// Error 实现error接口
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%d] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// Unwrap 返回包装的错误
func (e *AppError) Unwrap() error {
	return e.Err
}

// Is 判断错误类型
func (e *AppError) Is(target error) bool {
	t, ok := target.(*AppError)
	if !ok {
		return false
	}
	return e.Code == t.Code
}

// NotFound 创建未找到错误
func NotFound(resource string) *AppError {
	return &AppError{
		Code:    ErrNotFound,
		Message: fmt.Sprintf("%s not found", resource),
	}
}

// Validation 创建验证错误
func Validation(msg string) *AppError {
	return &AppError{
		Code:    ErrValidation,
		Message: msg,
	}
}

// Database 创建数据库错误
func Database(err error) *AppError {
	return &AppError{
		Code:    ErrDatabase,
		Message: "database error",
		Err:     err,
	}
}

// Scheduler 创建调度器错误
func Scheduler(err error) *AppError {
	return &AppError{
		Code:    ErrScheduler,
		Message: "scheduler error",
		Err:     err,
	}
}

// CircularDependency 创建循环依赖错误
func CircularDependency() *AppError {
	return &AppError{
		Code:    ErrCircularDependency,
		Message: "circular dependency detected in DAG",
	}
}

// Unauthorized 创建未授权错误
func Unauthorized(msg string) *AppError {
	return &AppError{
		Code:    ErrUnauthorized,
		Message: msg,
	}
}

// Internal 创建内部错误
func Internal(err error) *AppError {
	return &AppError{
		Code:    ErrInternal,
		Message: "internal error",
		Err:     err,
	}
}

// Wrap 包装错误
func Wrap(err error, msg string) *AppError {
	return &AppError{
		Code:    ErrInternal,
		Message: msg,
		Err:     err,
	}
}

// IsNotFound 判断是否为未找到错误
func IsNotFound(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code == ErrNotFound
	}
	return false
}

// IsValidation 判断是否为验证错误
func IsValidation(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code == ErrValidation
	}
	return false
}
