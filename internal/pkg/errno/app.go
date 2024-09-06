package errno

var (
	ErrAppAlreadyExist = &Errno{HTTP: 404, Code: "FailedOperation.AppAlreadyExist", Message: "App already exist."}
	ErrAppNotFound     = &Errno{HTTP: 404, Code: "ResourceNotFound.AppNotFound", Message: "App was not found."}
)
