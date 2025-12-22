// ABOUTME: Custom validators for Gin binding.
// ABOUTME: Registers additional validation rules like alphanumdash.

package validation

import (
	"regexp"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

var alphanumDashRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// Init registers custom validators with Gin's binding validator.
func Init() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		_ = v.RegisterValidation("alphanumdash", validateAlphanumDash)
	}
}

// validateAlphanumDash validates that a string contains only alphanumeric characters, underscores, and dashes.
func validateAlphanumDash(fl validator.FieldLevel) bool {
	return alphanumDashRegex.MatchString(fl.Field().String())
}
