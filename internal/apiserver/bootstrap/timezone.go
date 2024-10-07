package bootstrap

import (
	"bingo/internal/pkg/facade"
)

func InitTimezone() {
	facade.Config.Server.SetTimezone()
}
