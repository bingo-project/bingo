// ABOUTME: WebSocket auth method handlers.
// ABOUTME: Provides login and user-info endpoints for WS clients.

package ws

import (
	"context"

	"github.com/bingo-project/component-base/log"

	"bingo/internal/pkg/contextx"
	"bingo/internal/pkg/errno"
	"bingo/internal/pkg/store"
	"bingo/pkg/api/apiserver/v1"
)

// Login handles user login and returns JWT token.
func (h *Handler) Login(ctx context.Context, req any) (any, error) {
	log.C(ctx).Debugw("Login function called")

	r := req.(*v1.LoginRequest)

	return h.b.Auth().Login(ctx, r)
}

// UserInfo returns the current user's info.
func (h *Handler) UserInfo(ctx context.Context, req any) (any, error) {
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
