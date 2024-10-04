package facade

import (
	"github.com/hibiken/asynq"

	"bingo/internal/scheduler/config"
	"bingo/pkg/mail"
)

var (
	Config    config.Config
	Mail      *mail.Mailer
	Worker    *asynq.Server
	Scheduler *asynq.Scheduler
)
