// ABOUTME: Provider interface definition.
// ABOUTME: Abstracts AI service providers for multi-provider support.

package ai

import "context"

// Provider defines the interface for AI service providers
type Provider interface {
	// Name returns the provider identifier (e.g., "openai", "deepseek")
	Name() string

	// Chat performs a non-streaming chat completion
	Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error)

	// ChatStream performs a streaming chat completion
	ChatStream(ctx context.Context, req *ChatRequest) (*ChatStream, error)

	// Models returns the list of models supported by this provider
	Models() []ModelInfo
}

// ModelInfo contains model metadata
type ModelInfo struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Provider    string  `json:"provider"`
	MaxTokens   int     `json:"max_tokens"`
	InputPrice  float64 `json:"input_price"`
	OutputPrice float64 `json:"output_price"`
}
