package facade

import (
	"github.com/bingo-project/component-base/cache"
	"github.com/bingo-project/component-base/crypt"
	"github.com/redis/go-redis/v9"

	"bingo/internal/apiserver/config"
)

var Config *config.Config
var AES *crypt.AES
var Redis *redis.Client
var Cache *cache.CacheService
