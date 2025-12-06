// ABOUTME: WebSocket auth method handlers.
// ABOUTME: Provides login and user-info endpoints for WS clients.

package ws

import (
	"github.com/bingo-project/component-base/log"

	"bingo/internal/pkg/errno"
	"bingo/internal/pkg/store"
	"bingo/pkg/api/apiserver/v1"
	"bingo/pkg/jsonrpc"
	"bingo/pkg/ws"
)

// Login handles user login and returns JWT token.
func (h *Handler) Login(mc *ws.Context) *jsonrpc.Response {
	log.C(mc.Ctx).Debugw("Login function called")

	var req v1.LoginRequest
	if err := mc.BindValidate(&req); err != nil {
		return jsonrpc.NewErrorResponse(mc.Request.ID, errno.ErrBind.SetMessage(err.Error()))
	}

	resp, err := h.b.Auth().Login(mc.Ctx, &req)
	if err != nil {
		return jsonrpc.NewErrorResponse(mc.Request.ID, err)
	}

	return jsonrpc.NewResponse(mc.Request.ID, resp)
}

// UserInfo returns the current user's info.
func (h *Handler) UserInfo(mc *ws.Context) *jsonrpc.Response {
	uid := mc.UserID()
	if uid == "" {
		return jsonrpc.NewErrorResponse(mc.Request.ID, errno.ErrTokenInvalid)
	}

	user, err := store.S.User().GetByUID(mc.Ctx, uid)
	if err != nil {
		return jsonrpc.NewErrorResponse(mc.Request.ID, errno.ErrUserNotFound)
	}

	return jsonrpc.NewResponse(mc.Request.ID, &v1.UserInfo{
		UID:       user.UID,
		Username:  user.Username,
		Nickname:  user.Nickname,
		Email:     user.Email,
		Phone:     user.Phone,
		Status:    int32(user.Status),
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	})
}
