package task

import (
	"context"

	"github.com/bingo-project/component-base/log"
	"github.com/duke-git/lancet/v2/convertor"
	"github.com/hibiken/asynq"
)

type queue struct {
	Context context.Context
	Name    string
	Payload any
}

func (t *task) Queue(ctx context.Context, name string, payload any) *queue {
	return &queue{
		Context: ctx,
		Name:    name,
		Payload: payload,
	}
}

func (q *queue) Dispatch(opts ...asynq.Option) (*asynq.TaskInfo, error) {
	payload := convertor.ToString(q.Payload)

	defaultQueue := "default"

	opt := []asynq.Option{asynq.Queue(defaultQueue)}
	opt = append(opt, opts...)

	// TaskID
	if trace := q.Context.Value(log.KeyTrace); trace != nil {
		opt = append(opt, asynq.TaskID(convertor.ToString(trace)))
	}

	t := asynq.NewTask(q.Name, []byte(payload), opt...)

	return T.client.EnqueueContext(q.Context, t)
}
