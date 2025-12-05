// ABOUTME: gRPC server implementation for apiserver.
// ABOUTME: Implements Server interface with gRPC framework.

package server

import (
	"context"
	"net"

	"github.com/bingo-project/component-base/log"
	"google.golang.org/grpc"

	"bingo/internal/pkg/config"
)

// GRPCServer implements Server interface for gRPC protocol.
type GRPCServer struct {
	server *grpc.Server
	cfg    *config.GRPC
}

// NewGRPCServer creates a new gRPC server.
func NewGRPCServer(cfg *config.GRPC, server *grpc.Server) *GRPCServer {
	return &GRPCServer{
		server: server,
		cfg:    cfg,
	}
}

// Name returns the server name.
func (s *GRPCServer) Name() string {
	return "grpc"
}

// Run starts the gRPC server and blocks until context is cancelled or error occurs.
func (s *GRPCServer) Run(ctx context.Context) error {
	listen, err := net.Listen("tcp", s.cfg.Addr)
	if err != nil {
		return err
	}

	log.Infow("Starting gRPC server", "addr", s.cfg.Addr)

	errCh := make(chan error, 1)
	go func() {
		if err := s.server.Serve(listen); err != nil {
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

// Shutdown gracefully shuts down the gRPC server.
func (s *GRPCServer) Shutdown(ctx context.Context) error {
	log.Infow("Shutting down gRPC server", "addr", s.cfg.Addr)

	done := make(chan struct{})
	go func() {
		s.server.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		s.server.Stop()
		return ctx.Err()
	}
}
