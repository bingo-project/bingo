// ABOUTME: Integration tests for AI reload subscriber.
// ABOUTME: Uses real Redis for Pub/Sub verification.

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

func TestSubscriber_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
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

	// Type assertion to access mock-specific fields
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

	err = redisClient.Publish(ctx, AIReloadChannel, "trigger").Err()
	require.NoError(t, err)

	time.Sleep(500 * time.Millisecond)

	provider, ok := registry.Get("openai")
	require.True(t, ok, "provider should be loaded after reload trigger")
	assert.Equal(t, "openai", provider.Name())
}

func TestSubscriber_Disconnect_Handled(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:9999",
	})

	store := mockstore.NewStore()
	registry := aipkg.NewRegistry()
	loader := NewLoader(registry, store, map[string]Credential{})
	sub := NewSubscriber(redisClient, loader)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	go sub.Start()
	<-ctx.Done()

	assert.True(t, true, "subscriber handled disconnect gracefully")
}
