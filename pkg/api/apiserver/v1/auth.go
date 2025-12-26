package v1

import "time"

// RegisterRequest 注册请求
type RegisterRequest struct {
	Account  string `json:"account" binding:"required,min=5,max=255"`
	Password string `json:"password" binding:"required,min=6,max=18"`
	Code     string `json:"code"`                                       // 验证码（验证开启时必填）
	Nickname string `json:"nickname" binding:"omitempty,min=2,max=255"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Account  string `json:"account" binding:"required,min=5,max=255"`
	Password string `json:"password" binding:"required,min=6,max=18"`
}

// SendCodeRequest 发送验证码请求
type SendCodeRequest struct {
	Account string `json:"account" binding:"required,min=5,max=255"`
	Scene   string `json:"scene" binding:"required,oneof=register reset_password bind"`
}

// ResetPasswordRequest 重置密码请求
type ResetPasswordRequest struct {
	Account  string `json:"account" binding:"required,min=5,max=255"`
	Code     string `json:"code" binding:"required,len=6"`
	Password string `json:"password" binding:"required,min=6,max=18"`
}

// UpdateProfileRequest 更新用户资料请求（用户自助更新）
type UpdateProfileRequest struct {
	Email    *string `json:"email" binding:"omitempty,email"`
	Phone    *string `json:"phone" binding:"omitempty,min=11,max=11"`
	Code     string  `json:"code"`
	Nickname *string `json:"nickname" binding:"omitempty,min=2,max=255"`
}

// BindingInfo 社交账号绑定信息
type BindingInfo struct {
	Provider  string `json:"provider"`
	AccountID string `json:"accountId"`
	Username  string `json:"username"`
	Avatar    string `json:"avatar"`
	BindTime  string `json:"bindTime"`
}

// ListBindingsResponse 社交账号列表响应
type ListBindingsResponse struct {
	Data []BindingInfo `json:"data"`
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
