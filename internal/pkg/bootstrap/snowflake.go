package bootstrap

import (
	"github.com/bwmarrin/snowflake"

	"github.com/bingo-project/bingo/internal/pkg/facade"
	"github.com/bingo-project/bingo/internal/pkg/log"
)

func InitSnowflake() {
	var err error
	facade.Snowflake, err = snowflake.NewNode(1)
	if err != nil {
		log.Errorw("init snowflake failed", "err", err)

		return
	}
}
