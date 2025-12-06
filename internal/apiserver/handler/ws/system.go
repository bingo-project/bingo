// ABOUTME: WebSocket system method handlers.
// ABOUTME: Provides healthz and version endpoints for WS clients.

package ws

import (
	"context"
	"time"

	"github.com/bingo-project/component-base/version"
)

// Healthz returns server health status.
func (h *Handler) Healthz(ctx context.Context, req any) (any, error) {
	status, err := h.b.Servers().Status(ctx)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"status":     status,
		"serverTime": time.Now().Unix(),
	}, nil
}

// Version returns server version info.
func (h *Handler) Version(ctx context.Context, req any) (any, error) {
	return version.Get(), nil
}
