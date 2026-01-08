package xlogger

import (
	"github.com/jtolds/gls"
)

const (
	traceRequestIDKey     = "logger-trace-request-id"
	traceCorrelationIDKey = "logger-trace-correlation-id"
)

var traceContextManager = gls.NewContextManager()

// RunWithTrace executes fn within a goroutine-local context that stores
// request and correlation identifiers for later retrieval.
func RunWithTrace(requestID, correlationID string, fn func() error) error {
	if fn == nil {
		return nil
	}

	var result error
	traceContextManager.SetValues(gls.Values{
		traceRequestIDKey:     requestID,
		traceCorrelationIDKey: correlationID,
	}, func() {
		result = fn()
	})
	return result
}

// RunWithTraceVoid executes fn within the trace context when no error
// propagation is required.
func RunWithTraceVoid(requestID, correlationID string, fn func()) {
	if fn == nil {
		return
	}

	traceContextManager.SetValues(gls.Values{
		traceRequestIDKey:     requestID,
		traceCorrelationIDKey: correlationID,
	}, fn)
}

// TraceRequestID returns the goroutine-local request identifier.
func TraceRequestID() string {
	return getTraceValue(traceRequestIDKey)
}

// TraceCorrelationID returns the goroutine-local correlation identifier.
func TraceCorrelationID() string {
	return getTraceValue(traceCorrelationIDKey)
}

func getTraceValue(key string) string {
	value, ok := traceContextManager.GetValue(key)
	if !ok || value == nil {
		return ""
	}

	if str, ok := value.(string); ok {
		return str
	}
	return ""
}
