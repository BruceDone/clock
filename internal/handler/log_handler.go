package handler

import (
	"github.com/labstack/echo/v4"

	"clock/internal/logger"
	"clock/internal/repository"
	"clock/internal/service"
)

// LogHandler 日志处理器
type LogHandler struct {
	taskLogService service.TaskLogService
}

// NewLogHandler 创建日志处理器
func NewLogHandler(taskLogService service.TaskLogService) *LogHandler {
	return &LogHandler{
		taskLogService: taskLogService,
	}
}

// GetLogs 获取日志列表
func (h *LogHandler) GetLogs(c echo.Context) error {
	query := &repository.LogQuery{
		Page: repository.Page{
			Count:   getQueryIntDefault(c, "count", 10),
			Index:   getQueryIntDefault(c, "index", 1),
			LeftTs:  getQueryInt64Default(c, "left_ts", 0),
			RightTs: getQueryInt64Default(c, "right_ts", 0),
		},
		Tid: getQueryIntDefault(c, "tid", 0),
		Cid: getQueryIntDefault(c, "cid", 0),
	}

	result, err := h.taskLogService.List(query)
	if err != nil {
		logger.Errorf("[GetLogs] failed: %v", err)
		return HandleError(c, err)
	}

	return OK(c, result)
}

// DeleteLogs 删除日志
func (h *LogHandler) DeleteLogs(c echo.Context) error {
	query := &repository.LogQuery{
		Page: repository.Page{
			LeftTs:  getQueryInt64Default(c, "left_ts", 0),
			RightTs: getQueryInt64Default(c, "right_ts", 0),
		},
		Tid: getQueryIntDefault(c, "tid", 0),
		Cid: getQueryIntDefault(c, "cid", 0),
	}

	if err := h.taskLogService.Delete(query); err != nil {
		logger.Errorf("[DeleteLogs] failed: %v", err)
		return HandleError(c, err)
	}

	return OK(c, nil)
}

// DeleteLogByID 根据ID删除单条日志
func (h *LogHandler) DeleteLogByID(c echo.Context) error {
	lid := c.Param("lid")
	if lid == "" {
		return BadRequest(c, "lid is required")
	}

	if err := h.taskLogService.DeleteByID(lid); err != nil {
		logger.Errorf("[DeleteLogByID] failed: %v", err)
		return HandleError(c, err)
	}

	return OK(c, nil)
}

// DeleteAllLogs 删除所有日志
func (h *LogHandler) DeleteAllLogs(c echo.Context) error {
	if err := h.taskLogService.DeleteAll(); err != nil {
		logger.Errorf("[DeleteAllLogs] failed: %v", err)
		return HandleError(c, err)
	}

	return OK(c, nil)
}
