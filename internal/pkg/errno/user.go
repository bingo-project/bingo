package errno

import (
	"net/http"

	"github.com/bingo-project/bingo/pkg/errorsx"
)

var (
	// ErrUsernameInvalid 表示用户名不合法.
	ErrUsernameInvalid = &errorsx.ErrorX{
		Code:    http.StatusBadRequest,
		Reason:  "InvalidArgument.UsernameInvalid",
		Message: "Invalid username: Username must consist of letters, digits, and underscores only, and its length must be between 3 and 20 characters.",
	}

	ErrUserAlreadyExist        = &errorsx.ErrorX{Code: http.StatusBadRequest, Reason: "InvalidArgument.UserAlreadyExist", Message: "User already exist."}
	ErrUserAccountAlreadyExist = &errorsx.ErrorX{Code: http.StatusBadRequest, Reason: "InvalidArgument.UserAlreadyExist", Message: "User account already exist."}

	// ErrPasswordInvalid 表示密码不合法.
	ErrPasswordInvalid = &errorsx.ErrorX{
		Code:    http.StatusBadRequest,
		Reason:  "InvalidArgument.PasswordInvalid",
		Message: "Password is incorrect.",
	}

	ErrPasswordOldInvalid = &errorsx.ErrorX{
		Code:    http.StatusBadRequest,
		Reason:  "InvalidArgument.PasswordInvalid",
		Message: "Old password is incorrect.",
	}

	ErrUserNotFound = &errorsx.ErrorX{Code: http.StatusNotFound, Reason: "NotFound.UserNotFound", Message: "User was not found."}

	// ErrInvalidAccountFormat 账号格式错误
	ErrInvalidAccountFormat = &errorsx.ErrorX{
		Code:    http.StatusBadRequest,
		Reason:  "InvalidArgument.InvalidAccountFormat",
		Message: "Invalid account format, please enter email or phone number.",
	}

	// ErrInvalidCode 验证码错误
	ErrInvalidCode = &errorsx.ErrorX{
		Code:    http.StatusBadRequest,
		Reason:  "InvalidArgument.InvalidCode",
		Message: "Verification code is invalid or expired.",
	}

	// ErrAuthTypeNotAllowed 注册方式未开放
	ErrAuthTypeNotAllowed = &errorsx.ErrorX{
		Code:    http.StatusBadRequest,
		Reason:  "InvalidArgument.AuthTypeNotAllowed",
		Message: "This registration method is not allowed.",
	}

	// ErrAlreadyBound 已绑定该类型账号
	ErrAlreadyBound = &errorsx.ErrorX{
		Code:    http.StatusBadRequest,
		Reason:  "InvalidArgument.AlreadyBound",
		Message: "Already bound to this account type.",
	}

	// ErrAccountOccupied 账号已被占用
	ErrAccountOccupied = &errorsx.ErrorX{
		Code:    http.StatusBadRequest,
		Reason:  "InvalidArgument.AccountOccupied",
		Message: "This account is already in use by another user.",
	}

	// ErrSMSNotConfigured 短信服务未配置
	ErrSMSNotConfigured = &errorsx.ErrorX{
		Code:    http.StatusServiceUnavailable,
		Reason:  "InternalError.SMSNotConfigured",
		Message: "SMS service is not configured.",
	}

	// ErrNotBound 未绑定该账号
	ErrNotBound = &errorsx.ErrorX{
		Code:    http.StatusBadRequest,
		Reason:  "InvalidArgument.NotBound",
		Message: "This provider is not bound to your account.",
	}

	// ErrCannotUnbindLastLogin 不能解绑唯一登录方式
	ErrCannotUnbindLastLogin = &errorsx.ErrorX{
		Code:    http.StatusBadRequest,
		Reason:  "InvalidArgument.CannotUnbindLastLogin",
		Message: "Cannot unbind the only login method.",
	}

	// ErrPayPasswordInvalid 支付密码错误
	ErrPayPasswordInvalid = &errorsx.ErrorX{
		Code:    http.StatusBadRequest,
		Reason:  "InvalidArgument.PayPasswordInvalid",
		Message: "Pay password is incorrect.",
	}

	// ErrPayPasswordNotSet 未设置支付密码
	ErrPayPasswordNotSet = &errorsx.ErrorX{
		Code:    http.StatusBadRequest,
		Reason:  "InvalidArgument.PayPasswordNotSet",
		Message: "Pay password is not set.",
	}

	// ErrTOTPNotEnabled TOTP未启用
	ErrTOTPNotEnabled = &errorsx.ErrorX{
		Code:    http.StatusBadRequest,
		Reason:  "InvalidArgument.TOTPNotEnabled",
		Message: "TOTP is not enabled.",
	}

	// ErrTOTPAlreadyEnabled TOTP已启用
	ErrTOTPAlreadyEnabled = &errorsx.ErrorX{
		Code:    http.StatusBadRequest,
		Reason:  "InvalidArgument.TOTPAlreadyEnabled",
		Message: "TOTP is already enabled.",
	}

	// ErrTOTPInvalid TOTP验证码错误
	ErrTOTPInvalid = &errorsx.ErrorX{
		Code:    http.StatusBadRequest,
		Reason:  "InvalidArgument.TOTPInvalid",
		Message: "TOTP code is invalid.",
	}

	// ErrTOTPCodeRequired TOTP验证码必填
	ErrTOTPCodeRequired = &errorsx.ErrorX{
		Code:    http.StatusBadRequest,
		Reason:  "InvalidArgument.TOTPCodeRequired",
		Message: "TOTP code is required.",
	}

	// ErrTOTPRequired 该角色要求启用 TOTP
	ErrTOTPRequired = &errorsx.ErrorX{
		Code:    http.StatusBadRequest,
		Reason:  "InvalidArgument.TOTPRequired",
		Message: "This role requires TOTP to be enabled.",
	}

	// ErrTOTPTokenInvalid TOTP Token 无效或过期
	ErrTOTPTokenInvalid = &errorsx.ErrorX{
		Code:    http.StatusUnauthorized,
		Reason:  "Unauthenticated.TOTPTokenInvalid",
		Message: "TOTP token is invalid or expired.",
	}

	// ErrPasswordRequired 登录密码必填
	ErrPasswordRequired = &errorsx.ErrorX{
		Code:    http.StatusBadRequest,
		Reason:  "InvalidArgument.PasswordRequired",
		Message: "Login password is required.",
	}

	// ErrInvalidState OAuth state无效或过期
	ErrInvalidState = &errorsx.ErrorX{
		Code:    http.StatusUnauthorized,
		Reason:  "Unauthenticated.InvalidState",
		Message: "Invalid or expired OAuth state.",
	}

	// ErrOAuthCodeInvalid OAuth授权码无效或已过期
	ErrOAuthCodeInvalid = &errorsx.ErrorX{
		Code:    http.StatusBadRequest,
		Reason:  "InvalidArgument.OAuthCodeInvalid",
		Message: "Authorization code is invalid or expired.",
	}

	// ErrOAuthProviderError OAuth服务商返回错误
	ErrOAuthProviderError = &errorsx.ErrorX{
		Code:    http.StatusBadGateway,
		Reason:  "ExternalError.OAuthProviderError",
		Message: "OAuth provider returned an error.",
	}
)
