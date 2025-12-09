package scheduler

import (
	"github.com/hibiken/asynq"

	"github.com/bingo-project/bingo/internal/pkg/facade"
	"github.com/bingo-project/bingo/internal/pkg/log"
	"github.com/bingo-project/bingo/internal/scheduler/job"
	"github.com/bingo-project/bingo/internal/scheduler/middleware"
	"github.com/bingo-project/bingo/internal/scheduler/scheduler"
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
