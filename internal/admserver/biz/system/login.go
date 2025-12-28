package system

import (
	"context"
	"fmt"
	"time"

	"github.com/bingo-project/component-base/web/token"
	"github.com/google/uuid"

	"github.com/bingo-project/bingo/internal/pkg/auth"
	"github.com/bingo-project/bingo/internal/pkg/errno"
	"github.com/bingo-project/bingo/internal/pkg/facade"
	"github.com/bingo-project/bingo/internal/pkg/known"
	"github.com/bingo-project/bingo/internal/pkg/model"
	v1 "github.com/bingo-project/bingo/pkg/api/apiserver/v1"
)

func (b *adminBiz) Login(ctx context.Context, req *v1.LoginRequest) (*v1.LoginResponse, error) {
	// Get user
	user, err := b.ds.Admin().GetByUsername(ctx, req.Account)
	if err != nil {
		return nil, errno.ErrNotFound
	}

	// Check password
	err = auth.Compare(user.Password, req.Password)
	if err != nil {
		return nil, errno.ErrPasswordInvalid
	}

	// Get current role
	role, err := b.ds.SysRole().GetByName(ctx, user.RoleName)
	if err == nil && role.RequireTOTP {
		// Role requires TOTP
		if user.GoogleStatus != string(model.GoogleStatusEnabled) {
			return nil, errno.ErrTOTPRequired
		}

		// Generate temporary TOTP token
		totpToken := uuid.New().String()
		redisKey := fmt.Sprintf("admin:totp_token:%s", totpToken)
		facade.Cache.Set(redisKey, user.Username, time.Minute*5)

		return &v1.LoginResponse{
			RequireTOTP: true,
			TOTPToken:   totpToken,
		}, nil
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

func (b *adminBiz) LoginWithTOTP(ctx context.Context, req *v1.TOTPLoginRequest) (*v1.LoginResponse, error) {
	// Verify TOTP token
	redisKey := fmt.Sprintf("admin:totp_token:%s", req.TOTPToken)
	usernameValue := facade.Cache.Get(redisKey)
	if usernameValue == nil {
		return nil, errno.ErrTOTPTokenInvalid
	}

	username, ok := usernameValue.(string)
	if !ok {
		return nil, errno.ErrTOTPTokenInvalid
	}

	// Get user
	user, err := b.ds.Admin().GetByUsername(ctx, username)
	if err != nil {
		return nil, errno.ErrNotFound
	}

	// Verify TOTP code
	if user.GoogleStatus != string(model.GoogleStatusEnabled) || user.GoogleKey == "" {
		return nil, errno.ErrTOTPNotEnabled
	}

	secret, err := facade.AES.DecryptString(user.GoogleKey)
	if err != nil {
		return nil, err
	}

	if !auth.ValidateTOTP(req.Code, secret) {
		return nil, errno.ErrTOTPInvalid
	}

	// Delete TOTP token
	facade.Cache.Forget(redisKey)

	// Generate JWT
	t, err := token.Sign(user.Username, known.RoleAdmin)
	if err != nil {
		return nil, errno.ErrSignToken
	}

	return &v1.LoginResponse{
		AccessToken: t.AccessToken,
		ExpiresAt:   t.ExpiresAt,
	}, nil
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
