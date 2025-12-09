package middleware

import (
	"github.com/bingo-project/bingo/internal/pkg/log"
)

func Recover() {
	if err := recover(); err != nil {
		log.C(Ctx).Infow("recover from panic: ", "err", err)
	}
}
