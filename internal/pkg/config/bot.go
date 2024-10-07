package config

type Bot struct {
	Telegram string `mapstructure:"telegram" json:"telegram" yaml:"telegram"`
	Discord  string `mapstructure:"discord" json:"discord" yaml:"discord"`
}
