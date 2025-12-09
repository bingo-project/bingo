package syscfg

import (
	"context"

	"github.com/hibiken/asynq"

	"github.com/bingo-project/bingo/internal/pkg/model/syscfg"
	"github.com/bingo-project/bingo/internal/pkg/store"
	"github.com/bingo-project/bingo/pkg/store/where"
)

type ScheduleBiz interface {
	UserExpansion
}

// UserExpansion 定义用户操作的扩展方法.
type UserExpansion interface {
	GetConfigs() ([]*asynq.PeriodicTaskConfig, error)
}

type scheduleBiz struct {
	store store.IStore
}

var _ ScheduleBiz = (*scheduleBiz)(nil)

func NewSchedule(store store.IStore) *scheduleBiz {
	return &scheduleBiz{store: store}
}

func (b *scheduleBiz) GetConfigs() (ret []*asynq.PeriodicTaskConfig, err error) {
	whr := where.F("status", syscfg.ScheduleStatusEnabled)
	_, configs, err := b.store.Schedule().List(context.Background(), whr)
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
