package ws

import (
	"log"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512

	// 用户连接超时时间
	heartbeatExpirationTime = 30
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许所有来源
	},
}

type Login struct {
	AppID  uint32
	UserID string
	Client *Client
}

// GetKey 获取 key
func (l *Login) GetKey() (key string) {
	key = GetUserKey(l.AppID, l.UserID)

	return
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte

	Addr          string // 客户端地址
	AppID         uint32 // 登录的平台ID app/web/ios
	UserID        string // 用户ID，用户登录以后才有
	FirstTime     uint64 // 首次连接事件
	HeartbeatTime uint64 // 用户上次心跳时间
	LoginTime     uint64 // 登录时间
}

func NewClient(addr string, socket *websocket.Conn, firstTime uint64) (client *Client) {
	client = &Client{
		conn:          socket,
		send:          make(chan []byte, 100),
		Addr:          addr,
		FirstTime:     firstTime,
		HeartbeatTime: firstTime,
	}

	return
}

// GetKey 获取 key
func (c *Client) GetKey() (key string) {
	key = GetUserKey(c.AppID, c.UserID)

	return
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) readPump() {
	defer func() {
		if r := recover(); r != nil {
			log.Println("write stop", string(debug.Stack()), r)
		}
	}()

	defer func() {
		ClientManager.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))

		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}

			break
		}

		// message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))

		// 处理程序
		ProcessData(c, message)
	}
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) writePump() {
	defer func() {
		if r := recover(); r != nil {
			log.Println("write stop", string(debug.Stack()), r)
		}
	}()

	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ClientManager.unregister <- c
		_ = c.conn.Close()

		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				log.Println("Client发送数据 关闭连接", c.Addr, "ok", ok)
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})

				return
			}

			_ = c.conn.WriteMessage(websocket.TextMessage, message)
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// SendMsg 发送数据
func (c *Client) SendMsg(msg []byte) {
	if c == nil {
		return
	}

	defer func() {
		if r := recover(); r != nil {
			log.Println("SendMsg stop:", r, string(debug.Stack()))
		}
	}()

	c.send <- msg
}

// close 关闭客户端连接
func (c *Client) close() {
	close(c.send)
}

// Login 用户登录
func (c *Client) Login(appID uint32, userID string, loginTime uint64) {
	c.AppID = appID
	c.UserID = userID
	c.LoginTime = loginTime

	// 登录成功=心跳一次
	c.Heartbeat(loginTime)
}

// Heartbeat 用户心跳
func (c *Client) Heartbeat(currentTime uint64) {
	c.HeartbeatTime = currentTime

	return
}

// IsHeartbeatTimeout 心跳超时
func (c *Client) IsHeartbeatTimeout(currentTime uint64) (timeout bool) {
	if c.HeartbeatTime+heartbeatExpirationTime <= currentTime {
		timeout = true
	}

	return
}
