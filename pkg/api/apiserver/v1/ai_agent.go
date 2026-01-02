// ABOUTME: AI role API request and response structures.
// ABOUTME: Defines DTOs for AI role preset CRUD operations.

package v1

// CreateAiAgentRequest represents a request to create an AI agent.
type CreateAiAgentRequest struct {
	AgentID      string  `json:"agentId" binding:"required,max=32" example:"math_teacher"`
	Name         string  `json:"name" binding:"required,max=64" example:"数学老师"`
	Description  string  `json:"description,omitempty" binding:"max=255" example:"擅长小学数学辅导"`
	Icon         string  `json:"icon,omitempty" binding:"max=255" example:"https://example.com/icon.png"`
	Category     string  `json:"category,omitempty" binding:"omitempty,oneof=general education medical workplace creative" example:"education"`
	SystemPrompt string  `json:"systemPrompt" binding:"required" example:"你是一位经验丰富的小学数学老师..."`
	Model        string  `json:"model,omitempty" binding:"max=64" example:"gpt-4o"`
	Temperature  float64 `json:"temperature,omitempty" example:"0.7"`
	MaxTokens    int     `json:"maxTokens,omitempty" example:"2000"`
	Sort         int     `json:"sort,omitempty" example:"1"`
}

// UpdateAiAgentRequest represents a request to update an AI agent.
type UpdateAiAgentRequest struct {
	Name         string  `json:"name,omitempty" binding:"max=64"`
	Description  string  `json:"description,omitempty" binding:"max=255"`
	Icon         string  `json:"icon,omitempty" binding:"max=255"`
	Category     string  `json:"category,omitempty" binding:"omitempty,oneof=general education medical workplace creative"`
	SystemPrompt string  `json:"systemPrompt,omitempty"`
	Model        string  `json:"model,omitempty" binding:"max=64"`
	Temperature  float64 `json:"temperature,omitempty"`
	MaxTokens    int     `json:"maxTokens,omitempty"`
	Sort         int     `json:"sort,omitempty"`
	Status       string  `json:"status,omitempty" binding:"omitempty,oneof=active disabled"`
}

// ListAiAgentRequest represents a request to list AI agents.
type ListAiAgentRequest struct {
	Category string `form:"category" binding:"omitempty,oneof=general education medical workplace creative"`
	Status   string `form:"status" binding:"omitempty,oneof=active disabled"`
}

// AiAgentInfo represents AI agent information.
type AiAgentInfo struct {
	AgentID      string  `json:"agentId"`
	Name         string  `json:"name"`
	Description  string  `json:"description"`
	Icon         string  `json:"icon"`
	Category     string  `json:"category"`
	SystemPrompt string  `json:"systemPrompt,omitempty"`
	Model        string  `json:"model"`
	Temperature  float64 `json:"temperature"`
	MaxTokens    int     `json:"maxTokens"`
	Sort         int     `json:"sort"`
	Status       string  `json:"status"`
}

// ListAiAgentResponse represents a response containing a list of AI agents.
type ListAiAgentResponse struct {
	Total int64        `json:"total"`
	Data  []AiAgentInfo `json:"data"`
}
