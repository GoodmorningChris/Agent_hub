package repository

import (
	"context"

	"agent-hub/internal/model"
	"gorm.io/gorm"
)

// FollowRepository 关注关系数据访问层
type FollowRepository struct {
	db *gorm.DB
}

// NewFollowRepository 创建关注仓储
func NewFollowRepository(db *gorm.DB) *FollowRepository {
	return &FollowRepository{db: db}
}

// Exists 检查是否已关注
func (r *FollowRepository) Exists(ctx context.Context, followerID, followingID int64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Follow{}).
		Where("follower_id = ? AND following_id = ?", followerID, followingID).
		Count(&count).Error
	return count > 0, err
}

// Create 创建关注关系
func (r *FollowRepository) Create(ctx context.Context, f *model.Follow) error {
	return r.db.WithContext(ctx).Create(f).Error
}

// Delete 取消关注
func (r *FollowRepository) Delete(ctx context.Context, followerID, followingID int64) error {
	return r.db.WithContext(ctx).
		Where("follower_id = ? AND following_id = ?", followerID, followingID).
		Delete(&model.Follow{}).Error
}

func (r *FollowRepository) Ping(ctx context.Context) error {
	sqlDB, err := r.db.WithContext(ctx).DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}
