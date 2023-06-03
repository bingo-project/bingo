package user

import (
	"context"

	"github.com/go-redsync/redsync/v4"

	"bingo/internal/apiserver/store"
	"bingo/internal/pkg/log"
)

type UserWatcher struct {
	ctx   context.Context
	mutex *redsync.Mutex
}

// Run runs the watcher job.
func (w *UserWatcher) Run() {
	if err := w.mutex.Lock(); err != nil {
		log.C(w.ctx).Infow("UserWatcher already run.")

		return
	}
	defer func() {
		if _, err := w.mutex.Unlock(); err != nil {
			log.C(w.ctx).Errorw("could not release UserWatcher lock. err: %v", err)

			return
		}
	}()

	user, err := store.S.Users().Get(w.ctx, "test")
	if err != nil {
		log.Errorw(err.Error())

		return
	}

	log.Infow(user.Email)
}

// Spec is parsed using the time zone of clean Cron instance as the default.
func (w *UserWatcher) Spec() string {
	return "@every 1m"
}

// Init initializes the watcher for later execution.
func (w *UserWatcher) Init(ctx context.Context, rs *redsync.Mutex, config interface{}) error {
	*w = UserWatcher{
		ctx:   ctx,
		mutex: rs,
	}

	return nil
}
