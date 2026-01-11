package handler

import (
	"github.com/labstack/echo/v4"

	"clock/internal/logger"
	"clock/internal/service"
)

// SystemHandler 系统监控处理器
type SystemHandler struct {
	systemService service.SystemService
}

// NewSystemHandler 创建系统监控处理器
func NewSystemHandler(systemService service.SystemService) *SystemHandler {
	return &SystemHandler{
		systemService: systemService,
	}
}

// GetLoadAverage 获取系统负载
func (h *SystemHandler) GetLoadAverage(c echo.Context) error {
	load, err := h.systemService.GetLoadAverage()
	if err != nil {
		logger.Errorf("[GetLoadAverage] failed: %v", err)
		return HandleError(c, err)
	}

	return OK(c, load)
}

// GetMemoryUsage 获取内存使用率
func (h *SystemHandler) GetMemoryUsage(c echo.Context) error {
	usage, err := h.systemService.GetMemoryUsage()
	if err != nil {
		logger.Errorf("[GetMemoryUsage] failed: %v", err)
		return HandleError(c, err)
	}

	return OK(c, usage)
}

// GetCPUUsage 获取CPU使用率
func (h *SystemHandler) GetCPUUsage(c echo.Context) error {
	usage, err := h.systemService.GetCPUUsage()
	if err != nil {
		logger.Errorf("[GetCPUUsage] failed: %v", err)
		return HandleError(c, err)
	}

	return OK(c, usage)
}
