// ABOUTME: WebSocket HTTP handler for Gin.
// ABOUTME: Upgrades HTTP connections and manages WebSocket lifecycle.

package ws

import (
	"context"
	"net/http"

	"github.com/bingo-project/component-base/web/token"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"bingo/internal/pkg/config"
	"bingo/internal/pkg/contextx"
	"bingo/pkg/ws"
)

// Handler handles WebSocket connections.
type Handler struct {
	hub      *ws.Hub
	router   *ws.Router
	upgrader websocket.Upgrader
}

// NewHandler creates a new WebSocket handler.
func NewHandler(hub *ws.Hub, router *ws.Router, cfg *config.WebSocket) *Handler {
	h := &Handler{
		hub:    hub,
		router: router,
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
		// Allow requests without Origin header (non-browser clients like Postman, mobile SDKs)
		if origin == "" {
			return true
		}

		return cfg.IsOriginAllowed(origin)
	}
}

// tokenParser parses JWT token and returns user info.
func tokenParser(tokenStr string) (*ws.TokenInfo, error) {
	payload, err := token.Parse(tokenStr)
	if err != nil {
		return nil, err
	}
	return &ws.TokenInfo{
		UserID:    payload.Subject,
		ExpiresAt: payload.ExpiresAt.Unix(),
	}, nil
}

// contextUpdater updates context with user ID after login.
func contextUpdater(ctx context.Context, userID string) context.Context {
	return contextx.WithUserID(ctx, userID)
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

	// 3. Create anonymous client with router, token parser, and context updater
	client := ws.NewClient(h.hub, conn, ctx, nil,
		ws.WithRouter(h.router),
		ws.WithTokenParser(tokenParser),
		ws.WithContextUpdater(contextUpdater),
	)

	// 4. Register with hub (as anonymous)
	h.hub.Register <- client

	// 5. Start read/write pumps
	go client.WritePump()
	go client.ReadPump()
}
