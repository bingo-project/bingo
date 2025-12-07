// ABOUTME: Password encryption and comparison utilities.
// ABOUTME: Provides bcrypt-based password hashing functions.

package auth

import (
	"golang.org/x/crypto/bcrypt"
)

// Encrypt hashes a password using bcrypt.
func Encrypt(source string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(source), bcrypt.DefaultCost)

	return string(hashedBytes), err
}

// Compare verifies a password against a bcrypt hash.
func Compare(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
