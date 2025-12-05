// ABOUTME: Unified authenticator for all protocols.
// ABOUTME: Provides token verification for HTTP, gRPC, and WebSocket.

package auth

import (
	"context"
	"strings"

	"github.com/bingo-project/component-base/web/token"

	"bingo/internal/pkg/contextx"
	"bingo/pkg/errorsx"
)

// Authenticator provides unified authentication across protocols.
type Authenticator struct{}

// New creates a new Authenticator.
func New() *Authenticator {
	return &Authenticator{}
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
	ctx = contextx.WithUsername(ctx, payload.Subject)

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
