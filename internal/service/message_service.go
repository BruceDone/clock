package service

import (
	"context"

	"clock/internal/domain"
	"clock/internal/repository"
)

// messageService 消息服务实现
type messageService struct {
	hub           *StreamHub
	taskRepo      repository.TaskRepository
	containerRepo repository.ContainerRepository
}

// NewMessageService 创建消息服务
func NewMessageService(
	hub *StreamHub,
	taskRepo repository.TaskRepository,
	containerRepo repository.ContainerRepository,
) MessageService {
	return &messageService{
		hub:           hub,
		taskRepo:      taskRepo,
		containerRepo: containerRepo,
	}
}

func (s *messageService) Publish(event StreamEvent) {
	s.hub.Publish(event)
}

func (s *messageService) Subscribe(ctx context.Context) <-chan StreamEvent {
	return s.hub.Subscribe(ctx)
}

// GetCounters 获取任务统计
func (s *messageService) GetCounters() []domain.TaskCounter {
	// 获取所有任务
	tasks, _ := s.taskRepo.List(&repository.TaskQuery{
		Page: repository.Page{Count: 10000},
	})

	// 统计各状态数量
	var pendingCount, runningCount, successCount, failureCount int

	for _, task := range tasks {
		switch task.Status {
		case domain.StatusPending:
			pendingCount++
		case domain.StatusStart:
			runningCount++
		case domain.StatusSuccess:
			successCount++
		case domain.StatusFailure:
			failureCount++
		}
	}

	return []domain.TaskCounter{
		{Title: "等待运行", Icon: "pending", Count: pendingCount, Color: "#909399"},
		{Title: "正在运行", Icon: "running", Count: runningCount, Color: "#00ccff"},
		{Title: "运行成功", Icon: "success", Count: successCount, Color: "#00ff88"},
		{Title: "运行失败", Icon: "failure", Count: failureCount, Color: "#ff4466"},
	}
}
