package job

import (
	"context"
	"encoding/json"

	"github.com/hibiken/asynq"

	"bingo/internal/apiserver/facade"
)

const EmailVerificationCode = "email:verification"

type EmailVerificationCodePayload struct {
	To      string
	Subject string
	Content string
}

func HandleWelcomeEmailTask(ctx context.Context, t *asynq.Task) error {
	var payload EmailVerificationCodePayload
	err := json.Unmarshal(t.Payload(), &payload)
	if err != nil {
		return err
	}

	// Send email
	err = facade.Mail.Send(payload.To, payload.Subject, payload.Content)
	if err != nil {
		return err
	}

	return nil
}
