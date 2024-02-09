package config

import (
	"github.com/bingo-project/component-base/log"

	"bingo/pkg/db"
	"bingo/pkg/mail"
)

type Config struct {
	Server  *Server          `mapstructure:"server" json:"server" yaml:"server"`
	GRPC    *GRPC            `mapstructure:"grpc" json:"grpc" yaml:"grpc"`
	JWT     *JWT             `mapstructure:"jwt" json:"jwt" yaml:"jwt"`
	Feature *Feature         `mapstructure:"feature" json:"feature" yaml:"feature"`
	Mysql   *db.MySQLOptions `mapstructure:"mysql" json:"mysql" yaml:"mysql"`
	Redis   *Redis           `mapstructure:"redis" json:"redis" yaml:"redis"`
	Log     *log.Options     `mapstructure:"log" json:"log" yaml:"log"`
	Mail    *mail.Options    `mapstructure:"mail" json:"mail" yaml:"mail"`
	Code    Code             `mapstructure:"code" json:"code" yaml:"code"`
}
