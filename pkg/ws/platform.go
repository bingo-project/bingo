// ABOUTME: Platform constants for client identification.
// ABOUTME: Defines valid platforms and validation function.

package ws

// Platform constants.
const (
	PlatformWeb     = "web"
	PlatformIOS     = "ios"
	PlatformAndroid = "android"
	PlatformH5      = "h5"
	PlatformMiniApp = "miniapp"
	PlatformDesktop = "desktop"
)

// IsValidPlatform checks if the platform string is valid.
func IsValidPlatform(p string) bool {
	switch p {
	case PlatformWeb, PlatformIOS, PlatformAndroid, PlatformH5, PlatformMiniApp, PlatformDesktop:
		return true
	}

	return false
}
