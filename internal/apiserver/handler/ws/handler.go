// ABOUTME: WebSocket HTTP handler for Gin.
// ABOUTME: Upgrades HTTP connections and manages WebSocket lifecycle.

package ws

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"bingo/internal/pkg/config"
	"bingo/internal/pkg/contextx"
	"bingo/pkg/jsonrpc"
	"bingo/pkg/ws"
)

// Handler handles WebSocket connections.
type Handler struct {
	hub      *ws.Hub
	adapter  *jsonrpc.Adapter
	upgrader websocket.Upgrader
}

// NewHandler creates a new WebSocket handler.
func NewHandler(hub *ws.Hub, adapter *jsonrpc.Adapter, cfg *config.WebSocket) *Handler {
	h := &Handler{
		hub:     hub,
		adapter: adapter,
	}

	h.upgrader = websocket.Upgrader{
		ReadBufferSize:  4096,
		WriteBufferSize: 4096,
		CheckOrigin:     h.checkOrigin(cfg),
	}

	return h
}

// checkOrigin returns an origin checker function based on config.
func (h *Handler) checkOrigin(cfg *config.WebSocket) func(r *http.Request) bool {
	return func(r *http.Request) bool {
		if cfg == nil || cfg.AllowAllOrigins() {
			return true
		}
		origin := r.Header.Get("Origin")
		return cfg.IsOriginAllowed(origin)
	}
}

// ServeWS handles WebSocket upgrade requests.
func (h *Handler) ServeWS(c *gin.Context) {
	// 1. Create base context with request ID
	ctx := context.Background()
	ctx = contextx.WithRequestID(ctx, c.GetHeader("X-Request-ID"))

	// 2. Upgrade connection (no authentication at connect time)
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	// 3. Create anonymous client
	client := ws.NewClient(h.hub, conn, ctx, h.adapter)

	// 4. Register with hub (as anonymous)
	h.hub.Register <- client

	// 5. Start read/write pumps
	go client.WritePump()
	go client.ReadPump()
}
