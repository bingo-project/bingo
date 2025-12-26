// ABOUTME: Security settings API request/response types.
// ABOUTME: Covers pay password and TOTP operations.

package v1

// SetPayPasswordRequest sets or updates pay password.
type SetPayPasswordRequest struct {
	LoginPassword string `json:"loginPassword" binding:"required" example:"123456"`
	Code          string `json:"code" binding:"required" example:"123456"`
	PayPassword   string `json:"payPassword" binding:"required,min=6" example:"654321"`
}

// VerifyPayPasswordRequest verifies pay password.
type VerifyPayPasswordRequest struct {
	PayPassword string `json:"payPassword" binding:"required" example:"654321"`
}

// TOTPStatusResponse returns TOTP status.
type TOTPStatusResponse struct {
	Enabled bool `json:"enabled"`
}

// TOTPSetupResponse returns TOTP setup info.
type TOTPSetupResponse struct {
	Secret     string `json:"secret"`
	OtpauthURL string `json:"otpauthUrl"`
}

// TOTPEnableRequest enables TOTP with verification code.
type TOTPEnableRequest struct {
	Code string `json:"code" binding:"required,len=6" example:"123456"`
}

// TOTPVerifyRequest verifies TOTP code.
type TOTPVerifyRequest struct {
	Code string `json:"code" binding:"required,len=6" example:"123456"`
}

// TOTPDisableRequest disables TOTP.
type TOTPDisableRequest struct {
	VerifyCode string `json:"verifyCode" binding:"required" example:"123456"`
	TOTPCode   string `json:"totpCode" binding:"required,len=6" example:"123456"`
}

// SecurityStatusResponse returns user's security settings status.
type SecurityStatusResponse struct {
	PayPasswordSet bool `json:"payPasswordSet"`
	TOTPEnabled    bool `json:"totpEnabled"`
}
