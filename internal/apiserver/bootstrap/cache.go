package bootstrap

import (
	"sync"

	"github.com/bingo-project/component-base/cache"
	"github.com/bingo-project/component-base/log"
	"github.com/bingo-project/component-base/redis"

	"bingo/internal/apiserver/facade"
)

var (
	once   sync.Once
	prefix = "cache:"
)

func InitCache() {
	rds, err := redis.NewClient(facade.Config.Redis.Host, facade.Config.Redis.Password, facade.Config.Redis.Database)
	if err != nil {
		log.Errorw("init cache failed", "err", err)

		return
	}

	once.Do(func() {
		facade.Redis = rds.Client
		facade.Cache = cache.NewService(&cache.RedisStore{
			RedisClient: rds,
			KeyPrefix:   prefix,
		})
	})
}
