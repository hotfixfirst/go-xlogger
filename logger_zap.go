package xlogger

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	ConsoleTimeLayout = "2006-01-02 15:04:05 -07:00"
)

const (
	requestIDFieldKey     = "request_id"
	correlationIDFieldKey = "correlation_id"
)

// ZapLogger implements Logger interface using zap as the underlying logger
type ZapLogger struct {
	logger           *zap.Logger
	level            zapcore.Level
	mu               sync.RWMutex
	infraLogger      *ZapLogger
	gormLogger       *GORMLogger
	componentLoggers map[string]Logger
}

// determineEncoding extracts encoding determination logic
func determineEncoding(format LogFormat) string {
	normalized := format.Normalize()
	if normalized == FormatText {
		return "console"
	}
	return "json"
}

// createBaseEncoderConfig creates the base encoder configuration
func createBaseEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.RFC3339TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
}

// adjustEncoderForConsole adjusts encoder config for console format
func adjustEncoderForConsole(config *zap.Config) {
	if config.Encoding == "console" {
		config.EncoderConfig.EncodeLevel = emojiLevelEncoder
		config.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(ConsoleTimeLayout)
	}
}

// emojiLevelEncoder adds emoji to log levels for better visual distinction
func emojiLevelEncoder(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	switch level {
	case zapcore.DebugLevel:
		enc.AppendString("🧪 DEBUG")
	case zapcore.InfoLevel:
		enc.AppendString("📢 INFO ")
	case zapcore.WarnLevel:
		enc.AppendString("🚧 WARN ")
	case zapcore.ErrorLevel:
		enc.AppendString("❌ ERROR")
	case zapcore.DPanicLevel:
		enc.AppendString("🚨 DPANIC")
	case zapcore.PanicLevel:
		enc.AppendString("⛔ PANIC")
	case zapcore.FatalLevel:
		enc.AppendString("💀 FATAL")
	default:
		enc.AppendString(level.CapitalString())
	}
}

// NewZapLogger creates a ZapLogger with full configuration support
func NewZapLogger(cfg *Config) (*ZapLogger, error) {
	// Default configuration when no config provided
	if cfg == nil {
		cfg = DefaultLoggerConfig()
	}

	// Determine encoding using helper function
	encoding := determineEncoding(cfg.Format)
	config := zap.Config{
		Level:       zap.NewAtomicLevelAt(cfg.Level),
		Development: cfg.Development,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding:          encoding,
		EncoderConfig:     createBaseEncoderConfig(),
		OutputPaths:       []string{"stdout"},
		ErrorOutputPaths:  []string{"stderr"},
		DisableCaller:     cfg.DisableCaller,
		DisableStacktrace: cfg.DisableStacktrace,
	}
	adjustEncoderForConsole(&config)

	// Use CallerSkip from config for infrastructure logger
	var zapOptions []zap.Option
	if cfg.CallerSkip > 0 {
		zapOptions = append(zapOptions, zap.AddCallerSkip(cfg.CallerSkip))
	}

	zapLogger, err := config.Build(zapOptions...)
	if err != nil {
		return nil, err
	}

	baseLogger := &ZapLogger{
		logger:           zapLogger,
		level:            cfg.Level,
		componentLoggers: make(map[string]Logger),
	}

	// Pre-create infrastructure loggers for performance
	if err := baseLogger.initInfrastructureLoggers(cfg); err != nil {
		return nil, fmt.Errorf("failed to initialize infrastructure loggers: %w", err)
	}
	return baseLogger, nil
}

// initInfrastructureLoggers pre-creates infrastructure and GORM loggers for performance
func (l *ZapLogger) initInfrastructureLoggers(cfg *Config) error {
	// Determine encoding using helper function
	encoding := determineEncoding(cfg.Format)
	infraConfig := zap.Config{
		Level:       zap.NewAtomicLevelAt(cfg.Level),
		Development: cfg.Development,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding:          encoding,
		EncoderConfig:     createBaseEncoderConfig(),
		OutputPaths:       []string{"stdout"},
		ErrorOutputPaths:  []string{"stderr"},
		DisableCaller:     true,
		DisableStacktrace: true,
	}
	adjustEncoderForConsole(&infraConfig)

	// Use CallerSkip from config for infrastructure logger
	var infraOptions []zap.Option
	if cfg.CallerSkip > 0 {
		infraOptions = append(infraOptions, zap.AddCallerSkip(cfg.CallerSkip))
	}

	infraZapLogger, err := infraConfig.Build(infraOptions...)
	if err != nil {
		return fmt.Errorf("failed to create infrastructure logger: %w", err)
	}

	// Create simple infrastructure logger wrapper (no recursive initialization)
	l.infraLogger = &ZapLogger{
		logger: infraZapLogger,
		level:  cfg.Level,
	}

	// Pre-create GORM logger using infrastructure logger for performance
	l.gormLogger = NewGORMLogger(l.infraLogger)
	return nil
}

// convertFieldsToZap converts our Field slice to zap.Field slice with performance optimizations
func convertFieldsToZap(fields []Field) []zap.Field {
	fields = withTraceFields(fields)

	fieldCount := len(fields)
	if fieldCount == 0 {
		return nil
	}

	// Fast path for single field
	if fieldCount == 1 {
		field := fields[0]
		key := field.Key()
		// Direct type assertion without Type() method call
		switch v := field.Value().(type) {
		case string:
			return []zap.Field{zap.String(key, v)}
		case int:
			return []zap.Field{zap.Int(key, v)}
		case int64:
			return []zap.Field{zap.Int64(key, v)}
		case float64:
			return []zap.Field{zap.Float64(key, v)}
		case bool:
			return []zap.Field{zap.Bool(key, v)}
		case time.Time:
			return []zap.Field{zap.Time(key, v)}
		case time.Duration:
			return []zap.Field{zap.Duration(key, v)}
		case error:
			return []zap.Field{zap.NamedError(key, v)}
		default:
			return []zap.Field{zap.Any(key, v)}
		}
	}

	// Pre-allocate exact size for better memory efficiency
	zapFields := make([]zap.Field, fieldCount)

	// Optimized conversion loop
	for i, field := range fields {
		key := field.Key()
		// Direct type assertion eliminates the overhead of Type() method call
		switch v := field.Value().(type) {
		case string:
			zapFields[i] = zap.String(key, v)
		case int:
			zapFields[i] = zap.Int(key, v)
		case int64:
			zapFields[i] = zap.Int64(key, v)
		case float64:
			zapFields[i] = zap.Float64(key, v)
		case bool:
			zapFields[i] = zap.Bool(key, v)
		case time.Time:
			zapFields[i] = zap.Time(key, v)
		case time.Duration:
			zapFields[i] = zap.Duration(key, v)
		case error:
			zapFields[i] = zap.NamedError(key, v)
		default:
			// Fallback to Any type for unknown types
			zapFields[i] = zap.Any(key, v)
		}
	}
	return zapFields
}

// withTraceFields ensures request and correlation identifiers are appended
// to each log entry when they are not already present.
func withTraceFields(fields []Field) []Field {
	requestID := TraceRequestID()
	correlationID := TraceCorrelationID()

	needRequestID := requestID != ""
	needCorrelationID := correlationID != ""

	if !needRequestID && !needCorrelationID {
		return fields
	}

	hasRequestID := false
	hasCorrelationID := false
	for _, field := range fields {
		switch field.Key() {
		case requestIDFieldKey:
			hasRequestID = true
		case correlationIDFieldKey:
			hasCorrelationID = true
		}
	}

	if needRequestID && !hasRequestID {
		fields = append(fields, String(requestIDFieldKey, requestID))
	}
	if needCorrelationID && !hasCorrelationID {
		fields = append(fields, String(correlationIDFieldKey, correlationID))
	}

	return fields
}

// Debug logs a debug message with fields
func (l *ZapLogger) Debug(msg string, fields ...Field) {
	l.logger.Debug(msg, convertFieldsToZap(fields)...)
}

// Info logs an info message with fields
func (l *ZapLogger) Info(msg string, fields ...Field) {
	l.logger.Info(msg, convertFieldsToZap(fields)...)
}

// Warn logs a warning message with fields
func (l *ZapLogger) Warn(msg string, fields ...Field) {
	l.logger.Warn(msg, convertFieldsToZap(fields)...)
}

// Error logs an error message with fields
func (l *ZapLogger) Error(msg string, fields ...Field) {
	l.logger.Error(msg, convertFieldsToZap(fields)...)
}

// Panic logs a panic message with fields then calls panic()
func (l *ZapLogger) Panic(msg string, fields ...Field) {
	l.logger.Panic(msg, convertFieldsToZap(fields)...)
}

// Fatal logs a fatal message with fields then calls os.Exit(1)
func (l *ZapLogger) Fatal(msg string, fields ...Field) {
	l.logger.Fatal(msg, convertFieldsToZap(fields)...)
}

// With creates a new logger instance with additional fields pre-attached
func (l *ZapLogger) With(fields ...Field) Logger {
	newLogger := l.logger.With(convertFieldsToZap(fields)...)
	return &ZapLogger{
		logger:           newLogger,
		level:            l.level,
		mu:               sync.RWMutex{},
		infraLogger:      l.infraLogger,
		gormLogger:       l.gormLogger,
		componentLoggers: make(map[string]Logger),
	}
}

// ForInfra returns a logger optimized for infrastructure components
func (l *ZapLogger) ForInfra(component string) Logger {
	// Normalize component name with early return for empty
	if component == "" {
		component = "unknown"
	}

	// Fast read-only check first
	l.mu.RLock()
	if logger, exists := l.componentLoggers[component]; exists {
		l.mu.RUnlock()
		return logger
	}
	l.mu.RUnlock()

	// Upgrade to write lock only when needed
	l.mu.Lock()
	defer l.mu.Unlock()

	// Double-check pattern after acquiring write lock
	if logger, exists := l.componentLoggers[component]; exists {
		return logger
	}

	// Fast path: use pre-cached infrastructure logger if available
	if l.infraLogger != nil {
		componentLogger := l.infraLogger.With(String("component", component))
		l.componentLoggers[component] = componentLogger
		return componentLogger
	}

	// Fallback: create component logger from base logger
	componentLogger := l.With(String("component", component))
	l.componentLoggers[component] = componentLogger
	return componentLogger
}

// ForFxEvent returns a FX event logger that implements fxevent.Logger interface
func (l *ZapLogger) ForFxEvent() fxevent.Logger {
	return NewFxEventLogger(l.ForInfra("fx"))
}

// ForGORM returns a pre-cached logger optimized for GORM
func (l *ZapLogger) ForGORM() *GORMLogger {
	if l.gormLogger != nil {
		return l.gormLogger
	}
	// Fallback: create GORM logger if not pre-cached
	return NewGORMLogger(l)
}

// isIgnorableSyncError checks if a sync error can be safely ignored
// Common sync errors occur when stdout/stderr is redirected, piped, or in containers
func isIgnorableSyncError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	// Common sync errors that can be safely ignored
	ignorableErrors := []string{
		"sync /dev/stdout: inappropriate ioctl for device",
		"sync /dev/stderr: inappropriate ioctl for device",
		"sync /dev/stdout: invalid argument",
		"sync /dev/stderr: invalid argument",
		"sync stdout: inappropriate ioctl for device",
		"sync stderr: inappropriate ioctl for device",
	}

	for _, ignorable := range ignorableErrors {
		if strings.Contains(errStr, ignorable) {
			return true
		}
	}
	return false
}

// Sync implements Logger interface
func (l *ZapLogger) Sync() error {
	err := l.logger.Sync()
	if err != nil {
		// Ignore sync errors for stdout/stderr when output is redirected or piped
		// This commonly happens in containers, CI/CD, or when output is redirected
		if isIgnorableSyncError(err) {
			return nil
		}
	}
	return err
}

// Level returns the current logging level
func (l *ZapLogger) Level() zapcore.Level {
	return l.level
}

// NewNop creates a no-operation logger for testing purposes
// This logger discards all log entries and has minimal overhead
func NewNop() Logger {
	nopLogger := zap.NewNop()
	return &ZapLogger{
		logger:           nopLogger,
		level:            zapcore.InfoLevel,
		mu:               sync.RWMutex{},
		componentLoggers: make(map[string]Logger),
	}
}
