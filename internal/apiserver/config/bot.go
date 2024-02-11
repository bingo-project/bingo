package config

type Bot struct {
	Enabled  bool   `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
	Telegram string `mapstructure:"telegram" json:"telegram" yaml:"telegram"`
}
