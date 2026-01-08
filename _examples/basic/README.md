# Basic Example

This example demonstrates basic usage of the `xlogger` package.

## Run

```bash
cd _examples/basic
go run main.go
```

## Features Demonstrated

| # | Feature | Description |
| - | ------- | ----------- |
| 1 | Default Config | Using `DefaultLoggerConfig()` with JSON format |
| 2 | Custom Config | Using `NewLoggerConfig()` with text format and debug level |
| 3 | Contextual Fields | Using `With()` to add persistent fields |

## Sample Output

```text
=== xlogger Basic Examples ===

1. Default Config (JSON format, INFO level)
--------------------------------------------
{"level":"info","time":"2025-01-08T14:30:45.123+0700","caller":"basic/main.go:24","message":"Application started with default config"}
{"level":"info","time":"2025-01-08T14:30:45.123+0700","caller":"basic/main.go:25","message":"User logged in","user_id":"12345"}

2. Custom Config (Text format, DEBUG level)
--------------------------------------------
2025-01-08 14:30:45 +07:00 🧪 DEBUG basic/main.go:42 Debug message {"key": "value"}
2025-01-08 14:30:45 +07:00 📢 INFO  basic/main.go:43 Info message {"count": 42}
2025-01-08 14:30:45 +07:00 🚧 WARN  basic/main.go:44 Warning message {"active": true}
2025-01-08 14:30:45 +07:00 ❌ ERROR basic/main.go:45 Error message {"error": "something went wrong"}

3. Logger with Contextual Fields
---------------------------------
2025-01-08 14:30:45 +07:00 📢 INFO  basic/main.go:55 Request received {"service": "api-gateway", "version": "1.0.0"}
2025-01-08 14:30:45 +07:00 📢 INFO  basic/main.go:56 Request processed {"service": "api-gateway", "version": "1.0.0", "duration_ms": 150}

=== End of Basic Examples ===
```
