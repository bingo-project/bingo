// ABOUTME: Configuration for WebSocket Hub.
// ABOUTME: Defines timeout and cleanup intervals.

package ws

import "time"

// HubConfig holds configuration for the Hub.
type HubConfig struct {
	// Anonymous connection timeout (must login within this time)
	AnonymousTimeout time.Duration
	// Anonymous connection cleanup interval
	AnonymousCleanup time.Duration

	// Authenticated connection heartbeat timeout
	HeartbeatTimeout time.Duration
	// Authenticated connection cleanup interval
	HeartbeatCleanup time.Duration

	// WebSocket protocol ping period
	PingPeriod time.Duration
	// WebSocket protocol pong wait timeout
	PongWait time.Duration
}

// DefaultHubConfig returns default configuration.
func DefaultHubConfig() *HubConfig {
	return &HubConfig{
		AnonymousTimeout: 10 * time.Second,
		AnonymousCleanup: 2 * time.Second,
		HeartbeatTimeout: 60 * time.Second,
		HeartbeatCleanup: 30 * time.Second,
		PingPeriod:       54 * time.Second,
		PongWait:         60 * time.Second,
	}
}
