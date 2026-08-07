package logger

import (
	"log/slog"
	"sync"
	"testing"
	"time"
)

func TestOptions(t *testing.T) {
	config := &LogConfig{}

	levelOpt := &LogLevelOption{Level: "debug"}
	levelOpt.Apply(config)
	if config.Level != "debug" {
		t.Errorf("expected Level 'debug', got '%s'", config.Level)
	}

	outputOpt := &OutputTypeOption{OutputType: OutputTypeText}
	outputOpt.Apply(config)
	if config.OutputType != OutputTypeText {
		t.Errorf("expected OutputType OutputTypeText (%d), got %d", OutputTypeText, config.OutputType)
	}

	tzOpt := &TimeZoneOption{TimeZone: OutputTimeZoneLocal}
	tzOpt.Apply(config)
	if config.TimeZone != OutputTimeZoneLocal {
		t.Errorf("expected TimeZone OutputTimeZoneLocal (%d), got %d", OutputTimeZoneLocal, config.TimeZone)
	}
}

func TestGetLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected slog.Level
	}{
		{"debug", slog.LevelDebug},
		{"DEBUG", slog.LevelDebug},
		{"info", slog.LevelInfo},
		{"INFO", slog.LevelInfo},
		{"warn", slog.LevelWarn},
		{"WARN", slog.LevelWarn},
		{"error", slog.LevelError},
		{"ERROR", slog.LevelError},
		{"unknown", slog.LevelInfo},
		{"", slog.LevelInfo},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := getLevel(tt.input)
			if got != tt.expected {
				t.Errorf("getLevel(%q) = %v; want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestGetReplaceAttrFunc(t *testing.T) {
	fnUTC := getReplaceAttrFunc(OutputTimeZoneUTC)
	if fnUTC == nil {
		t.Error("expected non-nil replaceAttr function for UTC")
	}

	fnLocal := getReplaceAttrFunc(OutputTimeZoneLocal)
	if fnLocal == nil {
		t.Error("expected non-nil replaceAttr function for Local")
	}

	fnDefault := getReplaceAttrFunc(OutputTimeZone(999))
	if fnDefault == nil {
		t.Error("expected non-nil replaceAttr function for default")
	}
}

func TestFormatSource(t *testing.T) {
	tests := []struct {
		name     string
		source   *slog.Source
		expected string
	}{
		{
			name:     "nil source",
			source:   nil,
			expected: "",
		},
		{
			name: "more than 3 subdirectories (5 levels)",
			source: &slog.Source{
				File: "/a/b/c/d/e/file.go",
				Line: 100,
			},
			expected: "/c/d/e/file.go:100",
		},
		{
			name: "exactly 3 subdirectories",
			source: &slog.Source{
				File: "/dir1/dir2/dir3/file.go",
				Line: 50,
			},
			expected: "/dir1/dir2/dir3/file.go:50",
		},
		{
			name: "1 subdirectory (like /log/logger_test.go:229)",
			source: &slog.Source{
				File: "/log/logger_test.go",
				Line: 229,
			},
			expected: "/log/logger_test.go:229",
		},
		{
			name: "only filename (file on root)",
			source: &slog.Source{
				File: "filename.go",
				Line: 12,
			},
			expected: "filename.go:12",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatSource(tt.source)
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

	// Test UTC replaceAttr
	resUTC := replaceAttrUTC(nil, timeAttr)
	expectedUTCStr := testTime.UTC().Format(customTimeLayout)
	if resUTC.Value.String() != expectedUTCStr {
		t.Errorf("replaceAttrUTC time = %s, want %s", resUTC.Value.String(), expectedUTCStr)
	}

	resUTCStr := replaceAttrUTC(nil, strAttr)
	if resUTCStr.Key != "key" || resUTCStr.Value.String() != "val" {
		t.Errorf("replaceAttrUTC non-time attr modified: %v", resUTCStr)
	}

	resUTCSrc := replaceAttrUTC(nil, srcAttr)
	expectedSrcStr := "/log/logger_test.go:229"
	if resUTCSrc.Value.String() != expectedSrcStr {
		t.Errorf("replaceAttrUTC source = %s, want %s", resUTCSrc.Value.String(), expectedSrcStr)
	}

	// Test Local replaceAttr
	resLocal := replaceAttrLocal(nil, timeAttr)
	expectedLocalStr := testTime.Local().Format(customTimeLayout)
	if resLocal.Value.String() != expectedLocalStr {
		t.Errorf("replaceAttrLocal time = %s, want %s", resLocal.Value.String(), expectedLocalStr)
	}

	resLocalStr := replaceAttrLocal(nil, strAttr)
	if resLocalStr.Key != "key" || resLocalStr.Value.String() != "val" {
		t.Errorf("replaceAttrLocal non-time attr modified: %v", resLocalStr)
	}

	resLocalSrc := replaceAttrLocal(nil, srcAttr)
	if resLocalSrc.Value.String() != expectedSrcStr {
		t.Errorf("replaceAttrLocal source = %s, want %s", resLocalSrc.Value.String(), expectedSrcStr)
	}
}

func TestGetHandler(t *testing.T) {
	t.Run("JSON Handler", func(t *testing.T) {
		cfg := &LogConfig{
			Level:      "info",
			OutputType: OutputTypeJSON,
			TimeZone:   OutputTimeZoneUTC,
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
}

func TestLoggerInit(t *testing.T) {
	// Reset loggerOnce and logger for isolated testing
	loggerOnce = sync.Once{}
	logger = nil

	LoggerInit(
		&LogLevelOption{Level: "debug"},
		&OutputTypeOption{OutputType: OutputTypeJSON},
		&TimeZoneOption{TimeZone: OutputTimeZoneUTC},
	)

	if logger == nil {
		t.Fatal("expected logger to be initialized, but got nil")
	}

	logger.Info("TestLoggerInit", slog.String("test", "test1"))

	if GetLogger() != logger {
		t.Error("expected GetLogger() to return the initialized logger instance")
	}

	// Calling LoggerInit again should not panic or re-initialize due to sync.Once
	prevLogger := logger
	LoggerInit(&LogLevelOption{Level: "error"})
	if logger != prevLogger {
		t.Error("expected logger instance to remain unchanged on second LoggerInit call")
	}
}
