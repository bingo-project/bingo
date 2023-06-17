package cache

import (
	"sync"

	"github.com/bingo-project/component-base/cache"
	"github.com/bingo-project/component-base/redis"

	"bingo/internal/apiserver/facade"
)

var (
	once   sync.Once
	prefix = "cache:"
)

func NewCache(rds *redis.RedisClient) {
	once.Do(func() {
		facade.Redis = rds.Client
		facade.Cache = cache.NewService(&cache.RedisStore{
			RedisClient: rds,
			KeyPrefix:   prefix,
		})
	})
}
