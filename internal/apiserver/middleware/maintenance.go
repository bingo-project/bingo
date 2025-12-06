package middleware

import (
	"github.com/gin-gonic/gin"

	"bingo/internal/pkg/core"
	"bingo/internal/pkg/errno"
	"bingo/internal/pkg/log"
	model "bingo/internal/pkg/model/syscfg"
	"bingo/internal/pkg/store"
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
