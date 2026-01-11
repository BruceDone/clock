package handler

import (
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"

	"clock/internal/domain"
	"clock/internal/repository"
	"clock/internal/service"
)

// ContainerHandler 容器处理器
type ContainerHandler struct {
	containerService service.ContainerService
}

// NewContainerHandler 创建容器处理器
func NewContainerHandler(containerService service.ContainerService) *ContainerHandler {
	return &ContainerHandler{
		containerService: containerService,
	}
}

// GetContainers 获取容器列表
func (h *ContainerHandler) GetContainers(c echo.Context) error {
	query := &repository.ContainerQuery{
		Page: repository.Page{
			Count: getQueryIntDefault(c, "count", 10),
			Index: getQueryIntDefault(c, "index", 1),
			Order: c.QueryParam("order"),
		},
		Name: c.QueryParam("name"),
	}

	result, err := h.containerService.List(query)
	if err != nil {
		logrus.Errorf("[GetContainers] failed: %v", err)
		return HandleError(c, err)
	}

	return OK(c, result)
}

// GetContainer 获取单个容器
func (h *ContainerHandler) GetContainer(c echo.Context) error {
	cid, err := getPathInt(c, "cid")
	if err != nil {
		return BadRequest(c, err.Error())
	}

	container, err := h.containerService.Get(cid)
	if err != nil {
		logrus.Errorf("[GetContainer] failed: %v", err)
		return HandleError(c, err)
	}

	return OK(c, container)
}

// PutContainer 创建或更新容器
func (h *ContainerHandler) PutContainer(c echo.Context) error {
	var container domain.Container
	if err := c.Bind(&container); err != nil {
		return BadRequest(c, "invalid request body")
	}

	if err := h.containerService.Save(&container); err != nil {
		logrus.Errorf("[PutContainer] failed: %v", err)
		return HandleError(c, err)
	}

	return OK(c, container.Cid)
}

// DeleteContainer 删除容器
func (h *ContainerHandler) DeleteContainer(c echo.Context) error {
	cid, err := getPathInt(c, "cid")
	if err != nil {
		return BadRequest(c, err.Error())
	}

	if err := h.containerService.Delete(cid); err != nil {
		logrus.Errorf("[DeleteContainer] failed: %v", err)
		return HandleError(c, err)
	}

	return OK(c, nil)
}

// RunContainer 执行容器
func (h *ContainerHandler) RunContainer(c echo.Context) error {
	cid, err := getQueryInt(c, "cid")
	if err != nil {
		return BadRequest(c, err.Error())
	}

	if err := h.containerService.Run(cid); err != nil {
		logrus.Errorf("[RunContainer] failed: %v", err)
		return HandleError(c, err)
	}

	return OK(c, nil)
}
