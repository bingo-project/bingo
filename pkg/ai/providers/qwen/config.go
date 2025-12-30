// ABOUTME: Qwen provider configuration.
// ABOUTME: Defines Config for API key and model settings.

package qwen

import "github.com/bingo-project/bingo/pkg/ai"

// Config holds Qwen provider configuration
type Config struct {
	APIKey  string
	BaseURL string
	Models  []ai.ModelInfo
}

// DefaultConfig returns default configuration for Qwen
func DefaultConfig() *Config {
	return &Config{
		BaseURL: "https://dashscope.aliyuncs.com/compatible-mode/v1",
		Models: []ai.ModelInfo{
			{ID: "qwen-max", Name: "Qwen Max", Provider: "qwen", MaxTokens: 32000},
			{ID: "qwen-plus", Name: "Qwen Plus", Provider: "qwen", MaxTokens: 131072},
			{ID: "qwen-turbo", Name: "Qwen Turbo", Provider: "qwen", MaxTokens: 131072},
			{ID: "qwen-long", Name: "Qwen Long", Provider: "qwen", MaxTokens: 10000000},
		},
	}
}
