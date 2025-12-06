package log

import (
	"log"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger interface {
	Debug(msg string, fields ...zapcore.Field)
	Info(msg string, fields ...zapcore.Field)
	Warn(msg string, fields ...zapcore.Field)
	Error(msg string, fields ...zapcore.Field)
	Panic(msg string, fields ...zapcore.Field)
	Fatal(msg string, fields ...zapcore.Field)

	Debugf(format string, v ...interface{})
	Infof(format string, v ...interface{})
	Warnf(format string, v ...interface{})
	Errorf(format string, v ...interface{})
	Panicf(format string, v ...interface{})
	Fatalf(format string, v ...interface{})

	Debugw(msg string, keysAndValues ...interface{})
	Infow(msg string, keysAndValues ...interface{})
	Warnw(msg string, keysAndValues ...interface{})
	Errorw(msg string, keysAndValues ...interface{})
	Panicw(msg string, keysAndValues ...interface{})
	Fatalw(msg string, keysAndValues ...interface{})

	Sync()
}

type zapLogger struct {
	z *zap.Logger
}

var _ Logger = &zapLogger{}

var (
	mu  sync.Mutex
	std = NewLogger(NewDefaultOptions())
)

func Init(opts *Options) {
	mu.Lock()
	defer mu.Unlock()

	std = NewLogger(opts)
}

func NewLogger(opts *Options) *zapLogger {
	if opts == nil {
		opts = NewDefaultOptions()
	}

	z := NewChannel(opts)

	l := &zapLogger{z: z}

	zap.RedirectStdLog(z)

	return l
}

// SugaredLogger returns global sugared logger.
func SugaredLogger() *zap.SugaredLogger {
	return std.z.Sugar()
}

// StdErrLogger returns logger of standard library which writes to supplied zap
// logger at error level.
func StdErrLogger() *log.Logger {
	if std == nil {
		return nil
	}
	if l, err := zap.NewStdLogAt(std.z, zapcore.ErrorLevel); err == nil {
		return l
	}

	return nil
}

// StdInfoLogger returns logger of standard library which writes to supplied zap
// logger at info level.
func StdInfoLogger() *log.Logger {
	if std == nil {
		return nil
	}
	if l, err := zap.NewStdLogAt(std.z, zapcore.InfoLevel); err == nil {
		return l
	}

	return nil
}

func Sync() { std.Sync() }

func (l *zapLogger) Sync() {
	_ = l.z.Sync()
}

func Debug(msg string, fields ...zapcore.Field) {
	std.z.Debug(msg, fields...)
}

func (l *zapLogger) Debug(msg string, fields ...zapcore.Field) {
	l.z.Debug(msg, fields...)
}

func Info(msg string, fields ...zapcore.Field) {
	std.z.Info(msg, fields...)
}

func (l *zapLogger) Info(msg string, fields ...zapcore.Field) {
	l.z.Info(msg, fields...)
}

func Warn(msg string, fields ...zapcore.Field) {
	std.z.Warn(msg, fields...)
}

func (l *zapLogger) Warn(msg string, fields ...zapcore.Field) {
	l.z.Warn(msg, fields...)
}

func Error(msg string, fields ...zapcore.Field) {
	std.z.Error(msg, fields...)
}

func (l *zapLogger) Error(msg string, fields ...zapcore.Field) {
	l.z.Error(msg, fields...)
}

func Panic(msg string, fields ...zapcore.Field) {
	std.z.Panic(msg, fields...)
}

func (l *zapLogger) Panic(msg string, fields ...zapcore.Field) {
	l.z.Panic(msg, fields...)
}

func Fatal(msg string, fields ...zapcore.Field) {
	std.z.Fatal(msg, fields...)
}

func (l *zapLogger) Fatal(msg string, fields ...zapcore.Field) {
	l.z.Fatal(msg, fields...)
}

func Debugf(format string, v ...interface{}) {
	std.z.Sugar().Debugf(format, v...)
}

func (l *zapLogger) Debugf(format string, v ...interface{}) {
	l.z.Sugar().Debugf(format, v...)
}

func Infof(format string, v ...interface{}) {
	std.z.Sugar().Infof(format, v...)
}

func (l *zapLogger) Infof(format string, v ...interface{}) {
	l.z.Sugar().Infof(format, v...)
}

func Warnf(format string, v ...interface{}) {
	std.z.Sugar().Warnf(format, v...)
}

func (l *zapLogger) Warnf(format string, v ...interface{}) {
	l.z.Sugar().Warnf(format, v...)
}

func Errorf(format string, v ...interface{}) {
	std.z.Sugar().Errorf(format, v...)
}

func (l *zapLogger) Errorf(format string, v ...interface{}) {
	l.z.Sugar().Errorf(format, v...)
}

func Panicf(format string, v ...interface{}) {
	std.z.Sugar().Panicf(format, v...)
}

func (l *zapLogger) Panicf(format string, v ...interface{}) {
	l.z.Sugar().Panicf(format, v...)
}

func Fatalf(format string, v ...interface{}) {
	std.z.Sugar().Fatalf(format, v...)
}

func (l *zapLogger) Fatalf(format string, v ...interface{}) {
	l.z.Sugar().Fatalf(format, v...)
}

func Debugw(msg string, keysAndValues ...interface{}) {
	std.z.Sugar().Debugw(msg, keysAndValues...)
}

func (l *zapLogger) Debugw(msg string, keysAndValues ...interface{}) {
	l.z.Sugar().Debugw(msg, keysAndValues...)
}

func Infow(msg string, keysAndValues ...interface{}) {
	std.z.Sugar().Infow(msg, keysAndValues...)
}

func (l *zapLogger) Infow(msg string, keysAndValues ...interface{}) {
	l.z.Sugar().Infow(msg, keysAndValues...)
}

func Warnw(msg string, keysAndValues ...interface{}) {
	std.z.Sugar().Warnw(msg, keysAndValues...)
}

func (l *zapLogger) Warnw(msg string, keysAndValues ...interface{}) {
	l.z.Sugar().Warnw(msg, keysAndValues...)
}

func Errorw(msg string, keysAndValues ...interface{}) {
	std.z.Sugar().Errorw(msg, keysAndValues...)
}

func (l *zapLogger) Errorw(msg string, keysAndValues ...interface{}) {
	l.z.Sugar().Errorw(msg, keysAndValues...)
}

func Panicw(msg string, keysAndValues ...interface{}) {
	std.z.Sugar().Panicw(msg, keysAndValues...)
}

func (l *zapLogger) Panicw(msg string, keysAndValues ...interface{}) {
	l.z.Sugar().Panicw(msg, keysAndValues...)
}

func Fatalw(msg string, keysAndValues ...interface{}) {
	std.z.Sugar().Fatalw(msg, keysAndValues...)
}

func (l *zapLogger) Fatalw(msg string, keysAndValues ...interface{}) {
	l.z.Sugar().Fatalw(msg, keysAndValues...)
}
