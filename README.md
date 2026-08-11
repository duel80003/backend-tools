# Backend Tools

A lightweight, high-performance set of backend utilities for Go applications, providing structured logging built on Go's `log/slog` and generic numeric type conversion utilities.

---

## Table of Contents
- [Packages](#packages)
  - [`log` - Structured Logging Package](#log---structured-logging-package)
- [Installation](#installation)
- [Testing](#testing)

---

## Packages

### `log` - Structured Logging Package

The `log` package provides lock-free global logging functions (`logger.Info`, `logger.Debug`, `logger.Error`), context-aware logging (`logger.InfoContext`, `logger.Log`), Request ID HTTP middleware (`logger.RequestIDMiddleware`), as well as constructors for creating isolated custom loggers (`logger.New(...)`).

#### Features
- **Global Convenience Functions**: Log anywhere in your app via `logger.Info()`, `logger.Debug()`, `logger.Warn()`, `logger.Error()`.
- **Context-Aware Logging**: Log with context via `logger.InfoContext(ctx, ...)`, `logger.DebugContext(ctx, ...)`, `logger.WarnContext(ctx, ...)`, `logger.ErrorContext(ctx, ...)`.
- **Request ID Life Cycle & Middleware**: HTTP middleware (`logger.RequestIDMiddleware`) that extracts or generates a unique Request ID (UUID v4), injects it into `context.Context`, and sets `X-Request-ID` in HTTP response headers.
- **Automatic Request ID Log Injection**: `log/slog` handlers automatically extract `request_id` from `context.Context` and append it to all log entries (including database query logs via GORM logger).
- **Customizable Request ID Extractor**: Configure custom context keys or framework extractors via `logger.WithRequestIDExtractor(...)`.
- **Lock-Free Atomic Global State**: Global logger reads use lock-free `atomic.Pointer` for maximum performance.
- **Isolated Custom Loggers**: Constructing custom loggers via `logger.New(...)` (e.g., JSON loggers for Kibana) **never touches or interferes with** the global default logger.
- **Accurate Caller Depth**: Package-level log wrappers accurately report caller source line numbers (`main.go:20` instead of wrapper file lines).
- **Customizable Source Formatting**: Formats source file paths relative to the project working directory without leading slashes and caps path depth to configurable limits (e.g. `log/logger.go:127`).

#### Basic Usage Example

```go
package main

import (
	"context"
	"log/slog"

	logger "github.com/duel80003/backend-tools/log"
)

func main() {
	// 1. Initialize global default logger (for application-wide console logging)
	logger.LoggerInit(
		logger.WithLevel("debug"),
		logger.WithOutputType(logger.OutputTypeText),
		logger.WithTimeZone(logger.OutputTimeZoneLocal),
		logger.WithMaxParts(2),
	)

	// Direct global logging anywhere in your project
	logger.Info("server starting", slog.Int("port", 8080))
	logger.Debug("debugging request", slog.String("path", "/api/v1/health"))

	// 2. Create an independent custom logger (e.g. JSON format for Kibana)
	// Calling logger.New(...) NEVER modifies the global default logger!
	kibanaLogger := logger.New(
		logger.WithLevel("info"),
		logger.WithOutputType(logger.OutputTypeJSON),
		logger.WithTimeZone(logger.OutputTimeZoneUTC),
	)

	kibanaLogger.Info("user_payment",
		slog.String("user_id", "usr_123"),
		slog.Float64("amount", 99.95),
	)

	// Context-aware global logging
	ctx := context.Background()
	logger.Log(ctx, slog.LevelInfo, "context log", slog.String("trace_id", "abc-123"))
}
```

#### Request Life Cycle & Middleware Example

```go
package main

import (
	"net/http"

	logger "github.com/duel80003/backend-tools/log"
)

func main() {
	// Initialize global logger with JSON output
	logger.LoggerInit(logger.WithOutputType(logger.OutputTypeJSON))

	mux := http.NewServeMux()

	mux.HandleFunc("/api/users", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Output log automatically includes request_id from ctx:
		// {"time":"...","level":"INFO","msg":"fetching users","request_id":"c6f3796d-..."}
		logger.InfoContext(ctx, "fetching users")

		w.Write([]byte(`{"status":"ok"}`))
	})

	// Wrap server with RequestIDMiddleware
	http.ListenAndServe(":8080", logger.RequestIDMiddleware(mux))
}
```

---

### `tools` - Generic Number Utilities

The `tools` package provides generic utility functions for converting numeric types to strings and parsing strings to numbers using Go generics.

#### Features
- Supports all Go numeric primitives (`int`, `int8`, `int16`, `int32`, `int64`, `uint`, `uint8`, `uint16`, `uint32`, `uint64`, `float32`, `float64`, `complex64`, `complex128`).
- Supports custom defined underlying numeric types (`~int`, `~float64`, etc.).
- Safe functions with explicit `error` returns and safe `IgnoreError` fallbacks.

#### Usage Example

```go
package main

import (
	"fmt"

	"github.com/duel80003/backend-tools/tools"
)

func main() {
	// Convert any number to string
	str, err := tools.NumberToString(int64(9223372036854775807))
	fmt.Println(str) // "9223372036854775807"

	// Convert number to string ignoring errors
	strSafe := tools.NumberToStringIgnoreError(3.14159)
	fmt.Println(strSafe) // "3.14159"

	// Parse string to specific numeric type
	valInt, err := tools.StringToNumber[int]("42")
	fmt.Println(valInt) // 42

	valFloat, err := tools.StringToNumber[float64]("2.71828")
	fmt.Println(valFloat) // 2.71828

	// Parse string to number ignoring errors (returns 0 on failure)
	valSafe := tools.StringToNumberIgnoreError[uint16]("65535")
	fmt.Println(valSafe) // 65535
}
```

---

## Installation

Ensure Go 1.21+ is installed, then import packages into your project:

```go
import (
	"github.com/duel80003/backend-tools/log"
	"github.com/duel80003/backend-tools/tools"
)
```

---

## Testing

Run all unit tests across the repository:

```bash
go test -v -cover ./...
```

