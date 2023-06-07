package config

import (
	"bingo/internal/pkg/log"
	"bingo/pkg/db"
)

var (
	Cfg *Config
)

type Config struct {
	Server  *Server          `mapstructure:"server" json:"server" yaml:"server"`
	JWT     *JWT             `mapstructure:"jwt" json:"jwt" yaml:"jwt"`
	Feature *Feature         `mapstructure:"feature" json:"feature" yaml:"feature"`
	Mysql   *db.MySQLOptions `mapstructure:"mysql" json:"mysql" yaml:"mysql"`
	Redis   *Redis           `mapstructure:"redis" json:"redis" yaml:"redis"`
	Log     *log.Options     `mapstructure:"log" json:"log" yaml:"log"`
}

type Server struct {
	Mode string `mapstructure:"mode" json:"mode" yaml:"mode"`
	Addr string `mapstructure:"addr" json:"addr" yaml:"addr"`
}

type JWT struct {
	SecretKey string `mapstructure:"secretKey" json:"secretKey" yaml:"secretKey"`
	TTL       uint   `mapstructure:"ttl" json:"ttl" yaml:"ttl"`
}

type Feature struct {
	ApiDoc bool `mapstructure:"api-doc" json:"api-doc" yaml:"api-doc"`
}

type Redis struct {
	Host     string `mapstructure:"host" json:"host" yaml:"host"`
	Username string `mapstructure:"username" json:"username" yaml:"username"`
	Password string `mapstructure:"password" json:"-" yaml:"password"`
	Database int    `mapstructure:"database" json:"database" yaml:"database"`
}
