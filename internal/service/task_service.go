package service

import (
	"clock/internal/domain"
	apperrors "clock/internal/errors"
	"clock/internal/repository"
)

// taskService 任务服务实现
type taskService struct {
	taskRepo     repository.TaskRepository
	relationRepo repository.RelationRepository
	executor     *Executor
}

// NewTaskService 创建任务服务
func NewTaskService(
	taskRepo repository.TaskRepository,
	relationRepo repository.RelationRepository,
	executor *Executor,
) TaskService {
	return &taskService{
		taskRepo:     taskRepo,
		relationRepo: relationRepo,
		executor:     executor,
	}
}

// Get 获取任务
func (s *taskService) Get(tid int) (*domain.Task, error) {
	return s.taskRepo.GetByID(tid)
}

// List 查询任务列表
func (s *taskService) List(query *repository.TaskQuery) (*ListResult[*domain.Task], error) {
	tasks, err := s.taskRepo.List(query)
	if err != nil {
		return nil, err
	}

	return &ListResult[*domain.Task]{
		Items: tasks,
		Page:  &query.Page,
	}, nil
}

// Save 保存任务
func (s *taskService) Save(task *domain.Task) error {
	return s.taskRepo.Save(task)
}

// Delete 删除任务（同时删除关联关系）
func (s *taskService) Delete(tid int) error {
	// 删除任务
	if err := s.taskRepo.Delete(tid); err != nil {
		return err
	}

	// 删除相关关系（作为前置任务）
	if err := s.relationRepo.DeleteByTID(tid); err != nil {
		return err
	}

	// 删除相关关系（作为后续任务）
	if err := s.relationRepo.DeleteByNextTID(tid); err != nil {
		return err
	}

	return nil
}

// Run 执行任务
func (s *taskService) Run(tid int) error {
	task, err := s.taskRepo.GetByID(tid)
	if err != nil {
		return err
	}

	return s.executor.RunTask(task)
}

// UpdateNodes 批量更新节点坐标
func (s *taskService) UpdateNodes(nodes []domain.Node) error {
	for _, node := range nodes {
		if _, err := s.taskRepo.GetByID(node.ID); err != nil {
			if apperrors.IsNotFound(err) {
				continue
			}
			return err
		}

		if err := s.taskRepo.UpdateCoordinates(node.ID, node.X, node.Y); err != nil {
			return err
		}
	}
	return nil
}

// CancelTask 取消单个任务
func (s *taskService) CancelTask(tid int) error {
	return s.executor.CancelTask(tid)
}

// CancelRun 取消整个 run
func (s *taskService) CancelRun(runID string) error {
	return s.executor.CancelRun(runID)
}

// GetRunningTasks 获取运行中的任务列表
func (s *taskService) GetRunningTasks() []RunningTaskInfo {
	return s.executor.GetRunningTasks()
}
