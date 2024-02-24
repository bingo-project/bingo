package interceptor

import (
	"context"
	"time"

	"github.com/bingo-project/component-base/log"
	"google.golang.org/grpc"
)

func Logger(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()
	resp, err := handler(ctx, req)

	log.C(ctx).Infow(
		"interceptor.Logger request",
		"method", info.FullMethod,
		"cost", time.Since(start),
		"req", req,
		"resp", resp,
	)

	return resp, err
}
