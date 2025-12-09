// ABOUTME: Option functions for configuring App.
// ABOUTME: Use WithXxx pattern for optional configuration.

package app

import (
	"time"

	"gorm.io/gorm"
)

// Option configures an App.
type Option func(*App)

// WithShutdownTimeout sets the shutdown timeout.
func WithShutdownTimeout(d time.Duration) Option {
	return func(app *App) {
		app.shutdownTimeout = d
	}
}

// WithConfig sets the application configuration.
func WithConfig(cfg any) Option {
	return func(app *App) {
		app.config = cfg
	}
}

// WithDB sets the database connection.
func WithDB(db *gorm.DB) Option {
	return func(app *App) {
		app.db = db
	}
}

// WithInitFunc sets a custom initialization function.
// Used for testing to verify Init() behavior.
func WithInitFunc(fn func() error) Option {
	return func(app *App) {
		app.initFunc = fn
	}
}
