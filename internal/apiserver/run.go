// ABOUTME: Application entry point for apiserver.
// ABOUTME: Initializes HTTP, gRPC, and WebSocket servers based on configuration.

package apiserver

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/bingo-project/component-base/log"
	"github.com/gin-gonic/gin"
	gm "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"

	"bingo/internal/apiserver/biz"
	wshandler "bingo/internal/apiserver/handler/ws"
	"bingo/internal/apiserver/router"
	"bingo/internal/pkg/bootstrap"
	"bingo/internal/pkg/config"
	"bingo/internal/pkg/facade"
	interceptor "bingo/internal/pkg/middleware/grpc"
	"bingo/internal/pkg/server"
	"bingo/internal/pkg/store"
	"bingo/pkg/jsonrpc"
	"bingo/pkg/ws"
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

// initGinEngine initializes the Gin engine with routes.
func initGinEngine() *gin.Engine {
	g := bootstrap.InitGin()

	// Swagger
	if facade.Config.Feature.ApiDoc {
		router.MapSwagRouters(g)
	}

	// Common router
	router.MapCommonRouters(g)

	// Api
	router.MapApiRouters(g)

	return g
}

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

// initWebSocket initializes the WebSocket engine and hub.
func initWebSocket() (*gin.Engine, *ws.Hub) {
	// Create hub
	hub := ws.NewHub()

	// Create JSON-RPC adapter and register handlers
	adapter := jsonrpc.NewAdapter()
	bizInstance := biz.NewBiz(store.S)
	router.RegisterHandlers(adapter, bizInstance)

	// Create Gin engine for WebSocket
	engine := bootstrap.InitGin()

	// Register WebSocket route
	handler := wshandler.NewHandler(hub, adapter, facade.Config.WebSocket)
	engine.GET("/ws", handler.ServeWS)

	return engine, hub
}
