package bootstrap

import (
	"bingo/internal/apiserver/config"
	"bingo/internal/pkg/log"
)

func InitLog() {
	log.Init(config.Cfg.Log)
}
