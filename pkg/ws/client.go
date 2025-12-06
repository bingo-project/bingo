// ABOUTME: WebSocket client connection management.
// ABOUTME: Handles message read/write with JSON-RPC 2.0 protocol.

package ws

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"

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
	adapter        *jsonrpc.Adapter
	ctx            context.Context
	tokenParser    TokenParser
	contextUpdater ContextUpdater

	// Send channel for outbound messages
	Send      chan []byte
	closeOnce sync.Once

	// Client info
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

// NewClient creates a new WebSocket client.
func NewClient(hub *Hub, conn *websocket.Conn, ctx context.Context, adapter *jsonrpc.Adapter, opts ...ClientOption) *Client {
	now := time.Now().Unix()
	c := &Client{
		hub:           hub,
		conn:          conn,
		adapter:       adapter,
		ctx:           ctx,
		Send:          make(chan []byte, 256),
		Addr:          conn.RemoteAddr().String(),
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
	logger := c.hub.logger.WithContext(c.ctx)

	// Recover from panics in message handling
	defer func() {
		if r := recover(); r != nil {
			logger.Errorw("Panic in handleMessage", "addr", c.Addr, "panic", r)

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

	// Log message received
	logger.Debugw("WebSocket message received", "method", req.Method, "id", req.ID, "addr", c.Addr)

	// Update heartbeat for any message
	c.Heartbeat(time.Now().Unix())

	// Handle heartbeat
	if req.Method == "heartbeat" {
		c.sendJSON(jsonrpc.NewResponse(req.ID, map[string]any{
			"status":      "ok",
			"server_time": time.Now().Unix(),
		}))
		return
	}

	// Handle login (allowed without authentication)
	if req.Method == "login" || req.Method == "auth.login" {
		c.handleLogin(&req)
		return
	}

	// Require authentication for other methods
	if !c.IsAuthenticated() {
		resp := jsonrpc.NewErrorResponse(req.ID,
			errorsx.New(401, "Unauthorized", "Login required"))
		logger.Warnw("WebSocket unauthorized request", "method", req.Method, "id", req.ID, "addr", c.Addr)
		c.sendJSON(resp)
		return
	}

	// Handle subscribe/unsubscribe
	if req.Method == "subscribe" {
		c.handleSubscribe(&req)
		return
	}
	if req.Method == "unsubscribe" {
		c.handleUnsubscribe(&req)
		return
	}

	// Route through adapter for business methods
	resp := c.adapter.Handle(c.ctx, &req)
	if resp.Error != nil {
		logger.Warnw("WebSocket request failed", "method", req.Method, "id", req.ID, "error", resp.Error.Reason)
	} else {
		logger.Debugw("WebSocket request succeeded", "method", req.Method, "id", req.ID)
	}
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

func (c *Client) handleLogin(req *jsonrpc.Request) {
	logger := c.hub.logger.WithContext(c.ctx)

	// Parse params to get platform
	var params struct {
		Platform string `json:"platform"`
	}
	paramsBytes, _ := json.Marshal(req.Params)
	if err := json.Unmarshal(paramsBytes, &params); err != nil {
		logger.Warnw("WebSocket login invalid params", "addr", c.Addr, "error", err.Error())
		c.sendJSON(jsonrpc.NewErrorResponse(req.ID,
			errorsx.New(400, "InvalidParams", "Invalid login params")))

		return
	}

	if !IsValidPlatform(params.Platform) {
		logger.Warnw("WebSocket login invalid platform", "addr", c.Addr, "platform", params.Platform)
		c.sendJSON(jsonrpc.NewErrorResponse(req.ID,
			errorsx.New(400, "InvalidPlatform", "Invalid platform: %s", params.Platform)))

		return
	}

	// Route to auth.login handler via adapter
	loginReq := &jsonrpc.Request{
		JSONRPC: req.JSONRPC,
		Method:  "auth.login",
		Params:  req.Params,
		ID:      req.ID,
	}
	resp := c.adapter.Handle(c.ctx, loginReq)
	if resp.Error != nil {
		logger.Warnw("WebSocket login failed", "addr", c.Addr, "error", resp.Error.Reason)
		c.sendJSON(resp)
		return
	}

	// Extract token from response
	resultBytes, _ := json.Marshal(resp.Result)
	var loginResp struct {
		AccessToken string `json:"accessToken"`
	}
	if err := json.Unmarshal(resultBytes, &loginResp); err != nil {
		c.sendJSON(jsonrpc.NewErrorResponse(req.ID,
			errorsx.New(500, "InternalError", "Failed to parse login response")))
		return
	}

	// Parse token to get user info
	if c.tokenParser == nil {
		c.sendJSON(jsonrpc.NewErrorResponse(req.ID,
			errorsx.New(500, "InternalError", "Token parser not configured")))
		return
	}

	tokenInfo, err := c.tokenParser(loginResp.AccessToken)
	if err != nil {
		c.sendJSON(jsonrpc.NewErrorResponse(req.ID,
			errorsx.New(401, "InvalidToken", "Failed to parse token: %s", err.Error())))
		return
	}

	// Update client context with user info
	if c.contextUpdater != nil {
		c.ctx = c.contextUpdater(c.ctx, tokenInfo.UserID)
	}

	// Send login event to hub
	c.hub.Login <- &LoginEvent{
		Client:         c,
		UserID:         tokenInfo.UserID,
		Platform:       params.Platform,
		TokenExpiresAt: tokenInfo.ExpiresAt,
	}

	// Return the original login response (contains token for client to store)
	c.sendJSON(resp)
}

func (c *Client) handleSubscribe(req *jsonrpc.Request) {
	var params struct {
		Topics []string `json:"topics"`
	}

	paramsBytes, _ := json.Marshal(req.Params)
	if err := json.Unmarshal(paramsBytes, &params); err != nil {
		c.sendJSON(jsonrpc.NewErrorResponse(req.ID,
			errorsx.New(400, "InvalidParams", "Invalid subscribe params")))
		return
	}

	if len(params.Topics) == 0 {
		c.sendJSON(jsonrpc.NewErrorResponse(req.ID,
			errorsx.New(400, "InvalidParams", "Topics required")))
		return
	}

	result := make(chan []string, 1)
	c.hub.Subscribe <- &SubscribeEvent{
		Client: c,
		Topics: params.Topics,
		Result: result,
	}

	subscribed := <-result
	c.sendJSON(jsonrpc.NewResponse(req.ID, map[string]any{
		"subscribed": subscribed,
	}))
}

func (c *Client) handleUnsubscribe(req *jsonrpc.Request) {
	var params struct {
		Topics []string `json:"topics"`
	}

	paramsBytes, _ := json.Marshal(req.Params)
	if err := json.Unmarshal(paramsBytes, &params); err != nil {
		c.sendJSON(jsonrpc.NewErrorResponse(req.ID,
			errorsx.New(400, "InvalidParams", "Invalid unsubscribe params")))
		return
	}

	if len(params.Topics) == 0 {
		c.sendJSON(jsonrpc.NewErrorResponse(req.ID,
			errorsx.New(400, "InvalidParams", "Topics required")))
		return
	}

	c.hub.Unsubscribe <- &UnsubscribeEvent{
		Client: c,
		Topics: params.Topics,
	}

	c.sendJSON(jsonrpc.NewResponse(req.ID, map[string]any{
		"unsubscribed": params.Topics,
	}))
}
