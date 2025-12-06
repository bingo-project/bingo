// ABOUTME: WebSocket system method handlers.
// ABOUTME: Provides healthz and version endpoints for WS clients.

package ws

import (
	"context"
	"time"

	"github.com/bingo-project/component-base/version"

	"bingo/internal/apiserver/biz"
)

// SystemHandler handles system-related WebSocket methods.
type SystemHandler struct {
	b biz.IBiz
}

// NewSystemHandler creates a new SystemHandler.
func NewSystemHandler(b biz.IBiz) *SystemHandler {
	return &SystemHandler{b: b}
}

// Healthz returns server health status.
func (h *SystemHandler) Healthz(ctx context.Context, req any) (any, error) {
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
func (h *SystemHandler) Version(ctx context.Context, req any) (any, error) {
	return version.Get(), nil
}
