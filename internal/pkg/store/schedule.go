package store

import (
	"context"

	"bingo/internal/pkg/model/syscfg"
	genericstore "bingo/pkg/store"
	"bingo/pkg/store/where"
)

type ScheduleStore interface {
	Create(ctx context.Context, obj *syscfg.Schedule) error
	Update(ctx context.Context, obj *syscfg.Schedule, fields ...string) error
	Delete(ctx context.Context, opts *where.Options) error
	Get(ctx context.Context, opts *where.Options) (*syscfg.Schedule, error)
	List(ctx context.Context, opts *where.Options) (int64, []*syscfg.Schedule, error)

	UserExpansion
}

// UserExpansion 定义了用户操作的附加方法.
// nolint: iface
type UserExpansion interface {
}

type scheduleStore struct {
	*genericstore.Store[syscfg.Schedule]
}

var _ ScheduleStore = (*scheduleStore)(nil)

func NewScheduleStore(store *datastore) *scheduleStore {
	return &scheduleStore{
		Store: genericstore.NewStore[syscfg.Schedule](store, NewLogger()),
	}
}
