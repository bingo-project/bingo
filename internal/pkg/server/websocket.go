// ABOUTME: WebSocket server implementation.
// ABOUTME: Implements Server interface with WebSocket support via Gin.

package server

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/bingo-project/websocket"

	"github.com/bingo-project/bingo/internal/pkg/config"
	"github.com/bingo-project/bingo/internal/pkg/log"
)

// WebSocketServer implements Server interface for WebSocket protocol.
type WebSocketServer struct {
	server *http.Server
	cfg    *config.WebSocket
	hub    *websocket.Hub
}

// NewWebSocketServer creates a new WebSocket server.
// The caller is responsible for creating the Gin engine with routes and the Hub.
func NewWebSocketServer(cfg *config.WebSocket, engine *gin.Engine, hub *websocket.Hub) *WebSocketServer {
	return &WebSocketServer{
		server: &http.Server{
			Addr:    cfg.Addr,
			Handler: engine,
		},
		cfg: cfg,
		hub: hub,
	}
}

// Name returns the server name.
func (s *WebSocketServer) Name() string {
	return "websocket"
}

// Run starts the WebSocket server and blocks until context is canceled or error occurs.
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
func (s *WebSocketServer) Hub() *websocket.Hub {
	return s.hub
}
