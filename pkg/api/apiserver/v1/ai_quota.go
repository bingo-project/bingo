// ABOUTME: AI Quota API request and response structures.
// ABOUTME: Defines DTOs for AI User Quota management operations.

package v1

import "time"

// AiUserQuotaInfo represents user quota information.
type AiUserQuotaInfo struct {
	UID             string     `json:"uid"`
	Tier            string     `json:"tier"`
	RPM             int        `json:"rpm"`
	TPD             int        `json:"tpd"`
	UsedTokensToday int        `json:"usedTokensToday"`
	LastResetAt     *time.Time `json:"lastResetAt"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
}

// ListAiUserQuotaRequest represents a request to list user quotas.
type ListAiUserQuotaRequest struct {
	Tier     string `form:"tier" binding:"omitempty,oneof=free pro enterprise"`
	UID      string `form:"uid" binding:"omitempty,max=64"`
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"pageSize" binding:"omitempty,min=1,max=100"`
}

// ListAiUserQuotaResponse represents a response containing a list of user quotas.
type ListAiUserQuotaResponse struct {
	Total int64             `json:"total"`
	Data  []AiUserQuotaInfo `json:"data"`
}

// UpdateAiUserQuotaRequest represents a request to update user quota.
type UpdateAiUserQuotaRequest struct {
	Tier string `json:"tier,omitempty" binding:"omitempty,oneof=free pro enterprise"`
	RPM  *int   `json:"rpm,omitempty" binding:"omitempty,min=0"`
	TPD  *int   `json:"tpd,omitempty" binding:"omitempty,min=0"`
}
