package ws

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"

	"bingo/pkg/ws/model"
)

const (
	defaultAppID = 101 // 默认平台ID
)

var (
	ClientManager = NewHub()
	appIDs        = []uint32{defaultAppID, 102, 103, 104} // 全部的平台
	ServerIp      string
	ServerPort    string
)

// GetAppIDs 所有平台
func GetAppIDs() []uint32 {
	return appIDs
}

// GetServer 获取服务器
func GetServer() (server *model.Server) {
	server = model.NewServer(ServerIp, ServerPort)

	return
}

// IsLocal 判断是否为本机
func IsLocal(server *model.Server) (isLocal bool) {
	if server.Ip == ServerIp && server.Port == ServerPort {
		isLocal = true
	}

	return
}

// InAppIDs in app
func InAppIDs(appID uint32) (inAppID bool) {
	for _, value := range appIDs {
		if value == appID {
			inAppID = true

			return
		}
	}

	return
}

// GetDefaultAppID 获取默认 appID
func GetDefaultAppID() (appID uint32) {
	appID = defaultAppID

	return
}

// ServeWs handles websocket requests from the peer.
func ServeWs(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("WebSocket 升级失败:", err)

		return
	}

	// 创建新客户端
	addr := conn.RemoteAddr().String()
	currentTime := uint64(time.Now().Unix())
	client := NewClient(addr, conn, currentTime)

	ClientManager.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
}
