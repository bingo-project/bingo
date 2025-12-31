// ABOUTME: Unit tests for the AI retry mechanism.
// ABOUTME: Verifies exponential backoff and retriable error handling.

package ai

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDo_Retry(t *testing.T) {
	config := RetryConfig{
		MaxAttempts: 3,
		BaseDelay:   10 * time.Millisecond,
		MaxDelay:    50 * time.Millisecond,
		Multiplier:  2.0,
	}

	t.Run("success after retry", func(t *testing.T) {
		attempts := 0
		errRetriable := errors.New("503 service unavailable")

		resp, err := Do(context.Background(), config, func(ctx context.Context) (*ChatResponse, error) {
			attempts++
			if attempts < 2 {
				return nil, errRetriable
			}
			return &ChatResponse{ID: "test-id"}, nil
		})

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "test-id", resp.ID)
		assert.Equal(t, 2, attempts)
	})

	t.Run("max attempts reached", func(t *testing.T) {
		attempts := 0
		errRetriable := errors.New("429 too many requests")

		resp, err := Do(context.Background(), config, func(ctx context.Context) (*ChatResponse, error) {
			attempts++
			return nil, errRetriable
		})

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "429")
		assert.Equal(t, 3, attempts)
	})

	t.Run("non-retriable error", func(t *testing.T) {
		attempts := 0
		errPermanent := errors.New("401 unauthorized")

		resp, err := Do(context.Background(), config, func(ctx context.Context) (*ChatResponse, error) {
			attempts++
			return nil, errPermanent
		})

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, 1, attempts)
	})
}
