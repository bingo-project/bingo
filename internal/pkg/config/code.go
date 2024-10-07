package config

type Code struct {
	Length  uint `mapstructure:"length" json:"length" yaml:"length"`
	TTL     uint `mapstructure:"ttl" json:"ttl" yaml:"ttl"`
	Waiting uint `mapstructure:"waiting" json:"waiting" yaml:"waiting"`
}
