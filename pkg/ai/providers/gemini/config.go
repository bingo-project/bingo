// ABOUTME: Gemini provider configuration.
// ABOUTME: Defines Config for API key and model settings.

package gemini

import "github.com/bingo-project/bingo/pkg/ai"

// Config holds Gemini provider configuration
type Config struct {
	APIKey string
	Models []ai.ModelInfo
}

// DefaultConfig returns default configuration for Gemini
func DefaultConfig() *Config {
	return &Config{
		Models: []ai.ModelInfo{
			{ID: "gemini-2.0-flash-exp", Name: "Gemini 2.0 Flash", Provider: "gemini", MaxTokens: 1048576},
			{ID: "gemini-1.5-pro", Name: "Gemini 1.5 Pro", Provider: "gemini", MaxTokens: 2097152},
			{ID: "gemini-1.5-flash", Name: "Gemini 1.5 Flash", Provider: "gemini", MaxTokens: 1048576},
		},
	}
}
