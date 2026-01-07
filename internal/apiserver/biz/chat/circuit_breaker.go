// ABOUTME: Circuit breaker for AI providers to prevent cascading failures.
// ABOUTME: Implements three states: Closed, Open, Half-Open.

package chat

import (
	"context"
	"sync"
	"time"

	"github.com/bingo-project/bingo/internal/pkg/log"
)

// CircuitBreakerState represents the state of a circuit breaker.
type CircuitBreakerState int

const (
	// CircuitClosed is the normal state where the circuit is "closed" (like a closed electrical circuit),
	// meaning current can flow and requests are allowed through.
	//
	// Note: This follows the standard Circuit Breaker pattern terminology (Martin Fowler, 2002).
	// The naming is based on electrical circuit analogy, not natural language.
	// Closed circuit = connected = operational.
	CircuitClosed CircuitBreakerState = iota

	// CircuitOpen is the tripped state where the circuit is "open" (like an open electrical circuit),
	// meaning current cannot flow and requests are immediately rejected.
	//
	// Open circuit = disconnected = not operational.
	CircuitOpen

	// CircuitHalfOpen allows a test request to check if the service has recovered.
	// Success transitions back to Closed, failure trips back to Open.
	CircuitHalfOpen
)

func (s CircuitBreakerState) String() string {
	switch s {
	case CircuitClosed:
		return "closed"
	case CircuitOpen:
		return "open"
	case CircuitHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// CircuitBreakerConfig holds configuration for a circuit breaker.
type CircuitBreakerConfig struct {
	// MaxFailures is the number of consecutive failures before opening.
	MaxFailures int
	// OpenTimeout is how long to stay in Open state before trying Half-Open.
	OpenTimeout time.Duration
	// SuccessThreshold is consecutive successes needed to close in Half-Open.
	SuccessThreshold int
}

// DefaultCircuitBreakerConfig provides sensible defaults.
var DefaultCircuitBreakerConfig = CircuitBreakerConfig{
	MaxFailures:      5,
	OpenTimeout:      60 * time.Second,
	SuccessThreshold: 2,
}

// CircuitBreaker prevents cascading failures by tripping after consecutive errors.
type CircuitBreaker struct {
	mu    sync.Mutex
	state CircuitBreakerState

	// failure count in current state
	failures int
	// success count in Half-Open
	successes int

	// last state change time
	lastStateChange time.Time

	cfg  CircuitBreakerConfig
	name string
}

// NewCircuitBreaker creates a new circuit breaker with the given name.
func NewCircuitBreaker(name string, cfg CircuitBreakerConfig) *CircuitBreaker {
	if cfg.MaxFailures <= 0 {
		cfg.MaxFailures = DefaultCircuitBreakerConfig.MaxFailures
	}
	if cfg.OpenTimeout <= 0 {
		cfg.OpenTimeout = DefaultCircuitBreakerConfig.OpenTimeout
	}
	if cfg.SuccessThreshold <= 0 {
		cfg.SuccessThreshold = DefaultCircuitBreakerConfig.SuccessThreshold
	}

	return &CircuitBreaker{
		state:           CircuitClosed,
		lastStateChange: time.Now(),
		cfg:             cfg,
		name:            name,
	}
}

// Allow returns true if the request should be allowed through.
func (cb *CircuitBreaker) Allow(ctx context.Context) bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	// Check if we should transition from Open to Half-Open
	if cb.state == CircuitOpen && time.Since(cb.lastStateChange) >= cb.cfg.OpenTimeout {
		cb.setState(CircuitHalfOpen)
		log.C(ctx).Infow("circuit breaker half-open", "name", cb.name)
	}

	return cb.state != CircuitOpen
}

// RecordSuccess records a successful call.
func (cb *CircuitBreaker) RecordSuccess(ctx context.Context) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.state == CircuitHalfOpen {
		cb.successes++
		if cb.successes >= cb.cfg.SuccessThreshold {
			cb.setState(CircuitClosed)
			cb.failures = 0
			cb.successes = 0
			log.C(ctx).Infow("circuit breaker closed", "name", cb.name)
		}
	} else {
		// Reset failure count on success in Closed state
		cb.failures = 0
	}
}

// RecordFailure records a failed call.
func (cb *CircuitBreaker) RecordFailure(ctx context.Context, err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures++

	// Trip to Open if threshold exceeded
	if cb.failures >= cb.cfg.MaxFailures {
		if cb.state != CircuitOpen {
			cb.setState(CircuitOpen)
			log.C(ctx).Warnw("circuit breaker opened",
				"name", cb.name,
				"failures", cb.failures,
				"error", err.Error())
		}
	} else if cb.state == CircuitHalfOpen {
		// Immediately reopen on failure in Half-Open
		cb.setState(CircuitOpen)
		cb.successes = 0
		log.C(ctx).Warnw("circuit breaker reopened", "name", cb.name, "error", err.Error())
	}
}

// State returns the current state for monitoring.
func (cb *CircuitBreaker) State() CircuitBreakerState {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	return cb.state
}

func (cb *CircuitBreaker) setState(state CircuitBreakerState) {
	cb.state = state
	cb.lastStateChange = time.Now()

	// Update metrics for monitoring
	// Extract provider name from "provider:xxx" format
	providerName := cb.name
	if len(cb.name) > 9 && cb.name[:9] == "provider:" {
		providerName = cb.name[9:]
	}
	SetCircuitBreakerState(providerName, state)
}
