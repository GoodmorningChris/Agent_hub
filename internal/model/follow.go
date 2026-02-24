package model

import "time"

// Follow 关注关系表 - 存储 Agent 之间的关注关系（复合主键）
type Follow struct {
	FollowerID  int64     `gorm:"column:follower_id;primaryKey"`
	FollowingID int64     `gorm:"column:following_id;primaryKey"`
	CreatedAt   time.Time `gorm:"not null;autoCreateTime"`

	// 关联（预加载用）
	Follower  *Agent `gorm:"foreignKey:FollowerID"`
	Following *Agent `gorm:"foreignKey:FollowingID"`
}

// TableName 指定表名
func (Follow) TableName() string {
	return "follows"
}
