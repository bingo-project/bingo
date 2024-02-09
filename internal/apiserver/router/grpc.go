package router

import (
	"google.golang.org/grpc"

	"bingo/internal/apiserver/grpc/controller/v1/apiserver"
	v1 "bingo/internal/apiserver/grpc/proto/v1"
	"bingo/internal/apiserver/store"
)

func GRPC(g *grpc.Server) {
	// ApiServer
	v1.RegisterApiServerServer(g, apiserver.New(store.S))
}
