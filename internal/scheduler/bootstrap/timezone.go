package bootstrap

import "bingo/internal/scheduler/facade"

func InitTimezone() {
	facade.Config.Server.SetTimezone()
}
