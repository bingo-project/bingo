// ABOUTME: gRPC server configuration.
// ABOUTME: Defines enabled, address, and TLS settings for gRPC.

package config

// GRPC holds gRPC server configuration.
type GRPC struct {
	Enabled bool     `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
	Addr    string   `mapstructure:"addr" json:"addr" yaml:"addr"`
	TLS     *TLSConfig `mapstructure:"tls" json:"tls" yaml:"tls"`
}

// TLSConfig holds TLS configuration for secure connections.
type TLSConfig struct {
	Enabled  bool   `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
	CertFile string `mapstructure:"certFile" json:"certFile" yaml:"certFile"`
	KeyFile  string `mapstructure:"keyFile" json:"keyFile" yaml:"keyFile"`
}

// IsInsecure returns true if TLS is not enabled.
func (g *GRPC) IsInsecure() bool {
	return g.TLS == nil || !g.TLS.Enabled
}
