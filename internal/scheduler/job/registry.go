package job

import (
	"github.com/hibiken/asynq"

	"bingo/internal/pkg/task"
)

// Register jobs here.
func Register(mux *asynq.ServeMux) {
	// Send email.
	mux.HandleFunc(task.EmailVerificationCode, HandleEmailVerificationTask)
}
