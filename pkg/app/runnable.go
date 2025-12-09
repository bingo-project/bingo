// ABOUTME: Core component interfaces for the App lifecycle.
// ABOUTME: Runnable for long-running components, Registrar for setup-only components.

package app

import "context"

// Runnable is a component that runs for the lifetime of the App.
// Start blocks until ctx is canceled or an error occurs.
type Runnable interface {
	Start(ctx context.Context) error
}

// Registrar is a component that registers with the App during startup.
// Register is called once before any Runnable starts.
type Registrar interface {
	Register(app *App) error
}

// Named is an optional interface for components that have a name.
// Used for logging.
type Named interface {
	Name() string
}
