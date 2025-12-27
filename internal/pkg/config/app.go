// ABOUTME: Application-level configuration.
// ABOUTME: Defines name, timezone, and encryption key settings.

package config

import (
	"strings"
	"time"
)

// App holds application-level configuration.
type App struct {
	Name     string `mapstructure:"name" json:"name" yaml:"name"`
	URL      string `mapstructure:"url" json:"url" yaml:"url"`
	Timezone string `mapstructure:"timezone" json:"timezone" yaml:"timezone"`
	Key      string `mapstructure:"key" json:"key" yaml:"key"`
}

// SetTimezone sets the application timezone.
func (a App) SetTimezone() {
	time.Local, _ = time.LoadLocation(a.Timezone)
}

// AssetURL returns the full URL for a given path.
func (a App) AssetURL(path string) string {
	if path == "" {
		return ""
	}
	// Already absolute URL
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}

	base := strings.TrimRight(a.URL, "/")
	path = strings.TrimLeft(path, "/")

	return base + "/" + path
}
