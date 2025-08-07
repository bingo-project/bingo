package task

import (
	"log"
	"runtime/debug"
	"time"

	"bingo/pkg/ws/cache"
	ws "bingo/pkg/ws/server"
)

// ServerInit 服务初始化
func ServerInit() {
	Timer(2*time.Second, 60*time.Second, server, "", serverDefer, "")
}

// server 服务注册
func server(param interface{}) (result bool) {
	result = true

	defer func() {
		if r := recover(); r != nil {
			log.Println("服务注册 stop", r, string(debug.Stack()))
		}
	}()

	s := ws.GetServer()
	currentTime := uint64(time.Now().Unix())

	log.Println("定时任务，服务注册", param, s, currentTime)

	_ = cache.SetServerInfo(s, currentTime)

	return
}

// serverDefer 服务下线
func serverDefer(param interface{}) (result bool) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("服务下线 stop", r, string(debug.Stack()))
		}
	}()

	log.Println("服务下线", param)

	s := ws.GetServer()
	_ = cache.DelServerInfo(s)

	return
}
