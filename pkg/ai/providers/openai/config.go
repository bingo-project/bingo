// ABOUTME: OpenAI provider configuration.
// ABOUTME: Defines Config for API key, base URL, and model settings.

package openai

import "github.com/bingo-project/bingo/pkg/ai"

// Config holds OpenAI provider configuration
type Config struct {
	Name    string // Provider name (e.g., "openai", "deepseek", "moonshot")
	APIKey  string
	BaseURL string
	Models  []ai.ModelInfo
}

// DefaultConfig returns default configuration for OpenAI
func DefaultConfig() *Config {
	return &Config{
		Name:    "openai",
		BaseURL: "https://api.openai.com/v1",
		Models: []ai.ModelInfo{
			{ID: "gpt-4o", Name: "GPT-4o", Provider: "openai", MaxTokens: 128000},
			{ID: "gpt-4o-mini", Name: "GPT-4o Mini", Provider: "openai", MaxTokens: 128000},
			{ID: "gpt-4-turbo", Name: "GPT-4 Turbo", Provider: "openai", MaxTokens: 128000},
			{ID: "gpt-3.5-turbo", Name: "GPT-3.5 Turbo", Provider: "openai", MaxTokens: 16385},
		},
	}
}

// DeepSeekConfig returns configuration for DeepSeek (OpenAI-compatible)
func DeepSeekConfig() *Config {
	return &Config{
		Name:    "deepseek",
		BaseURL: "https://api.deepseek.com/v1",
		Models: []ai.ModelInfo{
			{ID: "deepseek-chat", Name: "DeepSeek Chat", Provider: "deepseek", MaxTokens: 64000},
			{ID: "deepseek-coder", Name: "DeepSeek Coder", Provider: "deepseek", MaxTokens: 64000},
		},
	}
}

// MoonshotConfig returns configuration for Moonshot (OpenAI-compatible)
func MoonshotConfig() *Config {
	return &Config{
		Name:    "moonshot",
		BaseURL: "https://api.moonshot.cn/v1",
		Models: []ai.ModelInfo{
			{ID: "moonshot-v1-8k", Name: "Moonshot V1 8K", Provider: "moonshot", MaxTokens: 8000},
			{ID: "moonshot-v1-32k", Name: "Moonshot V1 32K", Provider: "moonshot", MaxTokens: 32000},
			{ID: "moonshot-v1-128k", Name: "Moonshot V1 128K", Provider: "moonshot", MaxTokens: 128000},
		},
	}
}
