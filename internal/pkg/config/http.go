// ABOUTME: HTTP server configuration.
// ABOUTME: Defines enabled, address, gin mode, and protocol mode settings.

package config

// HTTP holds HTTP server configuration.
type HTTP struct {
	Enabled bool   `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
	Addr    string `mapstructure:"addr" json:"addr" yaml:"addr"`
	GinMode string `mapstructure:"ginMode" json:"ginMode" yaml:"ginMode"` // release, debug, test
	Mode    string `mapstructure:"mode" json:"mode" yaml:"mode"`          // standalone, gateway
}

// IsProduction returns true if gin mode is release.
func (h HTTP) IsProduction() bool {
	return h.GinMode == "release"
}
