package bootstrap

import (
	"github.com/bingo-project/component-base/log"
	"gorm.io/gorm"

	"bingo/internal/pkg/facade"
	"bingo/internal/scheduler/store"
	"bingo/pkg/db"
)

func InitDB() *gorm.DB {
	ins, err := db.NewMySQL(facade.Config.Mysql)
	if err != nil {
		log.Fatalw("init store failed", "err", err)
	}

	return ins
}

// InitStore 读取 db 配置，创建 gorm.DB 实例，并初始化 store 层.
func InitStore() {
	ins, err := db.NewMySQL(facade.Config.Mysql)
	if err != nil {
		log.Errorw("init store failed", "err", err)

		return
	}

	_ = store.NewStore(ins)
}
