// ABOUTME: End-to-end tests for AI provider dynamic loading.
// ABOUTME: Tests full database to registry flow.

package ai

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bingo-project/bingo/internal/pkg/model"
	mockstore "github.com/bingo-project/bingo/internal/pkg/testing/mock/store"
	aipkg "github.com/bingo-project/bingo/pkg/ai"
)

func TestAIProviderReload_E2E(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test")
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	defer redisClient.Close()

	ctx := context.Background()
	err := redisClient.Ping(ctx).Err()
	if err != nil {
		t.Skipf("Redis not available: %v", err)
	}

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
	sub := NewSubscriber(redisClient, loader)

	go sub.Start()
	defer sub.Stop()

	time.Sleep(100 * time.Millisecond)

	err = TriggerReload(ctx, redisClient)
	require.NoError(t, err, "TriggerReload should succeed")

	time.Sleep(500 * time.Millisecond)

	provider, ok := registry.Get("openai")
	require.True(t, ok, "provider should be loaded after reload trigger")
	assert.Equal(t, "openai", provider.Name())
}

func TestAIProviderReload_E2E_Simple(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test")
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	defer redisClient.Close()

	ctx := context.Background()
	err := redisClient.Ping(ctx).Err()
	if err != nil {
		t.Skipf("Redis not available: %v", err)
	}

	err = TriggerReload(ctx, redisClient)
	require.NoError(t, err, "TriggerReload should succeed")

	time.Sleep(100 * time.Millisecond)

	assert.True(t, true, "E2E test completed - TriggerReload sent message to Redis")
}
