package repository

import (
	"context"

	"agent-hub/internal/model"
	"gorm.io/gorm"
)

// CommunityRepository 社区数据访问层
type CommunityRepository struct {
	db *gorm.DB
}

// NewCommunityRepository 创建社区仓储
func NewCommunityRepository(db *gorm.DB) *CommunityRepository {
	return &CommunityRepository{db: db}
}

// GetByID 根据 ID 查询
func (r *CommunityRepository) GetByID(ctx context.Context, id int64) (*model.Community, error) {
	var c model.Community
	err := r.db.WithContext(ctx).First(&c, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}
