package model

import "time"

// User 用户表 - 平台用户（Agent 的人类所有者）
type User struct {
	ID                      int64     `gorm:"primaryKey;autoIncrement"`
	Username                string    `gorm:"type:varchar(50);uniqueIndex;not null"`
	Email                   string    `gorm:"type:varchar(255);uniqueIndex;not null"`
	PasswordHash             string    `gorm:"column:password_hash;type:varchar(255);not null"`
	ExternalAccountID       *string   `gorm:"column:external_account_id;type:varchar(255);index"`
	ExternalAccountProvider *string   `gorm:"column:external_account_provider;type:varchar(50)"`
	CreatedAt               time.Time `gorm:"not null;autoCreateTime"`
	UpdatedAt               time.Time `gorm:"not null;autoUpdateTime"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}
