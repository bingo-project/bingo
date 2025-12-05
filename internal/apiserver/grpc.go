package apiserver

import (
	"fmt"
	"net"

	"github.com/bingo-project/component-base/log"
	gm "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"

	"bingo/internal/apiserver/router"
	"bingo/internal/pkg/config"
	"bingo/internal/pkg/facade"
	"bingo/internal/pkg/grpc/interceptor"
)

type grpcAPIServer struct {
	*grpc.Server
	address string
}

// NewGRPC create a grpcAPIServer instance.
func NewGRPC() *grpcAPIServer {
	return NewGRPCWithConfig(facade.Config.GRPC)
}

// NewGRPCWithConfig creates a grpcAPIServer with explicit config.
func NewGRPCWithConfig(cfg *config.GRPC) *grpcAPIServer {
	// Build server options
	opts := buildServerOptions(cfg)

	// 创建 GRPC Server 实例
	srv := grpc.NewServer(opts...)

	// 注册 GRPC 路由
	router.GRPC(srv)

	// 启动反射（使用 grpcurl 调试）
	reflection.Register(srv)

	return &grpcAPIServer{srv, cfg.Addr}
}

func buildServerOptions(cfg *config.GRPC) []grpc.ServerOption {
	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(gm.ChainUnaryServer(
			interceptor.RequestID, // TraceID
			interceptor.ClientIP,  // Client IP
			interceptor.Logger,    // Log
			interceptor.Recovery,  // Panic recover
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

	return opts
}

func (s *grpcAPIServer) Run() {
	listen, err := net.Listen("tcp", s.address)
	if err != nil {
		log.Fatalw("Failed to listen: " + err.Error())
	}

	log.Infow("Start grpc server at " + s.address)

	go func() {
		if err := s.Serve(listen); err != nil {
			log.Fatalw("Failed to start grpc server: " + err.Error())
		}
	}()
}

func (s *grpcAPIServer) Close() {
	s.GracefulStop()

	log.Infow(fmt.Sprintf("GRPC server on %s stopped", s.address))
}
