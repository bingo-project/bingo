package logger

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/bingo-project/component-base/log"
	"github.com/duke-git/lancet/v2/slice"
	"go.uber.org/zap"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

var (
	opt       = zap.AddCallerSkip(3)
	blacklist = []string{
		"SELECT * FROM `sys_casbin_rule` ORDER BY ID",
		"SHOW STATUS",
	}
)

// Config defines a gorm logger configuration.
type Config struct {
	SlowThreshold time.Duration
	LogLevel      gormlogger.LogLevel
}

// New create a gorm logger instance.
func New(level int) gormlogger.Interface {
	config := Config{
		SlowThreshold: 200 * time.Millisecond,
		LogLevel:      gormlogger.LogLevel(level),
	}

	return &logger{
		Config: config,
	}
}

type logger struct {
	Config
}

// LogMode log mode.
func (l *logger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	newlogger := *l
	newlogger.LogLevel = level

	return &newlogger
}

// Info print info.
func (l logger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormlogger.Info {
		log.C(ctx).WithOption(opt).Infow(msg, "data", data)
	}
}

// Warn print warn messages.
func (l logger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormlogger.Warn {
		log.C(ctx).WithOption(opt).Warnw(msg, "data", data)
	}
}

// Error print error messages.
func (l logger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormlogger.Error {
		log.C(ctx).WithOption(opt).Errorw(msg, "data", data)
	}
}

// Trace print sql message.
func (l logger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel <= 0 {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	if slice.Contain(blacklist, sql) {
		return
	}

	switch {
	case err != nil && l.LogLevel >= gormlogger.Error && (!errors.Is(err, gorm.ErrRecordNotFound) || l.LogLevel >= gormlogger.Error):
		log.C(ctx).WithOption(opt).Errorw("SQL Error", "err", err, "elapsed", elapsed, "sql", sql, "rows", rows)
	case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= gormlogger.Warn:
		slowLog := fmt.Sprintf("SLOW SQL >= %v", l.SlowThreshold)
		log.C(ctx).WithOption(opt).Warnw("SQL Slow", "log", slowLog, "elapsed", elapsed, "sql", sql, "rows", rows)
	case l.LogLevel >= gormlogger.Info:
		log.C(ctx).WithOption(opt).Infow("SQL Info", "elapsed", elapsed, "sql", sql, "rows", rows)
	}
}
