package repository

import (
	"context"
	"time"

	"agent-hub/internal/model"
	"gorm.io/gorm"
)

// PointsRepository 积分与积分日志数据访问层
type PointsRepository struct {
	db *gorm.DB
}

// NewPointsRepository 创建积分仓储
func NewPointsRepository(db *gorm.DB) *PointsRepository {
	return &PointsRepository{db: db}
}

// CreateLog 写入积分日志
func (r *PointsRepository) CreateLog(ctx context.Context, log *model.PointsLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// SumTodayPointsByAgentAndReason 统计当日某 Agent 某 reason 的积分总和（用于每日上限）
func (r *PointsRepository) SumTodayPointsByAgentAndReason(ctx context.Context, agentID int64, reason string) (int, error) {
	start := time.Now().UTC().Truncate(24 * time.Hour)
	var sum int
	err := r.db.WithContext(ctx).Model(&model.PointsLog{}).
		Select("COALESCE(SUM(points_change), 0)").
		Where("agent_id = ? AND reason = ? AND created_at >= ?", agentID, reason, start).
		Scan(&sum).Error
	return sum, err
}

// HasReasonOnce 是否已有过该 reason 记录（一次性奖励用）
func (r *PointsRepository) HasReasonOnce(ctx context.Context, agentID int64, reason string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.PointsLog{}).
		Where("agent_id = ? AND reason = ?", agentID, reason).
		Limit(1).Count(&count).Error
	return count > 0, err
}

// AddAgentPoints 原子增加 Agent 积分（可为负）
func (r *PointsRepository) AddAgentPoints(ctx context.Context, agentID int64, delta int) error {
	return r.db.WithContext(ctx).Model(&model.Agent{}).Where("id = ?", agentID).
		UpdateColumn(
			"points",
			gorm.Expr("CASE WHEN points + ? < 0 THEN 0 ELSE points + ? END", delta, delta),
		).Error
}

func (r *PointsRepository) Ping(ctx context.Context) error {
	sqlDB, err := r.db.WithContext(ctx).DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}
