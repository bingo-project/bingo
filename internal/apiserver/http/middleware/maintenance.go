package middleware

import (
	"github.com/bingo-project/component-base/log"
	"github.com/gin-gonic/gin"

	model "bingo/internal/apiserver/model/syscfg"
	"bingo/internal/apiserver/store"
	"bingo/internal/pkg/core"
	"bingo/internal/pkg/errno"
)

func Maintenance() gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg, err := store.S.Configs().GetServerConfig(c)
		if err != nil {
			log.C(c).Errorw("Maintenance get server config error", log.KeyResult, err)
			c.Next()

			return
		}

		// Under maintenance.
		if cfg.Status == model.ServerStatusMaintenance {
			core.WriteResponse(c, errno.ErrServiceUnderMaintenance, nil)
			c.Abort()

			return
		}

		c.Next()
	}
}
