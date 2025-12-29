// ABOUTME: Database migration for ai_message table.
// ABOUTME: Creates table for AI chat messages within sessions.

package migration

import (
	"time"

	"github.com/bingo-project/bingoctl/pkg/migrate"
	"gorm.io/gorm"
)

type CreateAIMessageTable struct {
	ID        uint64    `gorm:"primaryKey"`
	SessionID string    `gorm:"type:varchar(64);index:idx_session_id;not null"`
	Role      string    `gorm:"type:varchar(16);not null"`
	Content   string    `gorm:"type:text;not null"`
	Tokens    int       `gorm:"type:int;not null;default:0"`
	Model     string    `gorm:"type:varchar(64);not null;default:''"`
	CreatedAt time.Time `gorm:"type:DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3)"`
}

func (CreateAIMessageTable) TableName() string {
	return "ai_message"
}

func (CreateAIMessageTable) Up(migrator gorm.Migrator) {
	_ = migrator.AutoMigrate(&CreateAIMessageTable{})
}

func (CreateAIMessageTable) Down(migrator gorm.Migrator) {
	_ = migrator.DropTable(&CreateAIMessageTable{})
}

func init() {
	migrate.Add("2025_12_29_100005_create_ai_message_table", CreateAIMessageTable{}.Up, CreateAIMessageTable{}.Down)
}
