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
func GetContainers(c echo.Context) (err error) {
	var query storage.ContainerQuery

	resp := param.ApiResponse{
		Code: 200,
		Msg:  "success",
		Data: nil,
	}

	if err := c.Bind(&query); err != nil {
		resp.Msg = fmt.Sprintf("[get containers] error to get the task param with: %v", err)
		logrus.Error(resp.Msg)

		return c.JSON(http.StatusBadRequest, resp)
	}

	containers, err := storage.GetContainers(&query)
	if err != nil {
		resp.Msg = fmt.Sprintf("[get containers] error to get the task from db with: %v", err)
		logrus.Error(resp.Msg)

		return c.JSON(http.StatusBadRequest, resp)
	}

	page := param.ListResponse{
		Items:     containers,
		PageQuery: query,
	}

	resp.Data = page
	return c.JSON(http.StatusOK, resp)
}

// 单个
func GetContainer(c echo.Context) (err error) {
	resp := param.ApiResponse{
		Code: 200,
		Msg:  "success",
		Data: nil,
	}

	taskId, err := getPathInt(c, "cid")

	if err != nil {
		resp.Msg = fmt.Sprintf("[get container] error to get the task tid with: %v", err)
		logrus.Error(resp.Msg)

		return c.JSON(http.StatusBadRequest, resp)
	}

	t, e := storage.GetContainer(taskId)
	if e != nil {
		resp.Msg = fmt.Sprintf("[get container] error to query the task with: %v", err)
		logrus.Error(resp.Msg)

		return c.JSON(http.StatusNotFound, resp)
	}

	resp.Data = t

	return c.JSON(http.StatusOK, resp)
}

// 删除 container
func DeleteContainer(c echo.Context) error {
	cid, err := getPathInt(c, "cid")

	resp := param.ApiResponse{
		Code: 200,
		Msg:  "success",
		Data: nil,
	}

	if err != nil {
		resp.Msg = fmt.Sprintf("[delete container] find empty tid %v", err)
		logrus.Error(resp.Msg)

		return c.JSON(http.StatusBadRequest, resp)
	}

	if err := storage.DeleteContainer(cid); err != nil {
		resp.Msg = fmt.Sprintf("[delete container] error to delete container with:%v", err)
		logrus.Error(err)

		return c.JSON(http.StatusBadRequest, resp)
	}

	return c.JSON(http.StatusOK, resp)
}

// 更新或新增一个task
func PutContainer(c echo.Context) error {
	resp := param.ApiResponse{
		Code: 200,
		Msg:  "success",
		Data: nil,
	}

	t := storage.Container{}

	if err := c.Bind(&t); err != nil {
		resp.Msg = fmt.Sprintf("[put container] invalidate param found: %v", err)
		logrus.Error(resp.Msg)

		return c.JSON(http.StatusBadRequest, resp)
	}

	if err := storage.PutContainer(t); err != nil {
		resp.Msg = fmt.Sprintf("[put container] error to query task from db with: %v", err)
		logrus.Error(resp.Msg)

		return c.JSON(http.StatusBadRequest, resp)
	}

	resp.Data = t.Cid

	return c.JSON(http.StatusOK, resp)
}

// 触发容器
func RunContainer(c echo.Context) error {
	resp := param.ApiResponse{
		Code: 200,
		Msg:  "success",
		Data: nil,
	}

	cid, err := getQueryInt(c, "cid")
	if err != nil {
		resp.Msg = fmt.Sprintf("[run container] error to get the container param with: %v", err)
		logrus.Error(resp.Msg)

		return c.JSON(http.StatusBadRequest, resp)
	}

	container, err := storage.GetContainer(cid)
	if err != nil {
		resp.Msg = fmt.Sprintf("[run container] error to get the container with cid %v and error: %v", cid, err)
		logrus.Error(resp.Msg)

		return c.JSON(http.StatusBadRequest, resp)
	}

	if err := storage.RunContainer(container); err != nil {
		resp.Msg = fmt.Sprintf("[run container] error run task with: %v", err)
		logrus.Error(resp.Msg)

		return c.JSON(http.StatusBadRequest, resp)
	}

	return c.JSON(http.StatusOK, resp)
}
