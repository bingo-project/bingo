package queue

import (
	"context"
	"time"

	"github.com/bingo-project/component-base/log"
	"github.com/hibiken/asynq"
)

func Logging(h asynq.Handler) asynq.Handler {
	return asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
		ctx = context.WithValue(ctx, "task", t.Type())
		start := time.Now()

		log.C(ctx).Infow("Start processing " + t.Type())

		// Process task
		err := h.ProcessTask(ctx, t)
		if err != nil {
			log.C(ctx).Errorw("Failed processing "+t.Type(), "err", err)

			return err
		}

		log.C(ctx).Infow("Finished processing "+t.Type(), "cost", time.Since(start), "payload", string(t.Payload()))

		return nil
	})
}
