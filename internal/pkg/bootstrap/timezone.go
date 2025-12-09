package bootstrap

import (
	"github.com/bingo-project/bingo/internal/pkg/facade"
)

func InitTimezone() {
	facade.Config.App.SetTimezone()
}
