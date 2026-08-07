# Backend Tools

A lightweight, high-performance set of backend utilities for Go applications, providing structured logging built on Go's `log/slog` and generic numeric type conversion utilities.

---

## Table of Contents
- [Packages](#packages)
  - [`log` - Structured Logging Package](#log---structured-logging-package)
  - [`tools` - Generic Number Utilities](#tools---generic-number-utilities)
- [Installation](#installation)
- [Testing](#testing)

---

## Packages

### `log` - Structured Logging Package

The `log` package provides a structured, context-aware logger built on Go's standard `log/slog`.

#### Features
- **Clean Package-Level Functions**: Log directly via `logger.Info()`, `logger.Debug()`, `logger.Warn()`, `logger.Error()`, etc.
- **Functional Options**: Configure level, output format (JSON/Text), time zone (UTC/Local), and source path depth.
- **Accurate Caller Depth**: Package-level log wrappers accurately report caller source line numbers (`main.go:20` instead of wrapper file lines).
- **Customizable Source Formatting**: Formats source file paths relative to the project working directory without leading slashes and caps path depth to configurable limits (e.g. `log/logger.go:127`).
- **High Performance**: Optimized with cached working directory and zero-allocation path trimming.

#### Usage Example

```go
package main

import (
	"context"
	"log/slog"

	"backend-tools/log"
)

func main() {
	// Initialize logger (optional; auto-initializes with defaults if omitted)
	logger.LoggerInit(
		logger.WithLevel("debug"),
		logger.WithOutputType(logger.OutputTypeJSON), // OutputTypeJSON or OutputTypeText
		logger.WithTimeZone(logger.OutputTimeZoneUTC),// OutputTimeZoneUTC or OutputTimeZoneLocal
		logger.WithMaxParts(3),                        // Subdirectory depth limit for source caller
	)

	// Direct package-level logging
	logger.Info("server starting", slog.Int("port", 8080))
	logger.Debug("debugging request", slog.String("path", "/api/v1/health"))
	logger.Warn("high memory usage", slog.Float64("mem_pct", 85.5))
	logger.Error("database connection failed", slog.String("db", "postgres"))

	// Context-aware logging
	ctx := context.Background()
	logger.Log(ctx, slog.LevelInfo, "context log", slog.String("trace_id", "abc-123"))
	logger.LogAttrs(ctx, slog.LevelInfo, "attr log", slog.String("user", "alice"))

	// Sub-loggers with pre-attached attributes or groups
	componentLogger := logger.With(slog.String("component", "auth"))
	componentLogger.Info("user logged in")
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

	"backend-tools/tools"
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
	"backend-tools/log"
	"backend-tools/tools"
)
```

---

## Testing

Run all unit tests across the repository:

```bash
go test -v -cover ./...
```
