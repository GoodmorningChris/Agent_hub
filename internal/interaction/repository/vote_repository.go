package repository

import (
	"context"

	"agent-hub/internal/model"
	"gorm.io/gorm"
)

// VoteRepository 投票记录数据访问层
type VoteRepository struct {
	db *gorm.DB
}

// NewVoteRepository 创建投票仓储
func NewVoteRepository(db *gorm.DB) *VoteRepository {
	return &VoteRepository{db: db}
}

// Get 查询是否已投票
func (r *VoteRepository) Get(ctx context.Context, agentID, targetID int64, targetType string) (*model.Vote, error) {
	var v model.Vote
	err := r.db.WithContext(ctx).
		Where("agent_id = ? AND target_id = ? AND target_type = ?", agentID, targetID, targetType).
		First(&v).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &v, nil
}

// Create 创建投票记录
func (r *VoteRepository) Create(ctx context.Context, v *model.Vote) error {
	return r.db.WithContext(ctx).Create(v).Error
}

// Update 更新投票类型（如从 upvote 改为 downvote）
func (r *VoteRepository) Update(ctx context.Context, v *model.Vote) error {
	return r.db.WithContext(ctx).Model(v).Update("vote_type", v.VoteType).Error
}

func (r *VoteRepository) Ping(ctx context.Context) error {
	sqlDB, err := r.db.WithContext(ctx).DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}
