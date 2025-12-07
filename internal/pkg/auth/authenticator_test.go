// ABOUTME: Tests for unified authenticator.
// ABOUTME: Verifies token extraction and authentication logic.

package auth

import (
	"context"
	"testing"
)

func TestExtractBearerToken(t *testing.T) {
	tests := []struct {
		name     string
		auth     string
		expected string
	}{
		{
			name:     "valid bearer token",
			auth:     "Bearer abc123",
			expected: "abc123",
		},
		{
			name:     "valid bearer token case insensitive",
			auth:     "bearer abc123",
			expected: "abc123",
		},
		{
			name:     "valid BEARER token uppercase",
			auth:     "BEARER abc123",
			expected: "abc123",
		},
		{
			name:     "empty string",
			auth:     "",
			expected: "",
		},
		{
			name:     "no bearer prefix",
			auth:     "abc123",
			expected: "",
		},
		{
			name:     "only bearer prefix",
			auth:     "Bearer ",
			expected: "",
		},
		{
			name:     "bearer without space",
			auth:     "Bearerabc123",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractBearerToken(tt.auth)
			if result != tt.expected {
				t.Errorf("ExtractBearerToken(%q) = %q, want %q", tt.auth, result, tt.expected)
			}
		})
	}
}

func TestAuthenticator_Verify_EmptyToken(t *testing.T) {
	a := New(nil)
	ctx := context.Background()

	_, err := a.Verify(ctx, "")
	if err == nil {
		t.Error("Verify with empty token should return error")
	}
}

func TestIsAuthenticated(t *testing.T) {
	tests := []struct {
		name     string
		ctx      context.Context
		expected bool
	}{
		{
			name:     "empty context",
			ctx:      context.Background(),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsAuthenticated(tt.ctx)
			if result != tt.expected {
				t.Errorf("IsAuthenticated() = %v, want %v", result, tt.expected)
			}
		})
	}
}
