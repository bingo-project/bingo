package syscfg

import (
	"context"

	"github.com/hibiken/asynq"

	"bingo/internal/scheduler/store"
)

type ScheduleBiz interface {
	GetConfigs() ([]*asynq.PeriodicTaskConfig, error)
}

type scheduleBiz struct {
	ds store.IStore
}

var _ ScheduleBiz = (*scheduleBiz)(nil)

func NewSchedule(ds store.IStore) *scheduleBiz {
	return &scheduleBiz{ds: ds}
}

func (b *scheduleBiz) GetConfigs() (ret []*asynq.PeriodicTaskConfig, err error) {
	configs, err := b.ds.Schedules().AllEnabled(context.Background())
	if err != nil {
		return
	}

	ret = make([]*asynq.PeriodicTaskConfig, 0, len(configs))
	for _, config := range configs {
		ret = append(ret, &asynq.PeriodicTaskConfig{
			Cronspec: config.Spec,
			Task:     asynq.NewTask(config.Job, nil),
		})
	}

	return
}
