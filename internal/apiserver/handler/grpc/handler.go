// ABOUTME: gRPC handler for apiserver service.
// ABOUTME: Defines the handler struct and constructor.

package grpc

import (
	"bingo/internal/apiserver/biz"
	"bingo/internal/pkg/store"
	v1 "bingo/pkg/proto/apiserver/v1/pb"
)

type Handler struct {
	b biz.IBiz
	v1.UnimplementedApiServerServer
}

func NewHandler(ds store.IStore) *Handler {
	return &Handler{b: biz.NewBiz(ds)}
}
