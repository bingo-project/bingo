package middleware

import (
	"context"
	"time"

	"github.com/bingo-project/component-base/log"
	"github.com/hibiken/asynq"
)

func Logging(h asynq.Handler) asynq.Handler {
	return asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
		// Context
		taskId, _ := asynq.GetTaskID(ctx)
		ctx = context.WithValue(ctx, log.KeyTask, t.Type())
		ctx = context.WithValue(ctx, log.KeyTrace, taskId)

		start := time.Now()

		log.C(ctx).Infow("Start processing " + t.Type())

		// Process task
		err := h.ProcessTask(ctx, t)
		if err != nil {
			log.C(ctx).Errorw("Failed processing "+t.Type(), log.KeyResult, err)

			return err
		}

		log.C(ctx).Infow("Finished processing "+t.Type(), "cost", time.Since(start), "payload", string(t.Payload()))

		return nil
	})
}
