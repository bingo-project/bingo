package bootstrap

import (
	"bingo/internal/apiserver/cache"
	"bingo/internal/apiserver/config"
	"bingo/internal/pkg/log"

	"github.com/goer-project/goer-core/redis"
)

func InitCache() {
	r, err := redis.NewClient(config.Cfg.Redis.Host, config.Cfg.Redis.Password, config.Cfg.Redis.Database)
	if err != nil {
		log.Errorw("init cache failed", "err", err)

		return
	}

	cache.NewCache(r)
}
