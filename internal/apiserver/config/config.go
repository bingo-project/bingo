package config

import (
	"github.com/bingo-project/component-base/log"

	"bingo/internal/pkg/config"
	"bingo/pkg/db"
	"bingo/pkg/mail"
)

type Config struct {
	Server  *config.Server   `mapstructure:"server" json:"server" yaml:"server"`
	GRPC    *config.GRPC     `mapstructure:"grpc" json:"grpc" yaml:"grpc"`
	Bot     *config.Bot      `mapstructure:"bot" json:"bot" yaml:"bot"`
	JWT     *config.JWT      `mapstructure:"jwt" json:"jwt" yaml:"jwt"`
	Feature *config.Feature  `mapstructure:"feature" json:"feature" yaml:"feature"`
	Mysql   *db.MySQLOptions `mapstructure:"mysql" json:"mysql" yaml:"mysql"`
	Redis   *config.Redis    `mapstructure:"redis" json:"redis" yaml:"redis"`
	Log     *log.Options     `mapstructure:"log" json:"log" yaml:"log"`
	Mail    *mail.Options    `mapstructure:"mail" json:"mail" yaml:"mail"`
	Code    config.Code      `mapstructure:"code" json:"code" yaml:"code"`
	OpenAPI config.OpenAPI   `mapstructure:"openapi" json:"openapi" yaml:"openapi"`
}
