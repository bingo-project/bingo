package auth

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"golang.org/x/crypto/bcrypt"
)

var (
	XRequestIDKey = "X-Request-ID"
	XForwardedKey = "X-Forwarded-For"
	XGuard        = "X-Guard"
	XUsernameKey  = "X-Username"
	XUserInfoKey  = "X-UserInfo"
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

func ID(c *gin.Context) any {
	info, exists := c.Get(XUsernameKey)
	if !exists {
		return nil
	}

	return info
}

func User(c *gin.Context, user any) error {
	info, exists := c.Get(XUserInfoKey)
	if !exists {
		return errors.New("not exists")
	}

	_ = copier.Copy(user, info)

	return nil
}
