package job

import "github.com/hibiken/asynq"

// Register jobs here.
func Register(mux *asynq.ServeMux) {
	// Send email.
	mux.HandleFunc(EmailVerificationCode, HandleEmailVerificationTask)
}
