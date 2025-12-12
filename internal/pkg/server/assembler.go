// ABOUTME: Server assembler based on configuration.
// ABOUTME: Creates Runner with enabled servers according to config.

package server

import (
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"

	"github.com/bingo-project/websocket"

	"github.com/bingo-project/bingo/internal/pkg/config"
)

// AssemblerOption configures the assembler.
type AssemblerOption func(*assemblerConfig)

type assemblerConfig struct {
	ginEngine  *gin.Engine
	grpcServer *grpc.Server
	wsEngine   *gin.Engine
	wsHub      *websocket.Hub
}

// WithGinEngine sets the Gin engine for HTTP server.
func WithGinEngine(engine *gin.Engine) AssemblerOption {
	return func(c *assemblerConfig) {
		c.ginEngine = engine
	}
}

// WithGRPCServer sets the gRPC server instance.
func WithGRPCServer(server *grpc.Server) AssemblerOption {
	return func(c *assemblerConfig) {
		c.grpcServer = server
	}
}

// WithWebSocket sets the Gin engine and Hub for WebSocket server.
// The caller is responsible for creating the engine with routes registered.
func WithWebSocket(engine *gin.Engine, hub *websocket.Hub) AssemblerOption {
	return func(c *assemblerConfig) {
		c.wsEngine = engine
		c.wsHub = hub
	}
}

// Assemble creates a Runner with servers based on configuration.
func Assemble(cfg *config.Config, opts ...AssemblerOption) *Runner {
	ac := &assemblerConfig{}
	for _, opt := range opts {
		opt(ac)
	}

	var servers []Server

	// gRPC Server (must start before gateway)
	if cfg.GRPC != nil && cfg.GRPC.Enabled && ac.grpcServer != nil {
		servers = append(servers, NewGRPCServer(cfg.GRPC, ac.grpcServer))
	}

	// HTTP Server: standalone mode or gateway mode
	if cfg.HTTP != nil && cfg.HTTP.Enabled {
		switch cfg.HTTP.Mode {
		case "gateway":
			// Gateway mode: proxy HTTP to gRPC
			if cfg.GRPC != nil && cfg.GRPC.Enabled {
				servers = append(servers, NewGatewayServer(cfg.HTTP, cfg.GRPC))
			}
		default:
			// Standalone mode: direct HTTP handling
			if ac.ginEngine != nil {
				servers = append(servers, NewHTTPServer(cfg.HTTP, ac.ginEngine))
			}
		}
	}

	// WebSocket Server
	if cfg.WebSocket != nil && cfg.WebSocket.Enabled && ac.wsEngine != nil && ac.wsHub != nil {
		servers = append(servers, NewWebSocketServer(cfg.WebSocket, ac.wsEngine, ac.wsHub))
	}

	return NewRunner(servers...)
}
