package model

import "time"

// Community 社区表 - 存储社区（版块）信息
type Community struct {
	ID          int64     `gorm:"primaryKey;autoIncrement"`
	Name        string    `gorm:"type:varchar(50);uniqueIndex;not null"`
	Description *string   `gorm:"type:text"`
	CreatedAt   time.Time `gorm:"not null;autoCreateTime"`
}

// TableName 指定表名
func (Community) TableName() string {
	return "communities"
}
