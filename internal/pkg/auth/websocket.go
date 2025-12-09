// ABOUTME: WebSocket authentication helpers using unified authenticator.
// ABOUTME: Provides context authentication for WebSocket connections.

package auth

import (
	"context"
	"net/http"

	"github.com/bingo-project/bingo/pkg/contextx"
)

// AuthenticateWebSocket authenticates a WebSocket connection from HTTP request.
// It extracts token from query parameter or Authorization header.
// Returns authenticated context, or original context if no token provided.
func (a *Authenticator) AuthenticateWebSocket(ctx context.Context, r *http.Request) (context.Context, error) {
	// Try token from query parameter first (common for WebSocket)
	tokenStr := r.URL.Query().Get("token")
	if tokenStr == "" {
		tokenStr = ExtractBearerToken(r.Header.Get("Authorization"))
	}

	// If no token, return original context (anonymous connection)
	if tokenStr == "" {
		return ctx, nil
	}

	// Verify token
	return a.Verify(ctx, tokenStr)
}

// AuthenticateWebSocketMessage authenticates based on token in JSON-RPC request.
// This is used for late authentication after connection is established.
func (a *Authenticator) AuthenticateWebSocketMessage(ctx context.Context, tokenStr string) (context.Context, error) {
	if tokenStr == "" {
		return ctx, nil
	}

	return a.Verify(ctx, tokenStr)
}

// IsAuthenticated checks if the context contains valid user info.
func IsAuthenticated(ctx context.Context) bool {
	return contextx.UserID(ctx) != ""
}
