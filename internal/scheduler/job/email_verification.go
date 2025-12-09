package job

import (
	"context"
	"encoding/json"

	"github.com/hibiken/asynq"

	"github.com/bingo-project/bingo/internal/pkg/facade"
	"github.com/bingo-project/bingo/internal/pkg/task"
)

func HandleEmailVerificationTask(ctx context.Context, t *asynq.Task) error {
	var payload task.EmailVerificationCodePayload
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
