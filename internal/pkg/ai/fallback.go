// ABOUTME: AI model fallback selection logic.
// ABOUTME: Handles model degradation when primary model is unavailable.

package ai

import (
	"context"

	"github.com/bingo-project/bingo/internal/pkg/log"
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

// SelectFallback returns the next available model for fallback.
// Returns empty string if no fallback model is available.
func (s *FallbackSelector) SelectFallback(ctx context.Context, originalModel string) string {
	models, err := s.store.ListActive(ctx)
	if err != nil {
		log.C(ctx).Warnw("Failed to list models for fallback", "err", err)

		return ""
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
		// Check if model is registered in Registry
		if _, ok := s.registry.GetByModel(m.Model); ok {
			log.C(ctx).Infow("AI model fallback selected",
				"original", originalModel,
				"fallback", m.Model)

			return m.Model
		}
	}

	return ""
}
