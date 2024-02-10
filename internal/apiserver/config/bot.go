package config

type Bot struct {
	Enabled bool   `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
	Token   string `mapstructure:"token" json:"token" yaml:"token"`
}
