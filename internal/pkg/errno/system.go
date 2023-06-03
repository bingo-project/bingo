package errno

var (
	ErrAdminAlreadyExist = &Errno{HTTP: 400, Code: "FailedOperation.AdminAlreadyExist", Message: "Admin already exist."}
	ErrAdminNotFound     = &Errno{HTTP: 404, Code: "ResourceNotFound.AdminNotFound", Message: "Admin was not found."}
)
