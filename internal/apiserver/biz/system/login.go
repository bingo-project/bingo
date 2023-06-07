package system

import (
	"context"

	"bingo/internal/pkg/errno"
	v1 "bingo/pkg/api/bingo/v1"
	"bingo/pkg/auth"
	"bingo/pkg/token"
)

func (b *adminBiz) Login(ctx context.Context, r *v1.LoginRequest) (*v1.LoginResponse, error) {
	// Get user
	user, err := b.ds.Admins().Get(ctx, r.Username)
	if err != nil {
		return nil, errno.ErrAdminNotFound
	}

	// Check password
	err = auth.Compare(user.Password, r.Password)
	if err != nil {
		return nil, errno.ErrPasswordIncorrect
	}

	// Generate token
	t, err := token.Sign(user.Username)
	if err != nil {
		return nil, errno.ErrSignToken
	}

	return &v1.LoginResponse{Token: t}, nil
}

func (b *adminBiz) ChangePassword(ctx context.Context, username string, r *v1.ChangePasswordRequest) error {
	userM, err := b.ds.Admins().Get(ctx, username)
	if err != nil {
		return errno.ErrAdminNotFound
	}

	// Check password
	if err := auth.Compare(userM.Password, r.OldPassword); err != nil {
		return errno.ErrPasswordIncorrect
	}

	// Update password
	userM.Password, _ = auth.Encrypt(r.NewPassword)
	if err := b.ds.Admins().Update(ctx, userM); err != nil {
		return err
	}

	return nil
}
