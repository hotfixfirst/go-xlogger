package xlogger

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zapcore"
)

// TestLogFormat tests the LogFormat type
func TestLogFormat(t *testing.T) {
	t.Run("String should return format string", func(t *testing.T) {
		assert.Equal(t, "json", FormatJSON.String())
		assert.Equal(t, "text", FormatText.String())
	})

	t.Run("IsValid should validate format", func(t *testing.T) {
		assert.True(t, FormatJSON.IsValid())
		assert.True(t, FormatText.IsValid())
		assert.True(t, LogFormat("JSON").IsValid())
		assert.True(t, LogFormat("TEXT").IsValid())
		assert.False(t, LogFormat("invalid").IsValid())
		assert.False(t, LogFormat("").IsValid())
	})

	t.Run("Normalize should return lowercase format", func(t *testing.T) {
		assert.Equal(t, FormatJSON, LogFormat("JSON").Normalize())
		assert.Equal(t, FormatJSON, LogFormat("Json").Normalize())
		assert.Equal(t, FormatText, LogFormat("TEXT").Normalize())
		assert.Equal(t, FormatText, LogFormat("Text").Normalize())
	})
}

// TestDefaultLoggerConfig tests the DefaultLoggerConfig function
func TestDefaultLoggerConfig(t *testing.T) {
	t.Run("should return default configuration", func(t *testing.T) {
		cfg := DefaultLoggerConfig()

		assert.NotNil(t, cfg)
		assert.Equal(t, zapcore.InfoLevel, cfg.Level)
		assert.Equal(t, FormatJSON, cfg.Format)
		assert.False(t, cfg.Development)
		assert.False(t, cfg.DisableCaller)
		assert.True(t, cfg.DisableStacktrace)
		assert.Empty(t, cfg.TimeFormat)
		assert.Equal(t, 1, cfg.CallerSkip)
	})
}

// TestNewLoggerConfig tests the NewLoggerConfig function with functional options
func TestNewLoggerConfig(t *testing.T) {
	t.Run("should return default config when no options provided", func(t *testing.T) {
		cfg := NewLoggerConfig()

		assert.NotNil(t, cfg)
		expected := DefaultLoggerConfig()
		assert.Equal(t, expected.Level, cfg.Level)
		assert.Equal(t, expected.Format, cfg.Format)
		assert.Equal(t, expected.Development, cfg.Development)
		assert.Equal(t, expected.DisableCaller, cfg.DisableCaller)
		assert.Equal(t, expected.DisableStacktrace, cfg.DisableStacktrace)
		assert.Equal(t, expected.CallerSkip, cfg.CallerSkip)
	})

	t.Run("should apply single option", func(t *testing.T) {
		cfg := NewLoggerConfig(WithLevel(zapcore.WarnLevel))

		assert.NotNil(t, cfg)
		assert.Equal(t, zapcore.WarnLevel, cfg.Level)
		// Other values should be default
		assert.Equal(t, FormatJSON, cfg.Format)
	})

	t.Run("should apply multiple options", func(t *testing.T) {
		cfg := NewLoggerConfig(
			WithLevel(zapcore.DebugLevel),
			WithFormat(FormatText),
			WithDevelopment(true),
		)

		assert.NotNil(t, cfg)
		assert.Equal(t, zapcore.DebugLevel, cfg.Level)
		assert.Equal(t, FormatText, cfg.Format)
		assert.True(t, cfg.Development)
		assert.False(t, cfg.DisableStacktrace) // Auto-enabled for debug level
	})

	t.Run("should apply all options", func(t *testing.T) {
		cfg := NewLoggerConfig(
			WithLevel(zapcore.WarnLevel),
			WithFormat(FormatText),
			WithDevelopment(true),
			WithDisableCaller(true),
			WithDisableStacktrace(true),
			WithTimeFormat("2006-01-02"),
			WithCallerSkip(2),
		)

		assert.NotNil(t, cfg)
		assert.Equal(t, zapcore.WarnLevel, cfg.Level)
		assert.Equal(t, FormatText, cfg.Format)
		assert.True(t, cfg.Development)
		assert.True(t, cfg.DisableCaller)
		assert.True(t, cfg.DisableStacktrace)
		assert.Equal(t, "2006-01-02", cfg.TimeFormat)
		assert.Equal(t, 2, cfg.CallerSkip)
	})

	t.Run("later options should override earlier ones", func(t *testing.T) {
		cfg := NewLoggerConfig(
			WithLevel(zapcore.InfoLevel),
			WithLevel(zapcore.ErrorLevel), // Override
		)

		assert.Equal(t, zapcore.ErrorLevel, cfg.Level)
	})
}

// TestWithLevel tests the WithLevel option
func TestWithLevel(t *testing.T) {
	t.Run("should set log level", func(t *testing.T) {
		levels := []zapcore.Level{
			zapcore.DebugLevel,
			zapcore.InfoLevel,
			zapcore.WarnLevel,
			zapcore.ErrorLevel,
		}

		for _, level := range levels {
			cfg := NewLoggerConfig(WithLevel(level))
			assert.Equal(t, level, cfg.Level, "Failed for level: %s", level.String())
		}
	})

	t.Run("should auto-enable stacktrace for debug level", func(t *testing.T) {
		cfg := NewLoggerConfig(WithLevel(zapcore.DebugLevel))
		assert.False(t, cfg.DisableStacktrace)
	})

	t.Run("should not change stacktrace for other levels", func(t *testing.T) {
		cfg := NewLoggerConfig(WithLevel(zapcore.InfoLevel))
		assert.True(t, cfg.DisableStacktrace) // Default is true
	})
}

// TestWithLevelString tests the WithLevelString option
func TestWithLevelString(t *testing.T) {
	t.Run("should set log level from string", func(t *testing.T) {
		tests := []struct {
			input    string
			expected zapcore.Level
		}{
			{"debug", zapcore.DebugLevel},
			{"DEBUG", zapcore.DebugLevel},
			{"info", zapcore.InfoLevel},
			{"INFO", zapcore.InfoLevel},
			{"warn", zapcore.WarnLevel},
			{"WARN", zapcore.WarnLevel},
			{"error", zapcore.ErrorLevel},
			{"ERROR", zapcore.ErrorLevel},
		}

		for _, tt := range tests {
			cfg := NewLoggerConfig(WithLevelString(tt.input))
			assert.Equal(t, tt.expected, cfg.Level, "Failed for input: %s", tt.input)
		}
	})

	t.Run("should auto-enable stacktrace for debug level string", func(t *testing.T) {
		cfg := NewLoggerConfig(WithLevelString("debug"))
		assert.False(t, cfg.DisableStacktrace)
	})

	t.Run("should ignore invalid level string", func(t *testing.T) {
		cfg := NewLoggerConfig(WithLevelString("invalid"))
		assert.Equal(t, zapcore.InfoLevel, cfg.Level) // Should keep default
	})

	t.Run("should ignore empty level string", func(t *testing.T) {
		cfg := NewLoggerConfig(WithLevelString(""))
		assert.Equal(t, zapcore.InfoLevel, cfg.Level) // Should keep default
	})
}

// TestWithFormat tests the WithFormat option
func TestWithFormat(t *testing.T) {
	t.Run("should set log format", func(t *testing.T) {
		formats := []struct {
			inputFormat    LogFormat
			expectedFormat LogFormat
		}{
			{FormatJSON, FormatJSON},
			{FormatText, FormatText},
			{LogFormat("JSON"), FormatJSON},
			{LogFormat("TEXT"), FormatText},
		}

		for _, format := range formats {
			cfg := NewLoggerConfig(WithFormat(format.inputFormat))
			assert.Equal(t, format.expectedFormat, cfg.Format, "Failed for format: %s", format.inputFormat)
		}
	})

	t.Run("should ignore invalid format", func(t *testing.T) {
		cfg := NewLoggerConfig(WithFormat(LogFormat("invalid")))
		assert.Equal(t, FormatJSON, cfg.Format) // Should keep default
	})
}

// TestWithDevelopment tests the WithDevelopment option
func TestWithDevelopment(t *testing.T) {
	t.Run("should enable development mode", func(t *testing.T) {
		cfg := NewLoggerConfig(WithDevelopment(true))
		assert.True(t, cfg.Development)
	})

	t.Run("should disable development mode", func(t *testing.T) {
		cfg := NewLoggerConfig(WithDevelopment(false))
		assert.False(t, cfg.Development)
	})
}

// TestWithDisableCaller tests the WithDisableCaller option
func TestWithDisableCaller(t *testing.T) {
	t.Run("should disable caller", func(t *testing.T) {
		cfg := NewLoggerConfig(WithDisableCaller(true))
		assert.True(t, cfg.DisableCaller)
	})

	t.Run("should enable caller", func(t *testing.T) {
		cfg := NewLoggerConfig(WithDisableCaller(false))
		assert.False(t, cfg.DisableCaller)
	})
}

// TestWithDisableStacktrace tests the WithDisableStacktrace option
func TestWithDisableStacktrace(t *testing.T) {
	t.Run("should disable stacktrace", func(t *testing.T) {
		cfg := NewLoggerConfig(WithDisableStacktrace(true))
		assert.True(t, cfg.DisableStacktrace)
	})

	t.Run("should enable stacktrace", func(t *testing.T) {
		cfg := NewLoggerConfig(WithDisableStacktrace(false))
		assert.False(t, cfg.DisableStacktrace)
	})
}

// TestWithTimeFormat tests the WithTimeFormat option
func TestWithTimeFormat(t *testing.T) {
	t.Run("should set time format", func(t *testing.T) {
		cfg := NewLoggerConfig(WithTimeFormat("2006-01-02 15:04:05"))
		assert.Equal(t, "2006-01-02 15:04:05", cfg.TimeFormat)
	})

	t.Run("should allow empty time format", func(t *testing.T) {
		cfg := NewLoggerConfig(WithTimeFormat(""))
		assert.Empty(t, cfg.TimeFormat)
	})
}

// TestWithCallerSkip tests the WithCallerSkip option
func TestWithCallerSkip(t *testing.T) {
	t.Run("should set caller skip", func(t *testing.T) {
		cfg := NewLoggerConfig(WithCallerSkip(3))
		assert.Equal(t, 3, cfg.CallerSkip)
	})

	t.Run("should allow zero caller skip", func(t *testing.T) {
		cfg := NewLoggerConfig(WithCallerSkip(0))
		assert.Equal(t, 0, cfg.CallerSkip)
	})
}

// TestConfigHelperMethods tests the helper methods on Config
func TestConfigHelperMethods(t *testing.T) {
	t.Run("GetLevel should return level string", func(t *testing.T) {
		cfg := &Config{Level: zapcore.DebugLevel}
		assert.Equal(t, "debug", cfg.GetLevel())

		cfg.Level = zapcore.InfoLevel
		assert.Equal(t, "info", cfg.GetLevel())

		cfg.Level = zapcore.WarnLevel
		assert.Equal(t, "warn", cfg.GetLevel())

		cfg.Level = zapcore.ErrorLevel
		assert.Equal(t, "error", cfg.GetLevel())
	})

	t.Run("GetFormat should return lowercase format string", func(t *testing.T) {
		cfg := &Config{Format: FormatJSON}
		assert.Equal(t, "json", cfg.GetFormat())

		cfg.Format = FormatText
		assert.Equal(t, "text", cfg.GetFormat())

		cfg.Format = LogFormat("JSON")
		assert.Equal(t, "json", cfg.GetFormat())
	})

	t.Run("IsDebugLevel should return correct value", func(t *testing.T) {
		cfg := &Config{Level: zapcore.DebugLevel}
		assert.True(t, cfg.IsDebugLevel())

		cfg.Level = zapcore.InfoLevel
		assert.False(t, cfg.IsDebugLevel())
	})

	t.Run("IsInfoLevel should return correct value", func(t *testing.T) {
		cfg := &Config{Level: zapcore.InfoLevel}
		assert.True(t, cfg.IsInfoLevel())

		cfg.Level = zapcore.DebugLevel
		assert.False(t, cfg.IsInfoLevel())
	})

	t.Run("IsWarnLevel should return correct value", func(t *testing.T) {
		cfg := &Config{Level: zapcore.WarnLevel}
		assert.True(t, cfg.IsWarnLevel())

		cfg.Level = zapcore.InfoLevel
		assert.False(t, cfg.IsWarnLevel())
	})

	t.Run("IsErrorLevel should return correct value", func(t *testing.T) {
		cfg := &Config{Level: zapcore.ErrorLevel}
		assert.True(t, cfg.IsErrorLevel())

		cfg.Level = zapcore.InfoLevel
		assert.False(t, cfg.IsErrorLevel())
	})

	t.Run("IsJSONFormat should return correct value", func(t *testing.T) {
		cfg := &Config{Format: FormatJSON}
		assert.True(t, cfg.IsJSONFormat())

		cfg.Format = LogFormat("JSON")
		assert.True(t, cfg.IsJSONFormat())

		cfg.Format = FormatText
		assert.False(t, cfg.IsJSONFormat())
	})

	t.Run("IsTextFormat should return correct value", func(t *testing.T) {
		cfg := &Config{Format: FormatText}
		assert.True(t, cfg.IsTextFormat())

		cfg.Format = LogFormat("TEXT")
		assert.True(t, cfg.IsTextFormat())

		cfg.Format = FormatJSON
		assert.False(t, cfg.IsTextFormat())
	})

	t.Run("IsDevelopment should return correct value", func(t *testing.T) {
		cfg := &Config{Development: true}
		assert.True(t, cfg.IsDevelopment())

		cfg.Development = false
		assert.False(t, cfg.IsDevelopment())
	})
}
