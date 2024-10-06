package bootstrap

import (
	"github.com/bingo-project/component-base/log"

	"bingo/internal/scheduler/facade"
	"bingo/internal/scheduler/store"
	"bingo/pkg/db"
)

// InitStore 读取 db 配置，创建 gorm.DB 实例，并初始化 store 层.
func InitStore() {
	ins, err := db.NewMySQL(facade.Config.Mysql)
	if err != nil {
		log.Errorw("init store failed", "err", err)

		return
	}

	_ = store.NewStore(ins)
}
