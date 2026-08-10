package logger

import (
	"context"
	"log/slog"
	"path/filepath"
	"testing"
	"time"
)

func TestOptions(t *testing.T) {
	config := &LogConfig{}

	WithLevel("debug")(config)
	if config.Level != "debug" {
		t.Errorf("expected Level 'debug', got '%s'", config.Level)
	}

	WithOutputType(OutputTypeText)(config)
	if config.OutputType != OutputTypeText {
		t.Errorf("expected OutputType OutputTypeText (%d), got %d", OutputTypeText, config.OutputType)
	}

	WithTimeZone(OutputTimeZoneLocal)(config)
	if config.TimeZone != OutputTimeZoneLocal {
		t.Errorf("expected TimeZone OutputTimeZoneLocal (%d), got %d", OutputTimeZoneLocal, config.TimeZone)
	}

	WithMaxParts(2)(config)
	if config.MaxParts != 2 {
		t.Errorf("expected MaxParts 2, got %d", config.MaxParts)
	}

	customReplace := func(groups []string, a slog.Attr) slog.Attr { return a }
	WithReplaceAttr(customReplace)(config)
	if config.ReplaceAttr == nil {
		t.Error("expected ReplaceAttr to be set")
	}
}

func TestGetLevel(t *testing.T) {
	tests := []struct {
		input    LogLevel
		expected slog.Level
	}{
		{LevelDebug, slog.LevelDebug},
		{LevenInfo, slog.LevelInfo},
		{LevelWarn, slog.LevelWarn},
		{LevelError, slog.LevelError},
		{LogLevel("error"), slog.LevelError},
	}

	for _, tt := range tests {
		t.Run(string(tt.input), func(t *testing.T) {
			got := getLevel(tt.input)
			if got != tt.expected {
				t.Errorf("getLevel(%q) = %v; want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestFormatSource(t *testing.T) {
	tests := []struct {
		name     string
		source   *slog.Source
		maxParts int
		expected string
	}{
		{
			name:     "nil source",
			source:   nil,
			maxParts: 4,
			expected: "",
		},
		{
			name: "maxParts 0 (fallback to default 4)",
			source: &slog.Source{
				File: "/a/b/c/d/e/file.go",
				Line: 100,
			},
			maxParts: 0,
			expected: "c/d/e/file.go:100",
		},
		{
			name: "more than 3 subdirectories (5 levels) with default maxParts 4",
			source: &slog.Source{
				File: "/a/b/c/d/e/file.go",
				Line: 100,
			},
			maxParts: 4,
			expected: "c/d/e/file.go:100",
		},
		{
			name: "custom maxParts 2 (1 subdirectory + filename)",
			source: &slog.Source{
				File: "/a/b/c/d/e/file.go",
				Line: 100,
			},
			maxParts: 2,
			expected: "e/file.go:100",
		},
		{
			name: "exactly 3 subdirectories with maxParts 4",
			source: &slog.Source{
				File: "/dir1/dir2/dir3/file.go",
				Line: 50,
			},
			maxParts: 4,
			expected: "dir1/dir2/dir3/file.go:50",
		},
		{
			name: "1 subdirectory (like log/logger_test.go:229)",
			source: &slog.Source{
				File: "/log/logger_test.go",
				Line: 229,
			},
			maxParts: 4,
			expected: "log/logger_test.go:229",
		},
		{
			name: "only filename (file on root)",
			source: &slog.Source{
				File: "filename.go",
				Line: 12,
			},
			maxParts: 4,
			expected: "filename.go:12",
		},
		{
			name: "file equals workDir",
			source: &slog.Source{
				File: workDir,
				Line: 5,
			},
			maxParts: 4,
			expected: filepath.Base(workDir) + ":5",
		},
		{
			name: "file with workDir prefix",
			source: &slog.Source{
				File: workDir + "/a/b/c/file.go",
				Line: 20,
			},
			maxParts: 4,
			expected: "a/b/c/file.go:20",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatSource(tt.source, tt.maxParts)
			if got != tt.expected {
				t.Errorf("formatSource() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestReplaceAttr(t *testing.T) {
	testTime := time.Date(2026, 8, 7, 12, 34, 56, 789000000, time.UTC)
	timeAttr := slog.Time(slog.TimeKey, testTime)
	strAttr := slog.String("key", "val")
	srcAttr := slog.Any(slog.SourceKey, &slog.Source{
		File: "/log/logger_test.go",
		Line: 229,
	})
	nonSrcAttr := slog.Any(slog.SourceKey, "invalid_source")
	nilSrcAttr := slog.Any(slog.SourceKey, (*slog.Source)(nil))

	cfgUTC := &LogConfig{TimeZone: OutputTimeZoneUTC, MaxParts: 4}
	cfgLocal := &LogConfig{TimeZone: OutputTimeZoneLocal, MaxParts: 4}

	// Test UTC replaceAttr (via getReplaceAttr())
	replaceAttrUTC := cfgUTC.getReplaceAttr()
	resUTC := replaceAttrUTC(nil, timeAttr)
	expectedUTCStr := testTime.UTC().Format(customTimeLayout)
	if resUTC.Value.String() != expectedUTCStr {
		t.Errorf("replaceAttr UTC time = %s, want %s", resUTC.Value.String(), expectedUTCStr)
	}

	resUTCStr := replaceAttrUTC(nil, strAttr)
	if resUTCStr.Key != "key" || resUTCStr.Value.String() != "val" {
		t.Errorf("replaceAttr non-time attr modified: %v", resUTCStr)
	}

	resUTCSrc := replaceAttrUTC(nil, srcAttr)
	expectedSrcStr := "log/logger_test.go:229"
	if resUTCSrc.Value.String() != expectedSrcStr {
		t.Errorf("replaceAttr UTC source = %s, want %s", resUTCSrc.Value.String(), expectedSrcStr)
	}

	// Test Local replaceAttr
	replaceAttrLocal := cfgLocal.getReplaceAttr()
	resLocal := replaceAttrLocal(nil, timeAttr)
	expectedLocalStr := testTime.Local().Format(customTimeLayout)
	if resLocal.Value.String() != expectedLocalStr {
		t.Errorf("replaceAttr Local time = %s, want %s", resLocal.Value.String(), expectedLocalStr)
	}

	resLocalStr := replaceAttrLocal(nil, strAttr)
	if resLocalStr.Key != "key" || resLocalStr.Value.String() != "val" {
		t.Errorf("replaceAttr Local non-time attr modified: %v", resLocalStr)
	}

	resLocalSrc := replaceAttrLocal(nil, srcAttr)
	if resLocalSrc.Value.String() != expectedSrcStr {
		t.Errorf("replaceAttr Local source = %s, want %s", resLocalSrc.Value.String(), expectedSrcStr)
	}

	// Test invalid / nil source attrs
	resNonSrc := replaceAttrUTC(nil, nonSrcAttr)
	if resNonSrc.Value.String() != "invalid_source" {
		t.Errorf("replaceAttr non-Source source key modified unexpectedly: %v", resNonSrc)
	}

	resNilSrc := replaceAttrUTC(nil, nilSrcAttr)
	if resNilSrc.Value.Any() != (*slog.Source)(nil) {
		t.Errorf("replaceAttr nil Source modified unexpectedly: %v", resNilSrc)
	}

	// Test Custom ReplaceAttr override
	customCalled := false
	customReplace := func(groups []string, a slog.Attr) slog.Attr {
		customCalled = true
		return slog.String(a.Key, "custom_"+a.Value.String())
	}
	cfgCustom := &LogConfig{ReplaceAttr: customReplace}
	resCustom := cfgCustom.getReplaceAttr()(nil, strAttr)
	if !customCalled || resCustom.Value.String() != "custom_val" {
		t.Errorf("expected custom ReplaceAttr to be invoked, got: %v", resCustom)
	}
}

func TestGetHandler(t *testing.T) {
	t.Run("JSON Handler", func(t *testing.T) {
		cfg := &LogConfig{
			Level:      "info",
			OutputType: OutputTypeJSON,
			TimeZone:   OutputTimeZoneUTC,
			MaxParts:   4,
		}
		handler := getHandler(cfg)
		if handler == nil {
			t.Fatal("expected non-nil handler for OutputTypeJSON")
		}
	})

	t.Run("Text Handler", func(t *testing.T) {
		cfg := &LogConfig{
			Level:      "debug",
			OutputType: OutputTypeText,
			TimeZone:   OutputTimeZoneLocal,
			MaxParts:   2,
		}
		handler := getHandler(cfg)
		if handler == nil {
			t.Fatal("expected non-nil handler for OutputTypeText")
		}
	})

	t.Run("Default Handler", func(t *testing.T) {
		cfg := &LogConfig{
			Level:      "warn",
			OutputType: OutputType(99),
			TimeZone:   OutputTimeZoneUTC,
		}
		handler := getHandler(cfg)
		if handler == nil {
			t.Fatal("expected non-nil handler for default OutputType")
		}
	})

	t.Run("Nil Config Handler", func(t *testing.T) {
		handler := getHandler(nil)
		if handler == nil {
			t.Fatal("expected non-nil handler for nil LogConfig")
		}
	})
}

func TestNewCustomLogger(t *testing.T) {
	custom1 := New(
		WithLevel("debug"),
		WithOutputType(OutputTypeText),
		WithTimeZone(OutputTimeZoneLocal),
		WithMaxParts(2),
	)
	if custom1 == nil {
		t.Fatal("expected non-nil custom logger from New()")
	}

	custom1.Info("test custom1", slog.String("key", "val"))
}

func TestLoggerInit(t *testing.T) {
	defaultLogger.Store(nil)

	lLazy := GetLogger()
	if lLazy == nil {
		t.Fatal("expected GetLogger() to lazily initialize default logger")
	}

	LoggerInit(
		WithLevel("debug"),
		WithOutputType(OutputTypeText),
		WithTimeZone(OutputTimeZoneLocal),
		WithMaxParts(2),
	)

	lInit := GetLogger()
	if lInit == nil {
		t.Fatal("expected GetLogger() to return non-nil after LoggerInit")
	}
}

func TestPackageLevelLogging(t *testing.T) {
	LoggerInit(WithLevel("debug"))
	ctx := context.Background()

	Info("info log", slog.String("key", "val"))
	Debug("debug log", slog.String("key", "val"))
	Warn("warn log", slog.String("key", "val"))
	Error("error log", slog.String("key", "val"))

	Log(ctx, slog.LevelInfo, "log msg", slog.String("key", "val"))
	LogAttrs(ctx, slog.LevelInfo, "log attrs msg", slog.String("key", "val"))

	wLogger := With(slog.String("component", "test"))
	if wLogger == nil {
		t.Error("expected non-nil logger from With()")
	}

	gLogger := WithGroup("group1")
	if gLogger == nil {
		t.Error("expected non-nil logger from WithGroup()")
	}

	// Test disabled log level branch
	LoggerInit(WithLevel("error"))
	Debug("disabled debug log")
	Log(ctx, slog.LevelDebug, "disabled log")
	LogAttrs(ctx, slog.LevelDebug, "disabled log attrs")
}

func TestDefaultAndCustomLoggersTogether(t *testing.T) {
	// Initialize default logger with TEXT output and INFO level
	LoggerInit(
		WithLevel(LevenInfo),
		WithOutputType(OutputTypeText),
		WithTimeZone(OutputTimeZoneLocal),
		WithMaxParts(2),
	)

	defaultPtrBefore := GetLogger()

	// Create custom logger with JSON output and DEBUG level
	customJSONLogger := New(
		WithLevel(LevelDebug),
		WithOutputType(OutputTypeJSON),
		WithTimeZone(OutputTimeZoneUTC),
		WithMaxParts(3),
	)

	// Verify creating custom logger did NOT change default logger
	if GetLogger() != defaultPtrBefore {
		t.Error("expected default logger instance to remain unchanged after calling New()")
	}

	ctx := context.Background()

	// Default logger at INFO level: Debug should be disabled
	if GetLogger().Enabled(ctx, slog.LevelDebug) {
		t.Error("expected default logger to have DEBUG disabled")
	}

	// Custom logger at DEBUG level: Debug should be enabled
	if !customJSONLogger.Enabled(ctx, slog.LevelDebug) {
		t.Error("expected custom logger to have DEBUG enabled")
	}

	// Log with default logger
	Info("default logger info event", slog.String("env", "prod"))

	// Log with custom logger
	customJSONLogger.Debug("custom logger debug event", slog.String("component", "kibana"))
}
