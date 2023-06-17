package config

import (
	"github.com/bingo-project/component-base/log"

	"bingo/pkg/db"
)

type Config struct {
	Server  *Server          `mapstructure:"server" json:"server" yaml:"server"`
	JWT     *JWT             `mapstructure:"jwt" json:"jwt" yaml:"jwt"`
	Feature *Feature         `mapstructure:"feature" json:"feature" yaml:"feature"`
	Mysql   *db.MySQLOptions `mapstructure:"mysql" json:"mysql" yaml:"mysql"`
	Redis   *Redis           `mapstructure:"redis" json:"redis" yaml:"redis"`
	Log     *log.Options     `mapstructure:"log" json:"log" yaml:"log"`
}
