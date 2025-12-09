// ABOUTME: Application entry point for admserver.
// ABOUTME: Assembles and runs all enabled servers.

package admserver

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/bingo-project/bingo/internal/pkg/facade"
	"github.com/bingo-project/bingo/internal/pkg/server"
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
