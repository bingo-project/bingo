// ABOUTME: WebSocket method registration for JSON-RPC.
// ABOUTME: Maps JSON-RPC methods to Biz layer handlers.

package router

import (
	"bingo/internal/admserver/biz"
	"bingo/pkg/jsonrpc"
)

// RegisterWSHandlers registers all JSON-RPC handlers with the adapter.
func RegisterWSHandlers(a *jsonrpc.Adapter, b biz.IBiz) {
	// Admin methods can be registered here as needed.
	// Example:
	// jsonrpc.Register(a, "admin.stats", b.System().Stats)
}
