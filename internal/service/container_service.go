package service

import (
	"clock/internal/domain"
	"clock/internal/repository"
)

// containerService 容器服务实现
type containerService struct {
	containerRepo repository.ContainerRepository
	taskRepo      repository.TaskRepository
	relationRepo  repository.RelationRepository
	scheduler     SchedulerService
	executor      *Executor
}

// NewContainerService 创建容器服务
func NewContainerService(
	containerRepo repository.ContainerRepository,
	taskRepo repository.TaskRepository,
	relationRepo repository.RelationRepository,
	scheduler SchedulerService,
	executor *Executor,
) ContainerService {
	return &containerService{
		containerRepo: containerRepo,
		taskRepo:      taskRepo,
		relationRepo:  relationRepo,
		scheduler:     scheduler,
		executor:      executor,
	}
}

// Get 获取容器
func (s *containerService) Get(cid int) (*domain.Container, error) {
	return s.containerRepo.GetByID(cid)
}

// List 查询容器列表
func (s *containerService) List(query *repository.ContainerQuery) (*ListResult[*domain.Container], error) {
	containers, err := s.containerRepo.List(query)
	if err != nil {
		return nil, err
	}

	return &ListResult[*domain.Container]{
		Items: containers,
		Page:  &query.Page,
	}, nil
}

// Save 保存容器
func (s *containerService) Save(container *domain.Container) error {
	// 移除旧的调度任务
	if container.EntryID > 0 {
		s.scheduler.RemoveJob(container.EntryID)
	}

	// 如果未禁用，添加新的调度任务
	if !container.Disable {
		if err := s.scheduler.AddJob(container); err != nil {
			return err
		}
	} else {
		container.EntryID = -1
	}

	return s.containerRepo.Save(container)
}

// Delete 删除容器（同时删除关联任务）
func (s *containerService) Delete(cid int) error {
	container, err := s.containerRepo.GetByID(cid)
	if err != nil {
		return err
	}

	// 移除调度任务
	s.scheduler.RemoveJob(container.EntryID)

	// 删除容器
	if err := s.containerRepo.Delete(cid); err != nil {
		return err
	}

	// 删除关联任务
	if err := s.taskRepo.DeleteByCID(cid); err != nil {
		return err
	}

	return nil
}

// Run 执行容器内所有任务
func (s *containerService) Run(cid int) error {
	container, err := s.containerRepo.GetByID(cid)
	if err != nil {
		return err
	}

	tasks, err := s.taskRepo.GetByCID(cid)
	if err != nil {
		return err
	}

	relations, err := s.relationRepo.GetByCID(cid)
	if err != nil {
		return err
	}

	return s.executor.RunContainer(container, tasks, relations)
}
