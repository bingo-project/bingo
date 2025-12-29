// ABOUTME: Database migration for ai_quota_tier table.
// ABOUTME: Creates table for AI usage quota tier definitions.

package migration

import (
	"time"

	"github.com/bingo-project/bingoctl/pkg/migrate"
	"gorm.io/gorm"
)

type CreateAIQuotaTierTable struct {
	ID          uint64    `gorm:"primaryKey"`
	Tier        string    `gorm:"type:varchar(32);uniqueIndex:uk_tier;not null"`
	DisplayName string    `gorm:"type:varchar(64)"`
	RPM         int       `gorm:"type:int;not null;default:10"`
	TPD         int       `gorm:"type:int;not null;default:100000"`
	CreatedAt   time.Time `gorm:"type:DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3)"`
	UpdatedAt   time.Time `gorm:"type:DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3)"`
}

func (CreateAIQuotaTierTable) TableName() string {
	return "ai_quota_tier"
}

func (CreateAIQuotaTierTable) Up(migrator gorm.Migrator) {
	_ = migrator.AutoMigrate(&CreateAIQuotaTierTable{})
}

func (CreateAIQuotaTierTable) Down(migrator gorm.Migrator) {
	_ = migrator.DropTable(&CreateAIQuotaTierTable{})
}

func init() {
	migrate.Add("2025_12_29_100002_create_ai_quota_tier_table", CreateAIQuotaTierTable{}.Up, CreateAIQuotaTierTable{}.Down)
}
