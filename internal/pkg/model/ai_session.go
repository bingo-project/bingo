// ABOUTME: AI session model definition.
// ABOUTME: Represents a chat session with message history.

package model

type AiSessionM struct {
	Base

	SessionID    string `gorm:"column:session_id;type:varchar(64);uniqueIndex:uk_session_id;not null" json:"sessionId"`
	UID          string `gorm:"column:uid;type:varchar(64);index:idx_uid;not null" json:"uid"`
	Title        string `gorm:"column:title;type:varchar(255);not null;default:''" json:"title"`
	Model        string `gorm:"column:model;type:varchar(64);not null" json:"model"`
	MessageCount int    `gorm:"column:message_count;type:int;not null;default:0" json:"messageCount"`
	TotalTokens  int    `gorm:"column:total_tokens;type:int;not null;default:0" json:"totalTokens"`
	Status       string `gorm:"column:status;type:varchar(16);not null;default:active" json:"status"`
}

func (*AiSessionM) TableName() string {
	return "ai_session"
}

// Session status constants.
const (
	AiSessionStatusActive   = "active"
	AiSessionStatusArchived = "archived"
	AiSessionStatusDeleted  = "deleted"
)
