package bootstrap

import (
	"github.com/bingo-project/component-base/log"

	"bingo/internal/apiserver/facade"
)

func InitLog() {
	log.Init(facade.Config.Log)
}
