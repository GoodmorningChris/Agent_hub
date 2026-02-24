package model

import "time"

// Agent Agent 表 - 存储 Agent 的核心信息
type Agent struct {
	ID              int64     `gorm:"primaryKey;autoIncrement"`
	UserID          int64     `gorm:"column:user_id;uniqueIndex;not null"`
	Name            string    `gorm:"type:varchar(50);uniqueIndex;not null"`
	AvatarURL       *string   `gorm:"column:avatar_url;type:varchar(512)"`
	Bio             *string   `gorm:"type:text"`
	Points          int       `gorm:"not null;default:0"`
	FollowersCount  int       `gorm:"column:followers_count;not null;default:0"`
	FollowingCount  int       `gorm:"column:following_count;not null;default:0"`
	IsVerified      bool      `gorm:"column:is_verified;not null;default:false"`
	IsFoundingAgent bool      `gorm:"column:is_founding_agent;not null;default:false"`
	CreatedAt       time.Time `gorm:"not null;autoCreateTime"`
	UpdatedAt       time.Time `gorm:"not null;autoUpdateTime"`

	// 关联（预加载用）
	User *User `gorm:"foreignKey:UserID"`
}

// TableName 指定表名
func (Agent) TableName() string {
	return "agents"
}
