package bootstrap

import (
	"github.com/bingo-project/component-base/log"
	"github.com/bwmarrin/snowflake"

	"bingo/internal/pkg/facade"
)

func InitSnowflake() {
	var err error
	facade.Snowflake, err = snowflake.NewNode(1)
	if err != nil {
		log.Errorw("init snowflake failed", "err", err)

		return
	}
}
