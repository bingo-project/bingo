// ABOUTME: gRPC auth method handlers.
// ABOUTME: Provides login and user-info endpoints for gRPC clients.

package grpc

import (
	"context"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/bingo-project/bingo/internal/pkg/errno"
	"github.com/bingo-project/bingo/internal/pkg/log"
	"github.com/bingo-project/bingo/internal/pkg/store"
	apiv1 "github.com/bingo-project/bingo/pkg/api/apiserver/v1"
	"github.com/bingo-project/bingo/pkg/contextx"
	v1 "github.com/bingo-project/bingo/pkg/proto/apiserver/v1/pb"
)

func (h *Handler) Login(ctx context.Context, req *v1.LoginRequest) (*v1.LoginReply, error) {
	log.C(ctx).Infow("Login function called.")

	loginReq := &apiv1.LoginRequest{
		Username: req.Username,
		Password: req.Password,
	}

	if err := validate.Struct(loginReq); err != nil {
		return nil, errno.ErrInvalidArgument
	}

	resp, err := h.b.Auth().Login(ctx, loginReq)
	if err != nil {
		return nil, err
	}

	return &v1.LoginReply{
		AccessToken: resp.AccessToken,
		TokenType:   "Bearer",
		ExpiresIn:   resp.ExpiresAt.Unix(),
	}, nil
}

func (h *Handler) UserInfo(ctx context.Context, req *v1.UserInfoRequest) (*v1.UserInfoReply, error) {
	log.C(ctx).Infow("UserInfo function called.")

	uid := contextx.UserID(ctx)
	if uid == "" {
		return nil, errno.ErrTokenInvalid
	}

	user, err := store.S.User().GetByUID(ctx, uid)
	if err != nil {
		return nil, errno.ErrUserNotFound
	}

	return &v1.UserInfoReply{
		Uid:       user.UID,
		Username:  user.Username,
		Nickname:  user.Nickname,
		Email:     user.Email,
		Phone:     user.Phone,
		Status:    int32(user.Status),
		CreatedAt: timestamppb.New(user.CreatedAt),
		UpdatedAt: timestamppb.New(user.UpdatedAt),
	}, nil
}
