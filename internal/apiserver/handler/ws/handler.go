// ABOUTME: WebSocket HTTP handler for Gin.
// ABOUTME: Upgrades HTTP connections and manages WebSocket lifecycle.

package ws

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"bingo/internal/pkg/auth"
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
	authn   *auth.Authenticator
}

// NewHandler creates a new WebSocket handler.
func NewHandler(hub *ws.Hub, adapter *jsonrpc.Adapter) *Handler {
	return &Handler{
		hub:     hub,
		adapter: adapter,
		authn:   auth.New(),
	}
}

// ServeWS handles WebSocket upgrade requests.
func (h *Handler) ServeWS(c *gin.Context) {
	// 1. Create base context with request ID
	ctx := context.Background()
	ctx = contextx.WithRequestID(ctx, c.GetHeader("X-Request-ID"))

	// 2. Authenticate using unified authenticator
	ctx, _ = h.authn.AuthenticateWebSocket(ctx, c.Request)

	// 3. Upgrade connection
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	// 4. Create client
	client := ws.NewClient(h.hub, conn, ctx, h.adapter)

	// 5. Register with hub
	h.hub.Register <- client

	// 6. If authenticated, also login
	if auth.IsAuthenticated(ctx) {
		h.hub.Login <- &ws.LoginEvent{
			Client: client,
			UserID: contextx.UserID(ctx),
			AppID:  0, // Default app ID
		}
	}

	// 7. Start read/write pumps
	go client.WritePump()
	go client.ReadPump()
}
