package middleware

import (
	"context"
	"time"

	"github.com/hibiken/asynq"

	"github.com/bingo-project/bingo/internal/pkg/log"
	"github.com/bingo-project/bingo/pkg/contextx"
)

func Logging(h asynq.Handler) asynq.Handler {
	return asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
		// Context
		taskID, _ := asynq.GetTaskID(ctx)
		ctx = contextx.WithTask(ctx, t.Type())
		ctx = contextx.WithRequestID(ctx, taskID)

		start := time.Now()

		log.C(ctx).Infow("Start processing " + t.Type())

		// Process task
		err := h.ProcessTask(ctx, t)
		if err != nil {
			log.C(ctx).Errorw("Failed processing "+t.Type(), log.KeyResult, err)

			return err
		}

		log.C(ctx).Infow("Finished processing "+t.Type(), log.KeyCost, time.Since(start), "payload", string(t.Payload()))

		return nil
	})
}
