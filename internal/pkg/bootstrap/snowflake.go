package bootstrap

import (
	"github.com/bwmarrin/snowflake"

	"bingo/internal/pkg/facade"
	"bingo/internal/pkg/log"
)

func InitSnowflake() {
	var err error
	facade.Snowflake, err = snowflake.NewNode(1)
	if err != nil {
		log.Errorw("init snowflake failed", "err", err)

		return
	}
}
