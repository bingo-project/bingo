// ABOUTME: WebSocket handler for admserver service.
// ABOUTME: Defines the handler struct and constructor.

package ws

import (
	"bingo/internal/admserver/biz"
	"bingo/internal/pkg/store"
)

// Handler handles WebSocket business methods.
type Handler struct {
	b biz.IBiz
}

// NewHandler creates a new WebSocket handler.
func NewHandler(ds store.IStore) *Handler {
	return &Handler{b: biz.NewBiz(ds)}
}
