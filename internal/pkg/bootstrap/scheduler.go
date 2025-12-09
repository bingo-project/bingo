package bootstrap

import (
	"time"

	"github.com/hibiken/asynq"

	"github.com/bingo-project/bingo/internal/pkg/facade"
	"github.com/bingo-project/bingo/internal/pkg/log"
	"github.com/bingo-project/bingo/internal/pkg/store"
	"github.com/bingo-project/bingo/internal/scheduler/biz/syscfg"
)

func InitScheduler() {
	opt := asynq.RedisClientOpt{
		Addr:     facade.Config.Redis.Host,
		Username: facade.Config.Redis.Username,
		Password: facade.Config.Redis.Password,
		DB:       facade.Config.Redis.Database,
	}

	// Timezone
	location, err := time.LoadLocation(facade.Config.App.Timezone)
	if err != nil {
		log.Fatalf(err.Error())
	}

	// Periodic task
	facade.Scheduler = asynq.NewScheduler(opt, &asynq.SchedulerOpts{
		Location: location,
	})

	// Periodic task (Dynamic)
	facade.TaskManager, err = asynq.NewPeriodicTaskManager(
		asynq.PeriodicTaskManagerOpts{
			RedisConnOpt:               opt,
			PeriodicTaskConfigProvider: syscfg.NewSchedule(store.S),
			SyncInterval:               time.Second * 10,
		})
	if err != nil {
		log.Fatalf(err.Error())
	}
}
