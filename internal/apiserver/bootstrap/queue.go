package bootstrap

import (
	"github.com/hibiken/asynq"

	"bingo/internal/apiserver/facade"
)

func InitQueue() {
	opt := asynq.RedisClientOpt{
		Addr:     facade.Config.Redis.Host,
		Username: facade.Config.Redis.Username,
		Password: facade.Config.Redis.Password,
		DB:       facade.Config.Redis.Database,
	}

	facade.Queue = asynq.NewClient(opt)
}
