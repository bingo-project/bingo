package bootstrap

import (
	"bingo/internal/apiserver/config"
	genericapiserver "bingo/internal/pkg/server"
)

var CfgFile string

const (
	// DefaultConfigName 指定了服务的默认配置文件名.
	DefaultConfigName = "bingo-apiserver.yaml"
)

// InitConfig reads in config file and ENV variables if set.
func InitConfig() {
	genericapiserver.LoadConfig(CfgFile, DefaultConfigName, &config.Cfg, Boot)
}
