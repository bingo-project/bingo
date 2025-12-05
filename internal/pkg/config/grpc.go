// ABOUTME: gRPC server configuration.
// ABOUTME: Defines enabled and address settings for gRPC.

package config

// GRPC holds gRPC server configuration.
type GRPC struct {
	Enabled bool   `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
	Addr    string `mapstructure:"addr" json:"addr" yaml:"addr"`
}
