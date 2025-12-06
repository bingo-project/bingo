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
func (h *Handler) Healthz(c *ws.Context) *jsonrpc.Response {
	status, err := h.b.Servers().Status(c.Ctx)
	if err != nil {
		return jsonrpc.NewErrorResponse(c.Request.ID, err)
	}

	return jsonrpc.NewResponse(c.Request.ID, map[string]any{
		"status":     status,
		"serverTime": time.Now().Unix(),
	})
}

// Version returns server version info.
func (h *Handler) Version(c *ws.Context) *jsonrpc.Response {
	return jsonrpc.NewResponse(c.Request.ID, version.Get())
}
