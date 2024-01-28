package errno

var (
	ErrUserAlreadyExist = &Errno{HTTP: 404, Code: "FailedOperation.UserAlreadyExist", Message: "User already exist."} // 代表用户已经存在.
	ErrUserNotFound     = &Errno{HTTP: 404, Code: "ResourceNotFound.UserNotFound", Message: "User was not found."}    // 表示未找到用户.

	ErrPasswordIncorrect    = &Errno{HTTP: 401, Code: "InvalidParameter.PasswordIncorrect", Message: "Password was incorrect."}     // 表示密码不正确.
	ErrPasswordOldIncorrect = &Errno{HTTP: 400, Code: "InvalidParameter.PasswordIncorrect", Message: "Old password was incorrect."} // 旧密码不正确
)
