// ABOUTME: AI model fallback selection logic.
// ABOUTME: Handles model degradation when primary model is unavailable.

package ai

import (
	"context"

	"github.com/bingo-project/bingo/internal/pkg/log"
	"github.com/bingo-project/bingo/internal/pkg/model"
	"github.com/bingo-project/bingo/internal/pkg/store"
	aipkg "github.com/bingo-project/bingo/pkg/ai"
)

// FallbackSelector handles model fallback selection
type FallbackSelector struct {
	store    store.AiModelStore
	registry *aipkg.Registry
}

// NewFallbackSelector creates a new FallbackSelector
func NewFallbackSelector(store store.AiModelStore, registry *aipkg.Registry) *FallbackSelector {
	return &FallbackSelector{
		store:    store,
		registry: registry,
	}
}

// SelectFallback returns the next available model configuration for fallback.
// Returns nil if no fallback model is available.
func (s *FallbackSelector) SelectFallback(ctx context.Context, originalModel string) *model.AiModelM {
	models, err := s.store.ListActive(ctx)
	if err != nil {
		log.C(ctx).Warnw("Failed to list models for fallback", "err", err)

		return nil
	}

	for _, m := range models {
		// Skip original model
		if m.Model == originalModel {
			continue
		}
		// Check if fallback is allowed
		if !m.AllowFallback {
			continue
		}
		// Check if provider is registered in Registry
		if _, ok := s.registry.Get(m.ProviderName); ok {
			log.C(ctx).Infow("AI model fallback selected",
				"original", originalModel,
				"fallback", m.Model,
				"provider", m.ProviderName)

			return m
		}
	}

	return nil
}
