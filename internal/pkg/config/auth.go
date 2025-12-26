// ABOUTME: Authentication configuration for multi-auth types.
// ABOUTME: Defines supported auth types (email/phone) and verification settings.

package config

// Auth holds authentication configuration.
type Auth struct {
	DefaultType       string   `mapstructure:"defaulttype" json:"defaulttype" yaml:"defaulttype"`
	AllowedTypes      []string `mapstructure:"allowedtypes" json:"allowedtypes" yaml:"allowedtypes"`
	EmailVerification bool     `mapstructure:"emailverification" json:"emailverification" yaml:"emailverification"`
	PhoneVerification bool     `mapstructure:"phoneverification" json:"phoneverification" yaml:"phoneverification"`
}
