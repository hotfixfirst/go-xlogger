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

// NewLoggerConfig creates logger configuration with custom settings.
// Falls back to DefaultLoggerConfig for nil input.
//
// Example:
//
//	cfg := xlogger.NewLoggerConfig(&xlogger.Config{
//	    Level:       zapcore.DebugLevel,
//	    Format:      xlogger.FormatText,
//	    Development: true,
//	})
//	logger, err := xlogger.NewZapLogger(cfg)
func NewLoggerConfig(cfg *Config) *Config {
	if cfg == nil {
		return DefaultLoggerConfig()
	}

	loggerCfg := DefaultLoggerConfig()

	loggerCfg.Level = cfg.Level
	loggerCfg.Development = cfg.Development
	loggerCfg.DisableCaller = cfg.DisableCaller
	loggerCfg.TimeFormat = cfg.TimeFormat
	loggerCfg.CallerSkip = cfg.CallerSkip

	if cfg.Format.IsValid() {
		loggerCfg.Format = cfg.Format.Normalize()
	}

	if cfg.IsDebugLevel() {
		loggerCfg.DisableStacktrace = false
	} else {
		loggerCfg.DisableStacktrace = cfg.DisableStacktrace
	}

	return loggerCfg
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
