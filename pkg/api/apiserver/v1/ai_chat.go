// ABOUTME: Chat API request and response structures.
// ABOUTME: Defines DTOs for chat completions and session management.

package v1

import "time"

// ChatCompletionRequest represents a chat completion request (OpenAI-compatible).
type ChatCompletionRequest struct {
	Model       string        `json:"model,omitempty" example:"glm-4-flash"`
	Messages    []ChatMessage `json:"messages" binding:"required,min=1"`
	MaxTokens   int           `json:"max_tokens,omitempty" example:"2048"`
	Temperature float64       `json:"temperature,omitempty" example:"0.7"`
	Stream      bool          `json:"stream,omitempty" example:"false"`
	// Extension fields
	SessionID string `json:"sessionId,omitempty"`
}

// ChatMessage represents a single message.
type ChatMessage struct {
	Role    string `json:"role" binding:"required,oneof=system user assistant" example:"user"`
	Content string `json:"content" binding:"required,max=32768" example:"你好"`
}

// ChatCompletionResponse represents a chat completion response (OpenAI-compatible).
type ChatCompletionResponse struct {
	ID      string       `json:"id"`
	Object  string       `json:"object"`
	Created int64        `json:"created"`
	Model   string       `json:"model"`
	Choices []ChatChoice `json:"choices"`
	Usage   ChatUsage    `json:"usage"`
}

// ChatChoice represents a completion choice.
type ChatChoice struct {
	Index        int          `json:"index"`
	Message      ChatMessage  `json:"message"`
	FinishReason string       `json:"finish_reason"`
	Delta        *ChatMessage `json:"delta,omitempty"`
}

// ChatUsage represents token usage.
type ChatUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ListModelsResponse represents the models list response.
type ListModelsResponse struct {
	Object string      `json:"object"`
	Data   []ModelInfo `json:"data"`
}

// ModelInfo represents model metadata.
type ModelInfo struct {
	ID          string  `json:"id"`
	Object      string  `json:"object"`
	Created     int64   `json:"created"`
	OwnedBy     string  `json:"ownedBy"`
	MaxTokens   int     `json:"maxTokens,omitempty"`
	InputPrice  float64 `json:"inputPrice,omitempty"`
	OutputPrice float64 `json:"outputPrice,omitempty"`
}

// CreateSessionRequest represents session creation request.
type CreateSessionRequest struct {
	AgentID string `json:"agentId,omitempty"` // Optional: bind agent to session
	Title   string `json:"title,omitempty"`   // Optional: defaults to agent name or "新对话"
	Model   string `json:"model,omitempty"`   // Optional: override agent's default model
}

// UpdateSessionRequest represents session update request.
type UpdateSessionRequest struct {
	Title string `json:"title,omitempty"`
	Model string `json:"model,omitempty"`
}

// SessionInfo represents session information.
type SessionInfo struct {
	SessionID    string    `json:"sessionId"`
	Title        string    `json:"title"`
	AgentID      string    `json:"agentId,omitempty"`
	AgentName    string    `json:"agentName,omitempty"`
	Model        string    `json:"model"`
	MessageCount int       `json:"messageCount"`
	TotalTokens  int       `json:"totalTokens"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// SessionHistoryResponse represents session history response.
type SessionHistoryResponse struct {
	SessionID string        `json:"sessionId"`
	Messages  []ChatMessage `json:"messages"`
}
