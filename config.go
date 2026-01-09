// Package xlogger provides advanced logging functionalities built on top of zap.
package xlogger

import (
	"strings"

	"go.uber.org/zap/zapcore"
)

// LogFormat represents the log output format.
type LogFormat string

const (
	// FormatJSON outputs logs in JSON format.
	FormatJSON LogFormat = "json"
	// FormatText outputs logs in human-readable text format.
	FormatText LogFormat = "text"
)

// String returns the string representation of LogFormat.
func (f LogFormat) String() string {
	return string(f)
}

// IsValid returns true if the format is valid (json or text).
func (f LogFormat) IsValid() bool {
	lower := strings.ToLower(string(f))
	return lower == "json" || lower == "text"
}

// Normalize returns the normalized lowercase format.
func (f LogFormat) Normalize() LogFormat {
	return LogFormat(strings.ToLower(string(f)))
}

// Config represents logger configuration options.
type Config struct {
	Level             zapcore.Level // Minimum log level
	Format            LogFormat     // Log format: FormatJSON or FormatText
	Development       bool          // Development mode (pretty printing)
	DisableCaller     bool          // Disable caller information
	DisableStacktrace bool          // Disable stacktrace in errors
	TimeFormat        string        // Time format (empty for default)
	CallerSkip        int           // Number of caller frames to skip
}

// DefaultLoggerConfig returns default logger configuration with INFO level and JSON format.
//
// Default values:
//   - Level: INFO
//   - Format: FormatJSON
//   - Development: false
//   - DisableCaller: false
//   - DisableStacktrace: true
//   - CallerSkip: 1
//
// Example:
//
//	cfg := xlogger.DefaultLoggerConfig()
//	logger, err := xlogger.NewZapLogger(cfg)
func DefaultLoggerConfig() *Config {
	return &Config{
		Level:             zapcore.InfoLevel,
		Format:            FormatJSON,
		Development:       false,
		DisableCaller:     false,
		DisableStacktrace: true,
		TimeFormat:        "",
		CallerSkip:        1,
	}
}

// NewLoggerConfig creates logger configuration with functional options.
// Starts with DefaultLoggerConfig and applies each option in order.
//
// Example:
//
//	cfg := xlogger.NewLoggerConfig(
//	    xlogger.WithLevel(zapcore.DebugLevel),
//	    xlogger.WithFormat(xlogger.FormatText),
//	    xlogger.WithDevelopment(true),
//	)
//	logger, err := xlogger.NewZapLogger(cfg)
func NewLoggerConfig(opts ...Option) *Config {
	cfg := DefaultLoggerConfig()
	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}

// GetLevel returns the log level as string.
func (c *Config) GetLevel() string {
	return c.Level.String()
}

// GetFormat returns normalized log format in lowercase.
func (c *Config) GetFormat() string {
	return string(c.Format.Normalize())
}

// IsDebugLevel returns true if the log level is debug.
func (c *Config) IsDebugLevel() bool {
	return c.Level == zapcore.DebugLevel
}

// IsInfoLevel returns true if the log level is info.
func (c *Config) IsInfoLevel() bool {
	return c.Level == zapcore.InfoLevel
}

// IsWarnLevel returns true if the log level is warn.
func (c *Config) IsWarnLevel() bool {
	return c.Level == zapcore.WarnLevel
}

// IsErrorLevel returns true if the log level is error.
func (c *Config) IsErrorLevel() bool {
	return c.Level == zapcore.ErrorLevel
}

// IsJSONFormat returns true if the format is JSON.
func (c *Config) IsJSONFormat() bool {
	return c.Format.Normalize() == FormatJSON
}

// IsTextFormat returns true if the format is text.
func (c *Config) IsTextFormat() bool {
	return c.Format.Normalize() == FormatText
}

// IsDevelopment returns true if development mode is enabled.
func (c *Config) IsDevelopment() bool {
	return c.Development
}

// Option is a function that modifies Config.
type Option func(*Config)

// WithLevel sets the log level.
//
// Example:
//
//	cfg := xlogger.NewLoggerConfig(
//	    xlogger.WithLevel(zapcore.DebugLevel),
//	)
func WithLevel(level zapcore.Level) Option {
	return func(c *Config) {
		c.Level = level
		// Auto-enable stacktrace for debug level
		if level == zapcore.DebugLevel {
			c.DisableStacktrace = false
		}
	}
}

// WithLevelString sets the log level from string.
// Supported values: "debug", "info", "warn", "error", "dpanic", "panic", "fatal"
// Invalid values are ignored and the default level (INFO) is used.
//
// Example:
//
//	cfg := xlogger.NewLoggerConfig(
//	    xlogger.WithLevelString("debug"),
//	)
func WithLevelString(level string) Option {
	return func(c *Config) {
		if parsed, err := zapcore.ParseLevel(level); err == nil {
			c.Level = parsed
			if parsed == zapcore.DebugLevel {
				c.DisableStacktrace = false
			}
		}
	}
}

// WithFormat sets the log format.
//
// Example:
//
//	cfg := xlogger.NewLoggerConfig(
//	    xlogger.WithFormat(xlogger.FormatText),
//	)
func WithFormat(format LogFormat) Option {
	return func(c *Config) {
		if format.IsValid() {
			c.Format = format.Normalize()
		}
	}
}

// WithDevelopment enables or disables development mode.
//
// Example:
//
//	cfg := xlogger.NewLoggerConfig(
//	    xlogger.WithDevelopment(true),
//	)
func WithDevelopment(dev bool) Option {
	return func(c *Config) {
		c.Development = dev
	}
}

// WithDisableCaller disables caller information.
//
// Example:
//
//	cfg := xlogger.NewLoggerConfig(
//	    xlogger.WithDisableCaller(true),
//	)
func WithDisableCaller(disable bool) Option {
	return func(c *Config) {
		c.DisableCaller = disable
	}
}

// WithDisableStacktrace disables stacktrace in errors.
//
// Example:
//
//	cfg := xlogger.NewLoggerConfig(
//	    xlogger.WithDisableStacktrace(false),
//	)
func WithDisableStacktrace(disable bool) Option {
	return func(c *Config) {
		c.DisableStacktrace = disable
	}
}

// WithTimeFormat sets the time format.
//
// Example:
//
//	cfg := xlogger.NewLoggerConfig(
//	    xlogger.WithTimeFormat("2006-01-02 15:04:05"),
//	)
func WithTimeFormat(format string) Option {
	return func(c *Config) {
		c.TimeFormat = format
	}
}

// WithCallerSkip sets the number of caller frames to skip.
//
// Example:
//
//	cfg := xlogger.NewLoggerConfig(
//	    xlogger.WithCallerSkip(2),
//	)
func WithCallerSkip(skip int) Option {
	return func(c *Config) {
		c.CallerSkip = skip
	}
}
