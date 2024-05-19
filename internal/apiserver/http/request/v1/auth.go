package v1

import "time"

type RegisterRequest struct {
	Nickname string `json:"nickname" binding:"alphanum,min=2,max=255" example:"Peter"`
	Username string `json:"username" binding:"required,alphanum,min=2,max=255" example:"peter"`
	Password string `json:"password" binding:"required,min=6,max=18" example:"123123"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required,alphanum,min=2,max=255" example:"peter"`
	Password string `json:"password" binding:"required,min=6,max=18" example:"123123"`
}

type LoginByProviderRequest struct {
	Code string `json:"code" form:"code"` // Auth code
}

type LoginResponse struct {
	AccessToken string    `json:"accessToken"`
	ExpiresAt   time.Time `json:"expiresAt"`
}

type AddressRequest struct {
	Address string `json:"address" form:"address" binding:"required,eth_addr"` // ETH Address
}

type NonceResponse struct {
	Nonce string `json:"nonce"` // Nonce
}

type LoginByAddressRequest struct {
	AddressRequest
	Sign string `json:"sign" form:"sign" binding:"required"` // Signature
}

type ChangePasswordRequest struct {
	PasswordOld string `json:"passwordOld" binding:"required,min=6,max=18"`
	PasswordNew string `json:"passwordNew" binding:"required,min=6,max=18"`
}
