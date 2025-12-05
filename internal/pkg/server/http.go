// ABOUTME: HTTP server implementation for apiserver.
// ABOUTME: Implements Server interface with Gin framework.

package server

import (
	"context"
	"errors"
	"net/http"

	"github.com/bingo-project/component-base/log"
	"github.com/gin-gonic/gin"

	"bingo/internal/pkg/config"
)

// HTTPServer implements Server interface for HTTP protocol.
type HTTPServer struct {
	server *http.Server
	cfg    *config.HTTP
}

// RouterInstaller is a function type for installing routes on Gin engine.
type RouterInstaller func(*gin.Engine)

// NewHTTPServer creates a new HTTP server.
func NewHTTPServer(cfg *config.HTTP, engine *gin.Engine) *HTTPServer {
	return &HTTPServer{
		server: &http.Server{
			Addr:    cfg.Addr,
			Handler: engine,
		},
		cfg: cfg,
	}
}

// Name returns the server name.
func (s *HTTPServer) Name() string {
	return "http"
}

// Run starts the HTTP server and blocks until context is cancelled or error occurs.
func (s *HTTPServer) Run(ctx context.Context) error {
	log.Infow("Starting HTTP server", "addr", s.cfg.Addr)

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

// Shutdown gracefully shuts down the HTTP server.
func (s *HTTPServer) Shutdown(ctx context.Context) error {
	log.Infow("Shutting down HTTP server", "addr", s.cfg.Addr)
	return s.server.Shutdown(ctx)
}
