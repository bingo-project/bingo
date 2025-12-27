// ABOUTME: Security settings business logic for pay password and TOTP.
// ABOUTME: Provides verification services for sensitive operations.

package auth

import (
	"context"

	"github.com/bingo-project/bingo/internal/pkg/auth"
	"github.com/bingo-project/bingo/internal/pkg/errno"
	"github.com/bingo-project/bingo/internal/pkg/facade"
	"github.com/bingo-project/bingo/internal/pkg/store"
	v1 "github.com/bingo-project/bingo/pkg/api/apiserver/v1"
)

const (
	// GoogleStatusUnbind TOTP未绑定
	GoogleStatusUnbind = "unbind"
	// GoogleStatusDisabled TOTP已禁用
	GoogleStatusDisabled = "disabled"
	// GoogleStatusEnabled TOTP已启用
	GoogleStatusEnabled = "enabled"

	// TOTPIssuer TOTP发行者名称
	TOTPIssuer = "Bingo"
)

// SecurityBiz defines security settings operations.
type SecurityBiz interface {
	// Pay password
	SetPayPassword(ctx context.Context, uid string, req *v1.SetPayPasswordRequest) error
	VerifyPayPassword(ctx context.Context, uid string, password string) error
	HasPayPassword(ctx context.Context, uid string) (bool, error)

	// TOTP
	GetTOTPStatus(ctx context.Context, uid string) (*v1.TOTPStatusResponse, error)
	SetupTOTP(ctx context.Context, uid string, email string) (*v1.TOTPSetupResponse, error)
	EnableTOTP(ctx context.Context, uid string, code string) error
	VerifyTOTP(ctx context.Context, uid string, code string) error
	DisableTOTP(ctx context.Context, uid string, verifyCode, totpCode string) error

	// Combined status
	GetSecurityStatus(ctx context.Context, uid string) (*v1.SecurityStatusResponse, error)
}

type securityBiz struct {
	ds      store.IStore
	codeBiz CodeBiz
}

var _ SecurityBiz = (*securityBiz)(nil)

// NewSecurityBiz creates a new SecurityBiz instance.
func NewSecurityBiz(ds store.IStore, codeBiz CodeBiz) SecurityBiz {
	return &securityBiz{ds: ds, codeBiz: codeBiz}
}

// SetPayPassword sets or updates pay password.
// Requires login password and verification code.
func (b *securityBiz) SetPayPassword(ctx context.Context, uid string, req *v1.SetPayPasswordRequest) error {
	user, err := b.ds.User().GetByUID(ctx, uid)
	if err != nil {
		return errno.ErrUserNotFound
	}

	// Verify login password
	if err := auth.Compare(user.Password, req.LoginPassword); err != nil {
		return errno.ErrPasswordInvalid
	}

	// Get user account for verification code
	account := user.Email
	if account == "" {
		account = user.Phone
	}
	if account == "" {
		return errno.ErrInvalidAccountFormat
	}

	// Verify code
	if err := b.codeBiz.Verify(ctx, account, CodeSceneSecurity, req.Code); err != nil {
		return err
	}

	// Encrypt and save pay password
	encrypted, err := auth.Encrypt(req.PayPassword)
	if err != nil {
		return err
	}

	user.PayPassword = encrypted

	return b.ds.User().Update(ctx, user, "pay_password")
}

// VerifyPayPassword verifies the pay password.
func (b *securityBiz) VerifyPayPassword(ctx context.Context, uid string, password string) error {
	user, err := b.ds.User().GetByUID(ctx, uid)
	if err != nil {
		return errno.ErrUserNotFound
	}

	if user.PayPassword == "" {
		return errno.ErrPayPasswordNotSet
	}

	if err := auth.Compare(user.PayPassword, password); err != nil {
		return errno.ErrPayPasswordInvalid
	}

	return nil
}

// HasPayPassword checks if user has set pay password.
func (b *securityBiz) HasPayPassword(ctx context.Context, uid string) (bool, error) {
	user, err := b.ds.User().GetByUID(ctx, uid)
	if err != nil {
		return false, errno.ErrUserNotFound
	}

	return user.PayPassword != "", nil
}

// GetTOTPStatus returns TOTP enabled status.
func (b *securityBiz) GetTOTPStatus(ctx context.Context, uid string) (*v1.TOTPStatusResponse, error) {
	user, err := b.ds.User().GetByUID(ctx, uid)
	if err != nil {
		return nil, errno.ErrUserNotFound
	}

	return &v1.TOTPStatusResponse{
		Enabled: user.GoogleStatus == GoogleStatusEnabled,
	}, nil
}

// SetupTOTP generates a new TOTP secret for binding.
func (b *securityBiz) SetupTOTP(ctx context.Context, uid string, email string) (*v1.TOTPSetupResponse, error) {
	user, err := b.ds.User().GetByUID(ctx, uid)
	if err != nil {
		return nil, errno.ErrUserNotFound
	}

	if user.GoogleStatus == GoogleStatusEnabled {
		return nil, errno.ErrTOTPAlreadyEnabled
	}

	// Generate TOTP secret
	accountName := email
	if accountName == "" {
		accountName = uid
	}

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

	user.GoogleKey = encryptedSecret
	user.GoogleStatus = GoogleStatusDisabled

	if err := b.ds.User().Update(ctx, user, "google_key", "google_status"); err != nil {
		return nil, err
	}

	return &v1.TOTPSetupResponse{
		Secret:     key.Secret(),
		OtpauthURL: key.URL(),
	}, nil
}

// EnableTOTP enables TOTP after verifying the code.
func (b *securityBiz) EnableTOTP(ctx context.Context, uid string, code string) error {
	user, err := b.ds.User().GetByUID(ctx, uid)
	if err != nil {
		return errno.ErrUserNotFound
	}

	if user.GoogleStatus == GoogleStatusEnabled {
		return errno.ErrTOTPAlreadyEnabled
	}

	if user.GoogleKey == "" {
		return errno.ErrTOTPNotEnabled
	}

	// Decrypt secret
	secret, err := facade.AES.DecryptString(user.GoogleKey)
	if err != nil {
		return err
	}

	// Verify TOTP code
	if !auth.ValidateTOTP(code, secret) {
		return errno.ErrTOTPInvalid
	}

	// Enable TOTP
	user.GoogleStatus = GoogleStatusEnabled

	return b.ds.User().Update(ctx, user, "google_status")
}

// VerifyTOTP verifies a TOTP code.
func (b *securityBiz) VerifyTOTP(ctx context.Context, uid string, code string) error {
	user, err := b.ds.User().GetByUID(ctx, uid)
	if err != nil {
		return errno.ErrUserNotFound
	}

	if user.GoogleStatus != GoogleStatusEnabled {
		return errno.ErrTOTPNotEnabled
	}

	// Decrypt secret
	secret, err := facade.AES.DecryptString(user.GoogleKey)
	if err != nil {
		return err
	}

	if !auth.ValidateTOTP(code, secret) {
		return errno.ErrTOTPInvalid
	}

	return nil
}

// DisableTOTP disables TOTP after verifying email code and current TOTP.
func (b *securityBiz) DisableTOTP(ctx context.Context, uid string, verifyCode, totpCode string) error {
	user, err := b.ds.User().GetByUID(ctx, uid)
	if err != nil {
		return errno.ErrUserNotFound
	}

	if user.GoogleStatus != GoogleStatusEnabled {
		return errno.ErrTOTPNotEnabled
	}

	// Get user account for verification code
	account := user.Email
	if account == "" {
		account = user.Phone
	}
	if account == "" {
		return errno.ErrInvalidAccountFormat
	}

	// Verify email/phone code
	if err := b.codeBiz.Verify(ctx, account, CodeSceneSecurity, verifyCode); err != nil {
		return err
	}

	// Decrypt and verify TOTP
	secret, err := facade.AES.DecryptString(user.GoogleKey)
	if err != nil {
		return err
	}

	if !auth.ValidateTOTP(totpCode, secret) {
		return errno.ErrTOTPInvalid
	}

	// Disable TOTP
	user.GoogleKey = ""
	user.GoogleStatus = GoogleStatusUnbind

	return b.ds.User().Update(ctx, user, "google_key", "google_status")
}

// GetSecurityStatus returns combined security settings status.
func (b *securityBiz) GetSecurityStatus(ctx context.Context, uid string) (*v1.SecurityStatusResponse, error) {
	user, err := b.ds.User().GetByUID(ctx, uid)
	if err != nil {
		return nil, errno.ErrUserNotFound
	}

	return &v1.SecurityStatusResponse{
		PayPasswordSet: user.PayPassword != "",
		TOTPEnabled:    user.GoogleStatus == GoogleStatusEnabled,
	}, nil
}
