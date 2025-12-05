package router

import (
	"google.golang.org/grpc"

	grpchandler "bingo/internal/apiserver/handler/grpc"
	"bingo/internal/pkg/store"
	v1 "bingo/pkg/proto/apiserver/v1/pb"
)

func GRPC(g *grpc.Server) {
	// ApiServer
	v1.RegisterApiServerServer(g, grpchandler.New(store.S))
}
