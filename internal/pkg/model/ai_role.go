// ABOUTME: AI role preset model definition.
// ABOUTME: Represents AI persona configurations with custom system prompts.

package model

import "time"

// AiRoleStatus represents the status of an AI role.
type AiRoleStatus string

const (
	AiRoleStatusActive   AiRoleStatus = "active"
	AiRoleStatusDisabled AiRoleStatus = "disabled"
)

// AiRoleCategory represents the category of an AI role.
type AiRoleCategory string

const (
	AiRoleCategoryGeneral   AiRoleCategory = "general"
	AiRoleCategoryEducation AiRoleCategory = "education"
	AiRoleCategoryMedical   AiRoleCategory = "medical"
	AiRoleCategoryWorkplace AiRoleCategory = "workplace"
	AiRoleCategoryCreative  AiRoleCategory = "creative"
)

// AiRoleM represents an AI role preset.
type AiRoleM struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	RoleID       string         `gorm:"column:role_id;type:varchar(32);uniqueIndex:uk_role_id;not null" json:"roleId"`
	Name         string         `gorm:"column:name;type:varchar(64);not null" json:"name"`
	Description  string         `gorm:"column:description;type:varchar(255)" json:"description"`
	Icon         string         `gorm:"column:icon;type:varchar(255)" json:"icon"`
	Category     AiRoleCategory `gorm:"column:category;type:varchar(32);not null;default:'general'" json:"category"`
	SystemPrompt string         `gorm:"column:system_prompt;type:text;not null" json:"systemPrompt"`
	Model        string         `gorm:"column:model;type:varchar(64)" json:"model"`
	Temperature  float64        `gorm:"column:temperature;type:decimal(3,2);not null;default:0.70" json:"temperature"`
	MaxTokens    int            `gorm:"column:max_tokens;type:int;not null;default:2000" json:"maxTokens"`
	Sort         int            `gorm:"column:sort;type:int;not null;default:0" json:"sort"`
	Status       AiRoleStatus   `gorm:"column:status;type:varchar(16);not null;default:'active'" json:"status"`

	CreatedAt time.Time `gorm:"type:DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3)" json:"createdAt"`
	UpdatedAt time.Time `gorm:"type:DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3)" json:"updatedAt"`
}

func (*AiRoleM) TableName() string {
	return "ai_role"
}
