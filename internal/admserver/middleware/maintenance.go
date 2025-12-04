package middleware

import (
	"github.com/bingo-project/component-base/log"
	"github.com/gin-gonic/gin"

	"bingo/internal/pkg/store"
	"bingo/internal/pkg/core"
	"bingo/internal/pkg/errno"
	model "bingo/internal/pkg/model/syscfg"
)

func Maintenance() gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg, err := store.S.SysConfig().GetServerConfig(c)
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
