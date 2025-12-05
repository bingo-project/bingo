// ABOUTME: WebSocket HTTP handler for Gin.
// ABOUTME: Upgrades HTTP connections and manages WebSocket lifecycle.

package ws

import (
	"context"
	"net/http"
	"strings"

	"github.com/bingo-project/component-base/web/token"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"bingo/internal/pkg/contextx"
	"bingo/pkg/jsonrpc"
	"bingo/pkg/ws"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// Handler handles WebSocket connections.
type Handler struct {
	hub     *ws.Hub
	adapter *jsonrpc.Adapter
}

// NewHandler creates a new WebSocket handler.
func NewHandler(hub *ws.Hub, adapter *jsonrpc.Adapter) *Handler {
	return &Handler{
		hub:     hub,
		adapter: adapter,
	}
}

// ServeWS handles WebSocket upgrade requests.
func (h *Handler) ServeWS(c *gin.Context) {
	// 1. Get token from query or header
	tokenStr := c.Query("token")
	if tokenStr == "" {
		tokenStr = extractBearerToken(c.GetHeader("Authorization"))
	}

	// 2. Create base context
	ctx := context.Background()
	ctx = contextx.WithRequestID(ctx, c.GetHeader("X-Request-ID"))

	// 3. Authenticate if token provided
	if tokenStr != "" {
		payload, err := token.Parse(tokenStr)
		if err == nil {
			ctx = contextx.WithUserID(ctx, payload.Subject)
		}
	}

	// 4. Upgrade connection
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	// 5. Create client
	client := ws.NewClient(h.hub, conn, ctx, h.adapter)

	// 6. Register with hub
	h.hub.Register <- client

	// 7. If authenticated, also login
	if userID := contextx.UserID(ctx); userID != "" {
		h.hub.Login <- &ws.LoginEvent{
			Client: client,
			UserID: userID,
			AppID:  0, // Default app ID
		}
	}

	// 8. Start read/write pumps
	go client.WritePump()
	go client.ReadPump()
}

func extractBearerToken(auth string) string {
	const prefix = "Bearer "
	if len(auth) > len(prefix) && strings.EqualFold(auth[:len(prefix)], prefix) {
		return auth[len(prefix):]
	}
	return ""
}
