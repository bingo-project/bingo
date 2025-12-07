package errno

import (
	"net/http"

	"bingo/pkg/errorsx"
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
)
