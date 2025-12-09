// ABOUTME: gRPC server implementation as Runnable.
// ABOUTME: Wraps grpc.Server with graceful shutdown support.

package server

import (
	"context"
	"net"

	"google.golang.org/grpc"
)

// GRPCServer is a gRPC server that implements Runnable.
type GRPCServer struct {
	addr     string
	server   *grpc.Server
	listener net.Listener
}

// NewGRPCServer creates a new gRPC server.
func NewGRPCServer(addr string, server *grpc.Server) *GRPCServer {
	return &GRPCServer{
		addr:   addr,
		server: server,
	}
}

// Start starts the gRPC server and blocks until ctx is canceled.
func (s *GRPCServer) Start(ctx context.Context) error {
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	s.listener = ln

	errCh := make(chan error, 1)
	go func() {
		errCh <- s.server.Serve(ln)
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		s.server.GracefulStop()
		return nil
	}
}

// Name returns the server name for logging.
func (s *GRPCServer) Name() string {
	return "grpc"
}

// Addr returns the actual listen address.
func (s *GRPCServer) Addr() string {
	if s.listener != nil {
		return s.listener.Addr().String()
	}
	return s.addr
}
