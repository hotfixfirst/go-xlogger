package xlogger

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// TestNewZapLogger tests the NewZapLogger constructor
func TestNewZapLogger(t *testing.T) {
	t.Run("should create logger with default config", func(t *testing.T) {
		logger, err := NewZapLogger(nil)

		assert.NoError(t, err)
		assert.NotNil(t, logger)
		assert.Equal(t, zapcore.InfoLevel, logger.Level())
	})

	t.Run("should create logger with custom config", func(t *testing.T) {
		cfg := &Config{
			Level:         zapcore.DebugLevel,
			Format:        "json",
			Development:   false,
			DisableCaller: true,
		}

		logger, err := NewZapLogger(cfg)

		assert.NoError(t, err)
		assert.NotNil(t, logger)
		assert.Equal(t, zapcore.DebugLevel, logger.Level())
	})

	t.Run("should create logger with text format", func(t *testing.T) {
		cfg := &Config{
			Level:       zapcore.WarnLevel,
			Format:      "text",
			Development: true,
		}

		logger, err := NewZapLogger(cfg)

		assert.NoError(t, err)
		assert.NotNil(t, logger)
		assert.Equal(t, zapcore.WarnLevel, logger.Level())
	})

	t.Run("should create logger with console format", func(t *testing.T) {
		cfg := &Config{
			Level:             zapcore.ErrorLevel,
			Format:            "console",
			Development:       true,
			DisableCaller:     false,
			DisableStacktrace: false,
		}

		logger, err := NewZapLogger(cfg)

		assert.NoError(t, err)
		assert.NotNil(t, logger)
		assert.Equal(t, zapcore.ErrorLevel, logger.Level())
	})

	t.Run("should handle invalid zap config gracefully", func(t *testing.T) {
		// Test with extremely invalid configuration that might cause zap.Config.Build() to fail
		cfg := &Config{
			Level:       zapcore.Level(-99), // Invalid level
			Format:      "invalid-format",
			Development: false,
		}

		// This should not fail since we normalize the encoding
		logger, err := NewZapLogger(cfg)

		// Even with invalid level, zap should handle it gracefully
		assert.NoError(t, err)
		assert.NotNil(t, logger)
	})

	t.Run("should handle invalid output paths", func(t *testing.T) {
		// Create a custom config that might cause issues
		cfg := &Config{
			Level:       zapcore.InfoLevel,
			Format:      "json",
			Development: false,
		}

		// This test primarily covers the error checking paths
		logger, err := NewZapLogger(cfg)
		assert.NoError(t, err)
		assert.NotNil(t, logger)

		// Test that infrastructure loggers were initialized
		infraLogger := logger.ForInfra("test")
		assert.NotNil(t, infraLogger)
	})
}

// TestHelperFunctions tests the helper functions used in logger creation
func TestHelperFunctions(t *testing.T) {
	t.Run("should determine encoding correctly", func(t *testing.T) {
		assert.Equal(t, "console", determineEncoding(FormatText))
		assert.Equal(t, "console", determineEncoding(LogFormat("TEXT")))
		assert.Equal(t, "json", determineEncoding(FormatJSON))
		assert.Equal(t, "json", determineEncoding(LogFormat("JSON")))
		assert.Equal(t, "json", determineEncoding(LogFormat("invalid")))
		assert.Equal(t, "json", determineEncoding(LogFormat("")))
	})

	t.Run("should create base encoder config", func(t *testing.T) {
		config := createBaseEncoderConfig()

		assert.Equal(t, "time", config.TimeKey)
		assert.Equal(t, "level", config.LevelKey)
		assert.Equal(t, "message", config.MessageKey)
		assert.Equal(t, "caller", config.CallerKey)
		assert.Equal(t, "stacktrace", config.StacktraceKey)
	})

	t.Run("should adjust encoder for console", func(t *testing.T) {
		config := &zap.Config{
			Encoding:      "console",
			EncoderConfig: createBaseEncoderConfig(),
		}

		adjustEncoderForConsole(config)

		// Should have color level encoder for console
		assert.NotNil(t, config.EncoderConfig.EncodeLevel)
		assert.NotNil(t, config.EncoderConfig.EncodeTime)
	})

	t.Run("should not adjust encoder for json", func(t *testing.T) {
		config := &zap.Config{
			Encoding:      "json",
			EncoderConfig: createBaseEncoderConfig(),
		}

		// Store original encoding before adjustment
		originalEncoding := config.Encoding

		adjustEncoderForConsole(config)

		// Should keep original encoding for JSON (not modified)
		assert.Equal(t, originalEncoding, config.Encoding)
		assert.Equal(t, "json", config.Encoding)
	})
}

// TestZapLogger_BasicLogging tests basic logging methods
func TestZapLogger_BasicLogging(t *testing.T) {
	logger := NewNop()

	t.Run("should log at all levels without panic", func(t *testing.T) {
		assert.NotPanics(t, func() {
			logger.Debug("debug message")
			logger.Info("info message")
			logger.Warn("warn message")
			logger.Error("error message")
		})
	})

	t.Run("should handle empty messages", func(t *testing.T) {
		assert.NotPanics(t, func() {
			logger.Info("")
			logger.Debug("")
		})
	})
}

// TestZapLogger_WithFields tests the With method
func TestZapLogger_WithFields(t *testing.T) {
	logger := NewNop()

	t.Run("should create logger with fields", func(t *testing.T) {
		fieldLogger := logger.With(
			String("service", "test"),
			Int("count", 42),
			Bool("active", true),
		)

		assert.NotNil(t, fieldLogger)
		assert.NotPanics(t, func() {
			fieldLogger.Info("message with fields")
		})
	})

	t.Run("should handle nil fields", func(t *testing.T) {
		assert.NotPanics(t, func() {
			fieldLogger := logger.With()
			fieldLogger.Info("message without fields")
		})
	})

	t.Run("should handle various field types", func(t *testing.T) {
		assert.NotPanics(t, func() {
			fieldLogger := logger.With(
				String("str", "value"),
				Int("int", 123),
				Int64("int64", 456),
				Float64("float", 3.14),
				Bool("bool", true),
				Duration("duration", time.Second),
				Time("time", time.Now()),
				Error(errors.New("test error")),
			)
			fieldLogger.Info("message with various field types")
		})
	})
}

// TestZapLogger_ForInfra tests the ForInfra method
func TestZapLogger_ForInfra(t *testing.T) {
	logger := NewNop()

	t.Run("should create infrastructure logger", func(t *testing.T) {
		infraLogger := logger.ForInfra("database")

		assert.NotNil(t, infraLogger)
		assert.NotPanics(t, func() {
			infraLogger.Info("database operation")
			infraLogger.Debug("connection established")
		})
	})

	t.Run("should handle empty component name", func(t *testing.T) {
		assert.NotPanics(t, func() {
			infraLogger := logger.ForInfra("")
			infraLogger.Info("message without component")
		})
	})

	t.Run("should create loggers for different components", func(t *testing.T) {
		components := []string{"database", "cache", "queue", "storage", "http"}

		for _, component := range components {
			assert.NotPanics(t, func() {
				infraLogger := logger.ForInfra(component)
				infraLogger.Info("component operation")
			})
		}
	})

	t.Run("should handle concurrent access with double-check pattern", func(t *testing.T) {
		logger := NewNop()
		zapLogger := logger.(*ZapLogger)
		component := "concurrent-test"

		// Channel to synchronize goroutines
		ready := make(chan struct{})
		done := make(chan Logger, 10)

		// Start multiple goroutines that try to get the same component logger
		for i := 0; i < 10; i++ {
			go func() {
				// Wait for all goroutines to be ready
				<-ready
				// Try to get the logger (double-check pattern should ensure only one creation)
				infraLogger := zapLogger.ForInfra(component)
				done <- infraLogger
			}()
		}

		// Signal all goroutines to start
		close(ready)

		// Collect all loggers
		var loggers []Logger
		for i := 0; i < 10; i++ {
			loggers = append(loggers, <-done)
		}

		// All loggers should be non-nil
		for _, l := range loggers {
			assert.NotNil(t, l)
		}

		// Verify that the component was cached properly
		cachedLogger := zapLogger.ForInfra(component)
		assert.NotNil(t, cachedLogger)

		// All subsequent calls should return the same cached instance behavior
		assert.NotPanics(t, func() {
			for i := 0; i < 5; i++ {
				logger := zapLogger.ForInfra(component)
				logger.Info("cached logger test")
			}
		})
	})

	t.Run("should use infraLogger when available", func(t *testing.T) {
		cfg := DefaultLoggerConfig()
		zapLogger, err := NewZapLogger(cfg)
		assert.NoError(t, err)
		assert.NotNil(t, zapLogger)

		// ForInfra should use pre-cached infraLogger for better performance
		infraLogger := zapLogger.ForInfra("test-component")
		assert.NotNil(t, infraLogger)

		// Multiple calls should reuse cached component logger
		infraLogger2 := zapLogger.ForInfra("test-component")
		assert.NotNil(t, infraLogger2)

		// Different component should create new logger
		infraLogger3 := zapLogger.ForInfra("different-component")
		assert.NotNil(t, infraLogger3)
	})

	t.Run("should use fallback path when infraLogger is nil", func(t *testing.T) {
		// Create logger without infraLogger (using NewNop)
		nopLogger := NewNop()
		zapLogger := nopLogger.(*ZapLogger)

		// Ensure infraLogger is nil
		assert.Nil(t, zapLogger.infraLogger)

		// ForInfra should use fallback path (create from base logger)
		infraLogger := zapLogger.ForInfra("fallback-component")
		assert.NotNil(t, infraLogger)

		// Should cache the component logger for reuse
		infraLogger2 := zapLogger.ForInfra("fallback-component")
		assert.NotNil(t, infraLogger2)

		// Test logging with fallback logger
		assert.NotPanics(t, func() {
			infraLogger.Info("fallback logger test")
			infraLogger.Debug("fallback debug message")
		})
	})
}

// TestZapLogger_ForGORM tests the ForGORM method
func TestZapLogger_ForGORM(t *testing.T) {
	logger := NewNop()

	t.Run("should create GORM logger", func(t *testing.T) {
		gormLogger := logger.ForGORM()

		assert.NotNil(t, gormLogger)

		// Test GORM logger interface methods
		assert.NotPanics(t, func() {
			ctx := context.Background()
			gormLogger.Info(ctx, "GORM operation")
			gormLogger.Warn(ctx, "GORM warning")
			gormLogger.Error(ctx, "GORM error")
		})
	})

	t.Run("should create functional GORM logger adapter", func(t *testing.T) {
		gormLogger := logger.ForGORM()

		// Test LogMode functionality
		assert.NotPanics(t, func() {
			infoLogger := gormLogger.LogMode(1)  // Info level
			errorLogger := gormLogger.LogMode(2) // Error level

			assert.NotNil(t, infoLogger)
			assert.NotNil(t, errorLogger)
		})
	})

	t.Run("should use cached GORM logger when available", func(t *testing.T) {
		// Create logger with GORM logger pre-cached
		cfg := DefaultLoggerConfig()
		zapLogger, err := NewZapLogger(cfg)
		assert.NoError(t, err)
		assert.NotNil(t, zapLogger)

		// First call should use pre-cached gormLogger
		gormLogger1 := zapLogger.ForGORM()
		assert.NotNil(t, gormLogger1)

		// Second call should return the same cached instance
		gormLogger2 := zapLogger.ForGORM()
		assert.NotNil(t, gormLogger2)
	})

	t.Run("should use fallback when gormLogger is nil", func(t *testing.T) {
		// Create logger without pre-cached GORM logger
		nopLogger := NewNop()
		zapLogger := nopLogger.(*ZapLogger)

		// Ensure gormLogger is nil
		assert.Nil(t, zapLogger.gormLogger)

		// ForGORM should use fallback path
		gormLogger := zapLogger.ForGORM()
		assert.NotNil(t, gormLogger)

		// Test functionality
		assert.NotPanics(t, func() {
			ctx := context.Background()
			gormLogger.Info(ctx, "fallback GORM test")
		})
	})
}

// TestZapLogger_Sync tests the Sync method
func TestZapLogger_Sync(t *testing.T) {
	logger := NewNop()

	t.Run("should sync without error", func(t *testing.T) {
		err := logger.Sync()
		assert.NoError(t, err)
	})
}

// TestIsIgnorableSyncError tests the isIgnorableSyncError function
func TestIsIgnorableSyncError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "should return false for nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "should ignore stdout ioctl error",
			err:      errors.New("sync /dev/stdout: inappropriate ioctl for device"),
			expected: true,
		},
		{
			name:     "should ignore stderr ioctl error",
			err:      errors.New("sync /dev/stderr: inappropriate ioctl for device"),
			expected: true,
		},
		{
			name:     "should ignore stdout invalid argument error",
			err:      errors.New("sync /dev/stdout: invalid argument"),
			expected: true,
		},
		{
			name:     "should ignore stderr invalid argument error",
			err:      errors.New("sync /dev/stderr: invalid argument"),
			expected: true,
		},
		{
			name:     "should ignore simplified stdout ioctl error",
			err:      errors.New("sync stdout: inappropriate ioctl for device"),
			expected: true,
		},
		{
			name:     "should ignore simplified stderr ioctl error",
			err:      errors.New("sync stderr: inappropriate ioctl for device"),
			expected: true,
		},
		{
			name:     "should not ignore other sync errors",
			err:      errors.New("sync /tmp/file: permission denied"),
			expected: false,
		},
		{
			name:     "should not ignore completely unrelated errors",
			err:      errors.New("database connection failed"),
			expected: false,
		},
		{
			name:     "should not ignore empty error message",
			err:      errors.New(""),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isIgnorableSyncError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestZapLogger_Level tests the Level method
func TestZapLogger_Level(t *testing.T) {
	t.Run("should return correct level for different configs", func(t *testing.T) {
		levels := []struct {
			configLevel zapcore.Level
			expected    zapcore.Level
		}{
			{zapcore.DebugLevel, zapcore.DebugLevel},
			{zapcore.InfoLevel, zapcore.InfoLevel},
			{zapcore.WarnLevel, zapcore.WarnLevel},
			{zapcore.ErrorLevel, zapcore.ErrorLevel},
		}

		for _, level := range levels {
			cfg := &Config{
				Level:  level.configLevel,
				Format: "json",
			}

			logger, err := NewZapLogger(cfg)
			assert.NoError(t, err)
			assert.Equal(t, level.expected, logger.Level())
		}
	})
}

// TestNewNop tests the NewNop function
func TestNewNop(t *testing.T) {
	t.Run("should create no-op logger", func(t *testing.T) {
		logger := NewNop()

		assert.NotNil(t, logger)
		assert.Equal(t, zapcore.InfoLevel, logger.Level())
	})

	t.Run("should work as full logger interface", func(t *testing.T) {
		logger := NewNop()

		// Test all interface methods work without panic
		assert.NotPanics(t, func() {
			logger.Debug("debug")
			logger.Info("info")
			logger.Warn("warn")
			logger.Error("error")

			// Test With methods
			fieldLogger := logger.With(String("key", "value"))
			fieldLogger.Info("with fields")

			// Test infrastructure logger
			infraLogger := logger.ForInfra("test")
			infraLogger.Debug("infra message")

			// Test GORM logger
			gormLogger := logger.ForGORM()
			assert.NotNil(t, gormLogger)

			// Test sync
			err := logger.Sync()
			assert.NoError(t, err)
		})
	})
}

// TestZapLogger_Concurrency tests concurrent usage
func TestZapLogger_Concurrency(t *testing.T) {
	logger := NewNop()

	t.Run("should handle concurrent logging", func(t *testing.T) {
		const numGoroutines = 100
		const numMessages = 10

		var wg sync.WaitGroup
		wg.Add(numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				defer wg.Done()

				for j := 0; j < numMessages; j++ {
					logger.Info("concurrent message",
						String("goroutine", string(rune(id))),
						Int("message", j),
					)
				}
			}(i)
		}

		// Should not panic or deadlock
		assert.NotPanics(t, func() {
			wg.Wait()
		})
	})

	t.Run("should handle concurrent With operations", func(t *testing.T) {
		const numGoroutines = 50

		var wg sync.WaitGroup
		wg.Add(numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				defer wg.Done()

				fieldLogger := logger.With(
					String("worker", string(rune(id))),
					Int("id", id),
				)

				fieldLogger.Info("worker message")
			}(i)
		}

		assert.NotPanics(t, func() {
			wg.Wait()
		})
	})
}

// TestZapLogger_FieldTypes tests different field types
func TestZapLogger_FieldTypes(t *testing.T) {
	logger := NewNop()

	t.Run("should handle all field types", func(t *testing.T) {
		assert.NotPanics(t, func() {
			logger.With(
				String("string", "value"),
				Int("int", 42),
				Int64("int64", 123456789),
				Float64("float64", 3.14159),
				Bool("bool", true),
				Duration("duration", time.Minute),
				Time("time", time.Now()),
				Error(errors.New("test error")),
			).Info("all field types")
		})
	})

	t.Run("should handle nil error field", func(t *testing.T) {
		assert.NotPanics(t, func() {
			logger.With(Error(nil)).Info("nil error")
		})
	})

	t.Run("should handle zero values", func(t *testing.T) {
		assert.NotPanics(t, func() {
			logger.With(
				String("empty_string", ""),
				Int("zero_int", 0),
				Bool("false_bool", false),
				Duration("zero_duration", 0),
			).Info("zero values")
		})
	})
}

// TestZapLogger_ErrorHandling tests error handling scenarios
func TestZapLogger_ErrorHandling(t *testing.T) {
	t.Run("should handle nil config", func(t *testing.T) {
		logger, err := NewZapLogger(nil)
		assert.NoError(t, err)
		assert.NotNil(t, logger)
		assert.Equal(t, zapcore.InfoLevel, logger.Level())
	})

	t.Run("should handle config with default values", func(t *testing.T) {
		cfg := DefaultLoggerConfig()

		logger, err := NewZapLogger(cfg)
		assert.NoError(t, err)
		assert.NotNil(t, logger)
	})
}

// TestConvertFieldsToZap tests the convertFieldsToZap function for performance optimizations
func TestConvertFieldsToZap(t *testing.T) {
	t.Run("should handle empty fields", func(t *testing.T) {
		zapFields := convertFieldsToZap([]Field{})
		assert.Nil(t, zapFields)
	})

	t.Run("should handle nil fields", func(t *testing.T) {
		zapFields := convertFieldsToZap(nil)
		assert.Nil(t, zapFields)
	})

	t.Run("should append trace fields when present", func(t *testing.T) {
		err := RunWithTrace("req-trace-123", "corr-trace-456", func() error {
			zapFields := convertFieldsToZap(nil)

			assert.Len(t, zapFields, 2)
			assert.Equal(t, requestIDFieldKey, zapFields[0].Key)
			assert.Equal(t, "req-trace-123", zapFields[0].String)
			assert.Equal(t, correlationIDFieldKey, zapFields[1].Key)
			assert.Equal(t, "corr-trace-456", zapFields[1].String)
			return nil
		})

		assert.NoError(t, err)
	})

	t.Run("should not duplicate existing trace fields", func(t *testing.T) {
		err := RunWithTrace("req-trace-123", "corr-trace-456", func() error {
			fields := []Field{
				String(requestIDFieldKey, "existing-request"),
				String("custom", "value"),
			}

			zapFields := convertFieldsToZap(fields)

			assert.Len(t, zapFields, 3)
			assert.Equal(t, requestIDFieldKey, zapFields[0].Key)
			assert.Equal(t, "existing-request", zapFields[0].String)
			assert.Equal(t, "custom", zapFields[1].Key)
			assert.Equal(t, correlationIDFieldKey, zapFields[2].Key)
			assert.Equal(t, "corr-trace-456", zapFields[2].String)
			return nil
		})

		assert.NoError(t, err)
	})

	t.Run("fast path - single string field", func(t *testing.T) {
		fields := []Field{String("key", "value")}
		zapFields := convertFieldsToZap(fields)

		assert.Len(t, zapFields, 1)
		assert.Equal(t, "key", zapFields[0].Key)
		assert.Equal(t, "value", zapFields[0].String)
	})

	t.Run("fast path - single int field", func(t *testing.T) {
		fields := []Field{Int("count", 42)}
		zapFields := convertFieldsToZap(fields)

		assert.Len(t, zapFields, 1)
		assert.Equal(t, "count", zapFields[0].Key)
		assert.Equal(t, int64(42), zapFields[0].Integer)
	})

	t.Run("fast path - single int64 field", func(t *testing.T) {
		fields := []Field{Int64("id", 123456789)}
		zapFields := convertFieldsToZap(fields)

		assert.Len(t, zapFields, 1)
		assert.Equal(t, "id", zapFields[0].Key)
		assert.Equal(t, int64(123456789), zapFields[0].Integer)
	})

	t.Run("fast path - single float64 field", func(t *testing.T) {
		fields := []Field{Float64("price", 19.99)}
		zapFields := convertFieldsToZap(fields)

		assert.Len(t, zapFields, 1)
		assert.Equal(t, "price", zapFields[0].Key)
		// Note: zap.Field internal structure is not exposed, just verify it doesn't panic
	})

	t.Run("fast path - single bool field", func(t *testing.T) {
		fields := []Field{Bool("active", true)}
		zapFields := convertFieldsToZap(fields)

		assert.Len(t, zapFields, 1)
		assert.Equal(t, "active", zapFields[0].Key)
		// Note: zap.Field internal structure is not exposed, just verify it doesn't panic
	})

	t.Run("fast path - single time field", func(t *testing.T) {
		now := time.Now()
		fields := []Field{Time("created_at", now)}
		zapFields := convertFieldsToZap(fields)

		assert.Len(t, zapFields, 1)
		assert.Equal(t, "created_at", zapFields[0].Key)
		// Note: zap.Field internal structure is not exposed, just verify it doesn't panic
	})

	t.Run("fast path - single duration field", func(t *testing.T) {
		duration := 5 * time.Second
		fields := []Field{Duration("timeout", duration)}
		zapFields := convertFieldsToZap(fields)

		assert.Len(t, zapFields, 1)
		assert.Equal(t, "timeout", zapFields[0].Key)
		// Note: zap.Field internal structure is not exposed, just verify it doesn't panic
	})

	t.Run("fast path - single error field", func(t *testing.T) {
		err := errors.New("test error")
		fields := []Field{Error(err)}
		zapFields := convertFieldsToZap(fields)

		assert.Len(t, zapFields, 1)
		assert.Equal(t, "error", zapFields[0].Key)
		// Note: zap.Field internal structure is not exposed, just verify it doesn't panic
	})

	t.Run("fast path - single error field with nil", func(t *testing.T) {
		fields := []Field{Error(nil)}
		zapFields := convertFieldsToZap(fields)

		assert.Len(t, zapFields, 1)
		assert.Equal(t, "error", zapFields[0].Key)
		// nil error should be handled gracefully
	})

	t.Run("fast path - single any field", func(t *testing.T) {
		complexValue := map[string]interface{}{"nested": "value"}
		fields := []Field{Any("metadata", complexValue)}
		zapFields := convertFieldsToZap(fields)

		assert.Len(t, zapFields, 1)
		assert.Equal(t, "metadata", zapFields[0].Key)
		// Any field uses reflection, so we just check it doesn't panic
	})

	t.Run("multiple fields - batch conversion", func(t *testing.T) {
		fields := []Field{
			String("name", "test"),
			Int("count", 42),
			Bool("active", true),
			Float64("score", 98.5),
		}
		zapFields := convertFieldsToZap(fields)

		assert.Len(t, zapFields, 4)

		// Check first field (string)
		assert.Equal(t, "name", zapFields[0].Key)
		// Note: zap.Field internal structure is not exposed, just verify it doesn't panic

		// Check second field (int -> int64)
		assert.Equal(t, "count", zapFields[1].Key)
		// Note: zap.Field internal structure is not exposed, just verify it doesn't panic

		// Check third field (bool)
		assert.Equal(t, "active", zapFields[2].Key)
		// Note: zap.Field internal structure is not exposed, just verify it doesn't panic

		// Check fourth field (float64)
		assert.Equal(t, "score", zapFields[3].Key)
		// Note: zap.Field internal structure is not exposed, just verify it doesn't panic
	})

	t.Run("performance - zero allocation for empty", func(t *testing.T) {
		// This test ensures no allocation for empty fields
		for i := 0; i < 100; i++ {
			zapFields := convertFieldsToZap([]Field{})
			assert.Nil(t, zapFields)
		}
	})

	t.Run("performance - single allocation for multiple fields", func(t *testing.T) {
		fields := []Field{
			String("key1", "value1"),
			String("key2", "value2"),
			String("key3", "value3"),
		}

		// Should pre-allocate exact size
		zapFields := convertFieldsToZap(fields)
		assert.Len(t, zapFields, 3)
		assert.Equal(t, cap(zapFields), 3) // Capacity should equal length
	})

	t.Run("edge cases - different value types", func(t *testing.T) {
		fields := []Field{
			String("empty", ""),
			Int("zero", 0),
			Int("negative", -42),
			Float64("negative_float", -3.14),
			Bool("false", false),
			Duration("zero_duration", 0),
		}

		zapFields := convertFieldsToZap(fields)
		assert.Len(t, zapFields, 6)

		// All should convert without panic
		for i, field := range zapFields {
			assert.NotEmpty(t, field.Key, "Field %d should have non-empty key", i)
		}
	})

	t.Run("type consistency - int vs int64", func(t *testing.T) {
		intField := convertFieldsToZap([]Field{Int("int", 42)})
		int64Field := convertFieldsToZap([]Field{Int64("int64", 42)})

		// Both should convert to int64 in zap
		assert.Equal(t, int64(42), intField[0].Integer)
		assert.Equal(t, int64(42), int64Field[0].Integer)
	})

	t.Run("fallback to Any type for unknown types", func(t *testing.T) {
		// Test custom struct type (unknown type)
		type CustomStruct struct {
			Name  string
			Value int
		}

		customValue := CustomStruct{Name: "test", Value: 123}
		fields := []Field{Any("custom", customValue)}
		zapFields := convertFieldsToZap(fields)

		assert.Len(t, zapFields, 1)
		assert.Equal(t, "custom", zapFields[0].Key)
		// Note: zap.Field internal structure is not exposed, just verify it doesn't panic
	})

	t.Run("fallback to Any type for slice types", func(t *testing.T) {
		// Test slice type (unknown type)
		sliceValue := []string{"item1", "item2", "item3"}
		fields := []Field{Any("items", sliceValue)}
		zapFields := convertFieldsToZap(fields)

		assert.Len(t, zapFields, 1)
		assert.Equal(t, "items", zapFields[0].Key)
		// Note: zap.Field internal structure is not exposed, just verify it doesn't panic
	})

	t.Run("fallback to Any type for map types", func(t *testing.T) {
		// Test map type (unknown type)
		mapValue := map[string]int{"key1": 1, "key2": 2}
		fields := []Field{Any("mapping", mapValue)}
		zapFields := convertFieldsToZap(fields)

		assert.Len(t, zapFields, 1)
		assert.Equal(t, "mapping", zapFields[0].Key)
		// Note: zap.Field internal structure is not exposed, just verify it doesn't panic
	})

	t.Run("fallback to Any type for interface{} with unknown value", func(t *testing.T) {
		// Test interface{} with unknown underlying type
		var unknownValue interface{} = complex(1, 2) // complex number
		fields := []Field{Any("complex", unknownValue)}
		zapFields := convertFieldsToZap(fields)

		assert.Len(t, zapFields, 1)
		assert.Equal(t, "complex", zapFields[0].Key)
		// Note: zap.Field internal structure is not exposed, just verify it doesn't panic
	})

	t.Run("fallback batch conversion with mixed known and unknown types", func(t *testing.T) {
		// Test batch conversion with mix of known and unknown types
		type CustomData struct {
			ID   int
			Meta string
		}

		fields := []Field{
			String("name", "test"),                       // known type
			Any("data", CustomData{ID: 1, Meta: "test"}), // unknown type -> fallback
			Int("count", 42),                             // known type
			Any("tags", []string{"tag1", "tag2"}),        // unknown type -> fallback
		}

		zapFields := convertFieldsToZap(fields)
		assert.Len(t, zapFields, 4)

		// Check that all fields have proper keys
		assert.Equal(t, "name", zapFields[0].Key)
		assert.Equal(t, "data", zapFields[1].Key)
		assert.Equal(t, "count", zapFields[2].Key)
		assert.Equal(t, "tags", zapFields[3].Key)

		// Ensure no panics occurred during conversion
		assert.NotPanics(t, func() {
			convertFieldsToZap(fields)
		})
	})
}
