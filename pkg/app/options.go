// ABOUTME: Option functions for configuring App.
// ABOUTME: Use WithXxx pattern for optional configuration.

package app

import "time"

// Option configures an App.
type Option func(*App)

// WithShutdownTimeout sets the shutdown timeout.
func WithShutdownTimeout(d time.Duration) Option {
	return func(app *App) {
		app.shutdownTimeout = d
	}
}
