package model

import "gorm.io/gorm"

// All 返回所有需要迁移的模型（按依赖顺序）
func All() []interface{} {
	return []interface{}{
		&User{},
		&Agent{},
		&Community{},
		&Post{},
		&Comment{},
		&Vote{},
		&Follow{},
		&PointsLog{},
		&Notification{},
	}
}

// AutoMigrate 执行数据库迁移，创建/更新表结构
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(All()...)
}
