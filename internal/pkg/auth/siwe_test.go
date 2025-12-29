// ABOUTME: Tests for SIWE domain validation utilities.
// ABOUTME: Verifies origin/domain validation against whitelist.

package auth

import (
	"testing"
)

func TestValidateOriginAndExtractDomain(t *testing.T) {
	tests := []struct {
		name           string
		origin         string
		allowedDomains []string
		wantDomain     string
		wantErr        bool
	}{
		{
			name:           "valid origin with matching domain",
			origin:         "https://example.com",
			allowedDomains: []string{"example.com", "test.com"},
			wantDomain:     "example.com",
			wantErr:        false,
		},
		{
			name:           "valid origin with port",
			origin:         "http://localhost:3000",
			allowedDomains: []string{"localhost:3000"},
			wantDomain:     "localhost:3000",
			wantErr:        false,
		},
		{
			name:           "case insensitive match",
			origin:         "https://Example.COM",
			allowedDomains: []string{"example.com"},
			wantDomain:     "Example.COM",
			wantErr:        false,
		},
		{
			name:           "empty origin",
			origin:         "",
			allowedDomains: []string{"example.com"},
			wantDomain:     "",
			wantErr:        true,
		},
		{
			name:           "domain not in whitelist",
			origin:         "https://evil.com",
			allowedDomains: []string{"example.com", "test.com"},
			wantDomain:     "",
			wantErr:        true,
		},
		{
			name:           "empty whitelist",
			origin:         "https://example.com",
			allowedDomains: []string{},
			wantDomain:     "",
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDomain, err := ValidateOriginAndExtractDomain(tt.origin, tt.allowedDomains)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateOriginAndExtractDomain() error = %v, wantErr %v", err, tt.wantErr)

				return
			}
			if gotDomain != tt.wantDomain {
				t.Errorf("ValidateOriginAndExtractDomain() = %v, want %v", gotDomain, tt.wantDomain)
			}
		})
	}
}

func TestIsDomainAllowed(t *testing.T) {
	tests := []struct {
		name           string
		domain         string
		allowedDomains []string
		want           bool
	}{
		{
			name:           "domain in list",
			domain:         "example.com",
			allowedDomains: []string{"example.com", "test.com"},
			want:           true,
		},
		{
			name:           "domain not in list",
			domain:         "evil.com",
			allowedDomains: []string{"example.com", "test.com"},
			want:           false,
		},
		{
			name:           "case insensitive",
			domain:         "EXAMPLE.COM",
			allowedDomains: []string{"example.com"},
			want:           true,
		},
		{
			name:           "empty list",
			domain:         "example.com",
			allowedDomains: []string{},
			want:           false,
		},
		{
			name:           "empty domain",
			domain:         "",
			allowedDomains: []string{"example.com"},
			want:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsDomainAllowed(tt.domain, tt.allowedDomains); got != tt.want {
				t.Errorf("IsDomainAllowed() = %v, want %v", got, tt.want)
			}
		})
	}
}
