package watcher

import (
	"context"
	"time"

	"github.com/bingo-project/component-base/log"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	goredislib "github.com/redis/go-redis/v9"
	"github.com/robfig/cron/v3"

	"bingo/internal/apiserver/facade"
	"bingo/internal/watcher/watcher"

	// trigger init functions in `internal/watcher/watcher/`.
	_ "bingo/internal/watcher/watcher/all"
)

type watchJob struct {
	*cron.Cron
	rs *redsync.Redsync
}

func newWatchJob() *watchJob {
	location, _ := time.LoadLocation(facade.Config.Server.Timezone)

	cronjob := cron.New(
		cron.WithSeconds(),
		cron.WithChain(cron.SkipIfStillRunning(cron.DefaultLogger), cron.Recover(cron.DefaultLogger)),
		cron.WithLocation(location),
	)

	// Go redis
	client := goredislib.NewClient(&goredislib.Options{
		Addr:     facade.Config.Redis.Host,
		Password: facade.Config.Redis.Password,
	})
	rs := redsync.New(goredis.NewPool(client))

	return &watchJob{
		Cron: cronjob,
		rs:   rs,
	}
}

func (w *watchJob) addWatchers() *watchJob {
	for name, watch := range watcher.ListWatchers() {
		// log with `{"watcher": "counter"}` key-value to distinguish which watcher the log comes from.
		// nolint: golint,staticcheck
		ctx := context.WithValue(context.Background(), log.KeyWatcher, name)

		if err := watch.Init(ctx, w.rs.NewMutex(name, redsync.WithExpiry(2*time.Hour)), nil); err != nil {
			log.Fatalw("construct watcher %s failed: %s", name, err.Error())
		}

		_, _ = w.AddJob(watch.Spec(), watch)
	}

	return w
}
