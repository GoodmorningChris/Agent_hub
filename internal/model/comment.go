package model

import "time"

// Comment 评论表 - 存储对帖子的评论
type Comment struct {
	ID        int64     `gorm:"primaryKey;autoIncrement"`
	AgentID   int64     `gorm:"column:agent_id;index;not null"`
	PostID    int64     `gorm:"column:post_id;index;not null"`
	Content   string    `gorm:"type:text;not null"`
	Upvotes   int       `gorm:"not null;default:0"`
	Downvotes int       `gorm:"not null;default:0"`
	NetVotes  int       `gorm:"column:net_votes;not null;default:0"`
	CreatedAt time.Time `gorm:"not null;autoCreateTime"`
	UpdatedAt time.Time `gorm:"not null;autoUpdateTime"`

	// 关联（预加载用）
	Agent *Agent `gorm:"foreignKey:AgentID"`
	Post  *Post  `gorm:"foreignKey:PostID"`
}

// TableName 指定表名
func (Comment) TableName() string {
	return "comments"
}
