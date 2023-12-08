package system

import (
	"context"

	"github.com/bingo-project/component-base/web/token"

	"bingo/internal/apiserver/global"
	"bingo/internal/pkg/errno"
	v1 "bingo/pkg/api/bingo/v1"
	"bingo/pkg/auth"
)

func (b *adminBiz) Login(ctx context.Context, req *v1.LoginRequest) (*v1.LoginResponse, error) {
	// Get user
	user, err := b.ds.Admins().Get(ctx, req.Username)
	if err != nil {
		return nil, errno.ErrResourceNotFound
	}

	// Check password
	err = auth.Compare(user.Password, req.Password)
	if err != nil {
		return nil, errno.ErrPasswordIncorrect
	}

	// Generate token
	t, err := token.Sign(user.Username, global.AuthAdmin)
	if err != nil {
		return nil, errno.ErrSignToken
	}

	return &v1.LoginResponse{Token: t.AccessToken}, nil
}

func (b *adminBiz) ChangePassword(ctx context.Context, username string, req *v1.ChangePasswordRequest) error {
	userM, err := b.ds.Admins().Get(ctx, username)
	if err != nil {
		return errno.ErrResourceNotFound
	}

	// Check password
	if err := auth.Compare(userM.Password, req.OldPassword); err != nil {
		return errno.ErrPasswordIncorrect
	}

	// Update password
	userM.Password, _ = auth.Encrypt(req.NewPassword)
	if err := b.ds.Admins().Update(ctx, userM); err != nil {
		return err
	}

	return nil
}
