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
	AI        AIConfig         `mapstructure:"ai" json:"ai" yaml:"ai"`
}

// AIConfig AI 模块配置
type AIConfig struct {
	DefaultModel string                  `mapstructure:"default-model" json:"defaultModel" yaml:"default-model"`
	Credentials  map[string]AICredential `mapstructure:"credentials" json:"credentials" yaml:"credentials"`
	Session      AISessionConfig         `mapstructure:"session" json:"session" yaml:"session"`
	Quota        AIQuotaConfig           `mapstructure:"quota" json:"quota" yaml:"quota"`
}

// AICredential Provider 凭证
type AICredential struct {
	APIKey  string `mapstructure:"api-key" json:"apiKey" yaml:"api-key"`
	BaseURL string `mapstructure:"base-url" json:"baseURL" yaml:"base-url"`
}

// AISessionConfig 会话配置
type AISessionConfig struct {
	MaxMessages   int `mapstructure:"max-messages" json:"maxMessages" yaml:"max-messages"`       // 单会话最大消息数
	MaxTokens     int `mapstructure:"max-tokens" json:"maxTokens" yaml:"max-tokens"`             // 单次请求最大 token
	ContextWindow int `mapstructure:"context-window" json:"contextWindow" yaml:"context-window"` // 上下文窗口大小
}

// AIQuotaConfig 配额配置
type AIQuotaConfig struct {
	Enabled    bool `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
	DefaultRPM int  `mapstructure:"default-rpm" json:"defaultRpm" yaml:"default-rpm"` // 默认 RPM
	DefaultTPD int  `mapstructure:"default-tpd" json:"defaultTpd" yaml:"default-tpd"` // 默认 TPD
}
