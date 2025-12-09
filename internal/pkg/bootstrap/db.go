package bootstrap

import (
	"gorm.io/gorm"

	"github.com/bingo-project/bingo/internal/pkg/facade"
	"github.com/bingo-project/bingo/internal/pkg/log"
	"github.com/bingo-project/bingo/internal/pkg/logger"
	"github.com/bingo-project/bingo/pkg/db"
)

func InitDB() *gorm.DB {
	ins, err := db.NewMySQL(facade.Config.Mysql, logger.New(facade.Config.Mysql.LogLevel))
	if err != nil {
		log.Fatalw("init store failed", "err", err)
	}

	return ins
}
