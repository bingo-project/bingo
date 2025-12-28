package errno

import (
	"net/http"

	"github.com/bingo-project/bingo/pkg/errorsx"
)

var (
	// OK 代表请求成功.
	OK                  = &errorsx.ErrorX{Code: http.StatusOK, Message: "ok"}
	ErrInternal         = errorsx.ErrInternal
	ErrNotFound         = errorsx.ErrNotFound
	ErrBind             = errorsx.ErrBind
	ErrInvalidArgument  = errorsx.ErrInvalidArgument
	ErrUnauthenticated  = errorsx.ErrUnauthenticated
	ErrPermissionDenied = errorsx.ErrPermissionDenied
	ErrOperationFailed  = errorsx.ErrOperationFailed

	ErrIllegalRequest        = &errorsx.ErrorX{Code: http.StatusBadRequest, Reason: "InvalidArgument.IllegalRequest", Message: "Illegal request."}                // 表示非法请求
	ErrResourceAlreadyExists = &errorsx.ErrorX{Code: http.StatusBadRequest, Reason: "InvalidArgument.ResourceAlreadyExists", Message: "Resource already exists."} // 表示资源已存在
	ErrPageNotFound          = &errorsx.ErrorX{Code: http.StatusNotFound, Reason: "NotFound.PageNotFound", Message: "Page not found."}
	ErrSignToken             = &errorsx.ErrorX{Code: http.StatusUnauthorized, Reason: "Unauthenticated.SignToken", Message: "Error occurred while signing the JSON web token."} // 表示签发 JWT Token 时出错.
	ErrTokenInvalid          = &errorsx.ErrorX{Code: http.StatusUnauthorized, Reason: "Unauthenticated.TokenInvalid", Message: "Token was invalid."}                            // 表示 JWT Token 格式错误.
	ErrTooManyRequests       = &errorsx.ErrorX{Code: http.StatusTooManyRequests, Reason: "TooManyRequests", Message: "Too Many Requests"}                                       // 请求过于频繁

	ErrDBRead                  = &errorsx.ErrorX{Code: http.StatusInternalServerError, Reason: "InternalError.DBRead", Message: "Database read failure."}
	ErrDBWrite                 = &errorsx.ErrorX{Code: http.StatusInternalServerError, Reason: "InternalError.DBWrite", Message: "Database write failure."}
	ErrAddRole                 = &errorsx.ErrorX{Code: http.StatusInternalServerError, Reason: "InternalError.AddRole", Message: "Error occurred while adding the role."}
	ErrRemoveRole              = &errorsx.ErrorX{Code: http.StatusInternalServerError, Reason: "InternalError.RemoveRole", Message: "Error occurred while removing the role."}
	ErrServiceUnderMaintenance = &errorsx.ErrorX{Code: http.StatusServiceUnavailable, Reason: "InternalError.ServiceUnderMaintenance", Message: "Server under maintenance."}

	// SIWE wallet login errors
	ErrInvalidOrigin      = &errorsx.ErrorX{Code: http.StatusBadRequest, Reason: "InvalidArgument.InvalidOrigin", Message: "Invalid request origin."}
	ErrInvalidDomain      = &errorsx.ErrorX{Code: http.StatusBadRequest, Reason: "InvalidArgument.InvalidDomain", Message: "Domain not allowed."}
	ErrInvalidSIWEMessage = &errorsx.ErrorX{Code: http.StatusBadRequest, Reason: "InvalidArgument.InvalidSIWEMessage", Message: "Invalid SIWE message format."}
	ErrNonceExpired       = &errorsx.ErrorX{Code: http.StatusBadRequest, Reason: "InvalidArgument.NonceExpired", Message: "Nonce has expired."}
	ErrInvalidNonce       = &errorsx.ErrorX{Code: http.StatusBadRequest, Reason: "InvalidArgument.InvalidNonce", Message: "Invalid or already used nonce."}
	ErrSignatureInvalid   = &errorsx.ErrorX{Code: http.StatusUnauthorized, Reason: "Unauthenticated.SignatureInvalid", Message: "Signature verification failed."}
	ErrWalletAlreadyBound = &errorsx.ErrorX{Code: http.StatusBadRequest, Reason: "InvalidArgument.WalletAlreadyBound", Message: "Wallet already bound to this account."}
	ErrWalletBoundToOther = &errorsx.ErrorX{Code: http.StatusBadRequest, Reason: "InvalidArgument.WalletBoundToOther", Message: "Wallet address already bound to another account."}
)
