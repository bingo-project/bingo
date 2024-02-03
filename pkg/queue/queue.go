package queue

import (
	"runtime"

	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

func NewQueue(rds *redis.Options) (client *asynq.Client, worker *asynq.Server) {
	opt := asynq.RedisClientOpt{
		Addr:     rds.Addr,
		Username: rds.Username,
		Password: rds.Password,
		DB:       rds.DB,
	}

	client = asynq.NewClient(opt)
	worker = asynq.NewServer(opt, asynq.Config{Concurrency: runtime.NumCPU()})

	return
}
