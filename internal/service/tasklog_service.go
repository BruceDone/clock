package service

import (
	"clock/internal/domain"
	"clock/internal/repository"
)

// taskLogService 任务日志服务实现
type taskLogService struct {
	taskLogRepo repository.TaskLogRepository
}

// NewTaskLogService 创建任务日志服务
func NewTaskLogService(taskLogRepo repository.TaskLogRepository) TaskLogService {
	return &taskLogService{
		taskLogRepo: taskLogRepo,
	}
}

// List 查询日志列表
func (s *taskLogService) List(query *repository.LogQuery) (*ListResult[*domain.TaskLog], error) {
	logs, err := s.taskLogRepo.List(query)
	if err != nil {
		return nil, err
	}

	return &ListResult[*domain.TaskLog]{
		Items: logs,
		Page:  &query.Page,
	}, nil
}

// Delete 删除日志
func (s *taskLogService) Delete(query *repository.LogQuery) error {
	return s.taskLogRepo.DeleteByTimeRange(query)
}

// DeleteByID 根据日志ID删除单条日志
func (s *taskLogService) DeleteByID(lid string) error {
	return s.taskLogRepo.DeleteByID(lid)
}

// DeleteAll 删除所有日志
func (s *taskLogService) DeleteAll() error {
	return s.taskLogRepo.DeleteAll()
}
