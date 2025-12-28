// ABOUTME: Asynq job handler registration.
// ABOUTME: Maps task types to their handler functions.

package job

import (
	"github.com/hibiken/asynq"

	"github.com/bingo-project/bingo/internal/pkg/task"
)

// Register jobs here.
func Register(mux *asynq.ServeMux) {
	// Send email.
	mux.HandleFunc(task.EmailVerificationCode, HandleEmailVerificationTask)

	// Publish announcement.
	mux.HandleFunc(task.AnnouncementPublish, HandleAnnouncementPublishTask)
}
