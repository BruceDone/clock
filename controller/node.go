package controller

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"

	"clock/param"
	"clock/storage"
)

// 得到当前所有的关系图
func PutNodes(c echo.Context) (err error) {
	var nodes []storage.Node

	resp := param.ApiResponse{
		Code: 200,
		Msg:  "success",
		Data: nil,
	}

	if err := c.Bind(&nodes); err != nil {
		resp.Msg = fmt.Sprintf("[put nodes] invalidate param found: %v", err)
		logrus.Error(resp.Msg)

		return c.JSON(http.StatusBadRequest, resp)
	}

	if err := storage.PutNodes(nodes); err != nil {
		resp.Msg = fmt.Sprintf("[put nodes] query db error: %v", err)
		logrus.Error(resp.Msg)

		return c.JSON(http.StatusBadRequest, resp)
	}

	return c.JSON(http.StatusOK, resp)
}
