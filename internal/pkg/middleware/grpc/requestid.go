package interceptor

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"

	"github.com/bingo-project/bingo/internal/pkg/known"
	"github.com/bingo-project/bingo/internal/pkg/log"
	"github.com/bingo-project/bingo/pkg/contextx"
	"github.com/bingo-project/bingo/pkg/errorsx"
)

func RequestID(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	var requestID string
	md, _ := metadata.FromIncomingContext(ctx)

	// 从请求中获取请求 ID
	if requestIDs := md[known.XRequestID]; len(requestIDs) > 0 {
		requestID = requestIDs[0]
	}

	if requestID == "" {
		requestID = uuid.New().String()
	}

	// 将元数据设置为新的 incoming context
	ctx = metadata.NewIncomingContext(ctx, md)

	// 将请求 ID 设置到响应的 Header Metadata 中
	// grpc.SetHeader 会在 gRPC 方法响应中添加元数据（Metadata），
	// 此处将包含请求 ID 的 Metadata 设置到 Header 中。
	// 注意：grpc.SetHeader 仅设置数据，它不会立即发送给客户端。
	// Header Metadata 会在 RPC 响应返回时一并发送。
	_ = grpc.SetHeader(ctx, md)

	ctx = contextx.WithRequestID(ctx, requestID)

	// 继续处理请求
	res, err := handler(ctx, req)

	// 错误处理，附加请求 ID
	if err != nil {
		return res, errorsx.FromError(err).WithRequestID(requestID)
	}

	return res, nil
}

func ClientIP(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	client, ok := peer.FromContext(ctx)
	if !ok {
		log.C(ctx).Errorw("failed to parse peer from context")

		return handler(ctx, req)
	}

	ip := strings.Split(client.Addr.String(), ":")[0]
	ctx = contextx.WithClientIP(ctx, ip)

	return handler(ctx, req)
}
