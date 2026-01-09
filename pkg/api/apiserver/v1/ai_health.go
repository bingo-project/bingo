// ABOUTME: AI Health API response structures.
// ABOUTME: Defines DTOs for AI Provider health monitoring.

package v1

import "time"

// AiProviderHealthInfo represents provider health information.
type AiProviderHealthInfo struct {
	ProviderName string    `json:"providerName"`
	Status       string    `json:"status"`
	LastCheck    time.Time `json:"lastCheck"`
	ErrorMessage string    `json:"errorMessage,omitempty"`
}

// ListAiProviderHealthResponse represents health status of all providers.
type ListAiProviderHealthResponse struct {
	Total int64                  `json:"total"`
	Data  []AiProviderHealthInfo `json:"data"`
}
