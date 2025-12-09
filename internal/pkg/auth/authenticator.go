// ABOUTME: Unified authenticator for all protocols.
// ABOUTME: Provides token verification for HTTP, gRPC, and WebSocket.

package auth

import (
	"context"
	"strings"

	"github.com/bingo-project/component-base/web/token"

	"github.com/bingo-project/bingo/pkg/contextx"
	"github.com/bingo-project/bingo/pkg/errorsx"
)

// UserLoader loads user information into context.
type UserLoader interface {
	LoadUser(ctx context.Context, userID string) (context.Context, error)
}

// Authenticator provides unified authentication across protocols.
type Authenticator struct {
	loader UserLoader
}

// New creates a new Authenticator with optional UserLoader.
func New(loader UserLoader) *Authenticator {
	return &Authenticator{loader: loader}
}

// Verify validates a token and returns a context with user info.
func (a *Authenticator) Verify(ctx context.Context, tokenStr string) (context.Context, error) {
	if tokenStr == "" {
		return ctx, errorsx.New(401, "Unauthenticated", "token is required")
	}

	payload, err := token.Parse(tokenStr)
	if err != nil {
		return ctx, errorsx.New(401, "Unauthenticated", "invalid token: %s", err.Error())
	}

	ctx = contextx.WithUserID(ctx, payload.Subject)

	if a.loader != nil {
		ctx, err = a.loader.LoadUser(ctx, payload.Subject)
		if err != nil {
			return ctx, err
		}
	}

	return ctx, nil
}

// ExtractBearerToken extracts token from "Bearer <token>" format.
func ExtractBearerToken(auth string) string {
	const prefix = "Bearer "
	if len(auth) > len(prefix) && strings.EqualFold(auth[:len(prefix)], prefix) {
		return auth[len(prefix):]
	}

	return ""
}
