// Package main demonstrates the basic usage of the xlogger package.
package main

import (
	"fmt"

	"go.uber.org/zap/zapcore"

	"github.com/hotfixfirst/go-xlogger"
)

func main() {
	fmt.Println("=== xlogger Basic Examples ===")
	fmt.Println()

	// 1. Using default config
	fmt.Println("1. Default Config (JSON format, INFO level)")
	fmt.Println("--------------------------------------------")
	defaultLogger, err := xlogger.NewZapLogger(xlogger.DefaultLoggerConfig())
	if err != nil {
		panic(err)
	}
	defer defaultLogger.Sync()

	defaultLogger.Info("Application started with default config")
	defaultLogger.Info("User logged in", xlogger.String("user_id", "12345"))

	// 2. Using custom config with functional options
	fmt.Println()
	fmt.Println("2. Custom Config (Text format, DEBUG level)")
	fmt.Println("--------------------------------------------")
	cfg := xlogger.NewLoggerConfig(
		xlogger.WithLevel(zapcore.DebugLevel),
		xlogger.WithFormat(xlogger.FormatText),
		xlogger.WithDevelopment(true),
	)

	textLogger, err := xlogger.NewZapLogger(cfg)
	if err != nil {
		panic(err)
	}
	defer textLogger.Sync()

	textLogger.Debug("Debug message", xlogger.String("key", "value"))
	textLogger.Info("Info message", xlogger.Int("count", 42))
	textLogger.Warn("Warning message", xlogger.Bool("active", true))
	textLogger.Error("Error message", xlogger.String("error", "something went wrong"))

	// 3. Using logger with fields
	fmt.Println()
	fmt.Println("3. Logger with Contextual Fields")
	fmt.Println("---------------------------------")
	contextLogger := textLogger.With(
		xlogger.String("service", "api-gateway"),
		xlogger.String("version", "1.0.0"),
	)

	contextLogger.Info("Request received")
	contextLogger.Info("Request processed", xlogger.Int("duration_ms", 150))

	fmt.Println()
	fmt.Println("=== End of Basic Examples ===")
}
