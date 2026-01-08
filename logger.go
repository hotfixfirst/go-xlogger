package xlogger

import (
	"time"

	"go.uber.org/fx/fxevent"
	"go.uber.org/zap/zapcore"
)

// Logger represents the main logging interface for structured logging
// This interface provides both application and infrastructure logging capabilities
type Logger interface {
	// Core logging methods for different levels
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)

	// These methods will terminate the application after logging
	Panic(msg string, fields ...Field)
	Fatal(msg string, fields ...Field)

	// Logger enhancement methods
	With(fields ...Field) Logger

	// Infrastructure optimization methods
	ForInfra(component string) Logger
	ForFxEvent() fxevent.Logger
	ForGORM() *GORMLogger

	// Logger configuration methods
	Level() zapcore.Level

	// Utility methods
	Sync() error
}

// Field represents a structured log field with key-value pairs
// This abstraction allows for type-safe logging without coupling to zap
type Field struct {
	key   string
	value interface{}
	typ   FieldType
}

// FieldType represents the type of a log field for optimization
type FieldType int

const (
	StringType FieldType = iota
	IntType
	Float64Type
	BoolType
	ErrorType
	DurationType
	TimeType
	AnyType
)

// String creates a string field
func String(key, value string) Field {
	return Field{key: key, value: value, typ: StringType}
}

// Int creates an integer field
func Int(key string, value int) Field {
	return Field{key: key, value: value, typ: IntType}
}

// Int64 creates an int64 field
func Int64(key string, value int64) Field {
	return Field{key: key, value: value, typ: IntType}
}

// Float64 creates a float64 field
func Float64(key string, value float64) Field {
	return Field{key: key, value: value, typ: Float64Type}
}

// Bool creates a boolean field
func Bool(key string, value bool) Field {
	return Field{key: key, value: value, typ: BoolType}
}

// Error creates an error field
func Error(err error) Field {
	return NamedError("error", err)
}

// NamedError creates an error field with a specific name
func NamedError(key string, err error) Field {
	return Field{key: key, value: err, typ: ErrorType}
}

// Duration creates a time.Duration field
func Duration(key string, value time.Duration) Field {
	return Field{key: key, value: value, typ: DurationType}
}

// Time creates a time.Time field
func Time(key string, value time.Time) Field {
	return Field{key: key, value: value, typ: TimeType}
}

// Any creates a field for any type (use sparingly for performance)
func Any(key string, value interface{}) Field {
	return Field{key: key, value: value, typ: AnyType}
}

// Getter methods for Field (for internal use)

// Key returns the field key
func (f Field) Key() string {
	return f.key
}

// Value returns the field value
func (f Field) Value() interface{} {
	return f.value
}

// Type returns the field type
func (f Field) Type() FieldType {
	return f.typ
}
