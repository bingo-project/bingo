// ABOUTME: WebSocket server initialization for apiserver.
// ABOUTME: Configures WebSocket upgrader, hub, and connection handling.

package apiserver

import (
	"context"
	"net/http"

	"github.com/bingo-project/component-base/web/token"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"bingo/internal/apiserver/router"
	"bingo/internal/pkg/bootstrap"
	"bingo/internal/pkg/config"
	"bingo/internal/pkg/facade"
	"bingo/internal/pkg/log"
	"bingo/pkg/contextx"
	"bingo/pkg/ws"
	"bingo/pkg/ws/middleware"
)

// initWebSocket initializes the WebSocket engine and hub.
func initWebSocket() (*gin.Engine, *ws.Hub) {
	// Create rate limiter store
	rateLimitStore := middleware.NewRateLimiterStore()

	// Create logger for WebSocket
	wsLogger := log.NewWSLogger()

	// Create hub with disconnect callback to cleanup rate limiters
	hub := ws.NewHub(
		ws.WithLogger(wsLogger),
		ws.WithClientDisconnectCallback(rateLimitStore.Remove),
	)

	// Create router and register handlers
	wsRouter := ws.NewRouter()
	router.RegisterWSHandlers(wsRouter, rateLimitStore, wsLogger)

	// Create Gin engine for WebSocket
	engine := bootstrap.InitGinForWebSocket()

	// Configure WebSocket upgrader
	upgrader := websocket.Upgrader{
		ReadBufferSize:  4096,
		WriteBufferSize: 4096,
		CheckOrigin:     checkOrigin(facade.Config.WebSocket),
	}

	// Register WebSocket route
	engine.GET("/ws", func(c *gin.Context) {
		serveWS(c, hub, wsRouter, upgrader)
	})

	return engine, hub
}

// checkOrigin returns an origin checker function based on config.
func checkOrigin(cfg *config.WebSocket) func(r *http.Request) bool {
	return func(r *http.Request) bool {
		if cfg == nil || cfg.AllowAllOrigins() {
			return true
		}

		origin := r.Header.Get("Origin")
		// Allow requests without Origin header (non-browser clients)
		if origin == "" {
			return true
		}

		return cfg.IsOriginAllowed(origin)
	}
}

// serveWS handles WebSocket upgrade requests.
func serveWS(c *gin.Context, hub *ws.Hub, router *ws.Router, upgrader websocket.Upgrader) {
	// Create base context with request ID
	ctx := context.Background()
	ctx = contextx.WithRequestID(ctx, c.GetHeader("X-Request-ID"))

	// Upgrade connection
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	// Create anonymous client with router, token parser, and context updater
	client := ws.NewClient(hub, conn, ctx,
		ws.WithRouter(router),
		ws.WithTokenParser(tokenParser),
		ws.WithContextUpdater(contextUpdater),
	)

	// Register with hub (as anonymous)
	hub.Register <- client

	// Start read/write pumps
	go client.WritePump()
	go client.ReadPump()
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

// RunnableFunc adapts a function to the Runnable interface.
// Useful for wrapping functions that don't return error.
type RunnableFunc func(ctx context.Context)

// Start implements Runnable.
func (f RunnableFunc) Start(ctx context.Context) error {
	f(ctx)
	return nil
}
