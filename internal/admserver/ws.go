// ABOUTME: WebSocket server initialization for admserver.
// ABOUTME: Configures WebSocket upgrader, hub, and connection handling.

package admserver

import (
	"context"
	"net/http"

	"github.com/bingo-project/component-base/web/token"
	"github.com/gin-gonic/gin"
	gorillaWS "github.com/gorilla/websocket"

	"github.com/bingo-project/websocket"
	"github.com/bingo-project/websocket/middleware"

	"github.com/bingo-project/bingo/internal/admserver/router"
	"github.com/bingo-project/bingo/internal/pkg/bootstrap"
	"github.com/bingo-project/bingo/internal/pkg/config"
	"github.com/bingo-project/bingo/internal/pkg/facade"
	"github.com/bingo-project/bingo/internal/pkg/log"
)

// initWebSocket initializes the WebSocket engine and hub.
func initWebSocket() (*gin.Engine, *websocket.Hub) {
	// Create rate limiter store
	rateLimitStore := middleware.NewRateLimiterStore()

	// Create logger for WebSocket
	wsLogger := log.NewWSLogger()

	// Create hub with disconnect callback to cleanup rate limiters
	hub := websocket.NewHub(
		websocket.WithLogger(wsLogger),
		websocket.WithClientDisconnectCallback(rateLimitStore.Remove),
	)

	// Create router and register handlers
	wsRouter := websocket.NewRouter()
	router.RegisterWSHandlers(wsRouter, rateLimitStore, wsLogger)

	// Create Gin engine for WebSocket
	engine := bootstrap.InitGinForWebSocket()

	// Configure WebSocket upgrader
	upgrader := gorillaWS.Upgrader{
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
func serveWS(c *gin.Context, hub *websocket.Hub, router *websocket.Router, upgrader gorillaWS.Upgrader) {
	// Create base context with request ID
	ctx := context.Background()
	ctx = websocket.WithRequestID(ctx, c.GetHeader("X-Request-ID"))

	// Upgrade connection
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	// Create anonymous client with router, token parser, and context updater
	client := websocket.NewClient(hub, conn, ctx,
		websocket.WithRouter(router),
		websocket.WithTokenParser(tokenParser),
		websocket.WithContextUpdater(contextUpdater),
	)

	// Register with hub (as anonymous)
	hub.Register <- client

	// Start read/write pumps
	go client.WritePump()
	go client.ReadPump()
}

// tokenParser parses JWT token and returns user info.
func tokenParser(tokenStr string) (*websocket.TokenInfo, error) {
	payload, err := token.Parse(tokenStr)
	if err != nil {
		return nil, err
	}

	return &websocket.TokenInfo{
		UserID:    payload.Subject,
		ExpiresAt: payload.ExpiresAt.Unix(),
	}, nil
}

// contextUpdater updates context with user ID after login.
func contextUpdater(ctx context.Context, userID string) context.Context {
	return websocket.WithUserID(ctx, userID)
}
