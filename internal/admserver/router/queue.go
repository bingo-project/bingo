package router

import (
	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"github.com/hibiken/asynqmon"

	"bingo/internal/pkg/facade"
)

func MapQueueRouters(g *gin.Engine) {
	mon := asynqmon.New(asynqmon.Options{
		RootPath: "/monitoring/tasks", // RootPath specifies the root for asynqmon app
		RedisConnOpt: asynq.RedisClientOpt{
			Addr:     facade.Config.Redis.Host,
			Username: facade.Config.Redis.Username,
			Password: facade.Config.Redis.Password,
			DB:       facade.Config.Redis.Database,
		},
	})

	g.Any("/monitoring/tasks/*any", gin.WrapH(mon))
}
