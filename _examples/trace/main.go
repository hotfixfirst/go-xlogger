// Package main demonstrates trace context functionality in xlogger.
package main

import (
	"fmt"

	"github.com/hotfixfirst/go-xlogger"
)

func main() {
	fmt.Println("=== Trace Context Examples ===")
	fmt.Println()

	// Create logger with default config
	config := xlogger.DefaultLoggerConfig()
	logger, err := xlogger.NewZapLogger(config)
	if err != nil {
		panic(err)
	}

	// Example 1: RunWithTrace with error propagation
	fmt.Println("1. RunWithTrace (with error return)")
	fmt.Println("------------------------------------")

	requestID := "req-abc123"
	correlationID := "corr-xyz789"

	err = xlogger.RunWithTrace(requestID, correlationID, func() error {
		// Inside this closure, TraceRequestID() and TraceCorrelationID() work
		fmt.Printf("Inside trace context:\n")
		fmt.Printf("  Request ID: %s\n", xlogger.TraceRequestID())
		fmt.Printf("  Correlation ID: %s\n", xlogger.TraceCorrelationID())

		// Log with trace context automatically included
		logger.Info("Processing request",
			xlogger.String("action", "user_login"),
			xlogger.String("user_id", "user-123"),
		)

		return nil
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	fmt.Println()

	// Example 2: RunWithTraceVoid (no error return)
	fmt.Println("2. RunWithTraceVoid (no error return)")
	fmt.Println("--------------------------------------")

	xlogger.RunWithTraceVoid("req-def456", "corr-uvw321", func() {
		fmt.Printf("Inside trace context (void):\n")
		fmt.Printf("  Request ID: %s\n", xlogger.TraceRequestID())
		fmt.Printf("  Correlation ID: %s\n", xlogger.TraceCorrelationID())

		logger.Info("Background task started",
			xlogger.String("task", "cleanup"),
		)
	})
	fmt.Println()

	// Example 3: Nested operations with trace
	fmt.Println("3. Trace Through Nested Operations")
	fmt.Println("-----------------------------------")

	xlogger.RunWithTraceVoid("req-nested-001", "corr-nested-001", func() {
		handleRequest(logger)
	})
	fmt.Println()

	// Example 4: Trace outside context (returns empty)
	fmt.Println("4. Trace Outside Context")
	fmt.Println("------------------------")

	fmt.Printf("Request ID outside context: '%s'\n", xlogger.TraceRequestID())
	fmt.Printf("Correlation ID outside context: '%s'\n", xlogger.TraceCorrelationID())
	fmt.Println()

	fmt.Println("=== End of Examples ===")
}

// handleRequest simulates processing an API request with trace context
func handleRequest(logger xlogger.Logger) {
	reqID := xlogger.TraceRequestID()
	corrID := xlogger.TraceCorrelationID()
	fmt.Printf("Processing request: reqID=%s, corrID=%s\n", reqID, corrID)

	// Log entry point
	logger.Info("Request received",
		xlogger.String("endpoint", "/api/users"),
		xlogger.String("method", "GET"),
	)

	// Simulate calling a service
	callService(logger)

	// Log completion
	logger.Info("Request completed",
		xlogger.Int("status", 200),
		xlogger.Int("duration_ms", 45),
	)
}

// callService simulates calling an internal service
func callService(logger xlogger.Logger) {
	// The trace IDs are automatically available via goroutine-local storage
	logger.Debug("Calling user service",
		xlogger.String("service", "user-service"),
		xlogger.String("request_id", xlogger.TraceRequestID()),
	)

	// Simulate database call
	logger.Debug("Database query executed",
		xlogger.String("query", "SELECT * FROM users"),
		xlogger.Int("rows", 10),
	)
}
