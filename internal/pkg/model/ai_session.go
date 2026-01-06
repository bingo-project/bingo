// ABOUTME: AI session model definition.
// ABOUTME: Represents a chat session with message history.

package model

import "time"

// AiSessionStatus represents the status of an AI session.
type AiSessionStatus string

const (
	AiSessionStatusActive   AiSessionStatus = "active"
	AiSessionStatusArchived AiSessionStatus = "archived"
	AiSessionStatusDeleted  AiSessionStatus = "deleted"
)

type AiSessionM struct {
	ID           uint            `gorm:"primaryKey" json:"id"`
	SessionID    string          `gorm:"column:session_id;type:varchar(64);uniqueIndex:uk_session_id;not null" json:"sessionId"`
	UID          string          `gorm:"column:uid;type:varchar(64);index:idx_uid;not null" json:"uid"`
	AgentID      string          `gorm:"column:agent_id;type:varchar(64);index:idx_agent_id" json:"agentId"`
	Title        string          `gorm:"column:title;type:varchar(255);not null;default:''" json:"title"`
	Model        string          `gorm:"column:model;type:varchar(64);not null;default:''" json:"model"`
	MessageCount int             `gorm:"column:message_count;type:int;not null;default:0" json:"messageCount"`
	TotalTokens  int             `gorm:"column:total_tokens;type:int;not null;default:0" json:"totalTokens"`
	Status       AiSessionStatus `gorm:"column:status;type:varchar(16);not null;default:'active'" json:"status"`

	CreatedAt time.Time `gorm:"type:DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3)" json:"createdAt"`
	UpdatedAt time.Time `gorm:"type:DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3)" json:"updatedAt"`
}

func (*AiSessionM) TableName() string {
	return "ai_session"
}
