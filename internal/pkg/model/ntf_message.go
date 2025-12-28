// ABOUTME: Notification message model for individual user notifications.
// ABOUTME: Stores personal notifications like login alerts, transaction updates.

package model

import (
	"time"

	"gorm.io/gorm"

	"github.com/bingo-project/bingo/internal/pkg/facade"
)

// NotificationCategory defines the notification category type.
type NotificationCategory string

const (
	NotificationCategorySystem      NotificationCategory = "system"
	NotificationCategorySecurity    NotificationCategory = "security"
	NotificationCategoryTransaction NotificationCategory = "transaction"
	NotificationCategorySocial      NotificationCategory = "social"
)

// NtfMessageM represents a notification message.
type NtfMessageM struct {
	ID        uint64     `gorm:"primaryKey" json:"id"`
	UUID      string     `gorm:"type:varchar(64);uniqueIndex:uk_uuid" json:"uuid"`
	UserID    string     `gorm:"type:varchar(64);index:idx_user_id;not null" json:"userId"`
	Category  string     `gorm:"type:varchar(32);index:idx_category;not null" json:"category"`
	Type      string     `gorm:"type:varchar(64);not null" json:"type"`
	Title     string     `gorm:"type:varchar(255);not null" json:"title"`
	Content   string     `gorm:"type:text" json:"content"`
	ActionURL string     `gorm:"type:varchar(512);not null;default:''" json:"actionUrl"`
	IsRead    bool       `gorm:"type:tinyint(1);not null;default:0" json:"isRead"`
	ReadAt    *time.Time `gorm:"type:timestamp;default:null" json:"readAt"`
	CreatedAt time.Time  `gorm:"type:DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3);index:idx_created_at" json:"createdAt"`
}

func (*NtfMessageM) TableName() string {
	return "ntf_message"
}

// BeforeCreate generates UUID before creating.
func (m *NtfMessageM) BeforeCreate(tx *gorm.DB) error {
	if m.UUID == "" {
		m.UUID = facade.Snowflake.Generate().String()
	}

	return nil
}
