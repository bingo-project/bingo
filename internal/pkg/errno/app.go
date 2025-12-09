package errno

import (
	"net/http"

	"github.com/bingo-project/bingo/pkg/errorsx"
)

var (
	ErrAppNotFound = &errorsx.ErrorX{Code: http.StatusNotFound, Reason: "NotFound.AppNotFound", Message: "App was not found."}
)
