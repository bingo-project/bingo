package apiserver

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/bingo-project/component-base/log"

	"bingo/internal/apiserver/bootstrap"
	"bingo/internal/apiserver/facade"
)

type httpAPIServer struct {
	insecureServer  *http.Server
	insecureAddress string
}

// NewHttp create a grpcAPIServer instance.
func NewHttp() *httpAPIServer {
	g := bootstrap.InitGin()
	srv := &http.Server{Addr: facade.Config.Server.Addr, Handler: g}

	return &httpAPIServer{insecureServer: srv, insecureAddress: facade.Config.Server.Addr}
}

func (s *httpAPIServer) Run() {
	go func() {
		// Initializing the server in a goroutine so that
		// it won't block the graceful shutdown handling below
		log.Infow("Start to listening the incoming requests on http address: " + s.insecureAddress)

		if err := s.insecureServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalw(err.Error())
		}

		log.Infow(fmt.Sprintf("Server on %s stopped", s.insecureAddress))
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
}
