package service

import (
	"github.com/robfig/cron/v3"

	"clock/internal/domain"
	apperrors "clock/internal/errors"
	"clock/internal/logger"
	"clock/internal/repository"
)

// schedulerService 调度器服务实现
type schedulerService struct {
	cron          *cron.Cron
	containerRepo repository.ContainerRepository
	taskRepo      repository.TaskRepository
	relationRepo  repository.RelationRepository
	executor      *Executor
}

// NewSchedulerService 创建调度器服务
func NewSchedulerService(
	containerRepo repository.ContainerRepository,
	taskRepo repository.TaskRepository,
	relationRepo repository.RelationRepository,
	executor *Executor,
) SchedulerService {
	cronLogger := cron.WithLogger(logger.NewCronLogger())

	return &schedulerService{
		cron:          cron.New(cronLogger),
		containerRepo: containerRepo,
		taskRepo:      taskRepo,
		relationRepo:  relationRepo,
		executor:      executor,
	}
}

// Start 启动调度器
func (s *schedulerService) Start() error {
	// 获取所有容器
	containers, err := s.containerRepo.FindAll()
	if err != nil {
		return err
	}

	// 初始化所有容器状态
	for _, container := range containers {
		container.Status = domain.StatusPending
		container.EntryID = -1

		if !container.Disable {
			if err := s.AddJob(container); err != nil {
				logger.Errorf("[scheduler] failed to add job for container %s: %v", container.Name, err)
			}
		}

		if err := s.containerRepo.Save(container); err != nil {
			logger.Errorf("[scheduler] failed to save container %s: %v", container.Name, err)
		}
	}

	logger.Info("[scheduler] starting cron scheduler")
	s.cron.Start()

	return nil
}

// Stop 停止调度器
func (s *schedulerService) Stop() {
	s.cron.Stop()
}

// AddJob 添加调度任务
func (s *schedulerService) AddJob(container *domain.Container) error {
	runFunc := func() {
		tasks, err := s.taskRepo.GetByCID(container.Cid)
		if err != nil {
			logger.Errorf("[scheduler] failed to get tasks for container %d: %v", container.Cid, err)
			return
		}

		relations, err := s.relationRepo.GetByCID(container.Cid)
		if err != nil {
			logger.Errorf("[scheduler] failed to get relations for container %d: %v", container.Cid, err)
			return
		}

		if err := s.executor.RunContainer(container, tasks, relations); err != nil {
			logger.Errorf("[scheduler] failed to run container %s: %v", container.Name, err)
		}
	}

	entryID, err := s.cron.AddFunc(container.Expression, runFunc)
	if err != nil {
		return apperrors.Scheduler(err)
	}

	container.EntryID = int(entryID)
	logger.Infof("[scheduler] added job for container %s with entry id %d", container.Name, container.EntryID)

	return nil
}

// RemoveJob 移除调度任务
func (s *schedulerService) RemoveJob(entryID int) {
	if entryID > 0 {
		s.cron.Remove(cron.EntryID(entryID))
	}
}
