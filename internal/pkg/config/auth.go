// ABOUTME: Authentication configuration for multi-auth types.
// ABOUTME: Defines supported auth types (email/phone) and verification settings.

package config

import "time"

// Auth holds authentication configuration.
type Auth struct {
	DefaultType       string   `mapstructure:"defaulttype" json:"defaulttype" yaml:"defaulttype"`
	AllowedTypes      []string `mapstructure:"allowedtypes" json:"allowedtypes" yaml:"allowedtypes"`
	EmailVerification bool     `mapstructure:"emailverification" json:"emailverification" yaml:"emailverification"`
	PhoneVerification bool     `mapstructure:"phoneverification" json:"phoneverification" yaml:"phoneverification"`
	SIWE              SIWE     `mapstructure:"siwe" json:"siwe" yaml:"siwe"`
}

// SIWE holds Sign-In with Ethereum configuration.
type SIWE struct {
	Enabled         bool          `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
	Domains         []string      `mapstructure:"domains" json:"domains" yaml:"domains"`
	Statement       string        `mapstructure:"statement" json:"statement" yaml:"statement"`
	ChainID         int           `mapstructure:"chainId" json:"chainId" yaml:"chainId"`
	NonceExpiration time.Duration `mapstructure:"nonceExpiration" json:"nonceExpiration" yaml:"nonceExpiration"`
}
