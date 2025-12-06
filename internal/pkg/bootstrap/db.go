package bootstrap

import (
	"gorm.io/gorm"

	"bingo/internal/pkg/facade"
	"bingo/internal/pkg/log"
	"bingo/pkg/db"
)

func InitDB() *gorm.DB {
	ins, err := db.NewMySQL(facade.Config.Mysql)
	if err != nil {
		log.Fatalw("init store failed", "err", err)
	}

	return ins
}
