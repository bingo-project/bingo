package job

import (
	"context"
	"encoding/json"
	"log"

	"github.com/hibiken/asynq"
)

type EmailTaskPayload struct {
	Email    string
	Username string
}

func HandleWelcomeEmailTask(ctx context.Context, t *asynq.Task) error {
	var payload EmailTaskPayload
	err := json.Unmarshal(t.Payload(), &payload)
	if err != nil {
		return err
	}

	log.Printf(" [*] Send Welcome Email to %s (%s)", payload.Username, payload.Email)

	return nil
}
