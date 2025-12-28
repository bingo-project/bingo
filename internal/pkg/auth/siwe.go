// ABOUTME: SIWE (Sign-In with Ethereum) domain validation utilities.
// ABOUTME: Provides helpers for validating origin and domain against whitelist.

package auth

import (
	"fmt"
	"net/url"
	"strings"
)

// ValidateOriginAndExtractDomain validates origin against whitelist and extracts domain.
func ValidateOriginAndExtractDomain(origin string, allowedDomains []string) (string, error) {
	if origin == "" {
		return "", fmt.Errorf("empty origin")
	}

	parsed, err := url.Parse(origin)
	if err != nil {
		return "", err
	}

	domain := parsed.Host
	for _, allowed := range allowedDomains {
		if strings.EqualFold(domain, allowed) {
			return domain, nil
		}
	}

	return "", fmt.Errorf("domain not in whitelist")
}

// IsDomainAllowed checks if domain is in the allowed list.
func IsDomainAllowed(domain string, allowedDomains []string) bool {
	for _, allowed := range allowedDomains {
		if strings.EqualFold(domain, allowed) {
			return true
		}
	}
	return false
}
