package bootstrap

import (
	genericapiserver "bingo/internal/pkg/server"
	"bingo/internal/scheduler/facade"
)

var CfgFile string

const (
	// DefaultConfigName 指定了服务的默认配置文件名.
	DefaultConfigName = "bingo-scheduler.yaml"
)

// InitConfig reads in config file and ENV variables if set.
func InitConfig() {
	genericapiserver.LoadConfig(CfgFile, DefaultConfigName, &facade.Config, Boot)
}
