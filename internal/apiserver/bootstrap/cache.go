package bootstrap

import (
	"bingo/internal/apiserver/cache"
	"bingo/internal/apiserver/facade"
	"bingo/internal/pkg/log"

	"github.com/goer-project/goer-core/redis"
)

func InitCache() {
	r, err := redis.NewClient(facade.Config.Redis.Host, facade.Config.Redis.Password, facade.Config.Redis.Database)
	if err != nil {
		log.Errorw("init cache failed", "err", err)

		return
	}

	cache.NewCache(r)
}
