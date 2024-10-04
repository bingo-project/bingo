package facade

import (
	"github.com/bingo-project/component-base/cache"
	"github.com/bingo-project/component-base/crypt"
	"github.com/bwmarrin/snowflake"
	"github.com/redis/go-redis/v9"

	"bingo/internal/apiserver/config"
	"bingo/pkg/mail"
)

var (
	Config    config.Config
	AES       *crypt.AES
	Redis     *redis.Client
	Cache     *cache.CacheService
	Snowflake *snowflake.Node
	Mail      *mail.Mailer
)
