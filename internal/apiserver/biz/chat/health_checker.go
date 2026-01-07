// ABOUTME: Provider health checker for monitoring AI service availability.
// ABOUTME: Periodically pings providers to detect failures early.

package chat

import (
	"context"
	"sync"
	"time"

	"github.com/bingo-project/bingo/internal/pkg/log"
	aipkg "github.com/bingo-project/bingo/pkg/ai"
)

// HealthStatus represents the health status of a provider.
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusUnknown   HealthStatus = "unknown"
)

// ProviderHealth tracks the health of a single provider.
type ProviderHealth struct {
	ProviderName string
	Status       HealthStatus
	LastCheck    time.Time
	LastError    error
}

// HealthChecker performs periodic health checks on AI providers.
type HealthChecker struct {
	registry *aipkg.Registry
	health   map[string]*ProviderHealth
	mu       sync.RWMutex

	checkInterval time.Duration
	checkTimeout  time.Duration
	stopCh        chan struct{}
}

// NewHealthChecker creates a new health checker.
func NewHealthChecker(registry *aipkg.Registry) *HealthChecker {
	return &HealthChecker{
		registry:      registry,
		health:        make(map[string]*ProviderHealth),
		checkInterval: 5 * time.Minute,
		checkTimeout:  30 * time.Second,
		stopCh:        make(chan struct{}),
	}
}

// Start begins periodic health checks.
func (h *HealthChecker) Start() {
	log.Info("Starting AI provider health checker")

	ticker := time.NewTicker(h.checkInterval)
	defer ticker.Stop()

	// Run first check immediately
	h.checkAll(context.Background())

	for {
		select {
		case <-ticker.C:
			h.checkAll(context.Background())
		case <-h.stopCh:
			log.Info("AI provider health checker stopped")

			return
		}
	}
}

// Stop stops the health checker.
func (h *HealthChecker) Stop() {
	close(h.stopCh)
}

// checkAll checks health of all registered providers.
func (h *HealthChecker) checkAll(ctx context.Context) {
	providers := h.registry.ListProviders()
	for _, providerName := range providers {
		if provider, ok := h.registry.Get(providerName); ok {
			status, lastErr := h.checkProvider(ctx, provider)
			h.updateHealth(providerName, status, lastErr)
		}
	}
}

// checkProvider performs a health check on a single provider.
// It sends a minimal request to verify the provider is responsive.
func (h *HealthChecker) checkProvider(ctx context.Context, provider aipkg.Provider) (HealthStatus, error) {
	checkCtx, cancel := context.WithTimeout(ctx, h.checkTimeout)
	defer cancel()

	// Send a minimal chat request to test connectivity
	// Use a very short message to minimize cost
	testReq := &aipkg.ChatRequest{
		Model: "test",
		Messages: []aipkg.Message{
			{Role: aipkg.RoleUser, Content: "hi"},
		},
		MaxTokens: 5,
	}

	_, err := provider.Chat(checkCtx, testReq)
	if err != nil {
		// Check if it's an actual error or just a "model not found" type error
		// Some providers return 404 for unknown models, which means the provider is up
		if isProviderAvailableError(err) {
			return HealthStatusHealthy, nil
		}

		return HealthStatusUnhealthy, err
	}

	return HealthStatusHealthy, nil
}

// updateHealth updates the health status for a provider.
func (h *HealthChecker) updateHealth(providerName string, status HealthStatus, lastErr error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, exists := h.health[providerName]; !exists {
		h.health[providerName] = &ProviderHealth{
			ProviderName: providerName,
		}
	}

	h.health[providerName].Status = status
	h.health[providerName].LastCheck = time.Now()
	h.health[providerName].LastError = lastErr

	if status == HealthStatusUnhealthy {
		log.Warnw("AI provider health check failed",
			"provider", providerName,
			"error", lastErr)
	}
}

// GetHealth returns the current health status of all providers.
func (h *HealthChecker) GetHealth() map[string]*ProviderHealth {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// Return a copy to avoid concurrent modification
	result := make(map[string]*ProviderHealth, len(h.health))
	for k, v := range h.health {
		result[k] = &ProviderHealth{
			ProviderName: v.ProviderName,
			Status:       v.Status,
			LastCheck:    v.LastCheck,
			LastError:    v.LastError,
		}
	}

	return result
}

// GetProviderHealth returns the health status of a specific provider.
func (h *HealthChecker) GetProviderHealth(providerName string) *ProviderHealth {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if h, exists := h.health[providerName]; exists {
		return &ProviderHealth{
			ProviderName: h.ProviderName,
			Status:       h.Status,
			LastCheck:    h.LastCheck,
			LastError:    h.LastError,
		}
	}

	return &ProviderHealth{
		ProviderName: providerName,
		Status:       HealthStatusUnknown,
	}
}

// isProviderAvailableError checks if an error indicates the provider is available
// but the request failed (e.g., model not found vs connection error).
func isProviderAvailableError(err error) bool {
	if err == nil {
		return true
	}

	errMsg := err.Error()
	// These errors indicate the provider is up but the request was invalid
	availableErrors := []string{
		"model not found",
		"invalid_model",
		"404",
	}

	for _, available := range availableErrors {
		if contains(errMsg, available) {
			return true
		}
	}

	return false
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsInner(s, substr))
}

func containsInner(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}

	return false
}
