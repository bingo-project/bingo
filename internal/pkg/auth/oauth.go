// ABOUTME: OAuth security utilities for PKCE and state validation.
// ABOUTME: Provides code verifier/challenge generation and state management via Redis.

package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	stateKeyPrefix = "oauth:state:"
	stateTTL       = 5 * time.Minute
)

// GenerateCodeVerifier generates a random code verifier for PKCE (43-128 chars).
func GenerateCodeVerifier() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// GenerateCodeChallenge computes S256 code challenge from verifier.
func GenerateCodeChallenge(verifier string) string {
	h := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h[:])
}

// GenerateState generates a random state string.
func GenerateState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// SaveState stores state in Redis with TTL.
func SaveState(ctx context.Context, rdb *redis.Client, state string) error {
	key := stateKeyPrefix + state
	return rdb.Set(ctx, key, "1", stateTTL).Err()
}

// ValidateAndDeleteState validates state exists and deletes it (one-time use).
func ValidateAndDeleteState(ctx context.Context, rdb *redis.Client, state string) error {
	key := stateKeyPrefix + state
	result, err := rdb.Del(ctx, key).Result()
	if err != nil {
		return err
	}
	if result == 0 {
		return fmt.Errorf("invalid or expired state")
	}
	return nil
}
