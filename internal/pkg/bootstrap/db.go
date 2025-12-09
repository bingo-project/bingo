package bootstrap

import (
	"gorm.io/gorm"

	"bingo/internal/pkg/facade"
	"bingo/internal/pkg/log"
	"bingo/internal/pkg/logger"
	"bingo/pkg/db"
)

func InitDB() *gorm.DB {
	ins, err := db.NewMySQL(facade.Config.Mysql, logger.New(facade.Config.Mysql.LogLevel))
	if err != nil {
		log.Fatalw("init store failed", "err", err)
	}

	return ins
}
