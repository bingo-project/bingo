package system

import (
	"context"

	"github.com/bingo-project/component-base/web/token"

	"github.com/bingo-project/bingo/internal/pkg/auth"
	"github.com/bingo-project/bingo/internal/pkg/errno"
	"github.com/bingo-project/bingo/internal/pkg/known"
	v1 "github.com/bingo-project/bingo/pkg/api/apiserver/v1"
)

func (b *adminBiz) Login(ctx context.Context, req *v1.LoginRequest) (*v1.LoginResponse, error) {
	// Get user
	user, err := b.ds.Admin().GetByUsername(ctx, req.Username)
	if err != nil {
		return nil, errno.ErrNotFound
	}

	// Check password
	err = auth.Compare(user.Password, req.Password)
	if err != nil {
		return nil, errno.ErrPasswordInvalid
	}

	// Generate token
	t, err := token.Sign(user.Username, known.RoleAdmin)
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
		return errno.ErrNotFound
	}

	// Check password
	if err := auth.Compare(userM.Password, req.PasswordOld); err != nil {
		return errno.ErrPasswordOldInvalid
	}

	// Update password
	userM.Password, _ = auth.Encrypt(req.PasswordNew)
	if err := b.ds.Admin().Update(ctx, userM, "password"); err != nil {
		return err
	}

	return nil
}
