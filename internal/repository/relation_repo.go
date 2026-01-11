package repository

import (
	"time"

	"gorm.io/gorm"

	"clock/internal/domain"
	apperrors "clock/internal/errors"
)

// relationRepository 关系仓储实现
type relationRepository struct {
	db *gorm.DB
}

// NewRelationRepository 创建关系仓储
func NewRelationRepository(db *gorm.DB) RelationRepository {
	return &relationRepository{db: db}
}

// GetByCID 根据容器ID获取关系列表
func (r *relationRepository) GetByCID(cid int) ([]*domain.Relation, error) {
	var relations []*domain.Relation
	if err := r.db.Where("cid = ?", cid).Find(&relations).Error; err != nil {
		return nil, apperrors.Database(err)
	}
	return relations, nil
}

// Save 保存关系
func (r *relationRepository) Save(relation *domain.Relation) error {
	relation.UpdateAt = time.Now().Unix()
	if err := r.db.Save(relation).Error; err != nil {
		return apperrors.Database(err)
	}
	return nil
}

// Delete 删除关系
func (r *relationRepository) Delete(rid int) error {
	if err := r.db.Where("rid = ?", rid).Delete(&domain.Relation{}).Error; err != nil {
		return apperrors.Database(err)
	}
	return nil
}

// DeleteByTID 根据任务ID删除关系
func (r *relationRepository) DeleteByTID(tid int) error {
	if err := r.db.Where("tid = ?", tid).Delete(&domain.Relation{}).Error; err != nil {
		return apperrors.Database(err)
	}
	return nil
}

// DeleteByNextTID 根据后续任务ID删除关系
func (r *relationRepository) DeleteByNextTID(nextTid int) error {
	if err := r.db.Where("next_tid = ?", nextTid).Delete(&domain.Relation{}).Error; err != nil {
		return apperrors.Database(err)
	}
	return nil
}
