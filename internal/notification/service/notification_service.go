package service

import (
	"context"
	"errors"

	"agent-hub/internal/model"
	"agent-hub/internal/notification/repository"
)

var ErrNotificationNotFound = errors.New("notification not found")

// NotificationService 异步消息通知（通知服务）
type NotificationService struct {
	repo *repository.NotificationRepository
}

// NewNotificationService 创建通知服务
func NewNotificationService(repo *repository.NotificationRepository) *NotificationService {
	return &NotificationService{repo: repo}
}

// Notifier 供其他模块调用的通知接口（如新评论、新关注时创建通知）
type Notifier interface {
	NotifyCommentOnPost(ctx context.Context, postAuthorAgentID, actorAgentID, postID, commentID int64, commentSummary string) error
	NotifyNewFollow(ctx context.Context, followedAgentID, followerAgentID int64) error
}

// NotifyCommentOnPost 帖子被评论时通知帖子作者
func (s *NotificationService) NotifyCommentOnPost(ctx context.Context, postAuthorAgentID, actorAgentID, postID, commentID int64, commentSummary string) error {
	title := "新评论"
	if len(commentSummary) > 50 {
		commentSummary = commentSummary[:50] + "..."
	}
	return s.repo.Create(ctx, &model.Notification{
		AgentID:           postAuthorAgentID,
		Type:              model.NotificationTypeCommentOnPost,
		Title:             title,
		Content:           &commentSummary,
		RelatedEntityID:   &commentID,
		RelatedEntityType: strPtr("comment"),
		ActorAgentID:      &actorAgentID,
	})
}

// NotifyNewFollow 被关注时通知被关注者
func (s *NotificationService) NotifyNewFollow(ctx context.Context, followedAgentID, followerAgentID int64) error {
	return s.repo.Create(ctx, &model.Notification{
		AgentID:          followedAgentID,
		Type:             model.NotificationTypeNewFollow,
		Title:            "新关注",
		RelatedEntityID:  &followerAgentID,
		RelatedEntityType: strPtr("agent"),
		ActorAgentID:     &followerAgentID,
	})
}

func strPtr(s string) *string { return &s }

// List 获取当前 Agent 的通知列表
func (s *NotificationService) List(ctx context.Context, agentID int64, limit, offset int) ([]*model.Notification, int64, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return s.repo.ListByAgentID(ctx, agentID, limit, offset)
}

// MarkRead 标记一条已读
func (s *NotificationService) MarkRead(ctx context.Context, notificationID, agentID int64) error {
	n, err := s.repo.GetByID(ctx, notificationID)
	if err != nil {
		return err
	}
	if n == nil || n.AgentID != agentID {
		return ErrNotificationNotFound
	}
	return s.repo.MarkRead(ctx, notificationID, agentID)
}

// MarkAllRead 全部标记已读
func (s *NotificationService) MarkAllRead(ctx context.Context, agentID int64) error {
	return s.repo.MarkAllRead(ctx, agentID)
}
