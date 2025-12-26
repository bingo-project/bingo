package bootstrap

import (
	"github.com/bingo-project/bingo/internal/pkg/i18n"
)

func Boot() {
	InitLog()
	InitTimezone()
	InitSnowflake()
	InitMail()
	InitCache()
	InitAES()
	InitQueue()
	i18n.Init()
}
