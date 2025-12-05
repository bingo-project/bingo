// ABOUTME: Tests for Hub configuration.
// ABOUTME: Validates default config values.

package ws

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultHubConfig(t *testing.T) {
	cfg := DefaultHubConfig()

	assert.Equal(t, 10*time.Second, cfg.AnonymousTimeout)
	assert.Equal(t, 2*time.Second, cfg.AnonymousCleanup)
	assert.Equal(t, 60*time.Second, cfg.HeartbeatTimeout)
	assert.Equal(t, 30*time.Second, cfg.HeartbeatCleanup)
	assert.Equal(t, 54*time.Second, cfg.PingPeriod)
	assert.Equal(t, 60*time.Second, cfg.PongWait)
}
