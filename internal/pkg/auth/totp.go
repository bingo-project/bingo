// ABOUTME: TOTP (Time-based One-Time Password) utilities for Google Authenticator.
// ABOUTME: Provides functions to generate secrets, verify codes, and create otpauth URLs.

package auth

import (
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

// TOTPOptions configures TOTP generation.
type TOTPOptions struct {
	Issuer      string
	AccountName string
}

// GenerateTOTPSecret creates a new TOTP secret and returns the key.
func GenerateTOTPSecret(opts TOTPOptions) (*otp.Key, error) {
	return totp.Generate(totp.GenerateOpts{
		Issuer:      opts.Issuer,
		AccountName: opts.AccountName,
	})
}

// ValidateTOTP verifies a TOTP code against the secret.
func ValidateTOTP(code string, secret string) bool {
	return totp.Validate(code, secret)
}
