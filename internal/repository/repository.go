package repository

import (
	"clock/internal/domain"
)

// Page 分页参数
type Page struct {
	Count   int    `json:"count" query:"count"`       // 每页数量
	Index   int    `json:"index" query:"index"`       // 页码
	Total   int64  `json:"total" query:"total"`       // 总数
	Order   string `json:"order" query:"order"`       // 排序
	LeftTs  int64  `json:"left_ts" query:"left_ts"`   // 左时间戳
	RightTs int64  `json:"right_ts" query:"right_ts"` // 右时间戳
}

// TaskQuery 任务查询参数
type TaskQuery struct {
	Page
	Cid  int    `json:"cid"`
	Name string `json:"name"`
}

// ContainerQuery 容器查询参数
type ContainerQuery struct {
	Page
	Name string `json:"name"`
}

// LogQuery 日志查询参数
type LogQuery struct {
	Page
	Tid int `json:"tid"`
	Cid int `json:"cid"`
}

// TaskRepository 任务仓储接口
type TaskRepository interface {
	GetByID(tid int) (*domain.Task, error)
	List(query *TaskQuery) ([]*domain.Task, error)
	GetByCID(cid int) ([]*domain.Task, error)
	Save(task *domain.Task) error
	Delete(tid int) error
	DeleteByCID(cid int) error
	UpdateCoordinates(tid int, x, y int) error
}

// ContainerRepository 容器仓储接口
type ContainerRepository interface {
	GetByID(cid int) (*domain.Container, error)
	List(query *ContainerQuery) ([]*domain.Container, error)
	FindAll() ([]*domain.Container, error)
	Save(container *domain.Container) error
	Delete(cid int) error
}

// RelationRepository 关系仓储接口
type RelationRepository interface {
	GetByCID(cid int) ([]*domain.Relation, error)
	Save(relation *domain.Relation) error
	Delete(rid int) error
	DeleteByTID(tid int) error
	DeleteByNextTID(nextTid int) error
}

// TaskLogRepository 任务日志仓储接口
type TaskLogRepository interface {
	List(query *LogQuery) ([]*domain.TaskLog, error)
	Save(log *domain.TaskLog) error
	DeleteByTimeRange(query *LogQuery) error
}
