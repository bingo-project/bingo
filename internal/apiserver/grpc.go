// ABOUTME: gRPC server initialization for apiserver.
// ABOUTME: Configures gRPC server with services, interceptors, and TLS.

package apiserver

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"

	bizauth "bingo/internal/apiserver/biz/auth"
	"bingo/internal/apiserver/router"
	"bingo/internal/pkg/auth"
	"bingo/internal/pkg/config"
	"bingo/internal/pkg/log"
	interceptor "bingo/internal/pkg/middleware/grpc"
	"bingo/internal/pkg/store"
)

// gRPC methods that don't require authentication.
var publicMethods = map[string]bool{
	"/apiserver.v1.ApiServer/Healthz": true,
	"/apiserver.v1.ApiServer/Version": true,
	"/apiserver.v1.ApiServer/Login":   true,
}

// initGRPCServer initializes the gRPC server with services and TLS support.
func initGRPCServer(cfg *config.GRPC) *grpc.Server {
	loader := bizauth.NewUserLoader(store.S)
	authn := auth.New(loader)

	opts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			interceptor.RequestID,
			interceptor.ClientIP,
			interceptor.Logger,
			interceptor.Recovery,
			interceptor.Validator,
			auth.UnaryInterceptor(authn, publicMethods),
		),
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
