// ABOUTME: WebSocket auth method handlers.
// ABOUTME: Provides login and user-info endpoints for WS clients.

package ws

import (
	"context"

	"bingo/internal/apiserver/biz"
	"bingo/internal/pkg/contextx"
	"bingo/internal/pkg/errno"
	"bingo/internal/pkg/store"
	"bingo/pkg/api/apiserver/v1"
)

// AuthHandler handles auth-related WebSocket methods.
type AuthHandler struct {
	b biz.IBiz
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(b biz.IBiz) *AuthHandler {
	return &AuthHandler{b: b}
}

// Login handles user login and returns JWT token.
func (h *AuthHandler) Login(ctx context.Context, req any) (any, error) {
	r := req.(*v1.LoginRequest)
	return h.b.Auth().Login(ctx, r)
}

// UserInfo returns the current user's info.
func (h *AuthHandler) UserInfo(ctx context.Context, req any) (any, error) {
	uid := contextx.UserID(ctx)
	if uid == "" {
		return nil, errno.ErrTokenInvalid
	}

	user, err := store.S.User().GetByUID(ctx, uid)
	if err != nil {
		return nil, errno.ErrUserNotFound
	}

	return &v1.UserInfo{
		UID:       user.UID,
		Username:  user.Username,
		Nickname:  user.Nickname,
		Email:     user.Email,
		Phone:     user.Phone,
		Status:    int32(user.Status),
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}
