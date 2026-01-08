package xlogger

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap/zapcore"
	gormlogger "gorm.io/gorm/logger"
)

// MockLogger implements the Logger interface for testing
type MockLogger struct {
	mock.Mock
	level zapcore.Level
}

func (m *MockLogger) Debug(msg string, fields ...Field) {
	args := []interface{}{msg}
	for _, field := range fields {
		args = append(args, field)
	}
	m.Called(args...)
}

func (m *MockLogger) Info(msg string, fields ...Field) {
	args := []interface{}{msg}
	for _, field := range fields {
		args = append(args, field)
	}
	m.Called(args...)
}

func (m *MockLogger) Warn(msg string, fields ...Field) {
	args := []interface{}{msg}
	for _, field := range fields {
		args = append(args, field)
	}
	m.Called(args...)
}

func (m *MockLogger) Error(msg string, fields ...Field) {
	args := []interface{}{msg}
	for _, field := range fields {
		args = append(args, field)
	}
	m.Called(args...)
}

func (m *MockLogger) Panic(msg string, fields ...Field) {
	args := []interface{}{msg}
	for _, field := range fields {
		args = append(args, field)
	}
	m.Called(args...)
}

func (m *MockLogger) Fatal(msg string, fields ...Field) {
	args := []interface{}{msg}
	for _, field := range fields {
		args = append(args, field)
	}
	m.Called(args...)
}

func (m *MockLogger) With(fields ...Field) Logger {
	args := []interface{}{}
	for _, field := range fields {
		args = append(args, field)
	}
	result := m.Called(args...)
	return result.Get(0).(Logger)
}

func (m *MockLogger) WithContext(ctx context.Context) Logger {
	result := m.Called(ctx)
	return result.Get(0).(Logger)
}

func (m *MockLogger) ForInfra(component string) Logger {
	result := m.Called(component)
	return result.Get(0).(Logger)
}

func (m *MockLogger) ForFxEvent() fxevent.Logger {
	args := m.Called()
	return args.Get(0).(fxevent.Logger)
}

func (m *MockLogger) ForGORM() *GORMLogger {
	result := m.Called()
	return result.Get(0).(*GORMLogger)
}

func (m *MockLogger) Level() zapcore.Level {
	return m.level
}

func (m *MockLogger) Sync() error {
	result := m.Called()
	return result.Error(0)
}

func (m *MockLogger) SetLevel(level zapcore.Level) {
	m.level = level
}

func TestNewGORMLogger(t *testing.T) {
	tests := []struct {
		name              string
		loggerLevel       zapcore.Level
		expectedGormLevel gormlogger.LogLevel
	}{
		{
			name:              "debug level maps to gorm info",
			loggerLevel:       zapcore.DebugLevel,
			expectedGormLevel: gormlogger.Info,
		},
		{
			name:              "info level maps to gorm warn",
			loggerLevel:       zapcore.InfoLevel,
			expectedGormLevel: gormlogger.Warn,
		},
		{
			name:              "warn level maps to gorm warn",
			loggerLevel:       zapcore.WarnLevel,
			expectedGormLevel: gormlogger.Warn,
		},
		{
			name:              "error level maps to gorm error",
			loggerLevel:       zapcore.ErrorLevel,
			expectedGormLevel: gormlogger.Error,
		},
		{
			name:              "fatal level maps to gorm silent",
			loggerLevel:       zapcore.FatalLevel,
			expectedGormLevel: gormlogger.Silent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLogger := &MockLogger{level: tt.loggerLevel}
			mockLogger.On("With", mock.Anything).Return(mockLogger)

			gormLogger := NewGORMLogger(mockLogger)

			assert.NotNil(t, gormLogger)
			assert.Equal(t, tt.expectedGormLevel, gormLogger.level)
			assert.Equal(t, 500*time.Millisecond, gormLogger.slowThreshold)
			assert.False(t, gormLogger.ignoreRecordNotFoundError)
			mockLogger.AssertExpectations(t)
		})
	}
}

func TestNewGORMLogger_NilLogger(t *testing.T) {
	// This test shows that NewGORMLogger will panic with nil logger
	// which is expected behavior - callers should not pass nil
	assert.Panics(t, func() {
		NewGORMLogger(nil)
	})
}

func TestGORMLogger_LogMode(t *testing.T) {
	mockLogger := &MockLogger{level: 0}
	mockLogger.On("With", mock.Anything).Return(mockLogger)

	gormLogger := NewGORMLogger(mockLogger)

	// Test when level is same
	result := gormLogger.LogMode(gormlogger.Warn)
	assert.Equal(t, gormLogger, result)

	// Test when level is different
	result = gormLogger.LogMode(gormlogger.Info)
	assert.NotEqual(t, gormLogger, result)
	assert.Equal(t, gormlogger.Info, result.(*GORMLogger).level)
}

func TestGORMLogger_Info(t *testing.T) {
	tests := []struct {
		name      string
		level     gormlogger.LogLevel
		msg       string
		data      []interface{}
		shouldLog bool
	}{
		{
			name:      "info level - should log (level >= info)",
			level:     gormlogger.Info,
			msg:       "test message",
			data:      []interface{}{},
			shouldLog: true,
		},
		{
			name:      "warn level - should not log (warn < info)",
			level:     gormlogger.Warn,
			msg:       "test message",
			data:      []interface{}{},
			shouldLog: false,
		},
		{
			name:      "error level - should not log (error < info)",
			level:     gormlogger.Error,
			msg:       "test message",
			data:      []interface{}{},
			shouldLog: false,
		},
		{
			name:      "silent level - should not log (silent < info)",
			level:     gormlogger.Silent,
			msg:       "test message",
			data:      []interface{}{},
			shouldLog: false,
		},
		{
			name:      "info with format data",
			level:     gormlogger.Info,
			msg:       "user %s created with id %d",
			data:      []interface{}{"john", 123},
			shouldLog: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLogger := &MockLogger{}
			gormLogger := &GORMLogger{
				logger: mockLogger,
				level:  tt.level,
			}

			if tt.shouldLog {
				mockLogger.On("Info", mock.AnythingOfType("string"), mock.MatchedBy(func(field Field) bool {
					return field.Key() == "file"
				})).Once()
			}

			gormLogger.Info(context.Background(), tt.msg, tt.data...)

			mockLogger.AssertExpectations(t)
		})
	}
}

func TestGORMLogger_Warn(t *testing.T) {
	tests := []struct {
		name      string
		level     gormlogger.LogLevel
		shouldLog bool
	}{
		{
			name:      "info level - should log (info >= warn)",
			level:     gormlogger.Info,
			shouldLog: true,
		},
		{
			name:      "warn level - should log (warn >= warn)",
			level:     gormlogger.Warn,
			shouldLog: true,
		},
		{
			name:      "error level - should not log (error < warn)",
			level:     gormlogger.Error,
			shouldLog: false,
		},
		{
			name:      "silent level - should not log (silent < warn)",
			level:     gormlogger.Silent,
			shouldLog: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLogger := &MockLogger{}
			gormLogger := &GORMLogger{
				logger: mockLogger,
				level:  tt.level,
			}

			if tt.shouldLog {
				mockLogger.On("Warn", mock.AnythingOfType("string"), mock.MatchedBy(func(field Field) bool {
					return field.Key() == "file"
				})).Once()
			}

			gormLogger.Warn(context.Background(), "test warning")

			mockLogger.AssertExpectations(t)
		})
	}
}

func TestGORMLogger_Error(t *testing.T) {
	tests := []struct {
		name      string
		level     gormlogger.LogLevel
		shouldLog bool
	}{
		{
			name:      "info level - should log (info >= error)",
			level:     gormlogger.Info,
			shouldLog: true,
		},
		{
			name:      "warn level - should log (warn >= error)",
			level:     gormlogger.Warn,
			shouldLog: true,
		},
		{
			name:      "error level - should log (error >= error)",
			level:     gormlogger.Error,
			shouldLog: true,
		},
		{
			name:      "silent level - should not log (silent < error)",
			level:     gormlogger.Silent,
			shouldLog: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLogger := &MockLogger{}
			gormLogger := &GORMLogger{
				logger: mockLogger,
				level:  tt.level,
			}

			if tt.shouldLog {
				mockLogger.On("Error", mock.AnythingOfType("string"), mock.MatchedBy(func(field Field) bool {
					return field.Key() == "file"
				})).Once()
			}

			gormLogger.Error(context.Background(), "test error")

			mockLogger.AssertExpectations(t)
		})
	}
}

func TestGORMLogger_Trace(t *testing.T) {
	ctx := context.Background()
	begin := time.Now().Add(-100 * time.Millisecond)

	tests := []struct {
		name                      string
		level                     gormlogger.LogLevel
		slowThreshold             time.Duration
		ignoreRecordNotFoundError bool
		err                       error
		elapsed                   time.Duration
		expectedLogLevel          string // "error", "warn", "info", "none"
	}{
		{
			name:             "silent level - no logging",
			level:            gormlogger.Silent,
			expectedLogLevel: "none",
		},
		{
			name:                      "error with record not found - ignore enabled",
			level:                     gormlogger.Error,
			err:                       gormlogger.ErrRecordNotFound,
			ignoreRecordNotFoundError: true,
			expectedLogLevel:          "none",
		},
		{
			name:                      "error with record not found - ignore disabled",
			level:                     gormlogger.Error,
			err:                       gormlogger.ErrRecordNotFound,
			ignoreRecordNotFoundError: false,
			expectedLogLevel:          "error",
		},
		{
			name:             "generic error",
			level:            gormlogger.Error,
			err:              errors.New("database error"),
			expectedLogLevel: "error",
		},
		{
			name:             "slow query",
			level:            gormlogger.Warn,
			slowThreshold:    50 * time.Millisecond,
			elapsed:          100 * time.Millisecond,
			expectedLogLevel: "warn",
		},
		{
			name:             "fast query below threshold",
			level:            gormlogger.Warn,
			slowThreshold:    200 * time.Millisecond,
			elapsed:          50 * time.Millisecond,
			expectedLogLevel: "none",
		},
		{
			name:             "normal info query",
			level:            gormlogger.Info,
			expectedLogLevel: "info",
		},
		{
			name:             "warn level with fast query",
			level:            gormlogger.Warn,
			slowThreshold:    200 * time.Millisecond,
			elapsed:          50 * time.Millisecond,
			expectedLogLevel: "none",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLogger := &MockLogger{}
			gormLogger := &GORMLogger{
				logger:                    mockLogger,
				level:                     tt.level,
				slowThreshold:             tt.slowThreshold,
				ignoreRecordNotFoundError: tt.ignoreRecordNotFoundError,
			}

			// Mock the fc function
			fc := func() (string, int64) {
				return "SELECT * FROM users WHERE id = ?", 1
			}

			// Set up expectations based on expected log level
			switch tt.expectedLogLevel {
			case "error":
				mockLogger.On("Error", mock.AnythingOfType("string"), mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once()
			case "warn":
				mockLogger.On("Warn", mock.AnythingOfType("string"), mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once()
			case "info":
				mockLogger.On("Debug", mock.AnythingOfType("string"), mock.Anything, mock.Anything, mock.Anything).Once()
			case "none":
				// No expectations set
			}

			// Calculate begin time based on elapsed
			testBegin := begin
			if tt.elapsed > 0 {
				testBegin = time.Now().Add(-tt.elapsed)
			}

			gormLogger.Trace(ctx, testBegin, fc, tt.err)

			mockLogger.AssertExpectations(t)
		})
	}
}

func TestGORMLogger_SetSlowThreshold(t *testing.T) {
	mockLogger := &MockLogger{}
	gormLogger := &GORMLogger{
		logger:                    mockLogger,
		level:                     gormlogger.Info,
		slowThreshold:             100 * time.Millisecond,
		ignoreRecordNotFoundError: false,
	}

	newThreshold := 200 * time.Millisecond
	result := gormLogger.SetSlowThreshold(newThreshold)

	assert.NotEqual(t, gormLogger, result)
	assert.Equal(t, newThreshold, result.slowThreshold)
	assert.Equal(t, gormLogger.logger, result.logger)
	assert.Equal(t, gormLogger.level, result.level)
	assert.Equal(t, gormLogger.ignoreRecordNotFoundError, result.ignoreRecordNotFoundError)
}

func TestGORMLogger_SetIgnoreRecordNotFoundError(t *testing.T) {
	mockLogger := &MockLogger{}
	gormLogger := &GORMLogger{
		logger:                    mockLogger,
		level:                     gormlogger.Info,
		slowThreshold:             100 * time.Millisecond,
		ignoreRecordNotFoundError: false,
	}

	result := gormLogger.SetIgnoreRecordNotFoundError(true)

	assert.NotEqual(t, gormLogger, result)
	assert.True(t, result.ignoreRecordNotFoundError)
	assert.Equal(t, gormLogger.logger, result.logger)
	assert.Equal(t, gormLogger.level, result.level)
	assert.Equal(t, gormLogger.slowThreshold, result.slowThreshold)
}

func TestMapLoggerLevelToGORM(t *testing.T) {
	tests := []struct {
		name          string
		logger        Logger
		expectedLevel gormlogger.LogLevel
	}{
		{
			name:          "nil logger",
			logger:        nil,
			expectedLevel: gormlogger.Warn,
		},
		{
			name:          "debug level (-1)",
			logger:        &MockLogger{level: zapcore.DebugLevel},
			expectedLevel: gormlogger.Info,
		},
		{
			name:          "info level (0)",
			logger:        &MockLogger{level: zapcore.InfoLevel},
			expectedLevel: gormlogger.Warn,
		},
		{
			name:          "warn level (1)",
			logger:        &MockLogger{level: zapcore.WarnLevel},
			expectedLevel: gormlogger.Warn,
		},
		{
			name:          "error level (2)",
			logger:        &MockLogger{level: zapcore.ErrorLevel},
			expectedLevel: gormlogger.Error,
		},
		{
			name:          "panic level (4)",
			logger:        &MockLogger{level: zapcore.PanicLevel},
			expectedLevel: gormlogger.Error,
		},
		{
			name:          "fatal level (5)",
			logger:        &MockLogger{level: zapcore.FatalLevel},
			expectedLevel: gormlogger.Silent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapLoggerLevelToGORM(tt.logger)
			assert.Equal(t, tt.expectedLevel, result)
		})
	}
}

func TestGORMLogger_FormatRowsInfo(t *testing.T) {
	gormLogger := &GORMLogger{}

	tests := []struct {
		name             string
		rows             int64
		expectedDisplay  interface{}
		expectedFieldKey string
	}{
		{
			name:             "normal rows",
			rows:             5,
			expectedDisplay:  int64(5),
			expectedFieldKey: "rows_affected",
		},
		{
			name:             "no rows affected",
			rows:             0,
			expectedDisplay:  int64(0),
			expectedFieldKey: "rows_affected",
		},
		{
			name:             "unknown rows (-1)",
			rows:             -1,
			expectedDisplay:  "-",
			expectedFieldKey: "rows_affected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			display, field := gormLogger.formatRowsInfo(tt.rows)
			assert.Equal(t, tt.expectedDisplay, display)

			// Check field key by converting to string representation
			fieldStr := fmt.Sprintf("%v", field)
			assert.Contains(t, fieldStr, tt.expectedFieldKey)
		})
	}
}

func TestGORMLogger_CreateBaseFields(t *testing.T) {
	gormLogger := &GORMLogger{}

	fileLocation := "test.go:10"
	duration := 1 * time.Second
	rowsField := String("rows_affected", "5")

	fields := gormLogger.createBaseFields(fileLocation, duration, rowsField)

	assert.Len(t, fields, 3)

	// Convert fields to string to check content
	fieldsStr := fmt.Sprintf("%v", fields)
	assert.Contains(t, fieldsStr, "file")
	assert.Contains(t, fieldsStr, "duration")
	assert.Contains(t, fieldsStr, "rows_affected")
}

func TestGORMLogger_Trace_WithDifferentRowsAffected(t *testing.T) {
	ctx := context.Background()
	begin := time.Now().Add(-10 * time.Millisecond)

	tests := []struct {
		name         string
		rowsAffected int64
	}{
		{
			name:         "multiple rows affected",
			rowsAffected: 5,
		},
		{
			name:         "single row affected",
			rowsAffected: 1,
		},
		{
			name:         "no rows affected",
			rowsAffected: 0,
		},
		{
			name:         "unknown rows affected",
			rowsAffected: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLogger := &MockLogger{}
			gormLogger := &GORMLogger{
				logger: mockLogger,
				level:  gormlogger.Info,
			}

			fc := func() (string, int64) {
				return "SELECT * FROM users", tt.rowsAffected
			}

			mockLogger.On("Debug", mock.AnythingOfType("string"), mock.Anything, mock.Anything, mock.Anything).Once()

			gormLogger.Trace(ctx, begin, fc, nil)

			mockLogger.AssertExpectations(t)
		})
	}
}

func TestGORMLogger_Trace_SlowThresholdZero(t *testing.T) {
	ctx := context.Background()
	begin := time.Now().Add(-1 * time.Second) // Very slow query

	mockLogger := &MockLogger{}
	gormLogger := &GORMLogger{
		logger:        mockLogger,
		level:         gormlogger.Warn,
		slowThreshold: 0, // Zero threshold means no slow query logging
	}

	fc := func() (string, int64) {
		return "SELECT * FROM users", 1
	}

	// Should not log as slow query because threshold is 0
	// No expectations set

	gormLogger.Trace(ctx, begin, fc, nil)

	mockLogger.AssertExpectations(t)
}
