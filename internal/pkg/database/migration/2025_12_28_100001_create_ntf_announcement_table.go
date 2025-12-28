// ABOUTME: Database migration for ntf_announcement tables.
// ABOUTME: Creates tables for announcements and read status tracking.

package migration

import (
	"time"

	"github.com/bingo-project/bingoctl/pkg/migrate"
	"gorm.io/gorm"
)

type CreateNtfAnnouncementTable struct {
	ID          uint64     `gorm:"primaryKey"`
	UUID        string     `gorm:"type:varchar(64);uniqueIndex:uk_uuid"`
	Title       string     `gorm:"type:varchar(255);not null"`
	Content     string     `gorm:"type:text"`
	ActionURL   string     `gorm:"type:varchar(512);not null;default:''"`
	Status      string     `gorm:"type:varchar(32);index:idx_status;not null;default:'draft'"`
	ScheduledAt *time.Time `gorm:"type:timestamp;default:null"`
	PublishedAt *time.Time `gorm:"type:timestamp;default:null"`
	ExpiresAt   *time.Time `gorm:"type:timestamp;default:null"`
	CreatedAt   time.Time  `gorm:"type:DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3);index:idx_created_at"`
	UpdatedAt   time.Time  `gorm:"type:DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3)"`
}

func (CreateNtfAnnouncementTable) TableName() string {
	return "ntf_announcement"
}

type CreateNtfAnnouncementReadTable struct {
	UserID         string    `gorm:"type:varchar(64);primaryKey"`
	AnnouncementID uint64    `gorm:"primaryKey"`
	ReadAt         time.Time `gorm:"type:DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3)"`
}

func (CreateNtfAnnouncementReadTable) TableName() string {
	return "ntf_announcement_read"
}

func (CreateNtfAnnouncementTable) Up(migrator gorm.Migrator) {
	_ = migrator.AutoMigrate(&CreateNtfAnnouncementTable{})
	_ = migrator.AutoMigrate(&CreateNtfAnnouncementReadTable{})
}

func (CreateNtfAnnouncementTable) Down(migrator gorm.Migrator) {
	_ = migrator.DropTable(&CreateNtfAnnouncementReadTable{})
	_ = migrator.DropTable(&CreateNtfAnnouncementTable{})
}

func init() {
	migrate.Add("2025_12_28_100001_create_ntf_announcement_table", CreateNtfAnnouncementTable{}.Up, CreateNtfAnnouncementTable{}.Down)
}
