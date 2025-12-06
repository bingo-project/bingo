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
func (h *Handler) Login(c *ws.Context) *jsonrpc.Response {
	log.C(c.Ctx).Debugw("Login function called")

	var req v1.LoginRequest
	if err := c.BindValidate(&req); err != nil {
		return jsonrpc.NewErrorResponse(c.Request.ID, errno.ErrBind.SetMessage(err.Error()))
	}

	resp, err := h.b.Auth().Login(c.Ctx, &req)
	if err != nil {
		return jsonrpc.NewErrorResponse(c.Request.ID, err)
	}

	return jsonrpc.NewResponse(c.Request.ID, resp)
}

// UserInfo returns the current user's info.
func (h *Handler) UserInfo(c *ws.Context) *jsonrpc.Response {
	uid := c.UserID()
	if uid == "" {
		return jsonrpc.NewErrorResponse(c.Request.ID, errno.ErrTokenInvalid)
	}

	user, err := store.S.User().GetByUID(c.Ctx, uid)
	if err != nil {
		return jsonrpc.NewErrorResponse(c.Request.ID, errno.ErrUserNotFound)
	}

	return jsonrpc.NewResponse(c.Request.ID, &v1.UserInfo{
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
