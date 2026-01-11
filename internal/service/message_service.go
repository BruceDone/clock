package service

import (
	"clock/internal/domain"
	"clock/internal/repository"
)

// messageService 消息服务实现
type messageService struct {
	message       *domain.Message
	taskRepo      repository.TaskRepository
	containerRepo repository.ContainerRepository
}

// NewMessageService 创建消息服务
func NewMessageService(
	message *domain.Message,
	taskRepo repository.TaskRepository,
	containerRepo repository.ContainerRepository,
) MessageService {
	return &messageService{
		message:       message,
		taskRepo:      taskRepo,
		containerRepo: containerRepo,
	}
}

// Send 发送消息
func (s *messageService) Send(msg string) {
	s.message.Send(msg)
}

// Receive 获取消息通道
func (s *messageService) Receive() <-chan string {
	return s.message.Receive()
}

// GetCounters 获取任务统计
func (s *messageService) GetCounters() []domain.TaskCounter {
	// 获取所有任务
	tasks, _ := s.taskRepo.List(&repository.TaskQuery{
		Page: repository.Page{Count: 10000},
	})

	// 获取所有容器
	containers, _ := s.containerRepo.List(&repository.ContainerQuery{
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
		{Title: "Containers", Icon: "container", Count: len(containers), Color: "#409EFF"},
		{Title: "Tasks", Icon: "task", Count: len(tasks), Color: "#67C23A"},
		{Title: "Running", Icon: "running", Count: runningCount, Color: "#E6A23C"},
		{Title: "Success", Icon: "success", Count: successCount, Color: "#67C23A"},
		{Title: "Failure", Icon: "failure", Count: failureCount, Color: "#F56C6C"},
		{Title: "Pending", Icon: "pending", Count: pendingCount, Color: "#909399"},
	}
}
