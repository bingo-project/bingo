// ABOUTME: Logger interface for WebSocket package.
// ABOUTME: Allows callers to inject their own logger implementation.

package ws

import (
	"context"

	"github.com/bingo-project/component-base/log"
	"go.uber.org/zap"
)

// callerSkip adjusts the caller frame to show the actual calling code instead of this wrapper.
var callerSkip = zap.AddCallerSkip(1)

// Logger defines the logging interface used by the ws package.
type Logger interface {
	Debugw(msg string, keysAndValues ...any)
	Infow(msg string, keysAndValues ...any)
	Warnw(msg string, keysAndValues ...any)
	Errorw(msg string, keysAndValues ...any)
	WithContext(ctx context.Context) Logger
}

// defaultLogger wraps the bingo log package.
type defaultLogger struct{}

func (defaultLogger) Debugw(msg string, keysAndValues ...any) {
	log.SugaredLogger().WithOptions(callerSkip).Debugw(msg, keysAndValues...)
}

func (defaultLogger) Infow(msg string, keysAndValues ...any) {
	log.SugaredLogger().WithOptions(callerSkip).Infow(msg, keysAndValues...)
}

func (defaultLogger) Warnw(msg string, keysAndValues ...any) {
	log.SugaredLogger().WithOptions(callerSkip).Warnw(msg, keysAndValues...)
}

func (defaultLogger) Errorw(msg string, keysAndValues ...any) {
	log.SugaredLogger().WithOptions(callerSkip).Errorw(msg, keysAndValues...)
}

func (defaultLogger) WithContext(ctx context.Context) Logger {
	return contextLogger{ctx: ctx}
}

// contextLogger wraps the bingo log package with context support.
type contextLogger struct {
	ctx context.Context
}

func (l contextLogger) Debugw(msg string, keysAndValues ...any) {
	log.C(l.ctx).WithOption(callerSkip).Debugw(msg, keysAndValues...)
}

func (l contextLogger) Infow(msg string, keysAndValues ...any) {
	log.C(l.ctx).WithOption(callerSkip).Infow(msg, keysAndValues...)
}

func (l contextLogger) Warnw(msg string, keysAndValues ...any) {
	log.C(l.ctx).WithOption(callerSkip).Warnw(msg, keysAndValues...)
}

func (l contextLogger) Errorw(msg string, keysAndValues ...any) {
	log.C(l.ctx).WithOption(callerSkip).Errorw(msg, keysAndValues...)
}

func (l contextLogger) WithContext(ctx context.Context) Logger {
	return contextLogger{ctx: ctx}
}

// nopLogger discards all log output.
type nopLogger struct{}

func (nopLogger) Debugw(string, ...any)              {}
func (nopLogger) Infow(string, ...any)               {}
func (nopLogger) Warnw(string, ...any)               {}
func (nopLogger) Errorw(string, ...any)              {}
func (nopLogger) WithContext(context.Context) Logger { return nopLogger{} }

// DefaultLogger returns the default logger implementation.
func DefaultLogger() Logger {
	return defaultLogger{}
}

// NopLogger returns a logger that discards all output.
func NopLogger() Logger {
	return nopLogger{}
}
