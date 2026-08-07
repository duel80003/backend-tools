package logger

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const (
	customTimeLayout = "2006-01-02 15:04:05.000"
)

var (
	logger     *slog.Logger
	loggerOnce sync.Once
)

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

type LogConfig struct {
	Level      string
	OutputType OutputType
	TimeZone   OutputTimeZone
	MaxParts   int
}

// LoggerInit
// level: debug, info, warn, error
// this logger output as an UTC 0 timestamp
func LoggerInit(opts ...LoggerOption) {
	loggerOnce.Do(func() {
		config := &LogConfig{}
		for _, opt := range opts {
			if opt != nil {
				opt(config)
			}
		}
		logger = slog.New(getHandler(config))
	})
}

func GetLogger() *slog.Logger {
	if logger == nil {
		LoggerInit()
	}
	return logger
}

func Info(msg string, args ...any) {
	GetLogger().Info(msg, args...)
}

func Debug(msg string, args ...any) {
	GetLogger().Debug(msg, args...)
}

func Warn(msg string, args ...any) {
	GetLogger().Warn(msg, args...)
}

func Error(msg string, args ...any) {
	GetLogger().Error(msg, args...)
}

func Log(ctx context.Context, level slog.Level, msg string, args ...any) {
	GetLogger().Log(ctx, level, msg, args...)
}

func LogAttrs(ctx context.Context, level slog.Level, msg string, attrs ...slog.Attr) {
	GetLogger().LogAttrs(ctx, level, msg, attrs...)
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
	file := source.File
	if wd, err := os.Getwd(); err == nil {
		if rel, err := filepath.Rel(wd, file); err == nil && !strings.HasPrefix(rel, "..") {
			file = rel
		}
	}

	cleanPath := filepath.ToSlash(filepath.Clean(file))
	parts := strings.Split(cleanPath, "/")

	nonEmpty := make([]string, 0, len(parts))
	for _, p := range parts {
		if p != "" && p != "." {
			nonEmpty = append(nonEmpty, p)
		}
	}

	if len(nonEmpty) > maxParts {
		nonEmpty = nonEmpty[len(nonEmpty)-maxParts:]
	}

	trimmedPath := strings.Join(nonEmpty, "/")
	return fmt.Sprintf("%s:%d", trimmedPath, source.Line)
}

func (c *LogConfig) replaceAttr(_ []string, a slog.Attr) slog.Attr {
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
		ReplaceAttr: config.replaceAttr,
	}
	switch config.OutputType {
	case OutputTypeJSON:
		return slog.NewJSONHandler(os.Stdout, &options)
	case OutputTypeText:
		return slog.NewTextHandler(os.Stdout, &options)
	default:
		return slog.NewJSONHandler(os.Stdout, &options)
	}
}
