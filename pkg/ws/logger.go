// ABOUTME: Logger interface for WebSocket package.
// ABOUTME: Allows callers to inject their own logger implementation.

package ws

import (
	"context"
)

// Logger defines the logging interface used by the ws package.
type Logger interface {
	Debugw(msg string, keysAndValues ...any)
	Infow(msg string, keysAndValues ...any)
	Warnw(msg string, keysAndValues ...any)
	Errorw(msg string, keysAndValues ...any)
	WithContext(ctx context.Context) Logger
}

// nopLogger discards all log output.
type nopLogger struct{}

func (nopLogger) Debugw(string, ...any)              {}
func (nopLogger) Infow(string, ...any)               {}
func (nopLogger) Warnw(string, ...any)               {}
func (nopLogger) Errorw(string, ...any)              {}
func (nopLogger) WithContext(context.Context) Logger { return nopLogger{} }

// NopLogger returns a logger that discards all output.
// This is the default logger used when no logger is provided.
func NopLogger() Logger {
	return nopLogger{}
}
