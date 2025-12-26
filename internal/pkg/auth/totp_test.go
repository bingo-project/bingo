// ABOUTME: Tests for TOTP utilities.
// ABOUTME: Verifies secret generation and code validation.

package auth

import (
	"testing"
	"time"

	"github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateTOTPSecret(t *testing.T) {
	key, err := GenerateTOTPSecret(TOTPOptions{
		Issuer:      "TestApp",
		AccountName: "user@example.com",
	})

	require.NoError(t, err)
	assert.NotEmpty(t, key.Secret())
	assert.Contains(t, key.URL(), "otpauth://totp/")
	assert.Contains(t, key.URL(), "TestApp")
	assert.Contains(t, key.URL(), "user@example.com")
}

func TestValidateTOTP(t *testing.T) {
	key, err := GenerateTOTPSecret(TOTPOptions{
		Issuer:      "TestApp",
		AccountName: "user@example.com",
	})
	require.NoError(t, err)

	// Generate a valid code
	code, err := totp.GenerateCode(key.Secret(), time.Now())
	require.NoError(t, err)

	// Valid code should pass
	assert.True(t, ValidateTOTP(code, key.Secret()))

	// Invalid code should fail
	assert.False(t, ValidateTOTP("000000", key.Secret()))
	assert.False(t, ValidateTOTP("invalid", key.Secret()))
}
