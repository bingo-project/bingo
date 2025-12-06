// ABOUTME: Logger interface for WebSocket package.
// ABOUTME: Allows callers to inject their own logger implementation.

package ws

import (
	"context"

	"github.com/bingo-project/component-base/log"
)

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
	log.Debugw(msg, keysAndValues...)
}

func (defaultLogger) Infow(msg string, keysAndValues ...any) {
	log.Infow(msg, keysAndValues...)
}

func (defaultLogger) Warnw(msg string, keysAndValues ...any) {
	log.Warnw(msg, keysAndValues...)
}

func (defaultLogger) Errorw(msg string, keysAndValues ...any) {
	log.Errorw(msg, keysAndValues...)
}

func (defaultLogger) WithContext(ctx context.Context) Logger {
	return contextLogger{ctx: ctx}
}

// contextLogger wraps the bingo log package with context support.
type contextLogger struct {
	ctx context.Context
}

func (l contextLogger) Debugw(msg string, keysAndValues ...any) {
	log.C(l.ctx).Debugw(msg, keysAndValues...)
}

func (l contextLogger) Infow(msg string, keysAndValues ...any) {
	log.C(l.ctx).Infow(msg, keysAndValues...)
}

func (l contextLogger) Warnw(msg string, keysAndValues ...any) {
	log.C(l.ctx).Warnw(msg, keysAndValues...)
}

func (l contextLogger) Errorw(msg string, keysAndValues ...any) {
	log.C(l.ctx).Errorw(msg, keysAndValues...)
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

// NopLogger returns a logger that discards all output.
func NopLogger() Logger {
	return nopLogger{}
}
