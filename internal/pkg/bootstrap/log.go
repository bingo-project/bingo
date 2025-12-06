package bootstrap

import (
	"bingo/internal/pkg/facade"
	"bingo/internal/pkg/log"
)

func InitLog() {
	log.Init(facade.Config.Log)
}
