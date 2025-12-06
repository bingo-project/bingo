// ABOUTME: Application entry point for apiserver.
// ABOUTME: Assembles and runs all enabled servers.

package apiserver

import (
	"context"
	"os/signal"
	"syscall"

	"bingo/internal/pkg/facade"
	"bingo/internal/pkg/server"
)

// run starts all enabled servers based on configuration.
func run() error {
	// Create context that listens for interrupt signals
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Initialize servers
	ginEngine := initGinEngine()
	grpcServer := initGRPCServer(facade.Config.GRPC)
	wsEngine, wsHub := initWebSocket()

	// Assemble servers based on configuration
	runner := server.Assemble(
		&facade.Config,
		server.WithGinEngine(ginEngine),
		server.WithGRPCServer(grpcServer),
		server.WithWebSocket(wsEngine, wsHub),
	)

	// Run all servers
	return runner.Run(ctx)
}
