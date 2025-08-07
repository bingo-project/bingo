package task

import (
	"log"
	"runtime/debug"
	"time"

	ws "bingo/pkg/ws/server"
)

func Init() {
	Timer(3*time.Second, 30*time.Second, cleanConnection, "", nil, nil)
}

// cleanConnection 清理超时连接
func cleanConnection(param any) (result bool) {
	result = true

	defer func() {
		if r := recover(); r != nil {
			log.Println("ClearTimeoutConnections stop", r, string(debug.Stack()))
		}
	}()

	log.Println("定时任务，清理超时连接", param)

	ws.ClearTimeoutConnections()

	return
}
