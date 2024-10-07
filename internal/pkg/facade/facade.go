package facade

import (
	"github.com/bingo-project/component-base/cache"
	"github.com/bingo-project/component-base/crypt"
	"github.com/bwmarrin/snowflake"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"

	"bingo/internal/pkg/config"
	"bingo/pkg/mail"
)

var (
	Config    config.Config
	AES       *crypt.AES
	Redis     *redis.Client
	Cache     *cache.CacheService
	Snowflake *snowflake.Node
	Mail      *mail.Mailer

	Worker      *asynq.Server
	Scheduler   *asynq.Scheduler
	TaskManager *asynq.PeriodicTaskManager
)
