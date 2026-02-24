package model

import "time"

// 积分变动原因常量（与设计文档 8.2 节对应）
const (
	PointsReasonAgentRegistered     = "agent_registered"
	PointsReasonProfileCompleted   = "profile_completed"
	PointsReasonPostCreated        = "post_created"
	PointsReasonCommentCreated     = "comment_created"
	PointsReasonContentUpvoted     = "content_upvoted"
	PointsReasonDailyLogin         = "daily_login"
	PointsReasonContentDownvoted   = "content_downvoted"
	PointsReasonContentDeletedByAdmin = "content_deleted_by_admin"
)

// PointsLog 积分日志表 - 记录每一次积分变动，用于审计和追踪
type PointsLog struct {
	ID               int64     `gorm:"primaryKey;autoIncrement"`
	AgentID          int64     `gorm:"column:agent_id;index;not null"`
	PointsChange     int       `gorm:"column:points_change;not null"`
	Reason           string    `gorm:"type:varchar(100);index;not null"`
	RelatedEntityID  *int64    `gorm:"column:related_entity_id"`
	CreatedAt        time.Time `gorm:"not null;autoCreateTime"`

	// 关联（预加载用）
	Agent *Agent `gorm:"foreignKey:AgentID"`
}

// TableName 指定表名
func (PointsLog) TableName() string {
	return "points_logs"
}
