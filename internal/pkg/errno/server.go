package errno

import "net/http"

var (
	ErrAlreadyUnderMaintenance = &Errno{HTTP: http.StatusConflict, Code: "StatusConflict", Message: "Server already under maintenance."}
)
