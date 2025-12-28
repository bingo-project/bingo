// ABOUTME: Root configuration structure for the application.
// ABOUTME: Aggregates all configuration sections including app, protocols, and services.

package config

import (
	"github.com/bingo-project/bingo/internal/pkg/log"
	"github.com/bingo-project/bingo/pkg/db"
	"github.com/bingo-project/bingo/pkg/mail"
)

type Config struct {
	App       *App             `mapstructure:"app" json:"app" yaml:"app"`
	HTTP      *HTTP            `mapstructure:"http" json:"http" yaml:"http"`
	GRPC      *GRPC            `mapstructure:"grpc" json:"grpc" yaml:"grpc"`
	WebSocket *WebSocket       `mapstructure:"websocket" json:"websocket" yaml:"websocket"`
	Bot       *Bot             `mapstructure:"bot" json:"bot" yaml:"bot"`
	Auth      *Auth            `mapstructure:"auth" json:"auth" yaml:"auth"`
	JWT       *JWT             `mapstructure:"jwt" json:"jwt" yaml:"jwt"`
	Feature   *Feature         `mapstructure:"feature" json:"feature" yaml:"feature"`
	Mysql     *db.MySQLOptions `mapstructure:"mysql" json:"mysql" yaml:"mysql"`
	Redis     *Redis           `mapstructure:"redis" json:"redis" yaml:"redis"`
	Log       *log.Options     `mapstructure:"log" json:"log" yaml:"log"`
	Mail      *mail.Options    `mapstructure:"mail" json:"mail" yaml:"mail"`
	Code      Code             `mapstructure:"code" json:"code" yaml:"code"`
	OpenAPI   OpenAPI          `mapstructure:"openapi" json:"openapi" yaml:"openapi"`
}
