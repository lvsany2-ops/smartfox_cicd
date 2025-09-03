// models/notification.go
package models

import (
	"time"

	"gorm.io/gorm"
)

// Notification 通知/公告主表
type Notification struct {
	ID           string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
	Title        string    `json:"title" gorm:"type:varchar(255);not null"`
	Content      string    `json:"content" gorm:"type:text;not null"`
	ExperimentID string    `json:"experiment_id" gorm:"type:varchar(36);default:''"` // 可选关联的实验ID
	CreatedAt    time.Time `json:"created_at"`
	IsImportant  bool      `json:"is_important" gorm:"default:false"` // 是否为重要公告（高亮显示）
}

// NotificationUser 通知与用户的关联表（记录已读状态）
type NotificationUser struct {
	UserID         string `gorm:"type:varchar(36);not null;index:idx_user_notification,unique"` // 用户ID（来自用户服务）
	NotificationID string `gorm:"type:varchar(36);not null;index:idx_user_notification,unique"` // 通知ID
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&Notification{}, &NotificationUser{})
}
