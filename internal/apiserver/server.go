package apiserver

import (
	"time"

	"github.com/fvbock/endless"
	"github.com/gin-gonic/gin"

	"bingo/internal/apiserver/facade"
)

func startInsecureServer(g *gin.Engine) error {
	s := endless.NewServer(facade.Config.Server.Addr, g)
	s.ReadHeaderTimeout = 20 * time.Second
	s.WriteTimeout = 20 * time.Second
	s.MaxHeaderBytes = 1 << 20

	err := s.ListenAndServe()
	if err != nil {
		return err
	}

	return nil
}
