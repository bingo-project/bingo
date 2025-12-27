// ABOUTME: Tests for Apple Sign In JWT generation.
// ABOUTME: Uses a test ECDSA key to verify JWT structure.

package auth

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"testing"

	"github.com/golang-jwt/jwt/v5"
)

func TestGenerateAppleClientSecret(t *testing.T) {
	privateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	pkcs8Bytes, _ := x509.MarshalPKCS8PrivateKey(privateKey)
	pemBlock := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: pkcs8Bytes,
	})

	config := AppleConfig{
		TeamID:     "TEAM123",
		KeyID:      "KEY456",
		PrivateKey: string(pemBlock),
	}

	secret, err := GenerateAppleClientSecret("com.example.app", config)
	if err != nil {
		t.Fatalf("GenerateAppleClientSecret failed: %v", err)
	}

	token, _, err := jwt.NewParser().ParseUnverified(secret, jwt.MapClaims{})
	if err != nil {
		t.Fatalf("Failed to parse JWT: %v", err)
	}

	claims := token.Claims.(jwt.MapClaims)
	if claims["iss"] != "TEAM123" {
		t.Errorf("Expected iss=TEAM123, got %v", claims["iss"])
	}
	if claims["sub"] != "com.example.app" {
		t.Errorf("Expected sub=com.example.app, got %v", claims["sub"])
	}
}

func TestParseAppleConfig(t *testing.T) {
	jsonStr := `{"team_id":"TEAM123","key_id":"KEY456","private_key":"-----BEGIN PRIVATE KEY-----\ntest\n-----END PRIVATE KEY-----"}`
	config, err := ParseAppleConfig(jsonStr)
	if err != nil {
		t.Fatalf("ParseAppleConfig failed: %v", err)
	}
	if config.TeamID != "TEAM123" {
		t.Errorf("Expected TeamID=TEAM123, got %s", config.TeamID)
	}
}
