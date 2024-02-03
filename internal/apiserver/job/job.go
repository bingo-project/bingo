package job

import "github.com/hibiken/asynq"

func AddJobs(mux *asynq.ServeMux) {
	// Add jobs here
	mux.HandleFunc(EmailVerificationCode, HandleWelcomeEmailTask) // Demo task
}
