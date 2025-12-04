package system

import (
	"context"

	"github.com/bingo-project/component-base/web/token"

	"bingo/internal/pkg/errno"
	"bingo/internal/pkg/global"
	v1 "bingo/pkg/api/apiserver/v1"
	"bingo/pkg/auth"
)

func (b *adminBiz) Login(ctx context.Context, req *v1.LoginRequest) (*v1.LoginResponse, error) {
	// Get user
	user, err := b.ds.Admin().GetByUsername(ctx, req.Username)
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

	resp := &v1.LoginResponse{
		AccessToken: t.AccessToken,
		ExpiresAt:   t.ExpiresAt,
	}

	return resp, nil
}

func (b *adminBiz) ChangePassword(ctx context.Context, username string, req *v1.ChangePasswordRequest) error {
	userM, err := b.ds.Admin().GetByUsername(ctx, username)
	if err != nil {
		return errno.ErrResourceNotFound
	}

	// Check password
	if err := auth.Compare(userM.Password, req.PasswordOld); err != nil {
		return errno.ErrPasswordOldIncorrect
	}

	// Update password
	userM.Password, _ = auth.Encrypt(req.PasswordNew)
	if err := b.ds.Admin().Update(ctx, userM, "password"); err != nil {
		return err
	}

	return nil
}
