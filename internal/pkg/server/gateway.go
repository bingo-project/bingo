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
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"bingo/internal/pkg/config"
	middleware "bingo/internal/pkg/middleware/http"
	pb "bingo/pkg/proto/apiserver/v1/pb"
)

// GatewayServer implements Server interface for gRPC-Gateway.
type GatewayServer struct {
	server   *http.Server
	httpCfg  *config.HTTP
	grpcCfg  *config.GRPC
	grpcAddr string
}

// NewGatewayServer creates a new gRPC-Gateway server.
func NewGatewayServer(httpCfg *config.HTTP, grpcCfg *config.GRPC) *GatewayServer {
	return &GatewayServer{
		httpCfg:  httpCfg,
		grpcCfg:  grpcCfg,
		grpcAddr: grpcCfg.Addr,
	}
}

// Name returns the server name.
func (s *GatewayServer) Name() string {
	return "grpc-gateway"
}

// Run starts the gRPC-Gateway server and blocks until context is cancelled.
func (s *GatewayServer) Run(ctx context.Context) error {
	mux := runtime.NewServeMux()

	// Build dial options based on gRPC TLS config
	opts := s.buildDialOptions()

	// Register all service handlers
	if err := pb.RegisterApiServerHandlerFromEndpoint(ctx, mux, s.grpcAddr, opts); err != nil {
		return err
	}

	s.server = &http.Server{
		Addr:    s.httpCfg.Addr,
		Handler: middleware.CorsHandler(mux),
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

// buildDialOptions creates gRPC dial options based on TLS config.
func (s *GatewayServer) buildDialOptions() []grpc.DialOption {
	// Use TLS if gRPC server has TLS enabled
	if s.grpcCfg != nil && s.grpcCfg.TLS != nil && s.grpcCfg.TLS.Enabled {
		creds, err := credentials.NewClientTLSFromFile(s.grpcCfg.TLS.CertFile, "")
		if err != nil {
			log.Warnw("Failed to load TLS credentials, falling back to insecure", "err", err)
			return []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
		}
		log.Infow("gRPC-Gateway using TLS to connect to gRPC backend")
		return []grpc.DialOption{grpc.WithTransportCredentials(creds)}
	}

	// Default to insecure for local development
	return []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
}
