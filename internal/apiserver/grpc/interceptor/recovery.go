package interceptor

import (
	"context"
	"fmt"
	"runtime/debug"

	"github.com/bingo-project/component-base/log"
	"google.golang.org/grpc"
)

// Recovery catch panic & recover.
func Recovery(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	defer func() {
		if err := recover(); err != nil {
			log.C(ctx).Errorw("recovery", "method", info.FullMethod, "req", req, "err", err)
			fmt.Println(string(debug.Stack()))
		}
	}()

	return handler(ctx, req)
}
