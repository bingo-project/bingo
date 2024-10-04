package task

import (
	"sync"

	"github.com/hibiken/asynq"
)

var (
	once sync.Once
	T    *task
)

type task struct {
	client *asynq.Client
}

func NewTask(opt asynq.RedisClientOpt) *task {
	once.Do(func() {
		T = &task{
			client: asynq.NewClient(opt),
		}
	})

	return T
}
