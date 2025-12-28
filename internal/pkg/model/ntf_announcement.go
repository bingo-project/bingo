// ABOUTME: Announcement model for system-wide broadcasts.
// ABOUTME: Supports draft, scheduled, and published states.

package model

import (
	"time"

	"gorm.io/gorm"

	"github.com/bingo-project/bingo/internal/pkg/facade"
)

// AnnouncementStatus defines the announcement status type.
type AnnouncementStatus string

const (
	AnnouncementStatusDraft     AnnouncementStatus = "draft"
	AnnouncementStatusScheduled AnnouncementStatus = "scheduled"
	AnnouncementStatusPublished AnnouncementStatus = "published"
)

// NtfAnnouncementM represents a system announcement.
type NtfAnnouncementM struct {
	ID          uint64     `gorm:"primaryKey" json:"id"`
	UUID        string     `gorm:"type:varchar(64);uniqueIndex:uk_uuid" json:"uuid"`
	Title       string     `gorm:"type:varchar(255);not null" json:"title"`
	Content     string     `gorm:"type:text" json:"content"`
	ActionURL   string     `gorm:"type:varchar(512);not null;default:''" json:"actionUrl"`
	Status      string     `gorm:"type:varchar(32);index:idx_status;not null;default:'draft'" json:"status"`
	ScheduledAt *time.Time `gorm:"type:timestamp;default:null" json:"scheduledAt"`
	PublishedAt *time.Time `gorm:"type:timestamp;default:null" json:"publishedAt"`
	ExpiresAt   *time.Time `gorm:"type:timestamp;default:null" json:"expiresAt"`
	CreatedAt   time.Time  `gorm:"type:DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3);index:idx_created_at" json:"createdAt"`
	UpdatedAt   time.Time  `gorm:"type:DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3)" json:"updatedAt"`
}

func (*NtfAnnouncementM) TableName() string {
	return "ntf_announcement"
}

// BeforeCreate generates UUID before creating.
func (m *NtfAnnouncementM) BeforeCreate(tx *gorm.DB) error {
	if m.UUID == "" {
		m.UUID = facade.Snowflake.Generate().String()
	}

	return nil
}

// NtfAnnouncementReadM represents a user's read status for an announcement.
type NtfAnnouncementReadM struct {
	UserID         string    `gorm:"type:varchar(64);primaryKey" json:"userId"`
	AnnouncementID uint64    `gorm:"primaryKey" json:"announcementId"`
	ReadAt         time.Time `gorm:"type:DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3)" json:"readAt"`
}

func (*NtfAnnouncementReadM) TableName() string {
	return "ntf_announcement_read"
}
