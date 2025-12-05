// ABOUTME: WebSocket server configuration.
// ABOUTME: Defines address, origin validation, and enabled settings for WebSocket.

package config

// WebSocket holds WebSocket server configuration.
type WebSocket struct {
	Enabled        bool     `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
	Addr           string   `mapstructure:"addr" json:"addr" yaml:"addr"`
	AllowedOrigins []string `mapstructure:"allowedOrigins" json:"allowedOrigins" yaml:"allowedOrigins"`
}

// AllowAllOrigins returns true if no origin restrictions are configured.
func (w *WebSocket) AllowAllOrigins() bool {
	return len(w.AllowedOrigins) == 0
}

// IsOriginAllowed checks if the given origin is in the allowed list.
func (w *WebSocket) IsOriginAllowed(origin string) bool {
	if w.AllowAllOrigins() {
		return true
	}
	for _, allowed := range w.AllowedOrigins {
		if allowed == "*" || allowed == origin {
			return true
		}
	}
	return false
}
