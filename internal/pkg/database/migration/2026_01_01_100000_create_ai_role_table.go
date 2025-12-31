// ABOUTME: Database migration for ai_role table.
// ABOUTME: Creates table for AI role presets.

package migration

import (
	"time"

	"github.com/bingo-project/bingoctl/pkg/migrate"
	"gorm.io/gorm"
)

type CreateAIRoleTable struct {
	ID           uint64    `gorm:"primaryKey"`
	RoleID       string    `gorm:"type:varchar(32);uniqueIndex:uk_role_id;not null"`
	Name         string    `gorm:"type:varchar(64);not null"`
	Description  string    `gorm:"type:varchar(255)"`
	Icon         string    `gorm:"type:varchar(255)"`
	Category     string    `gorm:"type:varchar(32);not null;default:'general';index:idx_category_status"`
	SystemPrompt string    `gorm:"type:text;not null"`
	Model        string    `gorm:"type:varchar(64)"`
	Temperature  float64   `gorm:"type:decimal(3,2);default:0.70"`
	MaxTokens    int       `gorm:"type:int;default:2000"`
	Sort         int       `gorm:"type:int;default:0;index:idx_status_sort"`
	Status       string    `gorm:"type:varchar(16);not null;default:'active';index:idx_category_status;index:idx_status_sort"`
	CreatedAt    time.Time `gorm:"type:DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3)"`
	UpdatedAt    time.Time `gorm:"type:DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3)"`
}

func (CreateAIRoleTable) TableName() string {
	return "ai_role"
}

func (CreateAIRoleTable) Up(migrator gorm.Migrator) {
	_ = migrator.AutoMigrate(&CreateAIRoleTable{})
}

func (CreateAIRoleTable) Down(migrator gorm.Migrator) {
	_ = migrator.DropTable(&CreateAIRoleTable{})
}

func init() {
	migrate.Add("2026_01_01_100000_create_ai_role_table", CreateAIRoleTable{}.Up, CreateAIRoleTable{}.Down)
}
