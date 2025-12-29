// ABOUTME: OpenAI provider configuration.
// ABOUTME: Defines Config for API key, base URL, and model settings.

package openai

import "github.com/bingo-project/bingo/pkg/ai"

// Config holds OpenAI provider configuration
type Config struct {
	APIKey  string
	BaseURL string
	Models  []ai.ModelInfo
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	return &Config{
		BaseURL: "https://api.openai.com/v1",
		Models: []ai.ModelInfo{
			{ID: "gpt-4o", Name: "GPT-4o", Provider: "openai", MaxTokens: 128000},
			{ID: "gpt-4o-mini", Name: "GPT-4o Mini", Provider: "openai", MaxTokens: 128000},
			{ID: "gpt-4-turbo", Name: "GPT-4 Turbo", Provider: "openai", MaxTokens: 128000},
			{ID: "gpt-3.5-turbo", Name: "GPT-3.5 Turbo", Provider: "openai", MaxTokens: 16385},
		},
	}
}
