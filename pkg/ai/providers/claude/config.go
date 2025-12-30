// ABOUTME: Claude provider configuration.
// ABOUTME: Defines Config for API key and model settings.

package claude

import "github.com/bingo-project/bingo/pkg/ai"

// Config holds Claude provider configuration
type Config struct {
	APIKey string
	Models []ai.ModelInfo
}

// DefaultConfig returns default configuration for Claude
func DefaultConfig() *Config {
	return &Config{
		Models: []ai.ModelInfo{
			{ID: "claude-sonnet-4-20250514", Name: "Claude Sonnet 4", Provider: "claude", MaxTokens: 200000},
			{ID: "claude-3-5-sonnet-20241022", Name: "Claude 3.5 Sonnet", Provider: "claude", MaxTokens: 200000},
			{ID: "claude-3-5-haiku-20241022", Name: "Claude 3.5 Haiku", Provider: "claude", MaxTokens: 200000},
			{ID: "claude-3-opus-20240229", Name: "Claude 3 Opus", Provider: "claude", MaxTokens: 200000},
		},
	}
}
