// ABOUTME: Database migration for ai_model table.
// ABOUTME: Creates table for AI model configurations with pricing.

package migration

import (
	"time"

	"github.com/bingo-project/bingoctl/pkg/migrate"
	"gorm.io/gorm"
)

type CreateAIModelTable struct {
	ID            uint64    `gorm:"primaryKey"`
	ProviderName  string    `gorm:"type:varchar(32);uniqueIndex:uk_provider_model;not null"`
	Model         string    `gorm:"type:varchar(64);uniqueIndex:uk_provider_model;not null"`
	DisplayName   string    `gorm:"type:varchar(64)"`
	MaxTokens     int       `gorm:"type:int;not null;default:4096"`
	InputPrice    float64   `gorm:"type:decimal(10,6);not null;default:0"`
	OutputPrice   float64   `gorm:"type:decimal(10,6);not null;default:0"`
	Status        string    `gorm:"type:varchar(16);not null;default:active"`
	IsDefault     bool      `gorm:"type:tinyint(1);not null;default:0"`
	Sort          int       `gorm:"type:int;not null;default:0"`
	AllowFallback bool      `gorm:"type:tinyint(1);not null;default:1;comment:是否允许作为降级目标"`
	CreatedAt     time.Time `gorm:"type:DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3)"`
	UpdatedAt     time.Time `gorm:"type:DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3)"`
}

func (CreateAIModelTable) TableName() string {
	return "ai_model"
}

func (CreateAIModelTable) Up(migrator gorm.Migrator) {
	_ = migrator.AutoMigrate(&CreateAIModelTable{})
}

func (CreateAIModelTable) Down(migrator gorm.Migrator) {
	_ = migrator.DropTable(&CreateAIModelTable{})
}

func init() {
	migrate.Add("2025_12_29_100001_create_ai_model_table", CreateAIModelTable{}.Up, CreateAIModelTable{}.Down)
}
