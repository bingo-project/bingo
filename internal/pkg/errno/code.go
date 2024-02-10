package errno

import "net/http"

var (
	OK = &Errno{HTTP: http.StatusOK, Code: "", Message: "OK"} // 代表请求成功

	ErrResourceAlreadyExists = &Errno{HTTP: 400, Code: "ResourceAlreadyExists", Message: "Resource already exists."}                                          // 表示资源已存在
	ErrBind                  = &Errno{HTTP: 400, Code: "InvalidParameter.BindError", Message: "Error occurred while binding the request body to the struct."} // 表示参数绑定错误
	ErrInvalidParameter      = &Errno{HTTP: 400, Code: "InvalidParameter", Message: "Parameter verification failed."}                                         // 表示所有验证失败的错误
	ErrResourceNotFound      = &Errno{HTTP: 404, Code: "ResourceNotFound", Message: "Resource not found."}                                                    // 表示资源不存在
	ErrTooManyRequests       = &Errno{HTTP: 429, Code: "TooManyRequests", Message: "Too Many Requests"}                                                       // 请求过于频繁

	ErrSignToken    = &Errno{HTTP: 401, Code: "AuthFailure.SignTokenError", Message: "Error occurred while signing the JSON web token."} // 表示签发 JWT Token 时出错.
	ErrTokenInvalid = &Errno{HTTP: 401, Code: "AuthFailure.TokenInvalid", Message: "Token was invalid."}                                 // 表示 JWT Token 格式错误.
	ErrUnauthorized = &Errno{HTTP: 401, Code: "AuthFailure.Unauthorized", Message: "Unauthorized."}                                      // 表示请求没有被授权.
	ErrForbidden    = &Errno{HTTP: 403, Code: "Forbidden", Message: "Forbidden."}                                                        // 表示请求没有被授权.

	InternalServerError        = &Errno{HTTP: 500, Code: "InternalError", Message: "Internal server error."}               // 表示所有未知的服务器端错误
	ErrServiceUnderMaintenance = &Errno{HTTP: 503, Code: "ServiceUnderMaintenance", Message: "Service under maintenance."} // 系统维护中
)
