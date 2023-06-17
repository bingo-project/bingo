package bootstrap

import "bingo/internal/apiserver/facade"

func InitTimezone() {
	facade.Config.Server.SetTimezone()
}
