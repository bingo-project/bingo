// ABOUTME: Implementation of the AI retry mechanism with exponential backoff.
// ABOUTME: Handles transient errors such as 429 (rate limit) and 503 (service unavailable).

package ai

import (
	"context"
	"math"
	"strings"
	"time"
)

// RetryConfig defines the configuration for the retry mechanism
type RetryConfig struct {
	MaxAttempts int           // Maximum number of attempts
	BaseDelay   time.Duration // Initial delay between retries
	MaxDelay    time.Duration // Maximum delay between retries
	Multiplier  float64       // Factor by which the delay increases after each retry
}

// DefaultRetryConfig provides a standard configuration for AI retries
var DefaultRetryConfig = RetryConfig{
	MaxAttempts: 3,
	BaseDelay:   500 * time.Millisecond,
	MaxDelay:    10 * time.Second,
	Multiplier:  2.0,
}

// Do executes an AI operation with retries based on the provided configuration
func Do(ctx context.Context, cfg RetryConfig, fn func(ctx context.Context) (*ChatResponse, error)) (*ChatResponse, error) {
	var lastErr error
	var resp *ChatResponse

	for attempt := 1; attempt <= cfg.MaxAttempts; attempt++ {
		// Check context before each attempt
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		resp, lastErr = fn(ctx)
		if lastErr == nil {
			return resp, nil
		}

		// If this is the last attempt or error is not retriable, stop retrying
		if attempt == cfg.MaxAttempts || !isRetriable(lastErr) {
			return nil, lastErr
		}

		// Calculate backoff delay
		delay := time.Duration(float64(cfg.BaseDelay) * math.Pow(cfg.Multiplier, float64(attempt-1)))
		if delay > cfg.MaxDelay {
			delay = cfg.MaxDelay
		}

		// Wait for delay or context cancellation
		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	return nil, lastErr
}

// isRetriable determines if an error is worth retrying
func isRetriable(err error) bool {
	if err == nil {
		return false
	}

	// Network errors or specific HTTP status codes
	errMsg := strings.ToLower(err.Error())
	retriableMessages := []string{
		"429",                   // Too Many Requests
		"503",                   // Service Unavailable
		"502",                   // Bad Gateway
		"504",                   // Gateway Timeout
		"timeout",               // General timeout
		"deadline exceeded",     // Context deadline exceeded
		"connection refused",    // Network issue
		"connection reset",      // Network issue
		"request_timeout_error", // Specific provider error
		"rate_limit_reached",    // Specific provider error
		"overloaded",            // Specific provider error (e.g. Claude)
	}

	for _, m := range retriableMessages {
		if strings.Contains(errMsg, m) {
			return true
		}
	}

	return false
}
