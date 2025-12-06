package auth

import (
	"golang.org/x/crypto/bcrypt"
)

var (
	XRequestIDKey = "x-request-id"
	XForwardedKey = "x-forwarded-for"
)

// Encrypt string by bcrypt.
func Encrypt(source string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(source), bcrypt.DefaultCost)

	return string(hashedBytes), err
}

// Compare password.
func Compare(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
