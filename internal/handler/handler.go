package handler

import (
	"errors"
	"strconv"

	"github.com/labstack/echo/v4"
)

// getPathInt 从路径参数获取整数
func getPathInt(c echo.Context, key string) (int, error) {
	value := c.Param(key)
	if value == "" {
		return 0, errors.New("missing path parameter: " + key)
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		return 0, errors.New("invalid path parameter: " + key)
	}

	return intValue, nil
}

// getQueryInt 从查询参数获取整数
func getQueryInt(c echo.Context, key string) (int, error) {
	value := c.QueryParam(key)
	if value == "" {
		return 0, errors.New("missing query parameter: " + key)
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		return 0, errors.New("invalid query parameter: " + key)
	}

	return intValue, nil
}

// getQueryIntDefault 从查询参数获取整数（带默认值）
func getQueryIntDefault(c echo.Context, key string, defaultValue int) int {
	value := c.QueryParam(key)
	if value == "" {
		return defaultValue
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}

	return intValue
}

// getQueryInt64Default 从查询参数获取int64（带默认值）
func getQueryInt64Default(c echo.Context, key string, defaultValue int64) int64 {
	value := c.QueryParam(key)
	if value == "" {
		return defaultValue
	}

	intValue, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return defaultValue
	}

	return intValue
}
