package apiserver

import (
	"fmt"
	"net"

	"github.com/bingo-project/component-base/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"bingo/internal/apiserver/facade"
	"bingo/internal/apiserver/router"
)

type grpcAPIServer struct {
	*grpc.Server
	address string
}

// NewGRPC create a grpcAPIServer instance.
func NewGRPC() *grpcAPIServer {
	// 创建 GRPC Server 实例
	srv := grpc.NewServer()

	// 注册 GRPC 路由
	router.GRPC(srv)

	// 启动反射（使用 grpcurl 调试）
	reflection.Register(srv)

	return &grpcAPIServer{srv, facade.Config.GRPC.Addr}
}

func (s *grpcAPIServer) Run() {
	listen, err := net.Listen("tcp", s.address)
	if err != nil {
		log.Fatalw("Failed to listen: " + err.Error())
	}

	go func() {
		log.Infow("Start grpc server at " + s.address)

		if err := s.Serve(listen); err != nil {
			log.Fatalw("Failed to start grpc server: " + err.Error())
		}
	}()
}

func (s *grpcAPIServer) Close() {
	s.GracefulStop()

	log.Infow(fmt.Sprintf("GRPC server on %s stopped", s.address))
}
