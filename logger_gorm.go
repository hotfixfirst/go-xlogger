package xlogger

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"go.uber.org/zap/zapcore"
	gormlogger "gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
)

// Pre-compiled regex for better performance
var whitespaceRegex = regexp.MustCompile(`\s+`)

// GORMLogger implements gorm.logger.Interface using our Logger
type GORMLogger struct {
	logger                    Logger
	level                     gormlogger.LogLevel
	slowThreshold             time.Duration
	ignoreRecordNotFoundError bool
	maxFilePathLevels         int
}

// NewGORMLogger creates a new GORM logger adapter with sensible defaults
func NewGORMLogger(logger Logger) *GORMLogger {
	gormLevel := mapLoggerLevelToGORM(logger)
	return &GORMLogger{
		logger:                    logger.With(String("component", "gorm")),
		level:                     gormLevel,
		slowThreshold:             500 * time.Millisecond,
		ignoreRecordNotFoundError: false,
		maxFilePathLevels:         3,
	}
}

// LogMode implements gorm.logger.Interface
func (l *GORMLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	if l.level == level {
		return l
	}
	return &GORMLogger{
		logger:                    l.logger,
		level:                     level,
		slowThreshold:             l.slowThreshold,
		ignoreRecordNotFoundError: l.ignoreRecordNotFoundError,
		maxFilePathLevels:         l.maxFilePathLevels,
	}
}

// Info implements gorm.logger.Interface
func (l *GORMLogger) Info(_ context.Context, msg string, data ...interface{}) {
	if l.level >= gormlogger.Info {
		l.logger.Info(fmt.Sprintf(msg, data...), String("file", l.shortFileLocation(utils.FileWithLineNum())))
	}
}

// Warn implements gorm.logger.Interface
func (l *GORMLogger) Warn(_ context.Context, msg string, data ...interface{}) {
	if l.level >= gormlogger.Warn {
		l.logger.Warn(fmt.Sprintf(msg, data...), String("file", l.shortFileLocation(utils.FileWithLineNum())))
	}
}

// Error implements gorm.logger.Interface
func (l *GORMLogger) Error(_ context.Context, msg string, data ...interface{}) {
	if l.level >= gormlogger.Error {
		l.logger.Error(fmt.Sprintf(msg, data...), String("file", l.shortFileLocation(utils.FileWithLineNum())))
	}
}

// shortFileLocation limits file path based on maxPathLevels configuration
func (l *GORMLogger) shortFileLocation(fileWithLine string) string {
	if fileWithLine == "" {
		return ""
	}

	// if maxPathLevels < 1 full path is returned
	if l.maxFilePathLevels < 1 {
		return fileWithLine
	}

	// Split path by separator to get parts
	parts := strings.Split(fileWithLine, "/")

	// If we have maxPathLevels or fewer parts, return as is
	if len(parts) <= l.maxFilePathLevels {
		return fileWithLine
	}

	// Take last maxPathLevels parts
	shortParts := parts[len(parts)-l.maxFilePathLevels:]
	return strings.Join(shortParts, "/")
}

// cleanSQLForLogging cleans SQL query for single-line logging by removing newlines and extra whitespace.
func (l *GORMLogger) cleanSQLForLogging(sql string) string {
	// Early return for empty strings to avoid unnecessary processing
	if sql == "" {
		return sql
	}

	// For very large SQL queries, use more efficient approach
	if len(sql) > 1024 {
		var builder strings.Builder
		builder.Grow(len(sql)) // Pre-allocate capacity
		for _, char := range sql {
			switch char {
			case '\n', '\r', '\t':
				builder.WriteByte(' ')
			default:
				builder.WriteRune(char)
			}
		}
		sql = builder.String()
	} else {
		// Use simple replace for shorter SQL
		sql = strings.ReplaceAll(sql, "\n", " ")
		sql = strings.ReplaceAll(sql, "\r", " ")
		sql = strings.ReplaceAll(sql, "\t", " ")
	}

	// Remove duplicate whitespace using pre-compiled regex
	sql = whitespaceRegex.ReplaceAllString(sql, " ")

	// Trim leading and trailing spaces
	return strings.TrimSpace(sql)
}

// Trace implements gorm.logger.Interface for SQL query logging
func (l *GORMLogger) Trace(_ context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.level <= gormlogger.Silent {
		return
	}

	duration := time.Since(begin)
	fileLocation := l.shortFileLocation(utils.FileWithLineNum())

	switch {
	case err != nil && l.level >= gormlogger.Error && (!errors.Is(err, gormlogger.ErrRecordNotFound) || !l.ignoreRecordNotFoundError):
		// Error case: get SQL only when needed
		sql, rows := fc()
		cleanSQL := l.cleanSQLForLogging(sql)
		rowsDisplay, rowsField := l.formatRowsInfo(rows)
		baseFields := l.createBaseFields(fileLocation, duration, rowsField)
		logMsg := fmt.Sprintf("[%s] [rows:%v] %s", duration.String(), rowsDisplay, cleanSQL)
		l.logger.Error(logMsg, append(baseFields, Error(err))...)

	case duration > l.slowThreshold && l.slowThreshold != 0 && l.level >= gormlogger.Warn:
		// Slow query case: get SQL only when needed
		sql, rows := fc()
		cleanSQL := l.cleanSQLForLogging(sql)
		rowsDisplay, rowsField := l.formatRowsInfo(rows)
		baseFields := l.createBaseFields(fileLocation, duration, rowsField)
		slowMsg := fmt.Sprintf("SLOW SQL >= %v", l.slowThreshold)
		logMsg := fmt.Sprintf("%s [%s] [rows:%v] %s", slowMsg, duration.String(), rowsDisplay, cleanSQL)
		l.logger.Warn(logMsg, append(baseFields, Duration("slow_threshold", l.slowThreshold), Bool("is_slow", true))...)

	case l.level == gormlogger.Info:
		// Normal case: get SQL only when needed
		sql, rows := fc()
		cleanSQL := l.cleanSQLForLogging(sql)
		rowsDisplay, rowsField := l.formatRowsInfo(rows)
		baseFields := l.createBaseFields(fileLocation, duration, rowsField)
		logMsg := fmt.Sprintf("[%s] [rows:%v] %s", duration.String(), rowsDisplay, cleanSQL)
		l.logger.Debug(logMsg, baseFields...)
	}
}

// SetSlowThreshold configures slow query threshold
func (l *GORMLogger) SetSlowThreshold(threshold time.Duration) *GORMLogger {
	return &GORMLogger{
		logger:                    l.logger,
		level:                     l.level,
		slowThreshold:             threshold,
		ignoreRecordNotFoundError: l.ignoreRecordNotFoundError,
		maxFilePathLevels:         l.maxFilePathLevels,
	}
}

// SetIgnoreRecordNotFoundError configures whether to ignore ErrRecordNotFound
func (l *GORMLogger) SetIgnoreRecordNotFoundError(ignore bool) *GORMLogger {
	return &GORMLogger{
		logger:                    l.logger,
		level:                     l.level,
		slowThreshold:             l.slowThreshold,
		ignoreRecordNotFoundError: ignore,
		maxFilePathLevels:         l.maxFilePathLevels,
	}
}

// SetMaxPathLevels configures maximum path levels to display (-1 = show "_", 0 = show full path)
func (l *GORMLogger) SetMaxPathLevels(levels int) *GORMLogger {
	return &GORMLogger{
		logger:                    l.logger,
		level:                     l.level,
		slowThreshold:             l.slowThreshold,
		ignoreRecordNotFoundError: l.ignoreRecordNotFoundError,
		maxFilePathLevels:         levels,
	}
}

// mapLoggerLevelToGORM maps logger level to GORM level
// Map zap levels to GORM levels
// DebugLevel(-1) -> Info (log all SQL queries)
// InfoLevel(0)   -> Warn (log slow queries and errors only)
// WarnLevel(1)   -> Warn (log slow queries and errors only)
// ErrorLevel(2)  -> Error (log errors only)
// DPanicLevel(3) -> Error (log errors only)
// PanicLevel(4)  -> Error (log errors only)
// FatalLevel(5)  -> Silent (no logging)
func mapLoggerLevelToGORM(logger Logger) gormlogger.LogLevel {
	if logger == nil {
		return gormlogger.Warn
	}
	loggerLevel := logger.Level()
	switch {
	case loggerLevel <= zapcore.DebugLevel:
		return gormlogger.Info
	case loggerLevel <= zapcore.WarnLevel:
		return gormlogger.Warn
	case loggerLevel <= zapcore.PanicLevel:
		return gormlogger.Error
	default:
		return gormlogger.Silent
	}
}

// formatRowsInfo formats rows display and creates appropriate field
func (l *GORMLogger) formatRowsInfo(rows int64) (interface{}, Field) {
	if rows == -1 {
		return "-", String("rows_affected", "-")
	}
	return rows, Int64("rows_affected", rows)
}

// createBaseFields creates base logging fields for structured logging
func (l *GORMLogger) createBaseFields(fileLocation string, duration time.Duration, rowsField Field) []Field {
	return []Field{
		String("file", fileLocation),
		Duration("duration", duration),
		rowsField,
	}
}
