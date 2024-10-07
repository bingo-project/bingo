package bootstrap

import (
	"github.com/bingo-project/component-base/log"

	"bingo/internal/pkg/facade"
)

func InitLog() {
	log.Init(facade.Config.Log)
}
