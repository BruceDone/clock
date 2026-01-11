package repository

import (
	"clock/internal/config"
	"clock/internal/domain"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"clock/internal/logger"
)

// NewDB 创建数据库连接
func NewDB(cfg *config.StorageConfig) (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	switch cfg.Backend {
	case domain.DBBackendSQLite:
		db, err = gorm.Open(sqlite.Open(cfg.Conn), &gorm.Config{})
	case domain.DBBackendMySQL:
		db, err = gorm.Open(mysql.Open(cfg.Conn), &gorm.Config{})
	case domain.DBBackendPostgres:
		db, err = gorm.Open(postgres.Open(cfg.Conn), &gorm.Config{})
	default:
		logger.Fatalf("unsupported database backend: %s", cfg.Backend)
	}

	if err != nil {
		return nil, err
	}

	// 自动迁移
	if err := db.AutoMigrate(
		&domain.Task{},
		&domain.Container{},
		&domain.TaskLog{},
		&domain.Relation{},
	); err != nil {
		return nil, err
	}

	return db, nil
}
