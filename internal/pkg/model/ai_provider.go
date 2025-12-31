// ABOUTME: AI provider model definition.
// ABOUTME: Represents AI service providers like OpenAI, DeepSeek.

package model

import "time"

type AiProviderM struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	Name        string `gorm:"column:name;type:varchar(32);uniqueIndex:uk_name;not null" json:"name"`
	DisplayName string `gorm:"column:display_name;type:varchar(64)" json:"displayName"`
	Status      string `gorm:"column:status;type:varchar(16);not null;default:active" json:"status"`
	// Models field removed - models are now stored in ai_model table
	IsDefault bool `gorm:"column:is_default;type:tinyint(1);not null;default:0" json:"isDefault"`
	Sort      int  `gorm:"column:sort;type:int;not null;default:0" json:"sort"`

	CreatedAt time.Time `gorm:"type:DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3)" json:"createdAt"`
	UpdatedAt time.Time `gorm:"type:DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3)" json:"updatedAt"`
}

func (*AiProviderM) TableName() string {
	return "ai_provider"
}

// Provider status constants.
const (
	AiProviderStatusActive   = "active"
	AiProviderStatusDisabled = "disabled"
)
