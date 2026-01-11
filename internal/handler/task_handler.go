package handler

import (
	"github.com/labstack/echo/v4"

	"clock/internal/domain"
	"clock/internal/logger"
	"clock/internal/repository"
	"clock/internal/service"
)

// TaskHandler 任务处理器
type TaskHandler struct {
	taskService service.TaskService
}

// NewTaskHandler 创建任务处理器
func NewTaskHandler(taskService service.TaskService) *TaskHandler {
	return &TaskHandler{
		taskService: taskService,
	}
}

// GetTasks 获取任务列表
func (h *TaskHandler) GetTasks(c echo.Context) error {
	query := &repository.TaskQuery{
		Page: repository.Page{
			Count: getQueryIntDefault(c, "count", 10),
			Index: getQueryIntDefault(c, "index", 1),
			Order: c.QueryParam("order"),
		},
		Cid:  getQueryIntDefault(c, "cid", 0),
		Name: c.QueryParam("name"),
	}

	result, err := h.taskService.List(query)
	if err != nil {
		logger.Errorf("[GetTasks] failed: %v", err)
		return HandleError(c, err)
	}

	return OK(c, result)
}

// GetTask 获取单个任务
func (h *TaskHandler) GetTask(c echo.Context) error {
	tid, err := getPathInt(c, "tid")
	if err != nil {
		return BadRequest(c, err.Error())
	}

	task, err := h.taskService.Get(tid)
	if err != nil {
		logger.Errorf("[GetTask] failed: %v", err)
		return HandleError(c, err)
	}

	return OK(c, task)
}

// PutTask 创建或更新任务
func (h *TaskHandler) PutTask(c echo.Context) error {
	var task domain.Task
	if err := c.Bind(&task); err != nil {
		return BadRequest(c, "invalid request body")
	}

	if err := h.taskService.Save(&task); err != nil {
		logger.Errorf("[PutTask] failed: %v", err)
		return HandleError(c, err)
	}

	return OK(c, task.Tid)
}

// DeleteTask 删除任务
func (h *TaskHandler) DeleteTask(c echo.Context) error {
	tid, err := getPathInt(c, "tid")
	if err != nil {
		return BadRequest(c, err.Error())
	}

	if err := h.taskService.Delete(tid); err != nil {
		logger.Errorf("[DeleteTask] failed: %v", err)
		return HandleError(c, err)
	}

	return OK(c, nil)
}

// RunTask 执行任务
func (h *TaskHandler) RunTask(c echo.Context) error {
	tid, err := getQueryInt(c, "tid")
	if err != nil {
		return BadRequest(c, err.Error())
	}

	if err := h.taskService.Run(tid); err != nil {
		logger.Errorf("[RunTask] failed: %v", err)
		return HandleError(c, err)
	}

	return OK(c, nil)
}

// PutNodes 更新节点坐标
func (h *TaskHandler) PutNodes(c echo.Context) error {
	var nodes []domain.Node
	if err := c.Bind(&nodes); err != nil {
		return BadRequest(c, "invalid request body")
	}

	if err := h.taskService.UpdateNodes(nodes); err != nil {
		logger.Errorf("[PutNodes] failed: %v", err)
		return HandleError(c, err)
	}

	return OK(c, nil)
}

// CancelTask 取消单个任务
func (h *TaskHandler) CancelTask(c echo.Context) error {
	var req struct {
		Tid int `json:"tid"`
	}
	if err := c.Bind(&req); err != nil {
		return BadRequest(c, "invalid request body")
	}

	if req.Tid <= 0 {
		return BadRequest(c, "tid is required")
	}

	if err := h.taskService.CancelTask(req.Tid); err != nil {
		logger.Errorf("[CancelTask] failed: %v", err)
		return HandleError(c, err)
	}

	return OK(c, nil)
}

// CancelRun 取消整个 run
func (h *TaskHandler) CancelRun(c echo.Context) error {
	var req struct {
		RunID string `json:"runId"`
	}
	if err := c.Bind(&req); err != nil {
		return BadRequest(c, "invalid request body")
	}

	if req.RunID == "" {
		return BadRequest(c, "runId is required")
	}

	if err := h.taskService.CancelRun(req.RunID); err != nil {
		logger.Errorf("[CancelRun] failed: %v", err)
		return HandleError(c, err)
	}

	return OK(c, nil)
}

// GetRunningTasks 获取运行中的任务列表
func (h *TaskHandler) GetRunningTasks(c echo.Context) error {
	tasks := h.taskService.GetRunningTasks()
	return OK(c, tasks)
}
