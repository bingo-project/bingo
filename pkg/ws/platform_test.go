// ABOUTME: Tests for platform validation.
// ABOUTME: Validates platform constants and IsValidPlatform function.

package ws

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidPlatform(t *testing.T) {
	tests := []struct {
		platform string
		valid    bool
	}{
		{PlatformWeb, true},
		{PlatformIOS, true},
		{PlatformAndroid, true},
		{PlatformH5, true},
		{PlatformMiniApp, true},
		{PlatformDesktop, true},
		{"unknown", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.platform, func(t *testing.T) {
			assert.Equal(t, tt.valid, IsValidPlatform(tt.platform))
		})
	}
}
