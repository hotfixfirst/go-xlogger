package xlogger

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunWithTrace(t *testing.T) {
	t.Run("should execute function with trace context", func(t *testing.T) {
		requestID := "req-123"
		correlationID := "corr-456"

		executed := false
		err := RunWithTrace(requestID, correlationID, func() error {
			executed = true
			// Verify trace IDs are accessible within the context
			assert.Equal(t, requestID, TraceRequestID())
			assert.Equal(t, correlationID, TraceCorrelationID())
			return nil
		})

		assert.NoError(t, err)
		assert.True(t, executed)
	})

	t.Run("should return error from function", func(t *testing.T) {
		expectedErr := fmt.Errorf("test error")

		err := RunWithTrace("req-1", "corr-1", func() error {
			return expectedErr
		})

		assert.Equal(t, expectedErr, err)
	})

	t.Run("should handle nil function", func(t *testing.T) {
		err := RunWithTrace("req-1", "corr-1", nil)
		assert.NoError(t, err)
	})

	t.Run("should not leak trace context outside RunWithTrace", func(t *testing.T) {
		// Before RunWithTrace
		assert.Empty(t, TraceRequestID())
		assert.Empty(t, TraceCorrelationID())

		err := RunWithTrace("req-leak", "corr-leak", func() error {
			// Inside RunWithTrace
			assert.Equal(t, "req-leak", TraceRequestID())
			assert.Equal(t, "corr-leak", TraceCorrelationID())
			return nil
		})

		assert.NoError(t, err)

		// After RunWithTrace - should be empty again
		assert.Empty(t, TraceRequestID())
		assert.Empty(t, TraceCorrelationID())
	})

	t.Run("should handle empty trace IDs", func(t *testing.T) {
		err := RunWithTrace("", "", func() error {
			assert.Empty(t, TraceRequestID())
			assert.Empty(t, TraceCorrelationID())
			return nil
		})

		assert.NoError(t, err)
	})
}

func TestRunWithTraceVoid(t *testing.T) {
	t.Run("should execute function with trace context", func(t *testing.T) {
		requestID := "req-void-123"
		correlationID := "corr-void-456"

		executed := false
		RunWithTraceVoid(requestID, correlationID, func() {
			executed = true
			assert.Equal(t, requestID, TraceRequestID())
			assert.Equal(t, correlationID, TraceCorrelationID())
		})

		assert.True(t, executed)
	})

	t.Run("should handle nil function", func(t *testing.T) {
		// Should not panic
		RunWithTraceVoid("req-1", "corr-1", nil)
	})

	t.Run("should not leak trace context", func(t *testing.T) {
		assert.Empty(t, TraceRequestID())

		RunWithTraceVoid("req-void", "corr-void", func() {
			assert.Equal(t, "req-void", TraceRequestID())
		})

		assert.Empty(t, TraceRequestID())
	})
}

func TestTraceRequestID(t *testing.T) {
	t.Run("should return empty string when no trace context", func(t *testing.T) {
		requestID := TraceRequestID()
		assert.Empty(t, requestID)
	})

	t.Run("should return request ID within trace context", func(t *testing.T) {
		expected := "req-test-123"

		err := RunWithTrace(expected, "corr-1", func() error {
			actual := TraceRequestID()
			assert.Equal(t, expected, actual)
			return nil
		})

		assert.NoError(t, err)
	})
}

func TestTraceCorrelationID(t *testing.T) {
	t.Run("should return empty string when no trace context", func(t *testing.T) {
		correlationID := TraceCorrelationID()
		assert.Empty(t, correlationID)
	})

	t.Run("should return correlation ID within trace context", func(t *testing.T) {
		expected := "corr-test-456"

		err := RunWithTrace("req-1", expected, func() error {
			actual := TraceCorrelationID()
			assert.Equal(t, expected, actual)
			return nil
		})

		assert.NoError(t, err)
	})
}

// TestConcurrentTraceIsolation tests that trace contexts are properly isolated between goroutines
func TestConcurrentTraceIsolation(t *testing.T) {
	t.Run("should isolate trace context between concurrent goroutines", func(t *testing.T) {
		const numGoroutines = 100
		var wg sync.WaitGroup
		results := make(chan struct {
			expected  string
			actual    string
			goroutine int
		}, numGoroutines)

		// Simulate concurrent HTTP requests with different trace IDs
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()

				requestID := fmt.Sprintf("req-%03d", id)
				correlationID := fmt.Sprintf("corr-%03d", id)

				err := RunWithTrace(requestID, correlationID, func() error {
					// Simulate some processing time
					time.Sleep(time.Millisecond * time.Duration(rand.Intn(10)))

					// Get trace IDs from goroutine-local storage
					gotRequestID := TraceRequestID()
					gotCorrelationID := TraceCorrelationID()

					// Store results for verification
					results <- struct {
						expected  string
						actual    string
						goroutine int
					}{
						expected:  requestID,
						actual:    gotRequestID,
						goroutine: id,
					}

					// Verify within the goroutine as well
					assert.Equal(t, requestID, gotRequestID,
						"Request ID mismatch in goroutine %d", id)
					assert.Equal(t, correlationID, gotCorrelationID,
						"Correlation ID mismatch in goroutine %d", id)

					return nil
				})

				assert.NoError(t, err)
			}(i)
		}

		// Wait for all goroutines to complete
		wg.Wait()
		close(results)

		// Verify no cross-contamination occurred
		for result := range results {
			assert.Equal(t, result.expected, result.actual,
				"Request ID should not be mixed between goroutines (goroutine %d)",
				result.goroutine)
		}
	})

	t.Run("should handle nested concurrent operations", func(t *testing.T) {
		const numParent = 10
		const numChild = 5
		var wg sync.WaitGroup
		errors := make(chan error, numParent*numChild)

		for i := 0; i < numParent; i++ {
			wg.Add(1)
			go func(parentID int) {
				defer wg.Done()

				parentRequestID := fmt.Sprintf("parent-req-%d", parentID)
				parentCorrelationID := fmt.Sprintf("parent-corr-%d", parentID)

				err := RunWithTrace(parentRequestID, parentCorrelationID, func() error {
					// Verify parent context
					if TraceRequestID() != parentRequestID {
						errors <- fmt.Errorf("parent %d: request ID mismatch", parentID)
					}

					// Spawn child goroutines
					var childWg sync.WaitGroup
					for j := 0; j < numChild; j++ {
						childWg.Add(1)
						go func(childID int) {
							defer childWg.Done()

							childRequestID := fmt.Sprintf("child-req-%d-%d", parentID, childID)
							childCorrelationID := fmt.Sprintf("child-corr-%d-%d", parentID, childID)

							err := RunWithTrace(childRequestID, childCorrelationID, func() error {
								time.Sleep(time.Millisecond * time.Duration(rand.Intn(5)))

								// Verify child has its own context
								if TraceRequestID() != childRequestID {
									return fmt.Errorf("child %d-%d: request ID mismatch", parentID, childID)
								}
								if TraceCorrelationID() != childCorrelationID {
									return fmt.Errorf("child %d-%d: correlation ID mismatch", parentID, childID)
								}
								return nil
							})

							if err != nil {
								errors <- err
							}
						}(j)
					}

					childWg.Wait()

					// Verify parent context still intact after child goroutines
					if TraceRequestID() != parentRequestID {
						return fmt.Errorf("parent %d: request ID changed after children", parentID)
					}

					return nil
				})

				if err != nil {
					errors <- err
				}
			}(i)
		}

		wg.Wait()
		close(errors)

		// Check for any errors
		for err := range errors {
			t.Error(err)
		}
	})

	t.Run("should handle high concurrency stress test", func(t *testing.T) {
		const numGoroutines = 1000
		const numOperations = 10
		var wg sync.WaitGroup
		errorCount := 0
		var errorMutex sync.Mutex

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()

				for op := 0; op < numOperations; op++ {
					requestID := fmt.Sprintf("stress-req-%d-%d", id, op)
					correlationID := fmt.Sprintf("stress-corr-%d-%d", id, op)

					err := RunWithTrace(requestID, correlationID, func() error {
						// Random processing time
						time.Sleep(time.Microsecond * time.Duration(rand.Intn(100)))

						// Verify trace IDs
						if TraceRequestID() != requestID {
							return fmt.Errorf("request ID mismatch: expected %s, got %s",
								requestID, TraceRequestID())
						}
						if TraceCorrelationID() != correlationID {
							return fmt.Errorf("correlation ID mismatch: expected %s, got %s",
								correlationID, TraceCorrelationID())
						}
						return nil
					})

					if err != nil {
						errorMutex.Lock()
						errorCount++
						errorMutex.Unlock()
					}
				}
			}(i)
		}

		wg.Wait()

		assert.Equal(t, 0, errorCount,
			"Expected no errors in stress test, but got %d errors", errorCount)
	})
}

// TestTraceContextPropagation tests trace context propagation patterns
func TestTraceContextPropagation(t *testing.T) {
	t.Run("should propagate trace context through function calls", func(t *testing.T) {
		requestID := "req-propagate-123"
		correlationID := "corr-propagate-456"

		level1Called := false
		level2Called := false
		level3Called := false

		level3 := func() {
			level3Called = true
			assert.Equal(t, requestID, TraceRequestID())
			assert.Equal(t, correlationID, TraceCorrelationID())
		}

		level2 := func() {
			level2Called = true
			assert.Equal(t, requestID, TraceRequestID())
			level3()
		}

		level1 := func() error {
			level1Called = true
			assert.Equal(t, requestID, TraceRequestID())
			level2()
			return nil
		}

		err := RunWithTrace(requestID, correlationID, level1)

		assert.NoError(t, err)
		assert.True(t, level1Called)
		assert.True(t, level2Called)
		assert.True(t, level3Called)
	})

	t.Run("should handle context.Context passing with trace IDs", func(t *testing.T) {
		requestID := "req-ctx-123"
		correlationID := "corr-ctx-456"

		err := RunWithTrace(requestID, correlationID, func() error {
			// Create a context with trace IDs
			ctx := context.Background()

			// Add values to context for demonstration
			type ctxKey string
			ctx = context.WithValue(ctx, ctxKey("request_id"), requestID)

			// Pass through function chain
			processRequest := func(ctx context.Context) error {
				// Verify context value is passed correctly
				if ctx.Value(ctxKey("request_id")) != requestID {
					return fmt.Errorf("context value not passed correctly")
				}

				// Trace IDs should still be accessible via goroutine-local storage
				assert.Equal(t, requestID, TraceRequestID())
				assert.Equal(t, correlationID, TraceCorrelationID())

				// Simulate database operation
				simulateDBQuery := func(ctx context.Context) error {
					// Verify context is propagated
					if ctx.Value(ctxKey("request_id")) != requestID {
						return fmt.Errorf("context value not propagated to DB query")
					}

					// Trace IDs still accessible
					assert.Equal(t, requestID, TraceRequestID())
					return nil
				}

				return simulateDBQuery(ctx)
			}

			return processRequest(ctx)
		})

		assert.NoError(t, err)
	})
}

// TestTraceContextEdgeCases tests edge cases and error scenarios
func TestTraceContextEdgeCases(t *testing.T) {
	t.Run("should handle panic within trace context", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				assert.Equal(t, "intentional panic", r)
			}
		}()

		err := RunWithTrace("req-panic", "corr-panic", func() error {
			panic("intentional panic")
		})

		// Should not reach here
		t.Error("Expected panic, but function returned:", err)
	})

	t.Run("should handle very long trace IDs", func(t *testing.T) {
		longRequestID := string(make([]byte, 10000))
		longCorrelationID := string(make([]byte, 10000))

		err := RunWithTrace(longRequestID, longCorrelationID, func() error {
			assert.Equal(t, longRequestID, TraceRequestID())
			assert.Equal(t, longCorrelationID, TraceCorrelationID())
			return nil
		})

		assert.NoError(t, err)
	})

	t.Run("should handle special characters in trace IDs", func(t *testing.T) {
		specialRequestID := "req-!@#$%^&*()_+-=[]{}|;:',.<>?/~`"
		specialCorrelationID := "corr-你好世界🚀💻"

		err := RunWithTrace(specialRequestID, specialCorrelationID, func() error {
			assert.Equal(t, specialRequestID, TraceRequestID())
			assert.Equal(t, specialCorrelationID, TraceCorrelationID())
			return nil
		})

		assert.NoError(t, err)
	})

	t.Run("should handle rapid context switching", func(t *testing.T) {
		const iterations = 1000
		var wg sync.WaitGroup

		for i := 0; i < iterations; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()

				requestID := fmt.Sprintf("rapid-req-%d", id)
				correlationID := fmt.Sprintf("rapid-corr-%d", id)

				err := RunWithTrace(requestID, correlationID, func() error {
					// Immediately check trace IDs
					assert.Equal(t, requestID, TraceRequestID())
					assert.Equal(t, correlationID, TraceCorrelationID())
					return nil
				})

				assert.NoError(t, err)
			}(i)
		}

		wg.Wait()
	})
}

// BenchmarkTraceContextOperations benchmarks trace context operations
func BenchmarkTraceContextOperations(b *testing.B) {
	b.Run("RunWithTrace", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = RunWithTrace("req-bench", "corr-bench", func() error {
				return nil
			})
		}
	})

	b.Run("TraceRequestID", func(b *testing.B) {
		_ = RunWithTrace("req-bench", "corr-bench", func() error {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = TraceRequestID()
			}
			return nil
		})
	})

	b.Run("TraceCorrelationID", func(b *testing.B) {
		_ = RunWithTrace("req-bench", "corr-bench", func() error {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = TraceCorrelationID()
			}
			return nil
		})
	})

	b.Run("ConcurrentAccess", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				requestID := fmt.Sprintf("req-%d", i)
				correlationID := fmt.Sprintf("corr-%d", i)

				_ = RunWithTrace(requestID, correlationID, func() error {
					_ = TraceRequestID()
					_ = TraceCorrelationID()
					return nil
				})

				i++
			}
		})
	})
}

// TestGetTraceValue tests the internal getTraceValue function
func TestGetTraceValue(t *testing.T) {
	t.Run("should return empty string for non-existent key", func(t *testing.T) {
		value := getTraceValue("non-existent-key")
		assert.Empty(t, value)
	})

	t.Run("should return empty string for nil value", func(t *testing.T) {
		// This test verifies the nil check in getTraceValue
		err := RunWithTrace("", "", func() error {
			value := getTraceValue(traceRequestIDKey)
			assert.Empty(t, value)
			return nil
		})
		assert.NoError(t, err)
	})

	t.Run("should handle type assertion failure gracefully", func(t *testing.T) {
		// getTraceValue should return empty string if type assertion fails
		// This is handled by the function's type assertion check
		value := getTraceValue("invalid-key")
		require.Equal(t, "", value)
	})
}
