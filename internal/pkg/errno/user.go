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
)
