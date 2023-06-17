package bootstrap

import (
	"bingo/internal/apiserver/facade"
	"bingo/internal/pkg/log"
)

func InitLog() {
	log.Init(facade.Config.Log)
}
