package user

import (
	"bingo/internal/apiserver/biz"
	"bingo/internal/apiserver/store"
	"bingo/pkg/auth"
)

const defaultMethods = "(GET)|(POST)|(PUT)|(DELETE)"

// UserController 是 user 模块在 Controller 层的实现，用来处理用户模块的请求.
type UserController struct {
	a *auth.Authz
	b biz.IBiz
}

// NewUserController 创建一个 user controller.
func NewUserController(ds store.IStore, a *auth.Authz) *UserController {
	return &UserController{a: a, b: biz.NewBiz(ds)}
}
