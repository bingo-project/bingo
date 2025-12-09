// ABOUTME: HTTP server implementation as Runnable.
// ABOUTME: Wraps gin.Engine with graceful shutdown support.

package server

import (
	"context"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
)

// HTTPServer is an HTTP server that implements Runnable.
type HTTPServer struct {
	addr     string
	engine   *gin.Engine
	server   *http.Server
	listener net.Listener
}

// NewHTTPServer creates a new HTTP server.
func NewHTTPServer(addr string, engine *gin.Engine) *HTTPServer {
	return &HTTPServer{
		addr:   addr,
		engine: engine,
	}
}

// Start starts the HTTP server and blocks until ctx is canceled.
func (s *HTTPServer) Start(ctx context.Context) error {
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
func (s *HTTPServer) Name() string {
	return "http"
}

// Addr returns the actual listen address.
// Only valid after Start is called.
func (s *HTTPServer) Addr() string {
	if s.listener != nil {
		return s.listener.Addr().String()
	}
	return s.addr
}
