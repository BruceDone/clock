package repository

import (
	"gorm.io/gorm"

	"clock/internal/domain"
	apperrors "clock/internal/errors"
)

// taskLogRepository 任务日志仓储实现
type taskLogRepository struct {
	db *gorm.DB
}

// NewTaskLogRepository 创建任务日志仓储
func NewTaskLogRepository(db *gorm.DB) TaskLogRepository {
	return &taskLogRepository{db: db}
}

// List 查询日志列表
func (r *taskLogRepository) List(query *LogQuery) ([]*domain.TaskLog, error) {
	var logs []*domain.TaskLog

	// 设置默认值
	if query.Count < 1 {
		query.Count = 10
	}
	if query.Index < 1 {
		query.Index = 1
	}

	db := r.db.Model(&domain.TaskLog{})

	// 条件过滤
	if query.Tid > 0 {
		db = db.Where("tid = ?", query.Tid)
	}
	if query.Cid > 0 {
		db = db.Where("cid = ?", query.Cid)
	}
	if query.LeftTs > 0 {
		db = db.Where("update_at > ?", query.LeftTs)
	}
	if query.RightTs > 0 {
		db = db.Where("update_at < ?", query.RightTs)
	}

	// 统计总数
	if err := db.Count(&query.Total).Error; err != nil {
		return nil, apperrors.Database(err)
	}

	// 分页和排序（默认按时间倒序）
	db = db.Offset((query.Index - 1) * query.Count).Limit(query.Count).Order("update_at desc")

	if err := db.Find(&logs).Error; err != nil {
		return nil, apperrors.Database(err)
	}

	return logs, nil
}

// Save 保存日志
func (r *taskLogRepository) Save(log *domain.TaskLog) error {
	if err := r.db.Save(log).Error; err != nil {
		return apperrors.Database(err)
	}
	return nil
}

// DeleteByTimeRange 根据时间范围删除日志
func (r *taskLogRepository) DeleteByTimeRange(query *LogQuery) error {
	db := r.db

	if query.Tid > 0 {
		db = db.Where("tid = ?", query.Tid)
	}
	if query.Cid > 0 {
		db = db.Where("cid = ?", query.Cid)
	}
	if query.LeftTs > 0 {
		db = db.Where("update_at > ?", query.LeftTs)
	}
	if query.RightTs > 0 {
		db = db.Where("update_at < ?", query.RightTs)
	}

	if err := db.Delete(&domain.TaskLog{}).Error; err != nil {
		return apperrors.Database(err)
	}
	return nil
}
