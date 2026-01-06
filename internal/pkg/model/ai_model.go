// ABOUTME: AI model configuration definition.
// ABOUTME: Represents available AI models with pricing information.

package model

import "time"

type AiModelM struct {
	ID            uint    `gorm:"primaryKey" json:"id"`
	ProviderName  string  `gorm:"column:provider_name;type:varchar(32);index:idx_provider_name;not null" json:"providerName"`
	Model         string  `gorm:"column:model;type:varchar(64);uniqueIndex:uk_model;not null" json:"model"`
	DisplayName   string  `gorm:"column:display_name;type:varchar(64)" json:"displayName"`
	MaxTokens     int     `gorm:"column:max_tokens;type:int;not null;default:4096" json:"maxTokens"`
	InputPrice    float64 `gorm:"column:input_price;type:decimal(10,6);not null;default:0" json:"inputPrice"`
	OutputPrice   float64 `gorm:"column:output_price;type:decimal(10,6);not null;default:0" json:"outputPrice"`
	Status        string  `gorm:"column:status;type:varchar(16);not null;default:active" json:"status"`
	IsDefault     bool    `gorm:"column:is_default;type:tinyint(1);not null;default:0" json:"isDefault"`
	Sort          int     `gorm:"column:sort;type:int;not null;default:0" json:"sort"`
	AllowFallback bool    `gorm:"column:allow_fallback;type:tinyint(1);not null;default:1" json:"allowFallback"`

	CreatedAt time.Time `gorm:"type:DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3)" json:"createdAt"`
	UpdatedAt time.Time `gorm:"type:DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3)" json:"updatedAt"`
}

func (*AiModelM) TableName() string {
	return "ai_model"
}

// Model status constants.
const (
	AiModelStatusActive   = "active"
	AiModelStatusDisabled = "disabled"
)
