package controller

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"

	"clock/param"
	"clock/storage"
)

// 列表
func GetLogs(c echo.Context) (err error) {
	var query storage.LogQuery

	resp := param.ApiResponse{
		Code: 200,
		Msg:  "success",
		Data: nil,
	}

	if err := c.Bind(&query); err != nil {
		resp.Msg = fmt.Sprintf("[get logs] error to get the query param with: %v", err)
		logrus.Error(resp.Msg)
		return c.JSON(http.StatusBadRequest, resp)
	}

	logs, err := storage.GetLogs(&query)
	if err != nil {
		resp.Msg = fmt.Sprintf("[get logs] error to get the logs: %v", err)
		logrus.Error(resp.Msg)
		return c.JSON(http.StatusBadRequest, resp)
	}

	page := param.ListResponse{
		Items:     logs,
		PageQuery: query,
	}

	resp.Data = page

	return c.JSON(http.StatusOK, resp)
}

// 清除多少天之前的日志
func DeleteLogs(c echo.Context) error {
	var query storage.LogQuery

	resp := param.ApiResponse{
		Code: 200,
		Msg:  "success",
		Data: nil,
	}

	if err := c.Bind(&query); err != nil {
		resp.Msg = fmt.Sprintf("[delete logs] error to get the query param with: %v", err)
		logrus.Error(resp.Msg)
		return c.JSON(http.StatusBadRequest, resp)
	}

	// 异步执行
	go storage.DeleteLogs(&query)
	return c.JSON(http.StatusOK, resp)
}
