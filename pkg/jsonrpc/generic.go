// ABOUTME: Generic handler registration for JSON-RPC adapter.
// ABOUTME: Provides type-safe registration using Go generics.

package jsonrpc

import (
	"context"
)

// TODO(方案B): 迁移到 proto.Message 类型约束
// 当前使用 any 类型约束以兼容现有 biz 层的 Go struct。
// 目标签名: func Register[Req, Resp proto.Message](...)

// Register provides type-safe handler registration using generics.
func Register[Req, Resp any](
	a *Adapter,
	method string,
	handler func(context.Context, Req) (Resp, error),
) {
	var zero Req
	a.RegisterHandler(method, func(ctx context.Context, req any) (any, error) {
		return handler(ctx, req.(Req))
	}, zero)
}
