package apiserver

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/bingo-project/component-base/log"
	"github.com/gin-gonic/gin"

	"bingo/internal/apiserver/router"
	"bingo/internal/pkg/bootstrap"
	"bingo/internal/pkg/facade"
)

type httpAPIServer struct {
	insecureServer  *http.Server
	insecureAddress string
}

// NewHttp create a grpcAPIServer instance.
func NewHttp() *httpAPIServer {
	g := bootstrap.InitGin()
	installRouters(g)

	srv := &http.Server{Addr: facade.Config.HTTP.Addr, Handler: g}

	return &httpAPIServer{insecureServer: srv, insecureAddress: facade.Config.HTTP.Addr}
}

func (s *httpAPIServer) Run() {
	// Initializing the server in a goroutine so that
	// it won't block the graceful shutdown handling below
	log.Infow("Start http server on " + s.insecureAddress)

	go func() {
		if err := s.insecureServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalw("Failed to listen: " + err.Error())
		}
	}()
}

// Close graceful shutdown the api server.
func (s *httpAPIServer) Close() {
	// The context is used to inform the server it has 10 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 10 秒内优雅关闭服务（将未处理完的请求处理完再关闭服务），超过 10 秒就超时退出
	if err := s.insecureServer.Shutdown(ctx); err != nil {
		log.Fatalw("Shutdown insecure server failed: " + err.Error())
	}

	log.Infow(fmt.Sprintf("HTTP server on %s stopped", s.insecureAddress))
}

func installRouters(g *gin.Engine) *gin.Engine {
	// Swagger
	if facade.Config.Feature.ApiDoc {
		router.MapSwagRouters(g)
	}

	// Common router
	router.MapCommonRouters(g)

	// Api
	router.MapApiRouters(g)

	return g
}
