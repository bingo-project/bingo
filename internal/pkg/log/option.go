package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Options struct {
	Level    string `mapstructure:"level" json:"level" yaml:"level"`
	Days     int    `mapstructure:"days" json:"days" yaml:"days"`
	Console  bool   `mapstructure:"console" json:"console" yaml:"console"`
	Format   string `mapstructure:"format" json:"format" yaml:"format"`
	MaxSize  int    `mapstructure:"maxSize" json:"maxSize" yaml:"maxSize"`
	Compress bool   `mapstructure:"compress" json:"compress" yaml:"compress"`
	Path     string `mapstructure:"path" json:"path" yaml:"path"`
}

func NewDefaultOptions() *Options {
	return &Options{
		Level:    zapcore.InfoLevel.String(),
		Days:     14,
		Console:  true,
		Format:   "console",
		MaxSize:  50,
		Compress: false,
		Path:     "/dev/null",
	}
}

func (l *zapLogger) WithOption(opts zap.Option) *zapLogger {
	lc := l.clone()

	lc.z = lc.z.WithOptions(opts)

	return lc
}
