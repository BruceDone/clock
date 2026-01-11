package handler

import (
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"

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
		logrus.Errorf("[GetLogs] failed: %v", err)
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
		logrus.Errorf("[DeleteLogs] failed: %v", err)
		return HandleError(c, err)
	}

	return OK(c, nil)
}
