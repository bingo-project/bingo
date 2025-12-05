package apiserver

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"bingo/internal/apiserver/router"
	"bingo/internal/apiserver/server"
	"bingo/internal/pkg/bootstrap"
	"bingo/internal/pkg/facade"
)

// run starts all enabled servers based on configuration.
func run() error {
	// Create context that listens for interrupt signals
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Initialize Gin engine and gRPC server
	ginEngine := initGinEngine()
	grpcServer := initGRPCServer()

	// Assemble servers based on configuration
	runner := server.Assemble(
		&facade.Config,
		server.WithGinEngine(ginEngine),
		server.WithGRPCServer(grpcServer),
	)

	// Run all servers
	return runner.Run(ctx)
}

// initGinEngine initializes the Gin engine with routes.
func initGinEngine() *gin.Engine {
	g := bootstrap.InitGin()
	installRouters(g)
	return g
}

// initGRPCServer initializes the gRPC server with services.
func initGRPCServer() *grpc.Server {
	opts := RegisterInterceptor()
	srv := grpc.NewServer(opts...)
	router.GRPC(srv)
	reflection.Register(srv)
	return srv
}
