package bootstrap

import (
	"github.com/hibiken/asynq"

	"bingo/internal/pkg/facade"
	"bingo/internal/pkg/task"
)

func InitQueue() {
	opt := asynq.RedisClientOpt{
		Addr:     facade.Config.Redis.Host,
		Username: facade.Config.Redis.Username,
		Password: facade.Config.Redis.Password,
		DB:       facade.Config.Redis.Database,
	}

	task.NewTask(opt)
}
