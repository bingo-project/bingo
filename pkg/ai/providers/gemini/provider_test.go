// ABOUTME: Gemini provider unit tests.
// ABOUTME: Tests provider creation and basic functionality.

package gemini

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
	assert.Equal(t, "gemini", p.Name())
}

func TestProvider_Models(t *testing.T) {
	cfg := DefaultConfig()
	cfg.APIKey = "test-key"

	p, err := New(cfg)
	require.NoError(t, err)

	models := p.Models()
	assert.GreaterOrEqual(t, len(models), 2)

	// Check that gemini-2.0-flash is included
	var found bool
	for _, m := range models {
		if m.ID == "gemini-2.0-flash-exp" {
			found = true

			break
		}
	}
	assert.True(t, found, "gemini-2.0-flash-exp should be in models list")
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

	assert.True(t, modelIDs["gemini-2.0-flash-exp"], "should have gemini-2.0-flash-exp")
}
