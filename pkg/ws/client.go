// ABOUTME: WebSocket client connection management.
// ABOUTME: Handles message read/write with JSON-RPC 2.0 protocol.

package ws

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"bingo/pkg/errorsx"
	"bingo/pkg/jsonrpc"
)

const (
	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Heartbeat timeout in seconds.
	heartbeatTimeout = 90
)

// TokenInfo contains parsed token information.
type TokenInfo struct {
	UserID    string
	ExpiresAt int64
}

// TokenParser parses a JWT token and returns user info.
type TokenParser func(token string) (*TokenInfo, error)

// ContextUpdater updates the client context with user info after login.
type ContextUpdater func(ctx context.Context, userID string) context.Context

// Client represents a WebSocket client connection.
type Client struct {
	hub            *Hub
	conn           *websocket.Conn
	router         *Router
	ctx            context.Context
	tokenParser    TokenParser
	contextUpdater ContextUpdater

	// Send channel for outbound messages
	Send      chan []byte
	closeOnce sync.Once

	// Client info
	ID             string // Unique client identifier
	Addr           string
	Platform       string
	UserID         string
	FirstTime      int64
	HeartbeatTime  int64
	LoginTime      int64
	TokenExpiresAt int64

	// Subscribed topics (managed by Hub, read-only for Client)
	topics     map[string]bool
	topicsLock sync.RWMutex
}

// ClientOption is a functional option for configuring Client.
type ClientOption func(*Client)

// WithTokenParser sets a token parser for the client.
func WithTokenParser(parser TokenParser) ClientOption {
	return func(c *Client) {
		c.tokenParser = parser
	}
}

// WithContextUpdater sets a context updater for the client.
func WithContextUpdater(updater ContextUpdater) ClientOption {
	return func(c *Client) {
		c.contextUpdater = updater
	}
}

// WithRouter sets a router for the client.
func WithRouter(r *Router) ClientOption {
	return func(c *Client) {
		c.router = r
	}
}

// NewClient creates a new WebSocket client.
func NewClient(hub *Hub, conn *websocket.Conn, ctx context.Context, opts ...ClientOption) *Client {
	now := time.Now().Unix()
	addr := ""
	if conn != nil {
		addr = conn.RemoteAddr().String()
	}
	c := &Client{
		hub:           hub,
		conn:          conn,
		ctx:           ctx,
		Send:          make(chan []byte, 256),
		ID:            uuid.New().String(),
		Addr:          addr,
		FirstTime:     now,
		HeartbeatTime: now,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// ReadPump pumps messages from the websocket connection to the hub.
func (c *Client) ReadPump() {
	defer func() {
		c.hub.Unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(c.hub.config.MaxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.hub.logger.Errorw("WebSocket read error", "addr", c.Addr, "err", err)
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
			c.conn.SetWriteDeadline(time.Now().Add(c.hub.config.WriteWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(c.hub.config.WriteWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) handleMessage(data []byte) {
	var req jsonrpc.Request
	if err := json.Unmarshal(data, &req); err != nil {
		resp := jsonrpc.NewErrorResponse(nil,
			errorsx.New(400, "ParseError", "Invalid JSON: %s", err.Error()))
		c.sendJSON(resp)
		return
	}

	// Update heartbeat for any message
	c.Heartbeat(time.Now().Unix())

	// Router is required
	if c.router == nil {
		resp := jsonrpc.NewErrorResponse(req.ID,
			errorsx.New(500, "InternalError", "Router not configured"))
		c.sendJSON(resp)
		return
	}

	ctx := &Context{
		Context:   c.ctx,
		Request:   &req,
		Client:    c,
		Method:    req.Method,
		StartTime: time.Now(),
	}
	resp := c.router.Dispatch(ctx)
	c.sendJSON(resp)
}

func (c *Client) sendJSON(v any) {
	data, err := json.Marshal(v)
	if err != nil {
		c.hub.logger.Errorw("JSON marshal error", "err", err)
		return
	}

	select {
	case c.Send <- data:
	default:
		c.hub.logger.Warnw("Client send buffer full", "addr", c.Addr)
	}
}

// SendJSON sends a JSON message to the client.
func (c *Client) SendJSON(v any) {
	c.sendJSON(v)
}

// Heartbeat updates the heartbeat time.
func (c *Client) Heartbeat(currentTime int64) {
	c.HeartbeatTime = currentTime
}

// IsHeartbeatTimeout returns true if heartbeat has timed out.
func (c *Client) IsHeartbeatTimeout(currentTime int64) bool {
	return c.HeartbeatTime+heartbeatTimeout <= currentTime
}

// IsAuthenticated returns true if the client has logged in.
func (c *Client) IsAuthenticated() bool {
	return c.UserID != "" && c.Platform != "" && c.LoginTime > 0
}

// Login sets the user info for this client.
func (c *Client) Login(platform, userID string, loginTime int64) {
	c.Platform = platform
	c.UserID = userID
	c.LoginTime = loginTime
	c.Heartbeat(loginTime)
}

// ParseToken parses a JWT token using the client's token parser.
func (c *Client) ParseToken(token string) (*TokenInfo, error) {
	if c.tokenParser == nil {
		return nil, errorsx.New(500, "InternalError", "Token parser not configured")
	}
	return c.tokenParser(token)
}

// UpdateContext updates the client context with user info.
func (c *Client) UpdateContext(userID string) {
	if c.contextUpdater != nil {
		c.ctx = c.contextUpdater(c.ctx, userID)
	}
}

// NotifyLogin sends a login event to the hub.
func (c *Client) NotifyLogin(userID, platform string, tokenExpiresAt int64) {
	c.hub.Login <- &LoginEvent{
		Client:         c,
		UserID:         userID,
		Platform:       platform,
		TokenExpiresAt: tokenExpiresAt,
	}
}
