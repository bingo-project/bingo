// ABOUTME: WebSocket system method handlers.
// ABOUTME: Provides healthz and version endpoints for WS clients.

package ws

import (
	"time"

	"github.com/bingo-project/component-base/version"

	"github.com/bingo-project/bingo/pkg/jsonrpc"
	"github.com/bingo-project/bingo/pkg/ws"
)

// Healthz returns server health status.
func (h *Handler) Healthz(c *ws.Context) *jsonrpc.Response {
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
func (h *Handler) Version(c *ws.Context) *jsonrpc.Response {
	return c.JSON(version.Get())
}
