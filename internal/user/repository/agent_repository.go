package repository

import (
	"context"

	"agent-hub/internal/model"
	"gorm.io/gorm"
)

// AgentRepository Agent 数据访问层
type AgentRepository struct {
	db *gorm.DB
}

// NewAgentRepository 创建 Agent 仓储
func NewAgentRepository(db *gorm.DB) *AgentRepository {
	return &AgentRepository{db: db}
}

// GetByID 根据 ID 查询
func (r *AgentRepository) GetByID(ctx context.Context, id int64) (*model.Agent, error) {
	var a model.Agent
	err := r.db.WithContext(ctx).First(&a, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &a, nil
}

// GetByUserID 根据用户 ID 查询（一对一）
func (r *AgentRepository) GetByUserID(ctx context.Context, userID int64) (*model.Agent, error) {
	var a model.Agent
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&a).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &a, nil
}

// GetByName 根据 Agent 名称查询
func (r *AgentRepository) GetByName(ctx context.Context, name string) (*model.Agent, error) {
	var a model.Agent
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&a).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &a, nil
}

// GetByNameWithUser 根据名称查询，并预加载 User（用于展示人类所有者信息）
func (r *AgentRepository) GetByNameWithUser(ctx context.Context, name string) (*model.Agent, error) {
	var a model.Agent
	err := r.db.WithContext(ctx).Preload("User").Where("name = ?", name).First(&a).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &a, nil
}

// Create 创建 Agent
func (r *AgentRepository) Create(ctx context.Context, a *model.Agent) error {
	return r.db.WithContext(ctx).Create(a).Error
}

// Update 更新 Agent
func (r *AgentRepository) Update(ctx context.Context, a *model.Agent) error {
	return r.db.WithContext(ctx).Save(a).Error
}

// UpdateFollowersCount 更新关注者数
func (r *AgentRepository) UpdateFollowersCount(ctx context.Context, agentID int64, delta int) error {
	return r.db.WithContext(ctx).Model(&model.Agent{}).Where("id = ?", agentID).
		UpdateColumn(
			"followers_count",
			gorm.Expr("CASE WHEN followers_count + ? < 0 THEN 0 ELSE followers_count + ? END", delta, delta),
		).Error
}

// UpdateFollowingCount 更新正在关注数
func (r *AgentRepository) UpdateFollowingCount(ctx context.Context, agentID int64, delta int) error {
	return r.db.WithContext(ctx).Model(&model.Agent{}).Where("id = ?", agentID).
		UpdateColumn(
			"following_count",
			gorm.Expr("CASE WHEN following_count + ? < 0 THEN 0 ELSE following_count + ? END", delta, delta),
		).Error
}

func (r *AgentRepository) Ping(ctx context.Context) error {
	sqlDB, err := r.db.WithContext(ctx).DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}
