// ABOUTME: WebSocket system method handlers.
// ABOUTME: Provides healthz and version endpoints for WS clients.

package ws

import (
	"time"

	"github.com/bingo-project/component-base/version"

	"bingo/pkg/jsonrpc"
	"bingo/pkg/ws"
)

// Healthz returns server health status.
func (h *Handler) Healthz(mc *ws.Context) *jsonrpc.Response {
	status, err := h.b.Servers().Status(mc.Ctx)
	if err != nil {
		return jsonrpc.NewErrorResponse(mc.Request.ID, err)
	}

	return jsonrpc.NewResponse(mc.Request.ID, map[string]any{
		"status":     status,
		"serverTime": time.Now().Unix(),
	})
}

// Version returns server version info.
func (h *Handler) Version(mc *ws.Context) *jsonrpc.Response {
	return jsonrpc.NewResponse(mc.Request.ID, version.Get())
}
