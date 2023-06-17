package log

import (
	"go.uber.org/zap/zapcore"
)

// Options 包含与日志相关的配置项.
type Options struct {
	Level    string `mapstructure:"level" json:"level" yaml:"level"` // 指定日志级别，可选值：debug, info, warn, error, dpanic, panic, fatal
	Days     int    `mapstructure:"days" json:"days" yaml:"days"`
	Console  bool   `mapstructure:"console" json:"console" yaml:"console"`
	Format   string `mapstructure:"format" json:"format" yaml:"format"`       // 指定日志显示格式，可选值：console, json
	MaxSize  int    `mapstructure:"maxSize" json:"maxSize" yaml:"maxSize"`    // 日志文件大小限制，M
	Compress bool   `mapstructure:"compress" json:"compress" yaml:"compress"` // 是否使用 gz 压缩历史日志文件
	Path     string `mapstructure:"path" json:"path" yaml:"path"`             // 日志文件位置
}

// NewOptions 创建一个带有默认参数的 Options 对象.
func NewOptions() *Options {
	return &Options{
		Level:    zapcore.InfoLevel.String(),
		Days:     14,
		Console:  true,
		Format:   "console",
		MaxSize:  100,
		Compress: true,
		Path:     "_output/logs/apiserver.log",
	}
}
