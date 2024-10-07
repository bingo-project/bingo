package router

import (
	"google.golang.org/grpc"

	"bingo/internal/apiserver/grpc/v1/apiserver"
	"bingo/internal/apiserver/store"
	v1 "bingo/pkg/proto/v1/pb"
)

func GRPC(g *grpc.Server) {
	// ApiServer
	v1.RegisterApiServerServer(g, apiserver.New(store.S))
}
