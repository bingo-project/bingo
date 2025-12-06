// ABOUTME: gRPC server initialization for admserver.
// ABOUTME: Configures gRPC server with services, interceptors, and TLS.

package admserver

import (
	"github.com/bingo-project/component-base/log"
	gm "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"

	"bingo/internal/admserver/router"
	"bingo/internal/pkg/config"
	interceptor "bingo/internal/pkg/middleware/grpc"
)

// initGRPCServer initializes the gRPC server with services and TLS support.
func initGRPCServer(cfg *config.GRPC) *grpc.Server {
	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(gm.ChainUnaryServer(
			interceptor.RequestID,
			interceptor.ClientIP,
			interceptor.Logger,
			interceptor.Recovery,
		)),
	}

	// Add TLS credentials if enabled
	if cfg != nil && cfg.TLS != nil && cfg.TLS.Enabled {
		creds, err := credentials.NewServerTLSFromFile(cfg.TLS.CertFile, cfg.TLS.KeyFile)
		if err != nil {
			log.Fatalw("Failed to load TLS credentials", "err", err)
		}
		opts = append(opts, grpc.Creds(creds))
		log.Infow("gRPC TLS enabled", "cert", cfg.TLS.CertFile)
	}

	srv := grpc.NewServer(opts...)

	// Register gRPC routes
	router.GRPC(srv)

	// Enable reflection for grpcurl debugging
	reflection.Register(srv)

	return srv
}
