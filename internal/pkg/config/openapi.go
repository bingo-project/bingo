package config

type OpenAPI struct {
	Enabled bool  `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
	Nonce   bool  `mapstructure:"nonce" json:"nonce" yaml:"nonce"`
	TTL     int64 `mapstructure:"ttl" json:"ttl" yaml:"ttl"`
}
