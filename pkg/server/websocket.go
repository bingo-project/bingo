// ABOUTME: WebSocket server implementation as Runnable.
// ABOUTME: Wraps gin.Engine for WebSocket connections.

package server

import (
	"context"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
)

// WebSocketServer is a WebSocket server that implements Runnable.
type WebSocketServer struct {
	addr     string
	engine   *gin.Engine
	server   *http.Server
	listener net.Listener
}

// NewWebSocketServer creates a new WebSocket server.
func NewWebSocketServer(addr string, engine *gin.Engine) *WebSocketServer {
	return &WebSocketServer{
		addr:   addr,
		engine: engine,
	}
}

// Start starts the WebSocket server and blocks until ctx is canceled.
func (s *WebSocketServer) Start(ctx context.Context) error {
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	s.listener = ln

	s.server = &http.Server{
		Handler: s.engine,
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- s.server.Serve(ln)
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		return s.server.Shutdown(context.Background())
	}
}

// Name returns the server name for logging.
func (s *WebSocketServer) Name() string {
	return "websocket"
}

// Addr returns the actual listen address.
func (s *WebSocketServer) Addr() string {
	if s.listener != nil {
		return s.listener.Addr().String()
	}
	return s.addr
}
