// ABOUTME: Database migration for adding role_id to ai_session table.
// ABOUTME: Adds foreign key to ai_role for session-role binding.

package migration

import (
	"time"

	"github.com/bingo-project/bingoctl/pkg/migrate"
	"gorm.io/gorm"
)

type AddRoleIDToAISession struct {
	ID           uint64    `gorm:"primaryKey"`
	SessionID    string    `gorm:"type:varchar(64);uniqueIndex:uk_session_id;not null"`
	UID          string    `gorm:"type:varchar(64);index:idx_uid;not null"`
	RoleID       string    `gorm:"type:varchar(64);index:idx_role_id"`
	Title        string    `gorm:"type:varchar(255);not null;default:''"`
	Model        string    `gorm:"type:varchar(64);not null"`
	MessageCount int       `gorm:"type:int;not null;default:0"`
	TotalTokens  int       `gorm:"type:int;not null;default:0"`
	Status       string    `gorm:"type:varchar(16);not null;default:active"`
	CreatedAt    time.Time `gorm:"type:DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3)"`
	UpdatedAt    time.Time `gorm:"type:DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3)"`
}

func (AddRoleIDToAISession) TableName() string {
	return "ai_session"
}

func (AddRoleIDToAISession) Up(migrator gorm.Migrator) {
	_ = migrator.AutoMigrate(&AddRoleIDToAISession{})
}

func (AddRoleIDToAISession) Down(migrator gorm.Migrator) {
	_ = migrator.DB().Exec("ALTER TABLE ai_session DROP COLUMN role_id")
	_ = migrator.DB().Exec("ALTER TABLE ai_session DROP INDEX idx_role_id")
}

func init() {
	migrate.Add("2026_01_02_100000_add_role_id_to_ai_session", AddRoleIDToAISession{}.Up, AddRoleIDToAISession{}.Down)
}
