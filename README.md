# go-xlogger

A lightweight Go logging SDK with JSON/text output, configurable log levels, and seamless integration with Zap for high-performance structured logging.

## Installation

```bash
go get github.com/hotfixfirst/go-xlogger
```

Or with a specific version:

```bash
go get github.com/hotfixfirst/go-xlogger@v1.0.0
```

## Quick Start

```go
package main

import (
    "github.com/hotfixfirst/go-xlogger"
    "go.uber.org/zap/zapcore"
)

func main() {
    // Using default config (JSON format, INFO level)
    logger, err := xlogger.NewZapLogger(xlogger.DefaultLoggerConfig())
    if err != nil {
        panic(err)
    }
    defer logger.Sync()

    logger.Info("Hello, World!")
    logger.Info("User action", xlogger.String("user_id", "12345"))
}
```

## Features

| Feature | Description |
| ------- | ----------- |
| Multiple Formats | JSON and Text output formats |
| Log Levels | Debug, Info, Warn, Error, Panic, Fatal |
| Structured Logging | Type-safe field constructors |
| Trace Context | Request/Correlation ID tracking |
| GORM Integration | Database query logging |
| Fx Integration | Uber Fx dependency injection support |

## Packages

| Package | Description | Documentation |
| ------- | ----------- | ------------- |
| [Config](#config) | Logger configuration | [Examples](./_examples/basic/) |
| [Logger](#logger) | Core logging interface | [Examples](./_examples/basic/) |
| [Trace](#trace-context) | Request tracking | [Examples](./_examples/trace/) |

## Config

### LogFormat

```go
// Available formats
xlogger.FormatJSON  // JSON output (default)
xlogger.FormatText  // Human-readable text output
```

### Config Struct

```go
type Config struct {
    Level             zapcore.Level // Minimum log level
    Format            LogFormat     // Log format: FormatJSON or FormatText
    Development       bool          // Development mode (pretty printing)
    DisableCaller     bool          // Disable caller information
    DisableStacktrace bool          // Disable stacktrace in errors
    TimeFormat        string        // Time format (empty for default)
    CallerSkip        int           // Number of caller frames to skip
}
```

### Config Functions

| Function | Description |
| -------- | ----------- |
| `DefaultLoggerConfig()` | Returns default config (INFO, JSON) |
| `NewLoggerConfig(opts...)` | Creates config with functional options |

### Option Functions

| Function | Description |
| -------- | ----------- |
| `WithLevel(level)` | Set log level |
| `WithFormat(format)` | Set output format (JSON/Text) |
| `WithDevelopment(bool)` | Enable development mode |
| `WithDisableCaller(bool)` | Disable caller info |
| `WithDisableStacktrace(bool)` | Disable stacktrace |
| `WithTimeFormat(format)` | Set time format |
| `WithCallerSkip(skip)` | Set caller skip frames |

### Config Example

```go
// Default config
cfg := xlogger.DefaultLoggerConfig()

// Custom config with functional options
cfg := xlogger.NewLoggerConfig(
    xlogger.WithLevel(zapcore.DebugLevel),
    xlogger.WithFormat(xlogger.FormatText),
    xlogger.WithDevelopment(true),
)
```

## Logger

### Creating Logger

```go
logger, err := xlogger.NewZapLogger(cfg)
if err != nil {
    panic(err)
}
defer logger.Sync()
```

### Logging Methods

```go
logger.Debug("Debug message", xlogger.String("key", "value"))
logger.Info("Info message", xlogger.Int("count", 42))
logger.Warn("Warning message", xlogger.Bool("active", true))
logger.Error("Error occurred", xlogger.Error(err))
```

### Field Constructors

| Function | Type | Example |
| -------- | ---- | ------- |
| `String(key, value)` | string | `xlogger.String("name", "John")` |
| `Int(key, value)` | int | `xlogger.Int("count", 42)` |
| `Int64(key, value)` | int64 | `xlogger.Int64("id", 123456)` |
| `Float64(key, value)` | float64 | `xlogger.Float64("price", 99.99)` |
| `Bool(key, value)` | bool | `xlogger.Bool("active", true)` |
| `Error(err)` | error | `xlogger.Error(err)` |
| `Duration(key, value)` | time.Duration | `xlogger.Duration("elapsed", time.Second)` |
| `Time(key, value)` | time.Time | `xlogger.Time("created", time.Now())` |
| `Any(key, value)` | any | `xlogger.Any("data", obj)` |

### Contextual Logger

```go
// Add persistent fields
contextLogger := logger.With(
    xlogger.String("service", "api"),
    xlogger.String("version", "1.0.0"),
)

contextLogger.Info("Request received")  // Includes service and version
```

## Trace Context

Track requests across function calls using goroutine-local storage.

### Trace Functions

| Function | Description |
| -------- | ----------- |
| `RunWithTrace(requestID, correlationID, fn)` | Execute function with trace context |
| `RunWithTraceVoid(requestID, correlationID, fn)` | Execute void function with trace context |
| `TraceRequestID()` | Get current request ID |
| `TraceCorrelationID()` | Get current correlation ID |

### Trace Example

```go
err := xlogger.RunWithTrace("req-123", "corr-456", func() error {
    // Trace IDs are automatically added to logs
    logger.Info("Processing request")

    // Access trace IDs anywhere in this scope
    fmt.Println(xlogger.TraceRequestID())      // "req-123"
    fmt.Println(xlogger.TraceCorrelationID())  // "corr-456"

    return nil
})
```

## GORM Integration

```go
gormLogger := logger.ForGORM()

db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
    Logger: gormLogger,
})
```

## Fx Integration

```go
import "go.uber.org/fx"

app := fx.New(
    fx.Provide(
        xlogger.DefaultLoggerConfig,
        xlogger.NewZapLogger,
    ),
    fx.Invoke(func(logger xlogger.Logger) {
        logger.Info("Application started")
    }),
)
```

## Examples

See the [_examples](./_examples/) directory for runnable examples.

| Example | Description |
| ------- | ----------- |
| [basic](./_examples/basic/) | Basic logger usage |
| [trace](./_examples/trace/) | Trace context for request tracking |

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
