// ABOUTME: WebSocket auth method handlers.
// ABOUTME: Provides login and user-info endpoints for WS clients.

package ws

import (
	"github.com/bingo-project/websocket"
	"github.com/bingo-project/websocket/jsonrpc"

	"github.com/bingo-project/bingo/internal/pkg/errno"
	"github.com/bingo-project/bingo/internal/pkg/log"
	"github.com/bingo-project/bingo/internal/pkg/store"
	v1 "github.com/bingo-project/bingo/pkg/api/apiserver/v1"
)

// Login handles user login and returns JWT token.
func (h *Handler) Login(c *websocket.Context) *jsonrpc.Response {
	log.C(c).Debugw("Login function called")

	var req v1.LoginRequest
	if err := c.BindValidate(&req); err != nil {
		return c.Error(errno.ErrInvalidArgument.WithMessage("%s", err.Error()))
	}

	// Validate platform for WebSocket login
	if !websocket.IsValidPlatform(req.Platform) {
		return c.Error(errno.ErrInvalidArgument.WithMessage("invalid platform: %s", req.Platform))
	}

	resp, err := h.b.Auth().Login(c, &req)
	if err != nil {
		return c.Error(err)
	}

	// Parse token and set login info for middleware
	tokenInfo, err := c.Client.ParseToken(resp.AccessToken)
	if err != nil {
		return c.Error(errno.ErrTokenInvalid)
	}
	c.SetLoginInfo(tokenInfo, req.Platform)

	return c.JSON(resp)
}

// UserInfo returns the current user's info.
func (h *Handler) UserInfo(c *websocket.Context) *jsonrpc.Response {
	uid := c.UserID()
	if uid == "" {
		return c.Error(errno.ErrTokenInvalid)
	}

	user, err := store.S.User().GetByUID(c, uid)
	if err != nil {
		return c.Error(errno.ErrUserNotFound)
	}

	return c.JSON(&v1.UserInfo{
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
