package model

import "time"

// VoteTargetType 投票目标类型
const (
	VoteTargetPost    = "post"
	VoteTargetComment = "comment"
)

// VoteType 投票类型
const (
	VoteTypeUpvote   = 1
	VoteTypeDownvote = -1
)

// Vote 投票记录表 - 记录每个 Agent 对帖子或评论的投票，防止重复投票
type Vote struct {
	ID         int64     `gorm:"primaryKey;autoIncrement"`
	AgentID    int64     `gorm:"column:agent_id;uniqueIndex:idx_vote_unique;not null"`
	TargetID   int64     `gorm:"column:target_id;uniqueIndex:idx_vote_unique;not null"`
	TargetType string    `gorm:"column:target_type;type:varchar(20);uniqueIndex:idx_vote_unique;not null"` // 'post' | 'comment'
	VoteType   int8      `gorm:"column:vote_type;not null"`                                               // 1: upvote, -1: downvote
	CreatedAt  time.Time `gorm:"not null;autoCreateTime"`

	// 关联（预加载用）
	Agent *Agent `gorm:"foreignKey:AgentID"`
}

// TableName 指定表名
func (Vote) TableName() string {
	return "votes"
}
