package apiserver

import (
	"fmt"
	"strings"

	"github.com/bingo-project/component-base/log"
	"github.com/gin-gonic/gin"

	"bingo/internal/apiserver/router"
	"bingo/internal/pkg/facade"
	"bingo/pkg/ws/cache"
	"bingo/pkg/ws/helper"
	ws "bingo/pkg/ws/server"
	"bingo/pkg/ws/task"
)

type websocketServer struct {
	address string
}

func NewWebsocket() *websocketServer {
	// Init redis
	cache.RedisClient = facade.Redis

	// Register websocket routers
	router.Websocket()

	// 定时任务
	task.Init()

	// 服务注册
	task.ServerInit()

	return &websocketServer{
		address: facade.Config.Websocket.Addr,
	}
}

func (s *websocketServer) Run() {
	ws.ServerIp = helper.GetServerIp()
	ws.ServerPort = strings.Split(s.address, ":")[1]

	log.Infow("Start websocket server at " + s.address)

	r := gin.Default()

	// 注册 WebSocket 路由
	r.GET("/", func(c *gin.Context) {
		ws.ServeWs(c)
	})

	go ws.ClientManager.Run()

	go func() {
		if err := r.Run(s.address); err != nil {
			log.Fatalw("Failed to start websocket server: " + err.Error())
		}
	}()
}

func (s *websocketServer) Close() {
	log.Infow(fmt.Sprintf("Websocket server on %s stopped", s.address))
}
