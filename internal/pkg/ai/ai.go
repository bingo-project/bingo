// ABOUTME: Centralized AI component initialization.
// ABOUTME: Provides InitAI for server startup and GetRegistry for access.

package ai

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/bingo-project/bingo/internal/pkg/log"
	"github.com/bingo-project/bingo/internal/pkg/store"
	aipkg "github.com/bingo-project/bingo/pkg/ai"
)

var (
	globalRegistry *aipkg.Registry
	globalLoader   *Loader
)

// InitAI initializes AI registry and starts reload mechanisms.
// Called from initConfig() in each server.
// Returns the registry (or nil if no credentials configured) and any error.
func InitAI(redisClient *redis.Client, st store.IStore, creds map[string]Credential) (*aipkg.Registry, error) {
	if len(creds) == 0 {
		log.Info("No AI credentials configured, skipping AI initialization")

		return nil, nil
	}

	globalRegistry = aipkg.NewRegistry()
	globalLoader = NewLoader(globalRegistry, st, creds)

	if err := globalLoader.Load(context.Background()); err != nil {
		log.Errorw("Failed to load AI providers", "err", err)
	}

	if redisClient != nil {
		sub := NewSubscriber(redisClient, globalLoader)
		go sub.Start()
		log.Info("AI reload subscriber started")
	} else {
		go startPeriodicReload(globalLoader)
	}

	return globalRegistry, nil
}

// GetRegistry returns the global AI registry.
func GetRegistry() *aipkg.Registry {
	return globalRegistry
}

// TriggerReload sends a Redis pub/sub message to trigger reload across all services.
func TriggerReload(ctx context.Context, redis *redis.Client) error {
	return redis.Publish(ctx, AIReloadChannel, "trigger").Err()
}

// startPeriodicReload starts periodic polling as fallback when Redis is unavailable.
func startPeriodicReload(loader *Loader) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	log.Info("AI periodic reload started (5 minute interval)")

	for range ticker.C {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		if err := loader.Reload(ctx); err != nil {
			log.Errorw("AI periodic reload failed", "err", err)
		}
		cancel()
	}
}
