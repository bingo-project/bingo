package user

import (
	"context"
	"errors"

	"github.com/bingo-project/component-base/log"
	"github.com/go-redsync/redsync/v4"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"bingo/internal/apiserver/store"
)

type UserWatcher struct {
	ctx   context.Context
	mutex *redsync.Mutex
}

// Run runs the watcher job.
func (w *UserWatcher) Run() {
	w.ctx = context.WithValue(w.ctx, log.KeyTrace, uuid.New().String())
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
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.C(w.ctx).Errorw(err.Error())

		return
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		log.C(w.ctx).Debug("not found")
	}

	log.C(w.ctx).Infow(user.Email)
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
