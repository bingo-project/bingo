package interceptor

import (
	"context"
	"strings"

	"github.com/bingo-project/component-base/log"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

func RequestID(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	rid := uuid.New().String()
	ctx = context.WithValue(ctx, log.KeyTrace, rid)

	return handler(ctx, req)
}

func ClientIP(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	client, ok := peer.FromContext(ctx)
	if !ok {
		log.C(ctx).Errorw("failed to parse peer from context")

		return handler(ctx, req)
	}

	ip := strings.Split(client.Addr.String(), ":")[0]
	ctx = context.WithValue(ctx, log.KeyIP, ip)

	return handler(ctx, req)
}
