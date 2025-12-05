// ABOUTME: WebSocket server configuration.
// ABOUTME: Defines address and enabled settings for WebSocket.

package config

// WebSocket holds WebSocket server configuration.
type WebSocket struct {
	Enabled bool   `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
	Addr    string `mapstructure:"addr" json:"addr" yaml:"addr"`
}
