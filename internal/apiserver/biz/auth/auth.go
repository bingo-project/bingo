package auth

import (
	"context"

	"github.com/bingo-project/component-base/web/token"

	v1 "bingo/internal/apiserver/http/request/v1"
	"bingo/internal/apiserver/store"
	"bingo/internal/pkg/errno"
	"bingo/pkg/auth"
)

// AuthBiz 定义了 user 模块在 biz 层所实现的方法.
type AuthBiz interface {
	Login(ctx context.Context, r *v1.LoginRequest) (*v1.LoginResponse, error)
	ChangePassword(ctx context.Context, username string, r *v1.ChangePasswordRequest) error
}

type authBiz struct {
	ds store.IStore
}

var _ AuthBiz = (*authBiz)(nil)

func NewAuth(ds store.IStore) *authBiz {
	return &authBiz{ds: ds}
}

func (b *authBiz) Login(ctx context.Context, req *v1.LoginRequest) (*v1.LoginResponse, error) {
	// Get user
	user, err := b.ds.Users().Get(ctx, req.Username)
	if err != nil {
		return nil, errno.ErrUserNotFound
	}

	// Check password
	err = auth.Compare(user.Password, req.Password)
	if err != nil {
		return nil, errno.ErrPasswordIncorrect
	}

	// Generate token
	t, err := token.Sign(user.Username, nil)
	if err != nil {
		return nil, errno.ErrSignToken
	}

	return &v1.LoginResponse{Token: t.AccessToken}, nil
}

func (b *authBiz) ChangePassword(ctx context.Context, username string, req *v1.ChangePasswordRequest) error {
	userM, err := b.ds.Users().Get(ctx, username)
	if err != nil {
		return err
	}

	// Check password
	if err := auth.Compare(userM.Password, req.PasswordOld); err != nil {
		return errno.ErrPasswordIncorrect
	}

	// Update password
	userM.Password, _ = auth.Encrypt(req.PasswordNew)
	if err := b.ds.Users().Update(ctx, userM); err != nil {
		return err
	}

	return nil
}
