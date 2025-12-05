// ABOUTME: Application-level configuration.
// ABOUTME: Defines name, timezone, and encryption key settings.

package config

import "time"

// App holds application-level configuration.
type App struct {
	Name     string `mapstructure:"name" json:"name" yaml:"name"`
	Timezone string `mapstructure:"timezone" json:"timezone" yaml:"timezone"`
	Key      string `mapstructure:"key" json:"key" yaml:"key"`
}

// SetTimezone sets the application timezone.
func (a App) SetTimezone() {
	time.Local, _ = time.LoadLocation(a.Timezone)
}
