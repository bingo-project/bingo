// ABOUTME: Tests for AI provider loader.
// ABOUTME: Verifies loading, reloading, and error handling.

package ai

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bingo-project/bingo/internal/pkg/model"
	mockstore "github.com/bingo-project/bingo/internal/pkg/testing/mock/store"
	aipkg "github.com/bingo-project/bingo/pkg/ai"
)

func TestLoader_Load_Success(t *testing.T) {
	store := mockstore.NewStore()
	registry := aipkg.NewRegistry()

	// Type assertion to access mock-specific fields
	store.AiProvider().(*mockstore.AiProviderStore).ListActiveResult = []*model.AiProviderM{
		{Name: "openai", Status: model.AiProviderStatusActive},
	}
	store.AiModel().(*mockstore.AiModelStore).ListActiveResult = []*model.AiModelM{
		{ProviderName: "openai", Model: "gpt-4o", Status: model.AiModelStatusActive, MaxTokens: 128000},
	}

	creds := map[string]Credential{
		"openai": {APIKey: "test-key"},
	}

	loader := NewLoader(registry, store, creds)
	err := loader.Load(context.Background())

	require.NoError(t, err)
	provider, ok := registry.Get("openai")
	require.True(t, ok, "openai provider should be registered")
	assert.Equal(t, "openai", provider.Name())
	assert.True(t, store.AiProvider().(*mockstore.AiProviderStore).ListActiveCalled, "ListActive should be called")
}

func TestLoader_Load_NoCredential_Skips(t *testing.T) {
	store := mockstore.NewStore()
	registry := aipkg.NewRegistry()

	store.AiProvider().(*mockstore.AiProviderStore).ListActiveResult = []*model.AiProviderM{
		{Name: "openai", Status: model.AiProviderStatusActive},
	}
	store.AiModel().(*mockstore.AiModelStore).ListActiveResult = []*model.AiModelM{}

	creds := map[string]Credential{}

	loader := NewLoader(registry, store, creds)
	err := loader.Load(context.Background())

	require.NoError(t, err)
	_, ok := registry.Get("openai")
	require.False(t, ok, "provider without credential should not be registered")
}

func TestLoader_Load_DBError_ReturnsError(t *testing.T) {
	store := mockstore.NewStore()
	registry := aipkg.NewRegistry()

	store.AiProvider().(*mockstore.AiProviderStore).ListActiveErr = assert.AnError

	loader := NewLoader(registry, store, map[string]Credential{})
	err := loader.Load(context.Background())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "list active providers")
}

func TestLoader_Reload_ClearsFirst(t *testing.T) {
	store := mockstore.NewStore()
	registry := aipkg.NewRegistry()

	store.AiProvider().(*mockstore.AiProviderStore).ListActiveResult = []*model.AiProviderM{
		{Name: "openai", Status: model.AiProviderStatusActive},
	}
	store.AiModel().(*mockstore.AiModelStore).ListActiveResult = []*model.AiModelM{
		{ProviderName: "openai", Model: "gpt-4o", Status: model.AiModelStatusActive},
	}

	creds := map[string]Credential{
		"openai": {APIKey: "test-key"},
	}

	loader := NewLoader(registry, store, creds)
	_ = loader.Load(context.Background())

	_, ok := registry.Get("openai")
	require.True(t, ok)

	store.AiProvider().(*mockstore.AiProviderStore).ListActiveResult = []*model.AiProviderM{}
	store.AiModel().(*mockstore.AiModelStore).ListActiveResult = []*model.AiModelM{}

	_ = loader.Reload(context.Background())

	_, ok = registry.Get("openai")
	require.False(t, ok, "provider should be removed after reload with empty data")
}
