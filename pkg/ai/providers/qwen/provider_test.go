// ABOUTME: Qwen provider unit tests.
// ABOUTME: Tests provider creation and basic functionality.

package qwen

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bingo-project/bingo/pkg/ai"
)

func TestProvider_Name(t *testing.T) {
	cfg := DefaultConfig()
	cfg.APIKey = "test-key"

	p, err := New(cfg)
	require.NoError(t, err)
	assert.Equal(t, "qwen", p.Name())
}

func TestProvider_Models(t *testing.T) {
	cfg := DefaultConfig()
	cfg.APIKey = "test-key"

	p, err := New(cfg)
	require.NoError(t, err)

	models := p.Models()
	assert.GreaterOrEqual(t, len(models), 1)

	// Check that qwen-plus is included
	var found bool
	for _, m := range models {
		if m.ID == "qwen-plus" {
			found = true
			break
		}
	}
	assert.True(t, found, "qwen-plus should be in models list")
}

func TestProvider_ImplementsInterface(t *testing.T) {
	cfg := DefaultConfig()
	cfg.APIKey = "test-key"

	p, err := New(cfg)
	require.NoError(t, err)

	// Verify it implements ai.Provider
	var _ ai.Provider = p
}

func TestConvertMessages(t *testing.T) {
	msgs := []ai.Message{
		{Role: ai.RoleSystem, Content: "You are helpful"},
		{Role: ai.RoleUser, Content: "Hello"},
		{Role: ai.RoleAssistant, Content: "Hi there"},
	}

	result := ai.ConvertMessages(msgs)

	assert.Len(t, result, 3)
	assert.Equal(t, "You are helpful", result[0].Content)
	assert.Equal(t, "Hello", result[1].Content)
	assert.Equal(t, "Hi there", result[2].Content)
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	assert.NotEmpty(t, cfg.Models)
	assert.Equal(t, "https://dashscope.aliyuncs.com/compatible-mode/v1", cfg.BaseURL)

	// Check default models
	modelIDs := make(map[string]bool)
	for _, m := range cfg.Models {
		modelIDs[m.ID] = true
	}

	assert.True(t, modelIDs["qwen-plus"], "should have qwen-plus")
}
