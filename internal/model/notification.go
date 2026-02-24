package model

import "time"

// 通知类型
const (
	NotificationTypeCommentOnPost = "comment_on_post" // 帖子被评论
	NotificationTypeNewFollow     = "new_follow"      // 被关注
	NotificationTypePostUpvoted   = "post_upvoted"   // 帖子被点赞
	NotificationTypeCommentUpvoted = "comment_upvoted" // 评论被点赞
)

// Notification 通知表 - 存储发给 Agent 的通知
type Notification struct {
	ID                 int64     `gorm:"primaryKey;autoIncrement"`
	AgentID            int64     `gorm:"column:agent_id;index;not null"`            // 接收者 Agent
	Type               string    `gorm:"type:varchar(50);index;not null"`
	Title              string    `gorm:"type:varchar(200);not null"`
	Content            *string   `gorm:"type:text"`
	RelatedEntityID    *int64    `gorm:"column:related_entity_id"`
	RelatedEntityType  *string   `gorm:"column:related_entity_type;type:varchar(20)"` // post, comment, agent
	ActorAgentID       *int64    `gorm:"column:actor_agent_id"`                      // 触发通知的 Agent（如评论者、关注者）
	IsRead             bool      `gorm:"column:is_read;not null;default:false"`
	CreatedAt          time.Time `gorm:"not null;autoCreateTime"`

	Agent *Agent `gorm:"foreignKey:AgentID"`
}

// TableName 指定表名
func (Notification) TableName() string {
	return "notifications"
}
