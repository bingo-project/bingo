package bootstrap

import (
	"runtime"

	"github.com/hibiken/asynq"

	"bingo/internal/scheduler/facade"
)

func InitQueue() {
	opt := asynq.RedisClientOpt{
		Addr:     facade.Config.Redis.Host,
		Username: facade.Config.Redis.Username,
		Password: facade.Config.Redis.Password,
		DB:       facade.Config.Redis.Database,
	}

	facade.Worker = asynq.NewServer(opt, asynq.Config{
		Concurrency: runtime.NumCPU(),
	})
}
