package repository

import (
	"time"

	"gorm.io/gorm"

	"clock/internal/domain"
	apperrors "clock/internal/errors"
)

// containerRepository 容器仓储实现
type containerRepository struct {
	db *gorm.DB
}

// NewContainerRepository 创建容器仓储
func NewContainerRepository(db *gorm.DB) ContainerRepository {
	return &containerRepository{db: db}
}

// GetByID 根据ID获取容器
func (r *containerRepository) GetByID(cid int) (*domain.Container, error) {
	var container domain.Container
	if err := r.db.Where("cid = ?", cid).First(&container).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.NotFound("container")
		}
		return nil, apperrors.Database(err)
	}
	return &container, nil
}

// List 查询容器列表
func (r *containerRepository) List(query *ContainerQuery) ([]*domain.Container, error) {
	var containers []*domain.Container

	// 设置默认值
	if query.Count < 1 {
		query.Count = 10
	}
	if query.Index < 1 {
		query.Index = 1
	}

	db := r.db.Model(&domain.Container{})

	// 条件过滤
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

	if err := db.Find(&containers).Error; err != nil {
		return nil, apperrors.Database(err)
	}

	return containers, nil
}

// FindAll 获取所有容器
func (r *containerRepository) FindAll() ([]*domain.Container, error) {
	var containers []*domain.Container
	if err := r.db.Find(&containers).Error; err != nil {
		return nil, apperrors.Database(err)
	}
	return containers, nil
}

// Save 保存容器
func (r *containerRepository) Save(container *domain.Container) error {
	container.UpdateAt = time.Now().Unix()
	if err := r.db.Save(container).Error; err != nil {
		return apperrors.Database(err)
	}
	return nil
}

// Delete 删除容器
func (r *containerRepository) Delete(cid int) error {
	if err := r.db.Where("cid = ?", cid).Delete(&domain.Container{}).Error; err != nil {
		return apperrors.Database(err)
	}
	return nil
}
