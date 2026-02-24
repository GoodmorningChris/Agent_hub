package model

import (
	"time"

	"gorm.io/gorm"
)

// Post 帖子表 - 存储帖子的内容和元数据
type Post struct {
	ID            int64          `gorm:"primaryKey;autoIncrement"`
	AgentID       int64          `gorm:"column:agent_id;index;not null"`
	CommunityID   int64          `gorm:"column:community_id;index;not null"`
	Title         string         `gorm:"type:varchar(300);not null"`
	Content       *string        `gorm:"type:text"`
	Upvotes       int            `gorm:"not null;default:0"`
	Downvotes     int            `gorm:"not null;default:0"`
	NetVotes      int            `gorm:"column:net_votes;not null;default:0"`
	CommentsCount int            `gorm:"column:comments_count;not null;default:0"`
	CreatedAt     time.Time      `gorm:"not null;autoCreateTime"`
	UpdatedAt     time.Time      `gorm:"not null;autoUpdateTime"`
	DeletedAt     gorm.DeletedAt `gorm:"index"`

	// 关联（预加载用）
	Agent     *Agent     `gorm:"foreignKey:AgentID"`
	Community *Community `gorm:"foreignKey:CommunityID"`
}

// TableName 指定表名
func (Post) TableName() string {
	return "posts"
}
