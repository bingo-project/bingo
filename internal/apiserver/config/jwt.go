package config

type JWT struct {
	SecretKey string `mapstructure:"secretKey" json:"secretKey" yaml:"secretKey"`
	TTL       uint   `mapstructure:"ttl" json:"ttl" yaml:"ttl"`
}
