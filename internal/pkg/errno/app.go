package errno

import (
	"net/http"

	"bingo/pkg/errorsx"
)

var (
	ErrAppNotFound = &errorsx.ErrorX{Code: http.StatusNotFound, Reason: "NotFound.AppNotFound", Message: "App was not found."}
)
