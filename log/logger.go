package logger

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

const (
	customTimeLayout = "2006-01-02 15:04:05.000"
)

var (
	defaultLogger atomic.Pointer[slog.Logger]
	workDir       string
)

func init() {
	if wd, err := os.Getwd(); err == nil {
		workDir = filepath.ToSlash(filepath.Clean(wd))
	}
}

var (
	OutputTypeJSON OutputType = 1
	OutputTypeText OutputType = 2

	OutputTimeZoneUTC   OutputTimeZone = 1
	OutputTimeZoneLocal OutputTimeZone = 2
)

type OutputType int
type OutputTimeZone int

type LoggerOption func(*LogConfig)

func WithLevel(level string) LoggerOption {
	return func(c *LogConfig) {
		c.Level = level
	}
}

func WithOutputType(outputType OutputType) LoggerOption {
	return func(c *LogConfig) {
		c.OutputType = outputType
	}
}

func WithTimeZone(timeZone OutputTimeZone) LoggerOption {
	return func(c *LogConfig) {
		c.TimeZone = timeZone
	}
}

func WithMaxParts(maxParts int) LoggerOption {
	return func(c *LogConfig) {
		c.MaxParts = maxParts
	}
}

func WithReplaceAttr(replaceAttr func(groups []string, a slog.Attr) slog.Attr) LoggerOption {
	return func(c *LogConfig) {
		c.ReplaceAttr = replaceAttr
	}
}

type LogConfig struct {
	Level       string
	OutputType  OutputType
	TimeZone    OutputTimeZone
	MaxParts    int
	ReplaceAttr func(groups []string, a slog.Attr) slog.Attr
}

// New creates and returns a brand-new independent *slog.Logger instance.
// Calling New NEVER modifies or affects the global default logger.
func New(opts ...LoggerOption) *slog.Logger {
	config := &LogConfig{}
	for _, opt := range opts {
		if opt != nil {
			opt(config)
		}
	}
	return slog.New(getHandler(config))
}

// LoggerInit initializes or updates the global default logger.
func LoggerInit(opts ...LoggerOption) {
	defaultLogger.Store(New(opts...))
}

// GetLogger returns the global default logger, lazily initializing it if not yet set.
func GetLogger() *slog.Logger {
	if l := defaultLogger.Load(); l != nil {
		return l
	}
	// Fallback lazy initialization if LoggerInit was not called
	l := New()
	if defaultLogger.CompareAndSwap(nil, l) {
		return l
	}
	return defaultLogger.Load()
}

func logAt(ctx context.Context, level slog.Level, msg string, args ...any) {
	l := GetLogger()
	if !l.Enabled(ctx, level) {
		return
	}
	var pcs [1]uintptr
	runtime.Callers(3, pcs[:])
	r := slog.NewRecord(time.Now(), level, msg, pcs[0])
	r.Add(args...)
	_ = l.Handler().Handle(ctx, r)
}

func logAttrsAt(ctx context.Context, level slog.Level, msg string, attrs ...slog.Attr) {
	l := GetLogger()
	if !l.Enabled(ctx, level) {
		return
	}
	var pcs [1]uintptr
	runtime.Callers(3, pcs[:])
	r := slog.NewRecord(time.Now(), level, msg, pcs[0])
	r.AddAttrs(attrs...)
	_ = l.Handler().Handle(ctx, r)
}

// Global convenience logging functions

func Info(msg string, args ...any) {
	logAt(context.Background(), slog.LevelInfo, msg, args...)
}

func Debug(msg string, args ...any) {
	logAt(context.Background(), slog.LevelDebug, msg, args...)
}

func Warn(msg string, args ...any) {
	logAt(context.Background(), slog.LevelWarn, msg, args...)
}

func Error(msg string, args ...any) {
	logAt(context.Background(), slog.LevelError, msg, args...)
}

func Log(ctx context.Context, level slog.Level, msg string, args ...any) {
	logAt(ctx, level, msg, args...)
}

func LogAttrs(ctx context.Context, level slog.Level, msg string, attrs ...slog.Attr) {
	logAttrsAt(ctx, level, msg, attrs...)
}

func With(args ...any) *slog.Logger {
	return GetLogger().With(args...)
}

func WithGroup(name string) *slog.Logger {
	return GetLogger().WithGroup(name)
}

func formatSource(source *slog.Source, maxParts int) string {
	if source == nil {
		return ""
	}
	if maxParts <= 0 {
		maxParts = 4 // default max 3 sub-directories + 1 filename
	}
	file := filepath.ToSlash(source.File)
	if workDir != "" && strings.HasPrefix(file, workDir+"/") {
		file = file[len(workDir)+1:]
	} else if workDir != "" && file == workDir {
		file = filepath.Base(file)
	}

	slashes := 0
	startIdx := 0
	for i := len(file) - 1; i >= 0; i-- {
		if file[i] == '/' {
			slashes++
			if slashes == maxParts {
				startIdx = i + 1
				break
			}
		}
	}

	trimmed := strings.TrimPrefix(file[startIdx:], "/")
	return trimmed + ":" + strconv.Itoa(source.Line)
}

func (c *LogConfig) defaultReplaceAttr(_ []string, a slog.Attr) slog.Attr {
	if a.Key == slog.TimeKey {
		t := a.Value.Time()
		if c.TimeZone == OutputTimeZoneLocal {
			t = t.Local()
		} else {
			t = t.UTC()
		}
		a.Value = slog.StringValue(t.Format(customTimeLayout))
	}
	if a.Key == slog.SourceKey {
		if source, ok := a.Value.Any().(*slog.Source); ok && source != nil {
			a.Value = slog.StringValue(formatSource(source, c.MaxParts))
		}
	}
	return a
}

func (c *LogConfig) getReplaceAttr() func(groups []string, a slog.Attr) slog.Attr {
	if c.ReplaceAttr != nil {
		return c.ReplaceAttr
	}
	return c.defaultReplaceAttr
}

func getLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "error":
		return slog.LevelError
	case "warn":
		return slog.LevelWarn
	default:
		return slog.LevelInfo
	}
}

func getHandler(config *LogConfig) slog.Handler {
	if config == nil {
		config = &LogConfig{}
	}
	options := slog.HandlerOptions{
		AddSource:   true,
		Level:       getLevel(config.Level),
		ReplaceAttr: config.getReplaceAttr(),
	}
	switch config.OutputType {
	case OutputTypeJSON:
		return slog.NewJSONHandler(os.Stdout, &options)
	case OutputTypeText:
		return slog.NewTextHandler(os.Stdout, &options)
	default:
		return slog.NewTextHandler(os.Stdout, &options)
	}
}
