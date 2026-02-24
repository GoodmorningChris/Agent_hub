package repository

import (
	"context"

	"agent-hub/internal/model"
	"gorm.io/gorm"
)

// NotificationRepository 通知数据访问层
type NotificationRepository struct {
	db *gorm.DB
}

// NewNotificationRepository 创建通知仓储
func NewNotificationRepository(db *gorm.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

// Create 创建通知
func (r *NotificationRepository) Create(ctx context.Context, n *model.Notification) error {
	return r.db.WithContext(ctx).Create(n).Error
}

// ListByAgentID 分页查询某 Agent 的通知，按创建时间倒序
func (r *NotificationRepository) ListByAgentID(ctx context.Context, agentID int64, limit, offset int) ([]*model.Notification, int64, error) {
	query := r.db.WithContext(ctx).Model(&model.Notification{}).Where("agent_id = ?", agentID)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []*model.Notification
	err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&list).Error
	return list, total, err
}

// GetByID 按 ID 查询
func (r *NotificationRepository) GetByID(ctx context.Context, id int64) (*model.Notification, error) {
	var n model.Notification
	err := r.db.WithContext(ctx).First(&n, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &n, nil
}

// MarkRead 标记已读
func (r *NotificationRepository) MarkRead(ctx context.Context, id, agentID int64) error {
	return r.db.WithContext(ctx).Model(&model.Notification{}).
		Where("id = ? AND agent_id = ?", id, agentID).Update("is_read", true).Error
}

// MarkAllRead 标记某 Agent 全部已读
func (r *NotificationRepository) MarkAllRead(ctx context.Context, agentID int64) error {
	return r.db.WithContext(ctx).Model(&model.Notification{}).
		Where("agent_id = ?", agentID).Update("is_read", true).Error
}
