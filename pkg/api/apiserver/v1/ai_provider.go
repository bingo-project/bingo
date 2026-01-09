// ABOUTME: AI Provider API request and response structures.
// ABOUTME: Defines DTOs for AI Provider management operations.

package v1

import "time"

// AiProviderInfo represents AI provider information.
type AiProviderInfo struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	DisplayName string    `json:"displayName"`
	Status      string    `json:"status"`
	IsDefault   bool      `json:"isDefault"`
	Sort        int       `json:"sort"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// ListAiProviderRequest represents a request to list AI providers.
type ListAiProviderRequest struct {
	Status string `form:"status" binding:"omitempty,oneof=active disabled"`
}

// ListAiProviderResponse represents a response containing a list of AI providers.
type ListAiProviderResponse struct {
	Total int64            `json:"total"`
	Data  []AiProviderInfo `json:"data"`
}

// UpdateAiProviderRequest represents a request to update an AI provider.
type UpdateAiProviderRequest struct {
	DisplayName string `json:"displayName,omitempty" binding:"omitempty,max=64"`
	Status      string `json:"status,omitempty" binding:"omitempty,oneof=active disabled"`
	IsDefault   *bool  `json:"isDefault,omitempty"`
	Sort        *int   `json:"sort,omitempty"`
}
