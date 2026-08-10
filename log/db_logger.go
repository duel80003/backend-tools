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

const Caller contextKey = "caller"

type GormToELK struct {
	Log *slog.Logger
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
	elapsed := time.Since(begin)
	sql, rows := fc()
	if len(sql) > 50 {
		sql = sql[:50]
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
		if alias, ok := ctx.Value("caller").(string); ok {
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

func NewGormToELK(opts ...LoggerOption) *GormToELK {
	logOpts := append([]LoggerOption{WithOutputType(OutputTypeJSON)}, opts...)
	return &GormToELK{
		Log: New(logOpts...),
		Config: logger.Config{
			SlowThreshold: 200 * time.Millisecond, // Slow SQL threshold
			LogLevel:      logger.Info,            // Log level
			Colorful:      false,
		},
	}
}

func getLogger() *GormToELK {
	return NewGormToELK()
}

func LogDBDefaultReplaceAttr(groups []string, a slog.Attr) slog.Attr {
	if a.Key == "func" || a.Key == "latency_ms" {
		return a
	}
	return slog.Attr{}
}
