package bootstrap

import (
	"time"

	"github.com/hibiken/asynq"

	"bingo/internal/scheduler/facade"
)

func InitScheduler() {
	opt := asynq.RedisClientOpt{
		Addr:     facade.Config.Redis.Host,
		Username: facade.Config.Redis.Username,
		Password: facade.Config.Redis.Password,
		DB:       facade.Config.Redis.Database,
	}

	// Timezone
	location, err := time.LoadLocation(facade.Config.Server.Timezone)
	if err != nil {
		panic(err)
	}

	facade.Scheduler = asynq.NewScheduler(opt, &asynq.SchedulerOpts{
		Location: location,
	})
}
