// ABOUTME: OpenAI provider unit tests.
// ABOUTME: Tests provider creation and basic functionality.

package openai

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
	assert.Equal(t, "openai", p.Name())
}

func TestProvider_Models(t *testing.T) {
	cfg := DefaultConfig()
	cfg.APIKey = "test-key"

	p, err := New(cfg)
	require.NoError(t, err)

	models := p.Models()
	assert.GreaterOrEqual(t, len(models), 4)

	// Check that gpt-4o is included
	var found bool
	for _, m := range models {
		if m.ID == "gpt-4o" {
			found = true

			break
		}
	}
	assert.True(t, found, "gpt-4o should be in models list")
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

	result := convertMessages(msgs)

	assert.Len(t, result, 3)
	assert.Equal(t, "You are helpful", result[0].Content)
	assert.Equal(t, "Hello", result[1].Content)
	assert.Equal(t, "Hi there", result[2].Content)
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	assert.Equal(t, "https://api.openai.com/v1", cfg.BaseURL)
	assert.NotEmpty(t, cfg.Models)

	// Check default models
	modelIDs := make(map[string]bool)
	for _, m := range cfg.Models {
		modelIDs[m.ID] = true
	}

	assert.True(t, modelIDs["gpt-4o"], "should have gpt-4o")
	assert.True(t, modelIDs["gpt-4o-mini"], "should have gpt-4o-mini")
}
