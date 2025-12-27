// ABOUTME: Apple Sign In specific utilities.
// ABOUTME: Generates JWT client_secret required for Apple OAuth token exchange.

package auth

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// AppleConfig holds Apple Sign In configuration from provider.Info.
type AppleConfig struct {
	TeamID     string `json:"team_id"`
	KeyID      string `json:"key_id"`
	PrivateKey string `json:"private_key"`
}

// GenerateAppleClientSecret generates a JWT client_secret for Apple OAuth.
// The JWT is valid for 6 months (Apple's maximum).
func GenerateAppleClientSecret(clientID string, config AppleConfig) (string, error) {
	block, _ := pem.Decode([]byte(config.PrivateKey))
	if block == nil {
		return "", fmt.Errorf("failed to parse private key PEM")
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("failed to parse private key: %w", err)
	}

	ecdsaKey, ok := key.(*ecdsa.PrivateKey)
	if !ok {
		return "", fmt.Errorf("private key is not ECDSA")
	}

	now := time.Now()
	claims := jwt.MapClaims{
		"iss": config.TeamID,
		"iat": now.Unix(),
		"exp": now.Add(time.Hour * 24 * 180).Unix(),
		"aud": "https://appleid.apple.com",
		"sub": clientID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	token.Header["kid"] = config.KeyID

	return token.SignedString(ecdsaKey)
}

// ParseAppleConfig parses AppleConfig from provider.Info JSON string.
func ParseAppleConfig(info string) (AppleConfig, error) {
	var config AppleConfig
	if err := json.Unmarshal([]byte(info), &config); err != nil {
		return config, err
	}
	return config, nil
}
