package bootstrap

import (
	"time"

	"github.com/bingo-project/component-base/log"
	"github.com/hibiken/asynq"

	"bingo/internal/pkg/facade"
	"bingo/internal/scheduler/biz/syscfg"
	"bingo/internal/scheduler/store"
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
