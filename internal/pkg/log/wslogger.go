// ABOUTME: Adapter that implements ws.Logger interface using internal log package.
// ABOUTME: Enables WebSocket package to use bingo's logging with context support.

package log

import (
	"context"

	"go.uber.org/zap"

	"github.com/bingo-project/bingo/pkg/ws"
)

// callerSkip adjusts the caller frame to show the actual calling code instead of this wrapper.
var callerSkip = zap.AddCallerSkip(1)

// wsLogger wraps the bingo log package to implement ws.Logger.
type wsLogger struct{}

// NewWSLogger returns a ws.Logger that uses the internal log package.
func NewWSLogger() ws.Logger {
	return wsLogger{}
}

func (wsLogger) Debugw(msg string, keysAndValues ...any) {
	std.z.WithOptions(callerSkip).Sugar().Debugw(msg, keysAndValues...)
}

func (wsLogger) Infow(msg string, keysAndValues ...any) {
	std.z.WithOptions(callerSkip).Sugar().Infow(msg, keysAndValues...)
}

func (wsLogger) Warnw(msg string, keysAndValues ...any) {
	std.z.WithOptions(callerSkip).Sugar().Warnw(msg, keysAndValues...)
}

func (wsLogger) Errorw(msg string, keysAndValues ...any) {
	std.z.WithOptions(callerSkip).Sugar().Errorw(msg, keysAndValues...)
}

func (wsLogger) WithContext(ctx context.Context) ws.Logger {
	return &wsContextLogger{ctx: ctx}
}

// wsContextLogger wraps the bingo log package with context support.
type wsContextLogger struct {
	ctx context.Context
}

func (l *wsContextLogger) Debugw(msg string, keysAndValues ...any) {
	C(l.ctx).z.WithOptions(callerSkip).Sugar().Debugw(msg, keysAndValues...)
}

func (l *wsContextLogger) Infow(msg string, keysAndValues ...any) {
	C(l.ctx).z.WithOptions(callerSkip).Sugar().Infow(msg, keysAndValues...)
}

func (l *wsContextLogger) Warnw(msg string, keysAndValues ...any) {
	C(l.ctx).z.WithOptions(callerSkip).Sugar().Warnw(msg, keysAndValues...)
}

func (l *wsContextLogger) Errorw(msg string, keysAndValues ...any) {
	C(l.ctx).z.WithOptions(callerSkip).Sugar().Errorw(msg, keysAndValues...)
}

func (l *wsContextLogger) WithContext(ctx context.Context) ws.Logger {
	return &wsContextLogger{ctx: ctx}
}
