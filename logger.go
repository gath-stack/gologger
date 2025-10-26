// Package logger provides a structured logging wrapper around Uber's Zap logger.
//
// This package offers a production-ready logging solution with:
//   - Environment-based configuration (development/production)
//   - Structured logging with strongly-typed fields
//   - Global logger instance with thread-safe initialization
//   - Convenient package-level functions
//   - Support for contextual loggers with pre-attached fields
//   - Integration-friendly design with OTEL support
//
// Basic usage:
//
//	func main() {
//	    logger.MustInitFromEnv()
//	    defer logger.Sync()
//
//	    logger.Info("application started")
//	}
//
// With contextual fields:
//
//	log := logger.With(
//	    zap.String("component", "auth"),
//	    zap.String("user_id", "12345"),
//	)
//	log.Info("user authenticated")
package logger

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gath-stack/gologger/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger wraps zap.Logger to provide additional functionality.
type Logger struct {
	*zap.Logger
}

var (
	globalLogger *Logger
	mu           sync.RWMutex
)

// buildLogger constructs a zap.Logger based on the provided configuration.
func buildLogger(cfg config.LoggerConfig) (*zap.Logger, error) {
	// Parse log level
	var level zapcore.Level
	switch cfg.Level {
	case config.LogLevelDebug:
		level = zapcore.DebugLevel
	case config.LogLevelInfo:
		level = zapcore.InfoLevel
	case config.LogLevelWarn:
		level = zapcore.WarnLevel
	case config.LogLevelError:
		level = zapcore.ErrorLevel
	default:
		return nil, fmt.Errorf("%w: %s", ErrInvalidLogLevel, cfg.Level)
	}

	// Build encoder config
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Choose encoder based on environment
	var encoder zapcore.Encoder
	if cfg.Environment == config.EnvProduction {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// Create core
	core := zapcore.NewCore(
		encoder,
		zapcore.AddSync(os.Stdout),
		level,
	)

	// Build logger with options
	logger := zap.New(core,
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.ErrorLevel),
		zap.Fields(zap.String("service", cfg.ServiceName)),
	)

	return logger, nil
}

// validateConfig validates the logger configuration.
func validateConfig(cfg config.LoggerConfig) error {
	if err := cfg.Level.Validate(); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidLogLevel, err)
	}
	if err := cfg.Environment.Validate(); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidEnvironment, err)
	}
	if strings.TrimSpace(cfg.ServiceName) == "" {
		return ErrMissingServiceName
	}
	return nil
}

// InitGlobal initializes the global logger with the provided configuration.
//
// This function can only be called once. Subsequent calls will return
// ErrAlreadyInitialized. The initialization is thread-safe.
//
// Example:
//
//	cfg := logger.LoggerConfig{
//	    Level:       logger.LogLevelInfo,
//	    Environment: logger.EnvProduction,
//	    ServiceName: "my-service",
//	}
//	if err := logger.InitGlobal(cfg); err != nil {
//	    log.Fatalf("failed to initialize logger: %v", err)
//	}
func InitGlobal(cfg config.LoggerConfig) error {
	mu.Lock()
	defer mu.Unlock()

	// Check if already initialized
	if globalLogger != nil {
		return ErrAlreadyInitialized
	}

	// Validate config
	if err := validateConfig(cfg); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidConfig, err)
	}

	// Build logger
	zapLogger, err := buildLogger(cfg)
	if err != nil {
		return fmt.Errorf("failed to build logger: %w", err)
	}

	globalLogger = &Logger{Logger: zapLogger}
	return nil
}

// Get returns the global logger instance.
//
// This function panics if the logger has not been initialized.
// Use TryGet() for a non-panicking version.
//
// Example:
//
//	log := logger.Get()
//	log.Info("application started")
func Get() *Logger {
	mu.RLock()
	defer mu.RUnlock()
	if globalLogger == nil {
		panic(ErrNotInitialized)
	}
	return globalLogger
}

// TryGet returns the global logger instance and an error if not initialized.
//
// This is a non-panicking alternative to Get() that is useful in library code
// or when you want to handle the uninitialized case gracefully.
//
// Example:
//
//	log, err := logger.TryGet()
//	if err != nil {
//	    if errors.Is(err, logger.ErrNotInitialized) {
//	        return fmt.Errorf("logger not initialized: %w", err)
//	    }
//	    return err
//	}
//	log.Info("doing something")
func TryGet() (*Logger, error) {
	mu.RLock()
	defer mu.RUnlock()
	if globalLogger == nil {
		return nil, ErrNotInitialized
	}
	return globalLogger, nil
}

// With returns a derived logger enriched with additional structured fields.
//
// The returned logger is a new instance and does not modify the original logger.
//
// Example:
//
//	log := logger.Get().With(zap.String("user_id", "abc123"))
//	log.Info("User login succeeded")
func (l *Logger) With(fields ...zap.Field) *Logger {
	return &Logger{Logger: l.Logger.With(fields...)}
}

// Sync flushes any buffered log entries to the underlying writer.
//
// This should be called before program exit to avoid data loss.
// The function ignores known benign sync errors on stderr/stdout.
//
// Example:
//
//	defer logger.Sync()
func Sync() error {
	log, err := TryGet()
	if err != nil {
		return err
	}
	return log.Sync()
}

// Sync flushes any buffered log entries for this logger instance.
func (l *Logger) Sync() error {
	if err := l.Logger.Sync(); err != nil {
		// Ignore known benign sync errors on stderr/stdout
		if !isIgnorableSyncError(err) {
			return fmt.Errorf("%w: %v", ErrSyncFailed, err)
		}
	}
	return nil
}

// isIgnorableSyncError returns true for sync errors that can be safely ignored.
// Zap can fail on /dev/stderr in some operating systems.
func isIgnorableSyncError(err error) bool {
	return errors.Is(err, syscall.EINVAL) ||
		errors.Is(err, syscall.ENOTTY) ||
		errors.Is(err, syscall.EBADF)
}

// Debug logs a message at the DEBUG level using the global logger.
func Debug(msg string, fields ...zap.Field) {
	Get().Debug(msg, fields...)
}

// Info logs a message at the INFO level using the global logger.
func Info(msg string, fields ...zap.Field) {
	Get().Info(msg, fields...)
}

// Warn logs a message at the WARN level using the global logger.
func Warn(msg string, fields ...zap.Field) {
	Get().Warn(msg, fields...)
}

// Error logs a message at the ERROR level using the global logger.
func Error(msg string, fields ...zap.Field) {
	Get().Error(msg, fields...)
}

// Fatal logs a message at the FATAL level and terminates the application.
//
// Use this sparingly—prefer returning errors whenever possible.
func Fatal(msg string, fields ...zap.Field) {
	Get().Fatal(msg, fields...)
}

// With creates a derived logger with pre-attached structured fields using the global logger.
//
// Example:
//
//	log := logger.With(zap.String("component", "auth"))
//	log.Info("Authentication service started")
func With(fields ...zap.Field) *Logger {
	return Get().With(fields...)
}

// UnderlyingLogger returns the underlying zap.Logger for advanced integrations.
//
// UNSTABLE API: This method exposes internal implementation details and may
// change in future versions. Use only when you need direct access to zap.Logger
// for advanced features like OTLP log export or custom cores.
//
// Example:
//
//	log := logger.Get()
//	zapLogger := log.UnderlyingLogger()
//	// Use zapLogger for advanced operations
func (l *Logger) UnderlyingLogger() *zap.Logger {
	return l.Logger
}

// WithCore creates a new logger with the specified core.
//
// UNSTABLE API: This method is for advanced use cases and may change.
// Use this when you need to replace or wrap the logger's core, such as
// adding additional outputs (like OTLP) while maintaining existing configuration.
//
// The returned logger is a new instance with the new core.
//
// Example:
//
//	log := logger.Get()
//	currentCore := log.UnderlyingLogger().Core()
//	otelCore := createOTELCore()
//	teeCore := zapcore.NewTee(currentCore, otelCore)
//	newLog := log.WithCore(teeCore)
//	newLog.Info("This goes to both console and OTLP")
func (l *Logger) WithCore(core zapcore.Core) *Logger {
	newLogger := zap.New(core,
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
	return &Logger{Logger: newLogger}
}

// WithOTELCore creates a new logger that sends logs to both console and OTLP.
//
// UNSTABLE API: This method is for advanced use cases and may change.
// This is a convenience method for adding OTLP export to the logger.
// The returned logger will write to both the original output and OTLP.
//
// Example:
//
//	log := logger.Get()
//	otelCore := createOTELCore()
//	newLog := log.WithOTELCore(otelCore)
//	newLog.Info("This goes to both console and Loki")
func (l *Logger) WithOTELCore(otelCore zapcore.Core) *Logger {
	currentCore := l.Logger.Core()
	teeCore := zapcore.NewTee(currentCore, otelCore)
	return l.WithCore(teeCore)
}

// InitFromEnv initializes the global logger using environment variables.
//
// This is a convenience function that loads configuration from environment
// and initializes the logger in one call. It's the recommended way to
// initialize the logger in main().
//
// Required environment variables:
//   - LOG_LEVEL: DEBUG, INFO, WARN, ERROR
//   - APP_ENV: development, production
//   - APP_NAME: your service name
//
// Returns an error if any required variable is missing or invalid.
//
// Example:
//
//	if err := logger.InitFromEnv(); err != nil {
//	    log.Fatalf("failed to initialize logger: %v", err)
//	}
//	defer logger.Sync()
func InitFromEnv() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}
	return InitGlobal(cfg.Logger)
}

// MustInitFromEnv initializes the global logger using environment variables.
//
// This function panics if initialization fails. Use this in main() for
// fail-fast behavior during application startup.
//
// Required environment variables:
//   - LOG_LEVEL: DEBUG, INFO, WARN, ERROR
//   - APP_ENV: development, production
//   - APP_NAME: your service name
//
// Example:
//
//	func main() {
//	    logger.MustInitFromEnv()
//	    defer logger.Sync()
//
//	    logger.Info("application started")
//	}
func MustInitFromEnv() {
	if err := InitFromEnv(); err != nil {
		panic(fmt.Sprintf("failed to initialize logger from environment: %v", err))
	}
}

// InitWithDefaults initializes the global logger with sensible defaults.
//
// This is useful for quick setup during development or testing.
// Default configuration:
//   - Level: INFO
//   - Environment: development
//   - ServiceName: "app"
//
// Example:
//
//	logger.InitWithDefaults()
//	defer logger.Sync()
func InitWithDefaults() error {
	cfg := config.LoggerConfig{
		Level:       config.LogLevelInfo,
		Environment: config.EnvDevelopment,
		ServiceName: "app",
	}
	return InitGlobal(cfg)
}

// SyncWithTimeout flushes log entries with a timeout.
//
// This is useful during graceful shutdown when you want to ensure
// logs are flushed but don't want to wait indefinitely.
//
// Example:
//
//	if err := logger.SyncWithTimeout(5 * time.Second); err != nil {
//	    fmt.Fprintf(os.Stderr, "sync timeout: %v\n", err)
//	}
func SyncWithTimeout(timeout time.Duration) error {
	done := make(chan error, 1)
	go func() {
		done <- Sync()
	}()

	select {
	case err := <-done:
		return err
	case <-time.After(timeout):
		return fmt.Errorf("sync timeout after %v", timeout)
	}
}

// SyncWithTimeout flushes log entries for this logger instance with a timeout.
func (l *Logger) SyncWithTimeout(timeout time.Duration) error {
	done := make(chan error, 1)
	go func() {
		done <- l.Sync()
	}()

	select {
	case err := <-done:
		return err
	case <-time.After(timeout):
		return fmt.Errorf("sync timeout after %v", timeout)
	}
}

// ReplaceGlobal replaces the global logger with a new instance.
//
// This is useful when you need to enhance the logger after initialization,
// such as adding OTLP export or other integrations.
//
// CAUTION: This function should be used sparingly and only during application
// startup or configuration phases. Replacing the logger during normal operation
// may cause unexpected behavior.
//
// Example:
//
//	log := logger.Get()
//	enhancedLog := log.WithOTELCore(otelCore)
//	logger.ReplaceGlobal(enhancedLog)
func ReplaceGlobal(newLogger *Logger) {
	mu.Lock()
	defer mu.Unlock()
	globalLogger = newLogger
}
