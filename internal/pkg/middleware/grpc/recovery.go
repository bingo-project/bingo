package interceptor

import (
	"context"
	"fmt"
	"runtime/debug"

	"google.golang.org/grpc"

	"github.com/bingo-project/bingo/internal/pkg/log"
)

// Recovery catch panic & recover.
func Recovery(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	defer func() {
		if err := recover(); err != nil {
			log.C(ctx).Errorw("recovery", "method", info.FullMethod, "req", req, "err", err)
			fmt.Println(string(debug.Stack()))
		}
	}()

	return handler(ctx, req)
}
