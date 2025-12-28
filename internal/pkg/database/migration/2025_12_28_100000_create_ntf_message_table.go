// ABOUTME: Database migration for ntf_message table.
// ABOUTME: Creates table for personal notification messages.

package migration

import (
	"time"

	"github.com/bingo-project/bingoctl/pkg/migrate"
	"gorm.io/gorm"
)

type CreateNtfMessageTable struct {
	ID        uint64     `gorm:"primaryKey"`
	UUID      string     `gorm:"type:varchar(64);uniqueIndex:uk_uuid"`
	UserID    string     `gorm:"type:varchar(64);index:idx_user_id;not null"`
	Category  string     `gorm:"type:varchar(32);index:idx_category;not null"`
	Type      string     `gorm:"type:varchar(64);not null"`
	Title     string     `gorm:"type:varchar(255);not null"`
	Content   string     `gorm:"type:text"`
	ActionURL string     `gorm:"type:varchar(512);not null;default:''"`
	IsRead    bool       `gorm:"type:tinyint(1);not null;default:0"`
	ReadAt    *time.Time `gorm:"type:timestamp;default:null"`
	CreatedAt time.Time  `gorm:"type:DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3);index:idx_created_at"`
}

func (CreateNtfMessageTable) TableName() string {
	return "ntf_message"
}

func (CreateNtfMessageTable) Up(migrator gorm.Migrator) {
	_ = migrator.AutoMigrate(&CreateNtfMessageTable{})
}

func (CreateNtfMessageTable) Down(migrator gorm.Migrator) {
	_ = migrator.DropTable(&CreateNtfMessageTable{})
}

func init() {
	migrate.Add("2025_12_28_100000_create_ntf_message_table", CreateNtfMessageTable{}.Up, CreateNtfMessageTable{}.Down)
}
