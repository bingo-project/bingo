package errno

var (
	ErrAdminAlreadyExist = &Errno{HTTP: 400, Code: "FailedOperation.AdminAlreadyExist", Message: "Admin already exist."}
	ErrAdminNotFound     = &Errno{HTTP: 404, Code: "ResourceNotFound.AdminNotFound", Message: "Admin was not found."}

	ErrRoleAlreadyExist = &Errno{HTTP: 400, Code: "FailedOperation.RoleAlreadyExist", Message: "Role already exist."}
	ErrRoleNotFound     = &Errno{HTTP: 404, Code: "ResourceNotFound.RoleNotFound", Message: "Role was not found."}

	ErrPermissionAlreadyExist = &Errno{HTTP: 400, Code: "FailedOperation.PermissionAlreadyExist", Message: "Permission already exist."}
	ErrPermissionNotFound     = &Errno{HTTP: 404, Code: "ResourceNotFound.PermissionNotFound", Message: "Permission was not found."}
)
