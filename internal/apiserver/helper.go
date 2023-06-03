package apiserver

import (
	"github.com/goer-project/goer-core/redis"

	"bingo/internal/apiserver/cache"
	"bingo/internal/apiserver/config"
	"bingo/internal/apiserver/store"
	genericapiserver "bingo/internal/pkg/server"
	"bingo/pkg/db"
)

var CfgFile string

const (
	// DefaultConfigName 指定了服务的默认配置文件名.
	DefaultConfigName = "bingo-apiserver.yaml"
)

// InitConfig reads in config file and ENV variables if set.
func InitConfig() {
	genericapiserver.LoadConfig(CfgFile, DefaultConfigName, &config.Cfg)
}

// InitStore 读取 db 配置，创建 gorm.DB 实例，并初始化 store 层.
func InitStore() error {
	ins, err := db.NewMySQL(config.Cfg.Mysql)
	if err != nil {
		return err
	}

	_ = store.NewStore(ins)

	return nil
}

func InitCache() error {
	r, err := redis.NewClient(config.Cfg.Redis.Host, config.Cfg.Redis.Password, config.Cfg.Redis.Database)
	if err != nil {
		return err
	}

	cache.NewCache(r)

	return nil
}
