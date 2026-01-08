# Trace Context Example

This example demonstrates trace context functionality for request tracking using goroutine-local storage.

## Run

```bash
cd _examples/trace
go run main.go
```

## Features Demonstrated

| # | Feature | Function |
| - | ------- | -------- |
| 1 | Run with error return | `RunWithTrace()` |
| 2 | Run without return value | `RunWithTraceVoid()` |
| 3 | Get request ID | `TraceRequestID()` |
| 4 | Get correlation ID | `TraceCorrelationID()` |

## Sample Output

```text
=== Trace Context Examples ===

1. RunWithTrace (with error return)
------------------------------------
Inside trace context:
  Request ID: req-abc123
  Correlation ID: corr-xyz789
{"level":"info","time":"...","message":"Processing request","action":"user_login","user_id":"user-123","request_id":"req-abc123","correlation_id":"corr-xyz789"}

2. RunWithTraceVoid (no error return)
--------------------------------------
Inside trace context (void):
  Request ID: req-def456
  Correlation ID: corr-uvw321
{"level":"info","time":"...","message":"Background task started","task":"cleanup","request_id":"req-def456","correlation_id":"corr-uvw321"}

3. Trace Through Nested Operations
-----------------------------------
Processing request: reqID=req-nested-001, corrID=corr-nested-001
{"level":"info","time":"...","message":"Request received",...}
{"level":"info","time":"...","message":"Request completed",...}

4. Trace Outside Context
------------------------
Request ID outside context: ''
Correlation ID outside context: ''

=== End of Examples ===
```

## Use Cases

- **HTTP Request Tracking**: Pass request ID through middleware
- **Distributed Tracing**: Correlation ID across microservices
- **Background Jobs**: Track async task execution
- **Error Debugging**: Trace request flow for debugging
