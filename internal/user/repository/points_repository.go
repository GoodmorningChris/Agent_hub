package repository

import (
	"context"

	"agent-hub/internal/model"
	"gorm.io/gorm"
)

// PointsRepository 积分日志数据访问（用户模块内用于注册时赠送积分）
type PointsRepository struct {
	db *gorm.DB
}

// NewPointsRepository 创建积分仓储
func NewPointsRepository(db *gorm.DB) *PointsRepository {
	return &PointsRepository{db: db}
}

// CreateLog 创建积分日志
func (r *PointsRepository) CreateLog(ctx context.Context, log *model.PointsLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}
