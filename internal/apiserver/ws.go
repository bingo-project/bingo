package apiserver

import (
	"fmt"

	"github.com/bingo-project/component-base/log"
	"github.com/gin-gonic/gin"

	"bingo/internal/apiserver/ws"
	"bingo/internal/pkg/facade"
)

type websocketServer struct {
	address string
}

func NewWebsocket() *websocketServer {
	return &websocketServer{
		address: facade.Config.Websocket.Addr,
	}
}

func (s *websocketServer) Run() {
	log.Infow("Start websocket server at " + s.address)

	// 启动 Hub
	hub := ws.NewHub()
	go hub.Run()

	r := gin.Default()

	// 注册 WebSocket 路由
	r.GET("/", func(c *gin.Context) {
		ws.ServeWs(hub, c)
	})

	go func() {
		if err := r.Run(s.address); err != nil {
			log.Fatalw("Failed to start websocket server: " + err.Error())
		}
	}()
}

func (s *websocketServer) Close() {
	log.Infow(fmt.Sprintf("Websocket server on %s stopped", s.address))
}
