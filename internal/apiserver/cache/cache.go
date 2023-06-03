package cache

import (
	"sync"

	"github.com/goer-project/goer-core/cache"
	"github.com/goer-project/goer-core/redis"
)

var (
	once   sync.Once
	C      *cache.CacheService
	prefix = "cache:"
)

func NewCache(rds *redis.RedisClient) {
	once.Do(func() {
		C = cache.NewService(&cache.RedisStore{
			RedisClient: rds,
			KeyPrefix:   prefix,
		})
	})
}
