// ABOUTME: Server assembler based on configuration.
// ABOUTME: Creates Runner with enabled servers according to config.

package server

import (
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"

	"bingo/internal/pkg/config"
	"bingo/pkg/jsonrpc"
)

// AssemblerOption configures the assembler.
type AssemblerOption func(*assemblerConfig)

type assemblerConfig struct {
	ginEngine   *gin.Engine
	grpcServer  *grpc.Server
	rpcAdapter  *jsonrpc.Adapter
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

// WithJSONRPCAdapter sets the JSON-RPC adapter for WebSocket.
func WithJSONRPCAdapter(adapter *jsonrpc.Adapter) AssemblerOption {
	return func(c *assemblerConfig) {
		c.rpcAdapter = adapter
	}
}

// Assemble creates a Runner with servers based on configuration.
func Assemble(cfg *config.Config, opts ...AssemblerOption) *Runner {
	ac := &assemblerConfig{}
	for _, opt := range opts {
		opt(ac)
	}

	var servers []Server

	// HTTP Server (first to start, last to stop)
	if cfg.HTTP != nil && cfg.HTTP.Enabled && ac.ginEngine != nil {
		servers = append(servers, NewHTTPServer(cfg.HTTP, ac.ginEngine))
	}

	// gRPC Server
	if cfg.GRPC != nil && cfg.GRPC.Enabled && ac.grpcServer != nil {
		servers = append(servers, NewGRPCServer(cfg.GRPC, ac.grpcServer))
	}

	// WebSocket Server (last to start, first to stop)
	if cfg.WebSocket != nil && cfg.WebSocket.Enabled {
		adapter := ac.rpcAdapter
		if adapter == nil {
			adapter = jsonrpc.NewAdapter()
		}
		servers = append(servers, NewWebSocketServer(cfg.WebSocket, adapter))
	}

	return NewRunner(servers...)
}
