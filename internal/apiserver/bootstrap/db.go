package bootstrap

import (
	"bingo/internal/apiserver/config"
	"bingo/internal/apiserver/store"
	"bingo/internal/pkg/log"
	"bingo/pkg/db"
)

// InitStore 读取 db 配置，创建 gorm.DB 实例，并初始化 store 层.
func InitStore() {
	ins, err := db.NewMySQL(config.Cfg.Mysql)
	if err != nil {
		log.Errorw("init store failed", "err", err)

		return
	}

	_ = store.NewStore(ins)
}
