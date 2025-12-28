// ABOUTME: Database migration for ntf_preference table.
// ABOUTME: Creates table for user notification preferences.

package migration

import (
	"time"

	"github.com/bingo-project/bingoctl/pkg/migrate"
	"gorm.io/gorm"
)

type CreateNtfPreferenceTable struct {
	ID          uint64    `gorm:"primaryKey"`
	UserID      string    `gorm:"type:varchar(64);uniqueIndex:uk_user_id"`
	Preferences string    `gorm:"type:json"`
	CreatedAt   time.Time `gorm:"type:DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3)"`
	UpdatedAt   time.Time `gorm:"type:DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3)"`
}

func (CreateNtfPreferenceTable) TableName() string {
	return "ntf_preference"
}

func (CreateNtfPreferenceTable) Up(migrator gorm.Migrator) {
	_ = migrator.AutoMigrate(&CreateNtfPreferenceTable{})
}

func (CreateNtfPreferenceTable) Down(migrator gorm.Migrator) {
	_ = migrator.DropTable(&CreateNtfPreferenceTable{})
}

func init() {
	migrate.Add("2025_12_28_100002_create_ntf_preference_table", CreateNtfPreferenceTable{}.Up, CreateNtfPreferenceTable{}.Down)
}
