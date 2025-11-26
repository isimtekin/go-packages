package logger

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

// Logger wraps logrus.Logger with additional functionality
type Logger struct {
	config *Config
	logger *logrus.Logger
	fields logrus.Fields
}

// Field represents a key-value pair for structured logging
type Field struct {
	Key   string
	Value interface{}
}

var (
	// globalLogger is the default logger instance
	globalLogger *Logger
)

func init() {
	// Initialize with default config
	globalLogger, _ = New(DefaultConfig())
}

// New creates a new logger instance with the given configuration
func New(config *Config) (*Logger, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	l := &Logger{
		config: config,
		logger: logrus.New(),
		fields: make(logrus.Fields),
	}

	// Set log level
	level, err := logrus.ParseLevel(config.Level)
	if err != nil {
		return nil, fmt.Errorf("invalid log level: %w", err)
	}
	l.logger.SetLevel(level)

	// Set formatter
	if config.Format == "json" {
		l.logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime:  "@timestamp",
				logrus.FieldKeyLevel: "level",
				logrus.FieldKeyMsg:   "message",
			},
		})
	} else {
		l.logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
			ForceColors:     config.ConsoleColors,
		})
	}

	// Set caller reporting
	l.logger.SetReportCaller(config.ReportCaller)

	// Add default fields
	if config.ServiceName != "" {
		l.fields["service"] = config.ServiceName
	}
	if config.Environment != "" {
		l.fields["environment"] = config.Environment
	}
	if config.Version != "" {
		l.fields["version"] = config.Version
	}
	if config.Host != "" {
		l.fields["host"] = config.Host
	} else {
		if hostname, err := os.Hostname(); err == nil {
			l.fields["host"] = hostname
		}
	}

	// Configure output
	if err := l.configureOutput(); err != nil {
		return nil, fmt.Errorf("failed to configure output: %w", err)
	}

	// If logging is disabled, set output to discard
	if !config.Enabled {
		l.logger.SetOutput(io.Discard)
	}

	return l, nil
}

// configureOutput sets up the output destination based on config
func (l *Logger) configureOutput() error {
	switch l.config.Output {
	case "console":
		l.logger.SetOutput(os.Stdout)
	case "logstash":
		return l.setupLogstashOutput()
	case "file":
		return l.setupFileOutput()
	case "multi":
		return l.setupMultiOutput()
	default:
		l.logger.SetOutput(os.Stdout)
	}
	return nil
}

// SetGlobal sets the global logger instance
func SetGlobal(logger *Logger) {
	globalLogger = logger
}

// GetGlobal returns the global logger instance
func GetGlobal() *Logger {
	return globalLogger
}

// WithFields returns a new logger with additional fields
func (l *Logger) WithFields(fields ...Field) *Logger {
	newFields := make(logrus.Fields)
	// Copy existing fields
	for k, v := range l.fields {
		newFields[k] = v
	}
	// Add new fields
	for _, f := range fields {
		newFields[f.Key] = f.Value
	}

	newLogger := &Logger{
		config: l.config,
		logger: l.logger,
		fields: newFields,
	}
	return newLogger
}

// WithField adds a single field to the logger
func (l *Logger) WithField(key string, value interface{}) *Logger {
	return l.WithFields(Field{Key: key, Value: value})
}

// WithContext adds request context fields
func (l *Logger) WithContext(ctx context.Context) *Logger {
	// Extract common context values
	newLogger := l
	if reqID := ctx.Value("request_id"); reqID != nil {
		newLogger = newLogger.WithField("request_id", reqID)
	}
	if userID := ctx.Value("user_id"); userID != nil {
		newLogger = newLogger.WithField("user_id", userID)
	}
	return newLogger
}

// IsEnabled returns whether logging is enabled
func (l *Logger) IsEnabled() bool {
	return l.config.Enabled
}

// IsDebugEnabled returns whether debug logging is enabled
func (l *Logger) IsDebugEnabled() bool {
	return l.config.Enabled && l.logger.IsLevelEnabled(logrus.DebugLevel)
}

// GetLogger returns the underlying logrus logger
func (l *Logger) GetLogger() *logrus.Logger {
	return l.logger
}

// entry returns a logrus entry with the logger's fields
func (l *Logger) entry() *logrus.Entry {
	if !l.config.Enabled {
		// Return a discarded entry
		discardLogger := logrus.New()
		discardLogger.SetOutput(io.Discard)
		return logrus.NewEntry(discardLogger)
	}
	return l.logger.WithFields(l.fields)
}

// Trace logs a message at trace level
func (l *Logger) Trace(msg string, fields ...Field) {
	if !l.config.Enabled {
		return
	}
	l.WithFields(fields...).entry().Trace(msg)
}

// Debug logs a message at debug level
func (l *Logger) Debug(msg string, fields ...Field) {
	if !l.config.Enabled {
		return
	}
	l.WithFields(fields...).entry().Debug(msg)
}

// Info logs a message at info level
func (l *Logger) Info(msg string, fields ...Field) {
	if !l.config.Enabled {
		return
	}
	l.WithFields(fields...).entry().Info(msg)
}

// Warn logs a message at warn level
func (l *Logger) Warn(msg string, fields ...Field) {
	if !l.config.Enabled {
		return
	}
	l.WithFields(fields...).entry().Warn(msg)
}

// Error logs a message at error level
func (l *Logger) Error(msg string, fields ...Field) {
	if !l.config.Enabled {
		return
	}
	l.WithFields(fields...).entry().Error(msg)
}

// Fatal logs a message at fatal level and exits
func (l *Logger) Fatal(msg string, fields ...Field) {
	if !l.config.Enabled {
		os.Exit(1)
		return
	}
	l.WithFields(fields...).entry().Fatal(msg)
}

// Panic logs a message at panic level and panics
func (l *Logger) Panic(msg string, fields ...Field) {
	if !l.config.Enabled {
		panic(msg)
	}
	l.WithFields(fields...).entry().Panic(msg)
}

// Helper functions for creating fields

// String creates a string field
func String(key, value string) Field {
	return Field{Key: key, Value: value}
}

// Int creates an int field
func Int(key string, value int) Field {
	return Field{Key: key, Value: value}
}

// Int64 creates an int64 field
func Int64(key string, value int64) Field {
	return Field{Key: key, Value: value}
}

// Float64 creates a float64 field
func Float64(key string, value float64) Field {
	return Field{Key: key, Value: value}
}

// Bool creates a bool field
func Bool(key string, value bool) Field {
	return Field{Key: key, Value: value}
}

// Err creates an error field
func Err(err error) Field {
	return Field{Key: "error", Value: err.Error()}
}

// Any creates a field with any value
func Any(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}

// Global logger convenience functions

// Trace logs to the global logger at trace level
func Trace(msg string, fields ...Field) {
	globalLogger.Trace(msg, fields...)
}

// Debug logs to the global logger at debug level
func Debug(msg string, fields ...Field) {
	globalLogger.Debug(msg, fields...)
}

// Info logs to the global logger at info level
func Info(msg string, fields ...Field) {
	globalLogger.Info(msg, fields...)
}

// Warn logs to the global logger at warn level
func Warn(msg string, fields ...Field) {
	globalLogger.Warn(msg, fields...)
}

// Error logs to the global logger at error level
func Error(msg string, fields ...Field) {
	globalLogger.Error(msg, fields...)
}

// Fatal logs to the global logger at fatal level
func Fatal(msg string, fields ...Field) {
	globalLogger.Fatal(msg, fields...)
}

// Panic logs to the global logger at panic level
func Panic(msg string, fields ...Field) {
	globalLogger.Panic(msg, fields...)
}
