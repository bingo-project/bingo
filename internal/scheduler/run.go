package scheduler

import (
	"github.com/bingo-project/component-base/log"
	"github.com/hibiken/asynq"

	"bingo/internal/scheduler/facade"
	"bingo/internal/scheduler/job"
	"bingo/internal/scheduler/middleware"
	"bingo/internal/scheduler/scheduler"
)

// run 函数是实际的业务代码入口函数.
func run() error {
	// Queue job
	go runQueueJobs()

	// Cron job
	go runCronJobs()

	// Cron job from database
	runDynamicCronJobs()

	return nil
}

func runQueueJobs() {
	mux := asynq.NewServeMux()
	mux.Use(middleware.Logging)

	job.Register(mux)

	err := facade.Worker.Run(mux)
	if err != nil {
		log.Fatalw("runQueueJobs failed", "err", err)
	}
}

func runCronJobs() {
	scheduler.RegisterPeriodicTasks()

	if err := facade.Scheduler.Run(); err != nil {
		log.Fatalw("runCronJobs failed", "err", err)
	}
}

func runDynamicCronJobs() {
	if err := facade.TaskManager.Run(); err != nil {
		log.Fatalw("runDynamicCronJobs failed", "err", err)
	}
}
