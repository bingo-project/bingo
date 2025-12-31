// ABOUTME: Database migration for ai_provider table.
// ABOUTME: Creates table for AI service providers configuration.

package migration

import (
	"time"

	"github.com/bingo-project/bingoctl/pkg/migrate"
	"gorm.io/gorm"
)

type CreateAIProviderTable struct {
	ID          uint64 `gorm:"primaryKey"`
	Name        string `gorm:"type:varchar(32);uniqueIndex:uk_name;not null"`
	DisplayName string `gorm:"type:varchar(64)"`
	Status      string `gorm:"type:varchar(16);not null;default:active"`
	// Models field removed - models are now stored in ai_model table
	IsDefault bool      `gorm:"type:tinyint(1);not null;default:0"`
	Sort      int       `gorm:"type:int;not null;default:0"`
	CreatedAt time.Time `gorm:"type:DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3)"`
	UpdatedAt time.Time `gorm:"type:DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3)"`
}

func (CreateAIProviderTable) TableName() string {
	return "ai_provider"
}

func (CreateAIProviderTable) Up(migrator gorm.Migrator) {
	_ = migrator.AutoMigrate(&CreateAIProviderTable{})
}

func (CreateAIProviderTable) Down(migrator gorm.Migrator) {
	_ = migrator.DropTable(&CreateAIProviderTable{})
}

func init() {
	migrate.Add("2025_12_29_100000_create_ai_provider_table", CreateAIProviderTable{}.Up, CreateAIProviderTable{}.Down)
}
