// ABOUTME: gRPC-Gateway server implementation for apiserver.
// ABOUTME: Proxies HTTP requests to gRPC backend using generated gateway handlers.

package server

import (
	"context"
	"errors"
	"net/http"

	"github.com/bingo-project/component-base/log"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"bingo/internal/pkg/config"
	pb "bingo/pkg/proto/apiserver/v1/pb"
)

// GatewayServer implements Server interface for gRPC-Gateway.
type GatewayServer struct {
	server   *http.Server
	httpCfg  *config.HTTP
	grpcAddr string
}

// NewGatewayServer creates a new gRPC-Gateway server.
func NewGatewayServer(httpCfg *config.HTTP, grpcAddr string) *GatewayServer {
	return &GatewayServer{
		httpCfg:  httpCfg,
		grpcAddr: grpcAddr,
	}
}

// Name returns the server name.
func (s *GatewayServer) Name() string {
	return "grpc-gateway"
}

// Run starts the gRPC-Gateway server and blocks until context is cancelled.
func (s *GatewayServer) Run(ctx context.Context) error {
	mux := runtime.NewServeMux()

	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	// Register all service handlers
	if err := pb.RegisterApiServerHandlerFromEndpoint(ctx, mux, s.grpcAddr, opts); err != nil {
		return err
	}

	s.server = &http.Server{
		Addr:    s.httpCfg.Addr,
		Handler: mux,
	}

	log.Infow("Starting gRPC-Gateway server", "addr", s.httpCfg.Addr, "grpc_backend", s.grpcAddr)

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

// Shutdown gracefully shuts down the gRPC-Gateway server.
func (s *GatewayServer) Shutdown(ctx context.Context) error {
	log.Infow("Shutting down gRPC-Gateway server", "addr", s.httpCfg.Addr)
	if s.server != nil {
		return s.server.Shutdown(ctx)
	}
	return nil
}
