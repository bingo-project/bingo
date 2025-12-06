// ABOUTME: WebSocket method registration for JSON-RPC.
// ABOUTME: Maps JSON-RPC methods to handler layer methods.

package router

import (
	"bingo/internal/apiserver/biz"
	wshandler "bingo/internal/apiserver/handler/ws"
	"bingo/pkg/api/apiserver/v1"
	"bingo/pkg/jsonrpc"
)

// TODO: 将 biz 层参数迁移到 proto.Message
// 当前使用 biz 层的 Go struct 请求/响应类型。
// 目标是迁移到 proto.Message，届时更新此处的类型注册。

// RegisterWSHandlers registers all JSON-RPC handlers with the adapter.
func RegisterWSHandlers(a *jsonrpc.Adapter, b biz.IBiz) {
	systemHandler := wshandler.NewSystemHandler(b)
	authHandler := wshandler.NewAuthHandler(b)

	// System methods (no auth required, handled specially in client.go)
	a.RegisterHandler("system.healthz", systemHandler.Healthz, &struct{}{})
	a.RegisterHandler("system.version", systemHandler.Version, &struct{}{})

	// Auth methods
	a.RegisterHandler("auth.login", authHandler.Login, &v1.LoginRequest{})
	a.RegisterHandler("auth.user-info", authHandler.UserInfo, &struct{}{})
}
