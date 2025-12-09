// ABOUTME: Application entry point for apiserver.
// ABOUTME: Assembles and runs all enabled servers using pkg/app.

package apiserver

import (
	"bingo/internal/pkg/facade"
	"bingo/pkg/app"
	"bingo/pkg/server"
)

// run starts all enabled servers based on configuration.
func run() error {
	cfg := &facade.Config

	application, err := app.New(
		app.WithConfig(cfg),
	)
	if err != nil {
		return err
	}

	// Add servers based on configuration
	if cfg.GRPC.Enabled {
		application.Add(server.NewGRPCServer(cfg.GRPC.Addr, initGRPCServer(cfg.GRPC)))
	}

	if cfg.HTTP.Enabled {
		application.Add(server.NewHTTPServer(cfg.HTTP.Addr, initGinEngine()))
	}

	if cfg.WebSocket.Enabled {
		wsEngine, hub := initWebSocket()
		application.Add(RunnableFunc(hub.Run))
		application.Add(server.NewWebSocketServer(cfg.WebSocket.Addr, wsEngine))
	}

	return application.Run(app.SetupSignalHandler())
}
