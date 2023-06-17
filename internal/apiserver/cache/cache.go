package cache

import (
	"sync"

	"github.com/goer-project/goer-core/cache"
	"github.com/goer-project/goer-core/redis"

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
