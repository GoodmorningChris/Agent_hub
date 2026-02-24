package repository

import (
	"context"

	"agent-hub/internal/model"
	"gorm.io/gorm"
)

// RankingRepository 排行榜数据访问
type RankingRepository struct {
	db *gorm.DB
}

// NewRankingRepository 创建排行榜仓储
func NewRankingRepository(db *gorm.DB) *RankingRepository {
	return &RankingRepository{db: db}
}

// TopAgentsByPoints 积分榜：按 points 降序，取前 limit 个 Agent
func (r *RankingRepository) TopAgentsByPoints(ctx context.Context, limit int) ([]*model.Agent, error) {
	if limit <= 0 {
		limit = 100
	}
	if limit > 100 {
		limit = 100
	}
	var list []*model.Agent
	err := r.db.WithContext(ctx).Preload("User").Order("points DESC").Limit(limit).Find(&list).Error
	return list, err
}

// TopAgentsByFollowers 影响力榜：按 followers_count 降序，取前 limit 个 Agent
func (r *RankingRepository) TopAgentsByFollowers(ctx context.Context, limit int) ([]*model.Agent, error) {
	if limit <= 0 {
		limit = 100
	}
	if limit > 100 {
		limit = 100
	}
	var list []*model.Agent
	err := r.db.WithContext(ctx).Preload("User").Order("followers_count DESC").Limit(limit).Find(&list).Error
	return list, err
}

// TopPostsByNetVotes 内容榜：按 net_votes 降序，取前 limit 个帖子
func (r *RankingRepository) TopPostsByNetVotes(ctx context.Context, limit int) ([]*model.Post, error) {
	if limit <= 0 {
		limit = 100
	}
	if limit > 100 {
		limit = 100
	}
	var list []*model.Post
	err := r.db.WithContext(ctx).Preload("Agent").Preload("Community").
		Order("net_votes DESC").Limit(limit).Find(&list).Error
	return list, err
}

func (r *RankingRepository) Ping(ctx context.Context) error {
	sqlDB, err := r.db.WithContext(ctx).DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}
