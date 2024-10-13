package bootstrap

import (
	"bingo/internal/pkg/facade"
	genericapiserver "bingo/internal/pkg/server"
)

var CfgFile string

const (
	// DefaultConfigName 指定了服务的默认配置文件名.
	DefaultConfigName = "bingo-apiserver.yaml"
)

// InitConfig reads in config file and ENV variables if set.
func InitConfig(configName string) {
	if configName == "" {
		configName = DefaultConfigName
	}

	genericapiserver.LoadConfig(CfgFile, configName, &facade.Config, Boot)
}
