// ABOUTME: Security settings business logic for admin TOTP management.
// ABOUTME: Provides TOTP verification services for administrator accounts.

package auth

import (
	"context"

	"github.com/bingo-project/bingo/internal/pkg/auth"
	"github.com/bingo-project/bingo/internal/pkg/errno"
	"github.com/bingo-project/bingo/internal/pkg/facade"
	"github.com/bingo-project/bingo/internal/pkg/model"
	"github.com/bingo-project/bingo/internal/pkg/store"
	v1 "github.com/bingo-project/bingo/pkg/api/apiserver/v1"
)

const (
	// TOTPIssuer TOTP发行者名称
	TOTPIssuer = "Bingo"
)

// SecurityBiz defines security settings operations.
type SecurityBiz interface {
	GetTOTPStatus(ctx context.Context, username string) (*v1.TOTPStatusResponse, error)
	SetupTOTP(ctx context.Context, username string) (*v1.TOTPSetupResponse, error)
	EnableTOTP(ctx context.Context, username string, code string) error
	VerifyTOTP(ctx context.Context, username string, code string) error
	DisableTOTP(ctx context.Context, username string, code string) error
}

type securityBiz struct {
	ds store.IStore
}

var _ SecurityBiz = (*securityBiz)(nil)

// NewSecurityBiz creates a new SecurityBiz instance.
func NewSecurityBiz(ds store.IStore) SecurityBiz {
	return &securityBiz{ds: ds}
}

// GetTOTPStatus returns TOTP enabled status.
func (b *securityBiz) GetTOTPStatus(ctx context.Context, username string) (*v1.TOTPStatusResponse, error) {
	admin, err := b.ds.Admin().GetByUsername(ctx, username)
	if err != nil {
		return nil, errno.ErrUserNotFound
	}

	return &v1.TOTPStatusResponse{
		Enabled: admin.GoogleStatus == string(model.GoogleStatusEnabled),
	}, nil
}

// SetupTOTP generates a new TOTP secret for binding.
func (b *securityBiz) SetupTOTP(ctx context.Context, username string) (*v1.TOTPSetupResponse, error) {
	admin, err := b.ds.Admin().GetByUsername(ctx, username)
	if err != nil {
		return nil, errno.ErrUserNotFound
	}

	if admin.GoogleStatus == string(model.GoogleStatusEnabled) {
		return nil, errno.ErrTOTPAlreadyEnabled
	}

	// Generate TOTP secret
	accountName := username

	key, err := auth.GenerateTOTPSecret(auth.TOTPOptions{
		Issuer:      TOTPIssuer,
		AccountName: accountName,
	})
	if err != nil {
		return nil, err
	}

	// Encrypt and store secret temporarily
	encryptedSecret, err := facade.AES.EncryptString(key.Secret())
	if err != nil {
		return nil, err
	}

	admin.GoogleKey = encryptedSecret
	admin.GoogleStatus = string(model.GoogleStatusDisabled)

	if err := b.ds.Admin().Update(ctx, admin, "google_key", "google_status"); err != nil {
		return nil, err
	}

	return &v1.TOTPSetupResponse{
		Secret:     key.Secret(),
		OtpauthURL: key.URL(),
	}, nil
}

// EnableTOTP enables TOTP after verifying the code.
func (b *securityBiz) EnableTOTP(ctx context.Context, username string, code string) error {
	admin, err := b.ds.Admin().GetByUsername(ctx, username)
	if err != nil {
		return errno.ErrUserNotFound
	}

	if admin.GoogleStatus == string(model.GoogleStatusEnabled) {
		return errno.ErrTOTPAlreadyEnabled
	}

	if admin.GoogleKey == "" {
		return errno.ErrTOTPNotEnabled
	}

	// Decrypt secret
	secret, err := facade.AES.DecryptString(admin.GoogleKey)
	if err != nil {
		return err
	}

	// Verify TOTP code
	if !auth.ValidateTOTP(code, secret) {
		return errno.ErrTOTPInvalid
	}

	// Enable TOTP
	admin.GoogleStatus = string(model.GoogleStatusEnabled)

	return b.ds.Admin().Update(ctx, admin, "google_status")
}

// VerifyTOTP verifies a TOTP code.
func (b *securityBiz) VerifyTOTP(ctx context.Context, username string, code string) error {
	admin, err := b.ds.Admin().GetByUsername(ctx, username)
	if err != nil {
		return errno.ErrUserNotFound
	}

	if admin.GoogleStatus != string(model.GoogleStatusEnabled) {
		return errno.ErrTOTPNotEnabled
	}

	// Decrypt secret
	secret, err := facade.AES.DecryptString(admin.GoogleKey)
	if err != nil {
		return err
	}

	if !auth.ValidateTOTP(code, secret) {
		return errno.ErrTOTPInvalid
	}

	return nil
}

// DisableTOTP disables TOTP after verifying current TOTP code.
func (b *securityBiz) DisableTOTP(ctx context.Context, username string, code string) error {
	admin, err := b.ds.Admin().GetByUsername(ctx, username)
	if err != nil {
		return errno.ErrUserNotFound
	}

	if admin.GoogleStatus != string(model.GoogleStatusEnabled) {
		return errno.ErrTOTPNotEnabled
	}

	// Decrypt and verify TOTP
	secret, err := facade.AES.DecryptString(admin.GoogleKey)
	if err != nil {
		return err
	}

	if !auth.ValidateTOTP(code, secret) {
		return errno.ErrTOTPInvalid
	}

	// Disable TOTP
	admin.GoogleKey = ""
	admin.GoogleStatus = string(model.GoogleStatusUnbind)

	return b.ds.Admin().Update(ctx, admin, "google_key", "google_status")
}
