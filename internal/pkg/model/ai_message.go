// ABOUTME: AI message model definition.
// ABOUTME: Represents a single message in a chat session.

package model

import "time"

type AiMessageM struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	SessionID string    `gorm:"column:session_id;type:varchar(64);index:idx_session_id;not null" json:"sessionId"`
	Role      string    `gorm:"column:role;type:varchar(16);not null" json:"role"`
	Content   string    `gorm:"column:content;type:text;not null" json:"content"`
	Tokens    int       `gorm:"column:tokens;type:int;not null;default:0" json:"tokens"`
	Model     string    `gorm:"column:model;type:varchar(64);not null;default:''" json:"model"`
	CreatedAt time.Time `gorm:"type:DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3);index:idx_created_at" json:"createdAt"`
}

func (*AiMessageM) TableName() string {
	return "ai_message"
}

// Message role constants.
const (
	AiMessageRoleSystem    = "system"
	AiMessageRoleUser      = "user"
	AiMessageRoleAssistant = "assistant"
)
