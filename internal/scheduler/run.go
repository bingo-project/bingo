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
	// Queue task
	go runJobs()

	// Scheduler
	runSchedulers()

	return nil
}

func runJobs() {
	mux := asynq.NewServeMux()
	mux.Use(middleware.Logging)

	job.Register(mux)

	err := facade.Worker.Run(mux)
	if err != nil {
		log.Fatalw("run worker failed", "err", err)
	}
}

func runSchedulers() {
	scheduler.Register()

	if err := facade.Scheduler.Run(); err != nil {
		log.Fatalw("run job failed", "err", err)
	}
}
