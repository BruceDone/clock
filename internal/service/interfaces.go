package service

import (
	"clock/internal/domain"
	"clock/internal/repository"
)

// ListResult 列表查询结果
type ListResult[T any] struct {
	Items interface{}      `json:"items"`
	Page  *repository.Page `json:"page"`
}

// TaskService 任务服务接口
type TaskService interface {
	Get(tid int) (*domain.Task, error)
	List(query *repository.TaskQuery) (*ListResult[*domain.Task], error)
	Save(task *domain.Task) error
	Delete(tid int) error
	Run(tid int) error
	UpdateNodes(nodes []domain.Node) error
}

// ContainerService 容器服务接口
type ContainerService interface {
	Get(cid int) (*domain.Container, error)
	List(query *repository.ContainerQuery) (*ListResult[*domain.Container], error)
	Save(container *domain.Container) error
	Delete(cid int) error
	Run(cid int) error
}

// RelationService 关系服务接口
type RelationService interface {
	GetGraph(cid int) (*domain.RelationGraph, error)
	Add(relation *domain.Relation) error
	Delete(rid int) error
	CheckCircle(tasks []*domain.Task, relations []*domain.Relation) bool
}

// TaskLogService 任务日志服务接口
type TaskLogService interface {
	List(query *repository.LogQuery) (*ListResult[*domain.TaskLog], error)
	Delete(query *repository.LogQuery) error
}

// SystemService 系统监控服务接口
type SystemService interface {
	GetLoadAverage() ([]float64, error)
	GetMemoryUsage() (float64, error)
	GetCPUUsage() (float64, error)
}

// SchedulerService 调度器服务接口
type SchedulerService interface {
	Start() error
	Stop()
	AddJob(container *domain.Container) error
	RemoveJob(entryID int)
}

// MessageService 消息服务接口
type MessageService interface {
	Send(msg string)
	Receive() <-chan string
	GetCounters() []domain.TaskCounter
}
