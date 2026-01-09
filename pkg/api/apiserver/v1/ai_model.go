// ABOUTME: AI Model API request and response structures.
// ABOUTME: Defines DTOs for AI Model management operations.

package v1

import "time"

// AiModelInfo represents AI model information.
type AiModelInfo struct {
	ID            uint      `json:"id"`
	ProviderName  string    `json:"providerName"`
	Model         string    `json:"model"`
	DisplayName   string    `json:"displayName"`
	MaxTokens     int       `json:"maxTokens"`
	InputPrice    float64   `json:"inputPrice"`
	OutputPrice   float64   `json:"outputPrice"`
	Status        string    `json:"status"`
	IsDefault     bool      `json:"isDefault"`
	Sort          int       `json:"sort"`
	AllowFallback bool      `json:"allowFallback"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

// ListAiModelRequest represents a request to list AI models.
type ListAiModelRequest struct {
	ProviderName string `form:"providerName" binding:"omitempty,max=32"`
	Status       string `form:"status" binding:"omitempty,oneof=active disabled"`
}

// ListAiModelResponse represents a response containing a list of AI models.
type ListAiModelResponse struct {
	Total int64         `json:"total"`
	Data  []AiModelInfo `json:"data"`
}

// UpdateAiModelRequest represents a request to update an AI model.
type UpdateAiModelRequest struct {
	DisplayName   string   `json:"displayName,omitempty" binding:"omitempty,max=64"`
	MaxTokens     *int     `json:"maxTokens,omitempty" binding:"omitempty,min=1"`
	InputPrice    *float64 `json:"inputPrice,omitempty" binding:"omitempty,min=0"`
	OutputPrice   *float64 `json:"outputPrice,omitempty" binding:"omitempty,min=0"`
	Status        string   `json:"status,omitempty" binding:"omitempty,oneof=active disabled"`
	IsDefault     *bool    `json:"isDefault,omitempty"`
	Sort          *int     `json:"sort,omitempty"`
	AllowFallback *bool    `json:"allowFallback,omitempty"`
}
