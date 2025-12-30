// ABOUTME: Database migration for ai_user_quota table.
// ABOUTME: Creates table for per-user AI usage quotas and tracking.

package migration

import (
	"time"

	"github.com/bingo-project/bingoctl/pkg/migrate"
	"gorm.io/gorm"
)

type CreateAIUserQuotaTable struct {
	ID              uint64     `gorm:"primaryKey"`
	UID             string     `gorm:"type:varchar(64);uniqueIndex:uk_uid;not null"`
	Tier            string     `gorm:"type:varchar(32);index:idx_tier;not null;default:free"`
	RPM             int        `gorm:"type:int;not null;default:0"`
	TPD             int        `gorm:"type:int;not null;default:0"`
	UsedTokensToday int        `gorm:"type:int;not null;default:0"`
	LastResetAt     *time.Time `gorm:"type:timestamp;default:null"`
	CreatedAt       time.Time `gorm:"type:DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3)"`
	UpdatedAt       time.Time `gorm:"type:DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3)"`
}

func (CreateAIUserQuotaTable) TableName() string {
	return "ai_user_quota"
}

func (CreateAIUserQuotaTable) Up(migrator gorm.Migrator) {
	_ = migrator.AutoMigrate(&CreateAIUserQuotaTable{})
}

func (CreateAIUserQuotaTable) Down(migrator gorm.Migrator) {
	_ = migrator.DropTable(&CreateAIUserQuotaTable{})
}

func init() {
	migrate.Add("2025_12_29_100003_create_ai_user_quota_table", CreateAIUserQuotaTable{}.Up, CreateAIUserQuotaTable{}.Down)
}
