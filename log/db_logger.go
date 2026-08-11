package logger

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"
	"strings"
	"time"

	"gorm.io/gorm/logger"
)

type contextKey string

const (
	Caller              contextKey = "caller"
	defaultMaxSQLLength int        = 500
)

type GormToELK struct {
	Log          *slog.Logger
	MaxSQLLength int
	logger.Config
}

func (l *GormToELK) LogMode(level logger.LogLevel) logger.Interface {
	newlogger := *l
	newlogger.LogLevel = level
	return &newlogger
}

func (l *GormToELK) Info(ctx context.Context, s string, i ...any) {
	if l.LogLevel >= logger.Info {
		log := l.Log
		if log == nil {
			log = GetLogger()
		}
		log.InfoContext(ctx, fmt.Sprintf(s, i...))
	}
}

func (l *GormToELK) Warn(ctx context.Context, s string, i ...any) {
	if l.LogLevel >= logger.Warn {
		log := l.Log
		if log == nil {
			log = GetLogger()
		}
		log.WarnContext(ctx, fmt.Sprintf(s, i...))
	}
}

func (l *GormToELK) Error(ctx context.Context, s string, i ...any) {
	if l.LogLevel >= logger.Error {
		log := l.Log
		if log == nil {
			log = GetLogger()
		}
		log.ErrorContext(ctx, fmt.Sprintf(s, i...))
	}
}

func (l *GormToELK) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel <= logger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	maxLen := l.MaxSQLLength
	if maxLen == 0 {
		maxLen = defaultMaxSQLLength
	}

	if maxLen > 0 && len(sql) > maxLen {
		runes := []rune(sql)
		if len(runes) > maxLen {
			sql = string(runes[:maxLen]) + "..."
		}
	}

	log := l.Log
	if log == nil {
		log = GetLogger()
	}

	// Only log if you want all queries, or set logic for SlowThreshold
	entry := log.With(
		"log_type", "sql_monitor",
		"latency_ms", float64(elapsed.Nanoseconds())/1e6,
		"rows_affected", rows,
		"query", sql,
		"func", getCallerFunctionName(ctx),
	)

	if err != nil {
		entry.ErrorContext(ctx, "SQL_ERROR", "error", err)
	} else if elapsed > l.SlowThreshold {
		entry.WarnContext(ctx, "SLOW_SQL")
	} else {
		entry.InfoContext(ctx, "SQL_OK")
	}
}

// getCallerFunctionName returns the name of the function that called the database operation.
func getCallerFunctionName(ctx context.Context) string {
	if ctx != nil {
		if alias, ok := ctx.Value(Caller).(string); ok {
			return alias
		}
	}

	pc := make([]uintptr, 15)
	n := runtime.Callers(3, pc) // Skip: getCallerFunctionName, Trace, and gorm internals
	frames := runtime.CallersFrames(pc[:n])

	for {
		frame, more := frames.Next()
		// Filter out GORM internal frames and standard library frames if necessary
		if !strings.Contains(frame.Function, "gorm.io") && !strings.Contains(frame.Function, "runtime.") {
			return frame.Function
		}
		if !more {
			break
		}
	}
	return "unknown"
}

type GormOption func(*GormToELK)

// WithMaxSQLLength creates a GormOption to set MaxSQLLength.
// Set to 0 for default (500), or -1 for unlimited length.
func WithMaxSQLLength(maxLen int) GormOption {
	return func(l *GormToELK) {
		l.MaxSQLLength = maxLen
	}
}

// WithSlowThreshold sets the slow query threshold.
func WithSlowThreshold(d time.Duration) GormOption {
	return func(l *GormToELK) {
		l.SlowThreshold = d
	}
}

// WithGormLogLevel sets the GORM log level.
func WithGormLogLevel(level logger.LogLevel) GormOption {
	return func(l *GormToELK) {
		l.LogLevel = level
	}
}

// WithSlogLogger sets a custom *slog.Logger instance.
func WithSlogLogger(log *slog.Logger) GormOption {
	return func(l *GormToELK) {
		l.Log = log
	}
}

// NewGormToELK initializes a GormToELK logger instance.
// Accepts both GormOption (e.g. WithMaxSQLLength, WithSlowThreshold) and LoggerOption (e.g. WithLevel).
func NewGormToELK(opts ...any) *GormToELK {
	var loggerOpts []LoggerOption
	var gormOpts []GormOption

	for _, opt := range opts {
		switch o := opt.(type) {
		case GormOption:
			gormOpts = append(gormOpts, o)
		case func(*GormToELK):
			gormOpts = append(gormOpts, o)
		case LoggerOption:
			loggerOpts = append(loggerOpts, o)
		case func(*LogConfig):
			loggerOpts = append(loggerOpts, o)
		}
	}

	logOpts := append([]LoggerOption{WithOutputType(OutputTypeJSON)}, loggerOpts...)
	g := &GormToELK{
		Log:          New(logOpts...),
		MaxSQLLength: defaultMaxSQLLength,
		Config: logger.Config{
			SlowThreshold: 200 * time.Millisecond, // Slow SQL threshold
			LogLevel:      logger.Info,            // Log level
			Colorful:      false,
		},
	}

	for _, opt := range gormOpts {
		if opt != nil {
			opt(g)
		}
	}

	return g
}

func getLogger() *GormToELK {
	return NewGormToELK()
}

func LogDBDefaultReplaceAttr(groups []string, a slog.Attr) slog.Attr {
	if a.Key == "func" || a.Key == "latency_ms" || a.Key == "query" {
		return a
	}
	return slog.Attr{}
}

// SetMaxSQLLength sets the max length of SQL string logged in Trace.
// Set to 0 for default (500), or -1 for unlimited length.
func (l *GormToELK) SetMaxSQLLength(maxLen int) *GormToELK {
	l.MaxSQLLength = maxLen
	return l
}

// WithGormOptions applies GormOptions to GormToELK.
func (l *GormToELK) WithGormOptions(opts ...GormOption) *GormToELK {
	for _, opt := range opts {
		if opt != nil {
			opt(l)
		}
	}
	return l
}
