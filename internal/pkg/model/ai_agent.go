// ABOUTME: AI agent preset model definition.
// ABOUTME: Represents AI agent configurations with custom system prompts.

package model

import "time"

// AiAgentStatus represents the status of an AI agent.
type AiAgentStatus string

const (
	AiAgentStatusActive   AiAgentStatus = "active"
	AiAgentStatusDisabled AiAgentStatus = "disabled"
)

// AiAgentCategory represents the category of an AI agent.
type AiAgentCategory string

const (
	AiAgentCategoryGeneral   AiAgentCategory = "general"
	AiAgentCategoryEducation AiAgentCategory = "education"
	AiAgentCategoryMedical   AiAgentCategory = "medical"
	AiAgentCategoryWorkplace AiAgentCategory = "workplace"
	AiAgentCategoryCreative  AiAgentCategory = "creative"
)

// AiAgentM represents an AI agent preset.
type AiAgentM struct {
	ID           uint            `gorm:"primaryKey" json:"id"`
	AgentID      string          `gorm:"column:agent_id;type:varchar(32);uniqueIndex:uk_agent_id;not null" json:"agentId"`
	Name         string          `gorm:"column:name;type:varchar(64);not null" json:"name"`
	Description  string          `gorm:"column:description;type:varchar(255)" json:"description"`
	Icon         string          `gorm:"column:icon;type:varchar(255)" json:"icon"`
	Category     AiAgentCategory `gorm:"column:category;type:varchar(32);not null;default:'general'" json:"category"`
	SystemPrompt string          `gorm:"column:system_prompt;type:text;not null" json:"systemPrompt"`
	Model        string          `gorm:"column:model;type:varchar(64)" json:"model"`
	Temperature  float64         `gorm:"column:temperature;type:decimal(3,2);not null;default:0.70" json:"temperature"`
	MaxTokens    int             `gorm:"column:max_tokens;type:int;not null;default:2000" json:"maxTokens"`
	Sort         int             `gorm:"column:sort;type:int;not null;default:0" json:"sort"`
	Status       AiAgentStatus   `gorm:"column:status;type:varchar(16);not null;default:'active'" json:"status"`

	CreatedAt time.Time `gorm:"type:DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3)" json:"createdAt"`
	UpdatedAt time.Time `gorm:"type:DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3)" json:"updatedAt"`
}

func (AiAgentM) TableName() string {
	return "ai_agent"
}
