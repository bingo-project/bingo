// ABOUTME: WebSocket method registration for JSON-RPC.
// ABOUTME: Maps JSON-RPC methods to Biz layer handlers.

package ws

import (
	"bingo/internal/apiserver/biz"
	"bingo/pkg/jsonrpc"
)

// TODO(方案B): 迁移到 proto.Message
// 当前使用 biz 层的 Go struct 请求/响应类型。
// 目标是迁移到 proto.Message，届时更新此处的类型注册。

// RegisterHandlers registers all JSON-RPC handlers with the adapter.
func RegisterHandlers(a *jsonrpc.Adapter, b biz.IBiz) {
	// Auth methods
	// jsonrpc.Register(a, "auth.login", b.Auth().Login)
	// jsonrpc.Register(a, "auth.register", b.Auth().Register)

	// User methods
	// jsonrpc.Register(a, "user.list", b.Users().List)
	// jsonrpc.Register(a, "user.get", b.Users().Get)

	// Note: Uncomment and adapt handlers as needed.
	// The biz layer methods need to match the jsonrpc.Register signature:
	// func(context.Context, *RequestType) (*ResponseType, error)

	// Example placeholder handler
	// jsonrpc.Register(a, "system.ping", func(ctx context.Context, req *struct{}) (*struct{ Pong string }, error) {
	// 	return &struct{ Pong string }{Pong: "pong"}, nil
	// })
}
