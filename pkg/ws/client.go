// ABOUTME: WebSocket client connection management.
// ABOUTME: Handles message read/write with JSON-RPC 2.0 protocol.

package ws

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gorilla/websocket"

	"github.com/bingo-project/component-base/log"

	"bingo/pkg/errorsx"
	"bingo/pkg/jsonrpc"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 4096

	// Heartbeat timeout in seconds.
	heartbeatTimeout = 90
)

// Client represents a WebSocket client connection.
type Client struct {
	hub     *Hub
	conn    *websocket.Conn
	adapter *jsonrpc.Adapter
	ctx     context.Context

	// Send channel for outbound messages
	Send chan []byte

	// Client info
	Addr          string
	AppID         uint32
	UserID        string
	FirstTime     uint64
	HeartbeatTime uint64
	LoginTime     uint64
}

// NewClient creates a new WebSocket client.
func NewClient(hub *Hub, conn *websocket.Conn, ctx context.Context, adapter *jsonrpc.Adapter) *Client {
	now := uint64(time.Now().Unix())
	return &Client{
		hub:           hub,
		conn:          conn,
		adapter:       adapter,
		ctx:           ctx,
		Send:          make(chan []byte, 256),
		Addr:          conn.RemoteAddr().String(),
		FirstTime:     now,
		HeartbeatTime: now,
	}
}

// ReadPump pumps messages from the websocket connection to the hub.
func (c *Client) ReadPump() {
	defer func() {
		c.hub.Unregister <- c
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
				log.Errorw("WebSocket read error", "addr", c.Addr, "err", err)
			}
			break
		}

		c.handleMessage(message)
	}
}

// WritePump pumps messages from the hub to the websocket connection.
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) handleMessage(data []byte) {
	// Recover from panics in message handling
	defer func() {
		if r := recover(); r != nil {
			log.Errorw("Panic in handleMessage", "addr", c.Addr, "panic", r)
			resp := jsonrpc.NewErrorResponse(nil,
				errorsx.New(500, "InternalError", "Message handler crashed"))
			c.sendJSON(resp)
		}
	}()

	var req jsonrpc.Request
	if err := json.Unmarshal(data, &req); err != nil {
		resp := jsonrpc.NewErrorResponse(nil,
			errorsx.New(400, "ParseError", "Invalid JSON: %s", err.Error()))
		c.sendJSON(resp)
		return
	}

	// Handle heartbeat specially
	if req.Method == "heartbeat" {
		c.Heartbeat(uint64(time.Now().Unix()))
		c.sendJSON(jsonrpc.NewResponse(req.ID, map[string]string{"status": "ok"}))
		return
	}

	// Route through adapter
	resp := c.adapter.Handle(c.ctx, &req)
	c.sendJSON(resp)
}

func (c *Client) sendJSON(v any) {
	data, err := json.Marshal(v)
	if err != nil {
		log.Errorw("JSON marshal error", "err", err)
		return
	}

	select {
	case c.Send <- data:
	default:
		log.Warnw("Client send buffer full", "addr", c.Addr)
	}
}

// SendJSON sends a JSON message to the client.
func (c *Client) SendJSON(v any) {
	c.sendJSON(v)
}

// Heartbeat updates the heartbeat time.
func (c *Client) Heartbeat(currentTime uint64) {
	c.HeartbeatTime = currentTime
}

// IsHeartbeatTimeout returns true if heartbeat has timed out.
func (c *Client) IsHeartbeatTimeout(currentTime uint64) bool {
	return c.HeartbeatTime+heartbeatTimeout <= currentTime
}

// Login sets the user info for this client.
func (c *Client) Login(appID uint32, userID string, loginTime uint64) {
	c.AppID = appID
	c.UserID = userID
	c.LoginTime = loginTime
	c.Heartbeat(loginTime)
}
