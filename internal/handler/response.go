package handler

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	apperrors "clock/internal/errors"
)

// Response API响应结构
type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

// Success 成功响应
func Success(data interface{}) *Response {
	return &Response{
		Code: 200,
		Msg:  "success",
		Data: data,
	}
}

// Error 错误响应
func Error(msg string) *Response {
	return &Response{
		Code: 400,
		Msg:  msg,
		Data: nil,
	}
}

// ErrorWithCode 带错误码的错误响应
func ErrorWithCode(code int, msg string) *Response {
	return &Response{
		Code: code,
		Msg:  msg,
		Data: nil,
	}
}

// HandleError 统一错误处理
func HandleError(c echo.Context, err error) error {
	var appErr *apperrors.AppError
	if errors.As(err, &appErr) {
		switch appErr.Code {
		case apperrors.ErrNotFound:
			return c.JSON(http.StatusNotFound, ErrorWithCode(int(appErr.Code), appErr.Message))
		case apperrors.ErrValidation:
			return c.JSON(http.StatusBadRequest, ErrorWithCode(int(appErr.Code), appErr.Message))
		case apperrors.ErrUnauthorized:
			return c.JSON(http.StatusUnauthorized, ErrorWithCode(int(appErr.Code), appErr.Message))
		default:
			return c.JSON(http.StatusInternalServerError, ErrorWithCode(int(appErr.Code), appErr.Error()))
		}
	}
	return c.JSON(http.StatusInternalServerError, Error(err.Error()))
}

// OK 返回成功响应
func OK(c echo.Context, data interface{}) error {
	return c.JSON(http.StatusOK, Success(data))
}

// Created 返回创建成功响应
func Created(c echo.Context, data interface{}) error {
	return c.JSON(http.StatusCreated, Success(data))
}

// BadRequest 返回400错误
func BadRequest(c echo.Context, msg string) error {
	return c.JSON(http.StatusBadRequest, Error(msg))
}

// NotFound 返回404错误
func NotFound(c echo.Context, msg string) error {
	return c.JSON(http.StatusNotFound, Error(msg))
}

// Unauthorized 返回401错误
func Unauthorized(c echo.Context, msg string) error {
	return c.JSON(http.StatusUnauthorized, Error(msg))
}

// InternalError 返回500错误
func InternalError(c echo.Context, msg string) error {
	return c.JSON(http.StatusInternalServerError, Error(msg))
}
