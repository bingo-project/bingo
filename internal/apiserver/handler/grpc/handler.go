// ABOUTME: gRPC handler for apiserver service.
// ABOUTME: Defines the handler struct and constructor.

package grpc

import (
	"github.com/go-playground/validator/v10"

	"github.com/bingo-project/bingo/internal/apiserver/biz"
	"github.com/bingo-project/bingo/internal/pkg/store"
	v1 "github.com/bingo-project/bingo/pkg/proto/apiserver/v1/pb"
)

var validate = validator.New(validator.WithRequiredStructEnabled())

func init() {
	validate.SetTagName("binding")
}

type Handler struct {
	b biz.IBiz
	v1.UnimplementedApiServerServer
}

func NewHandler(ds store.IStore) *Handler {
	return &Handler{b: biz.NewBiz(ds)}
}
