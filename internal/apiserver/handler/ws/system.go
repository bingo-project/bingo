// ABOUTME: WebSocket system method handlers.
// ABOUTME: Provides healthz and version endpoints for WS clients.

package ws

import (
	"time"

	"github.com/bingo-project/component-base/version"
	"github.com/bingo-project/websocket"
	"github.com/bingo-project/websocket/jsonrpc"
)

// Healthz returns server health status.
func (h *Handler) Healthz(c *websocket.Context) *jsonrpc.Response {
	status, err := h.b.Servers().Status(c)
	if err != nil {
		return c.Error(err)
	}

	return c.JSON(map[string]any{
		"status":     status,
		"serverTime": time.Now().Unix(),
	})
}

// Version returns server version info.
func (h *Handler) Version(c *websocket.Context) *jsonrpc.Response {
	return c.JSON(version.Get())
}
