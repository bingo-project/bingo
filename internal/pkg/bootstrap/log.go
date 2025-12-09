package bootstrap

import (
	"github.com/bingo-project/bingo/internal/pkg/facade"
	"github.com/bingo-project/bingo/internal/pkg/log"
)

func InitLog() {
	log.Init(facade.Config.Log)
}
