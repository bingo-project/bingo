// ABOUTME: WebSocket server implementation for apiserver.
// ABOUTME: Implements Server interface with WebSocket support via Gin.

package server

import (
	"context"
	"errors"
	"net/http"

	"github.com/bingo-project/component-base/log"
	"github.com/gin-gonic/gin"

	"bingo/internal/apiserver/handler/ws"
	"bingo/internal/pkg/config"
	"bingo/pkg/jsonrpc"
	pkgws "bingo/pkg/ws"
)

// WebSocketServer implements Server interface for WebSocket protocol.
type WebSocketServer struct {
	server  *http.Server
	cfg     *config.WebSocket
	hub     *pkgws.Hub
	adapter *jsonrpc.Adapter
}

// NewWebSocketServer creates a new WebSocket server.
func NewWebSocketServer(cfg *config.WebSocket, adapter *jsonrpc.Adapter) *WebSocketServer {
	hub := pkgws.NewHub()

	// Create Gin engine for WebSocket
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(gin.Recovery())

	// Create handler
	handler := ws.NewHandler(hub, adapter)

	// Register WebSocket route
	engine.GET("/ws", handler.ServeWS)

	return &WebSocketServer{
		server: &http.Server{
			Addr:    cfg.Addr,
			Handler: engine,
		},
		cfg:     cfg,
		hub:     hub,
		adapter: adapter,
	}
}

// Name returns the server name.
func (s *WebSocketServer) Name() string {
	return "websocket"
}

// Run starts the WebSocket server and blocks until context is cancelled or error occurs.
func (s *WebSocketServer) Run(ctx context.Context) error {
	// Start hub in background with context for graceful shutdown
	go s.hub.Run(ctx)

	log.Infow("Starting WebSocket server", "addr", s.cfg.Addr)

	errCh := make(chan error, 1)
	go func() {
		if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		return nil
	}
}

// Shutdown gracefully shuts down the WebSocket server.
func (s *WebSocketServer) Shutdown(ctx context.Context) error {
	log.Infow("Shutting down WebSocket server", "addr", s.cfg.Addr)
	return s.server.Shutdown(ctx)
}

// Hub returns the WebSocket hub for external access.
func (s *WebSocketServer) Hub() *pkgws.Hub {
	return s.hub
}

// Adapter returns the JSON-RPC adapter for handler registration.
func (s *WebSocketServer) Adapter() *jsonrpc.Adapter {
	return s.adapter
}
