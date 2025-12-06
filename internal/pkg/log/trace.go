package log

import (
	"context"

	"go.uber.org/zap"

	"bingo/pkg/contextx"
)

// 日志字段名常量.
var (
	KeyTrace    = "trace"
	KeySubject  = "subject"
	KeyIP       = "ip"
	KeyTask     = "task"
	KeyCost     = "cost"
	KeyResult   = "result"
	KeyCode     = "code"
	KeyMessage  = "message"
	KeyObject   = "object"
	KeyInstance = "instance"
	KeyInfo     = "info"
)

// C Parse context.
func C(ctx context.Context) *zapLogger {
	return std.C(ctx)
}

func (l *zapLogger) C(ctx context.Context) *zapLogger {
	lc := l.clone()

	// 定义一个映射，关联 context 提取函数和日志字段名.
	contextExtractors := map[string]func(context.Context) string{
		KeyTrace:    contextx.RequestID,
		KeySubject:  contextx.UserID,
		KeyIP:       contextx.ClientIP,
		KeyTask:     contextx.Task,
		KeyObject:   contextx.Object,
		KeyInstance: contextx.Instance,
		KeyInfo:     contextx.Info,
	}

	// 遍历映射，从 context 中提取值并添加到日志中.
	for fieldName, extractor := range contextExtractors {
		if val := extractor(ctx); val != "" {
			lc.z = lc.z.With(zap.String(fieldName, val))
		}
	}

	return lc
}

func (l *zapLogger) clone() *zapLogger {
	lc := *l

	return &lc
}
