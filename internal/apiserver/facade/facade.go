package facade

import (
	"github.com/go-redis/redis/v8"
	"github.com/goer-project/goer-core/cache"

	"bingo/internal/apiserver/config"
)

var Config *config.Config
var Redis *redis.Client
var Cache *cache.CacheService
