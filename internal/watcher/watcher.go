package watcher

import (
	"context"
	"time"

	goredislib "github.com/go-redis/redis/v8"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v8"
	"github.com/robfig/cron/v3"

	"bingo/internal/apiserver/config"
	"bingo/internal/pkg/log"
	"bingo/internal/watcher/watcher"

	// trigger init functions in `internal/watcher/watcher/`.
	_ "bingo/internal/watcher/watcher/all"
)

type watchJob struct {
	*cron.Cron
	rs *redsync.Redsync
}

func newWatchJob() *watchJob {
	client := goredislib.NewClient(&goredislib.Options{
		Addr:     config.Cfg.Redis.Host,
		Password: config.Cfg.Redis.Password,
	})

	rs := redsync.New(goredis.NewPool(client))

	cronjob := cron.New(
		cron.WithSeconds(),
		cron.WithChain(cron.SkipIfStillRunning(nil), cron.Recover(nil)),
	)

	return &watchJob{
		Cron: cronjob,
		rs:   rs,
	}
}

func (w *watchJob) addWatchers() *watchJob {
	for name, watch := range watcher.ListWatchers() {
		// log with `{"watcher": "counter"}` key-value to distinguish which watcher the log comes from.
		// nolint: golint,staticcheck
		ctx := context.WithValue(context.Background(), "watcher", name)

		if err := watch.Init(ctx, w.rs.NewMutex(name, redsync.WithExpiry(2*time.Hour)), nil); err != nil {
			log.Fatalw("construct watcher %s failed: %s", name, err.Error())
		}

		_, _ = w.AddJob(watch.Spec(), watch)
	}

	return w
}
