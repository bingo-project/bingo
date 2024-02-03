package job

import "github.com/hibiken/asynq"

func AddJobs(mux *asynq.ServeMux) {
	// Add jobs here
	mux.HandleFunc("demo:task", HandleWelcomeEmailTask) // Demo task
}
