// ABOUTME: Tests for OAuth PKCE and state utilities.
// ABOUTME: Verifies code verifier/challenge generation correctness.

package auth

import (
	"testing"
)

func TestGenerateCodeVerifier(t *testing.T) {
	verifier, err := GenerateCodeVerifier()
	if err != nil {
		t.Fatalf("GenerateCodeVerifier failed: %v", err)
	}
	if len(verifier) < 43 {
		t.Errorf("code verifier too short: got %d chars", len(verifier))
	}
}

func TestGenerateCodeChallenge(t *testing.T) {
	verifier := "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	challenge := GenerateCodeChallenge(verifier)
	if challenge == "" {
		t.Error("code challenge should not be empty")
	}
	if challenge == verifier {
		t.Error("code challenge should differ from verifier")
	}
}

func TestGenerateState(t *testing.T) {
	state1, err := GenerateState()
	if err != nil {
		t.Fatalf("GenerateState failed: %v", err)
	}
	state2, _ := GenerateState()
	if state1 == state2 {
		t.Error("states should be unique")
	}
}
