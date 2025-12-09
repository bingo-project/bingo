package router

import (
	"google.golang.org/grpc"

	grpchandler "github.com/bingo-project/bingo/internal/apiserver/handler/grpc"
	"github.com/bingo-project/bingo/internal/pkg/store"
	v1 "github.com/bingo-project/bingo/pkg/proto/apiserver/v1/pb"
)

func GRPC(g *grpc.Server) {
	// ApiServer
	v1.RegisterApiServerServer(g, grpchandler.NewHandler(store.S))
}
