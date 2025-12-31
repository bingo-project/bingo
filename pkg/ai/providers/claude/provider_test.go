// ABOUTME: Claude provider unit tests.
// ABOUTME: Tests provider creation and basic functionality.

package claude

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
	assert.Equal(t, "claude", p.Name())
}

func TestProvider_Name_Custom(t *testing.T) {
	cfg := DefaultConfig()
	cfg.APIKey = "test-key"
	cfg.Name = "anthropic"

	p, err := New(cfg)
	require.NoError(t, err)
	assert.Equal(t, "anthropic", p.Name())
}

func TestProvider_Models(t *testing.T) {
	cfg := DefaultConfig()
	cfg.APIKey = "test-key"

	p, err := New(cfg)
	require.NoError(t, err)

	models := p.Models()
	assert.GreaterOrEqual(t, len(models), 2)

	// Check that claude-sonnet-4 is included
	var found bool
	for _, m := range models {
		if m.ID == "claude-sonnet-4-20250514" {
			found = true
			break
		}
	}
	assert.True(t, found, "claude-sonnet-4-20250514 should be in models list")
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

	// Check default models
	modelIDs := make(map[string]bool)
	for _, m := range cfg.Models {
		modelIDs[m.ID] = true
	}

	assert.True(t, modelIDs["claude-sonnet-4-20250514"], "should have claude-sonnet-4-20250514")
	assert.True(t, modelIDs["claude-3-5-sonnet-20241022"], "should have claude-3-5-sonnet-20241022")
}
