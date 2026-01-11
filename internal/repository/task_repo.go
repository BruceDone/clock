package repository

import (
	"time"

	"gorm.io/gorm"

	"clock/internal/domain"
	apperrors "clock/internal/errors"
)

// taskRepository 任务仓储实现
type taskRepository struct {
	db *gorm.DB
}

// NewTaskRepository 创建任务仓储
func NewTaskRepository(db *gorm.DB) TaskRepository {
	return &taskRepository{db: db}
}

// GetByID 根据ID获取任务
func (r *taskRepository) GetByID(tid int) (*domain.Task, error) {
	var task domain.Task
	if err := r.db.Where("tid = ?", tid).First(&task).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.NotFound("task")
		}
		return nil, apperrors.Database(err)
	}
	return &task, nil
}

// List 查询任务列表
func (r *taskRepository) List(query *TaskQuery) ([]*domain.Task, error) {
	var tasks []*domain.Task

	// 设置默认值
	if query.Count < 1 {
		query.Count = 10
	}
	if query.Index < 1 {
		query.Index = 1
	}

	db := r.db.Model(&domain.Task{})

	// 条件过滤
	if query.Cid > 0 {
		db = db.Where("cid = ?", query.Cid)
	}
	if query.Name != "" {
		db = db.Where("name LIKE ?", "%"+query.Name+"%")
	}

	// 统计总数
	if err := db.Count(&query.Total).Error; err != nil {
		return nil, apperrors.Database(err)
	}

	// 分页和排序
	db = db.Offset((query.Index - 1) * query.Count).Limit(query.Count)
	if query.Order != "" {
		db = db.Order(query.Order)
	}

	if err := db.Find(&tasks).Error; err != nil {
		return nil, apperrors.Database(err)
	}

	return tasks, nil
}

// GetByCID 根据容器ID获取任务列表
func (r *taskRepository) GetByCID(cid int) ([]*domain.Task, error) {
	var tasks []*domain.Task
	if err := r.db.Where("cid = ?", cid).Find(&tasks).Error; err != nil {
		return nil, apperrors.Database(err)
	}
	return tasks, nil
}

// Save 保存任务
func (r *taskRepository) Save(task *domain.Task) error {
	task.UpdateAt = time.Now().Unix()
	if err := r.db.Save(task).Error; err != nil {
		return apperrors.Database(err)
	}
	return nil
}

// Delete 删除任务
func (r *taskRepository) Delete(tid int) error {
	if err := r.db.Where("tid = ?", tid).Delete(&domain.Task{}).Error; err != nil {
		return apperrors.Database(err)
	}
	return nil
}

// DeleteByCID 根据容器ID删除任务
func (r *taskRepository) DeleteByCID(cid int) error {
	if err := r.db.Where("cid = ?", cid).Delete(&domain.Task{}).Error; err != nil {
		return apperrors.Database(err)
	}
	return nil
}

// UpdateCoordinates 更新坐标
func (r *taskRepository) UpdateCoordinates(tid int, x, y int) error {
	if err := r.db.Model(&domain.Task{}).Where("tid = ?", tid).
		Updates(map[string]interface{}{
			"point_x":   x,
			"point_y":   y,
			"update_at": time.Now().Unix(),
		}).Error; err != nil {
		return apperrors.Database(err)
	}
	return nil
}
