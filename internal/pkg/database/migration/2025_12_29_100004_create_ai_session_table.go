// ABOUTME: Database migration for ai_session table.
// ABOUTME: Creates table for AI chat sessions.

package migration

import (
	"time"

	"github.com/bingo-project/bingoctl/pkg/migrate"
	"gorm.io/gorm"
)

type CreateAISessionTable struct {
	ID           uint64    `gorm:"primaryKey"`
	SessionID    string    `gorm:"type:varchar(64);uniqueIndex:uk_session_id;not null"`
	UID          string    `gorm:"type:varchar(64);index:idx_uid;not null"`
	RoleID       string    `gorm:"type:varchar(64);index:idx_role_id"`
	Title        string    `gorm:"type:varchar(255);not null;default:''"`
	Model        string    `gorm:"type:varchar(64);not null;default:''"`
	MessageCount int       `gorm:"type:int;not null;default:0"`
	TotalTokens  int       `gorm:"type:int;not null;default:0"`
	Status       string    `gorm:"type:varchar(16);not null;default:'active'"`
	CreatedAt    time.Time `gorm:"type:DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3)"`
	UpdatedAt    time.Time `gorm:"type:DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3)"`
}

func (CreateAISessionTable) TableName() string {
	return "ai_session"
}

func (CreateAISessionTable) Up(migrator gorm.Migrator) {
	_ = migrator.AutoMigrate(&CreateAISessionTable{})
}

func (CreateAISessionTable) Down(migrator gorm.Migrator) {
	_ = migrator.DropTable(&CreateAISessionTable{})
}

func init() {
	migrate.Add("2025_12_29_100004_create_ai_session_table", CreateAISessionTable{}.Up, CreateAISessionTable{}.Down)
}
