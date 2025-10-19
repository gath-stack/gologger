// Package logger initializes and manages structured logging for the application.
//
// This package provides a unified and high-performance logging abstraction built on top of
// Uber's zap library. It supports both human-readable console logs (for development)
// and JSON-formatted structured logs (for production), which can be directly ingested
// by observability backends such as Loki or Elasticsearch.
//
// The logger is environment-aware, with configurable log levels and output encoders,
// and supports both global and contextual loggers for flexible usage across modules.
//
// Key features:
//   - Fast, structured, leveled logging using zap.
//   - JSON output in production for seamless integration with Loki and other log pipelines.
//   - Colorized console output in development for easier debugging.
//   - Global singleton logger for simple application-wide access.
//   - Contextual field injection for structured log enrichment.
//   - Strict configuration validation to prevent runtime issues in production.
//
// Example usage:
//
//	func main() {
//	    cfg, err := logger.FromEnv()
//	    if err != nil {
//	        panic(fmt.Sprintf("invalid logger configuration: %v", err))
//	    }
//	    if err := logger.InitGlobal(cfg); err != nil {
//	        panic(fmt.Sprintf("failed to initialize logger: %v", err))
//	    }
//	    defer logger.Get().Sync()
//
//	    log := logger.Get()
//	    log.Info("application started", zap.String("version", "1.0.0"))
//	}
//
// Production recommendations:
//   - Always initialize the global logger early in application startup.
//   - In production, prefer JSON encoding for structured log ingestion.
//   - Call `Sync()` before process exit to flush any buffered log entries.
//   - Ensure all required environment variables are set before starting the application.
package logger

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger wraps zap.Logger to provide application-specific structured logging functionality.
//
// It supports contextual enrichment via `WithContext()` and integrates with the
// global logger pattern used throughout the application.
type Logger struct {
	*zap.Logger
}

// LogLevel represents the verbosity level for the logger.
type LogLevel string

const (
	// LevelDebug enables detailed debug and above level logging.
	LevelDebug LogLevel = "DEBUG"
	// LevelInfo enables informational and above level logging (default).
	LevelInfo LogLevel = "INFO"
	// LevelWarn enables warning and above level logging.
	LevelWarn LogLevel = "WARN"
	// LevelError enables error level logging only.
	LevelError LogLevel = "ERROR"
)

// Environment represents the deployment environment.
type Environment string

const (
	EnvDevelopment Environment = "development"
	EnvProduction  Environment = "production"
)

var (
	// globalLogger is the shared singleton logger instance for the application.
	globalLogger *Logger

	// ErrInvalidLogLevel is returned when an unsupported log level is provided.
	ErrInvalidLogLevel = errors.New("invalid log level: must be DEBUG, INFO, WARN, or ERROR")

	// ErrInvalidEnvironment is returned when an unsupported environment is provided.
	ErrInvalidEnvironment = errors.New("invalid environment: must be 'development' or 'production'")

	// ErrMissingServiceName is returned when the service name is empty.
	ErrMissingServiceName = errors.New("service name is required and cannot be empty")

	// ErrMissingRequiredEnvVar is returned when a required environment variable is not set.
	ErrMissingRequiredEnvVar = errors.New("required environment variable is not set")
)

// Config defines the configuration parameters for the logger.
//
// All fields are validated before creating a logger instance.
type Config struct {
	Level       LogLevel
	Environment Environment
	ServiceName string
}

// Validate checks if the configuration is valid for production use.
//
// Returns an error if any required field is missing or contains invalid values.
func (c Config) Validate() error {
	// Validate log level
	switch c.Level {
	case LevelDebug, LevelInfo, LevelWarn, LevelError:
		// Valid
	default:
		return fmt.Errorf("%w: got '%s'", ErrInvalidLogLevel, c.Level)
	}

	// Validate environment
	switch c.Environment {
	case EnvDevelopment, EnvProduction:
		// Valid
	default:
		return fmt.Errorf("%w: got '%s'", ErrInvalidEnvironment, c.Environment)
	}

	// Validate service name
	if strings.TrimSpace(c.ServiceName) == "" {
		return ErrMissingServiceName
	}

	return nil
}

// New creates a new logger instance according to the given configuration.
//
// The configuration is validated before creating the logger. If validation fails,
// an error is returned and the application should not proceed.
//
// In production mode, logs are formatted as structured JSON suitable for ingestion by Loki,
// FluentBit, or Elasticsearch. In development mode, logs use a colorized console encoder.
//
// Example:
//
//	logger, err := logger.New(logger.Config{
//	    Level:       logger.LevelDebug,
//	    Environment: logger.EnvProduction,
//	    ServiceName: "api-service",
//	})
//	if err != nil {
//	    panic(err)
//	}
func New(cfg Config) (*Logger, error) {
	// Validate configuration first
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid logger configuration: %w", err)
	}

	var zapLevel zapcore.Level

	// Map custom log levels to zap internal levels
	switch cfg.Level {
	case LevelDebug:
		zapLevel = zapcore.DebugLevel
	case LevelInfo:
		zapLevel = zapcore.InfoLevel
	case LevelWarn:
		zapLevel = zapcore.WarnLevel
	case LevelError:
		zapLevel = zapcore.ErrorLevel
	default:
		// This should never happen due to validation, but included for safety
		zapLevel = zapcore.InfoLevel
	}

	var zapConfig zap.Config
	if cfg.Environment == EnvProduction {
		zapConfig = zap.Config{
			Level:            zap.NewAtomicLevelAt(zapLevel),
			Development:      false,
			Encoding:         "json",
			EncoderConfig:    productionEncoderConfig(),
			OutputPaths:      []string{"stdout"},
			ErrorOutputPaths: []string{"stderr"},
		}
	} else {
		zapConfig = zap.Config{
			Level:            zap.NewAtomicLevelAt(zapLevel),
			Development:      true,
			Encoding:         "console",
			EncoderConfig:    developmentEncoderConfig(),
			OutputPaths:      []string{"stdout"},
			ErrorOutputPaths: []string{"stderr"},
		}
	}

	zapLogger, err := zapConfig.Build(
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to build logger: %w", err)
	}

	zapLogger = zapLogger.With(
		zap.String("service", cfg.ServiceName),
		zap.String("environment", string(cfg.Environment)),
	)

	return &Logger{Logger: zapLogger}, nil
}

// productionEncoderConfig defines the encoder settings for production JSON logs.
//
// The output schema is compatible with Loki and other structured logging systems.
func productionEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
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
}

// developmentEncoderConfig defines the encoder settings for development console logs.
//
// The output is colorized and human-readable for local debugging convenience.
func developmentEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		TimeKey:        "T",
		LevelKey:       "L",
		NameKey:        "N",
		CallerKey:      "C",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "M",
		StacktraceKey:  "S",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
}

// InitGlobal initializes the global singleton logger.
//
// This should be called during application startup to make the logger globally accessible.
// It replaces any existing global logger instance.
//
// The configuration is validated before initialization. If validation fails, an error
// is returned and the application should terminate.
func InitGlobal(cfg Config) error {
	logger, err := New(cfg)
	if err != nil {
		return err
	}
	globalLogger = logger
	return nil
}

// Get retrieves the global logger instance.
//
// Panics if no global logger has been initialized via InitGlobal().
// This is intentional to catch configuration errors early in production environments.
func Get() *Logger {
	if globalLogger == nil {
		panic("logger not initialized: call logger.InitGlobal() during application startup")
	}
	return globalLogger
}

// WithContext returns a derived logger enriched with additional structured fields.
//
// Example:
//
//	log := logger.Get().WithContext(zap.String("user_id", "abc123"))
//	log.Info("User login succeeded")
func (l *Logger) WithContext(fields ...zap.Field) *Logger {
	return &Logger{Logger: l.With(fields...)}
}

// FromEnv builds a logger configuration using environment variables.
//
// Required environment variables:
//   - LOG_LEVEL: sets log level (DEBUG, INFO, WARN, ERROR)
//   - APP_ENV: defines environment ("development" or "production")
//   - APP_NAME: sets the service name field
//
// Returns an error if any required variable is missing or invalid.
// The application should not start if this function returns an error.
func FromEnv() (Config, error) {
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		return Config{}, fmt.Errorf("%w: LOG_LEVEL", ErrMissingRequiredEnvVar)
	}

	appEnv := os.Getenv("APP_ENV")
	if appEnv == "" {
		return Config{}, fmt.Errorf("%w: APP_ENV", ErrMissingRequiredEnvVar)
	}

	appName := os.Getenv("APP_NAME")
	if appName == "" {
		return Config{}, fmt.Errorf("%w: APP_NAME", ErrMissingRequiredEnvVar)
	}

	cfg := Config{
		Level:       LogLevel(strings.ToUpper(logLevel)),
		Environment: Environment(strings.ToLower(appEnv)),
		ServiceName: appName,
	}

	// Validate the configuration before returning
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

// MustFromEnv builds a logger configuration from environment variables.
//
// Panics if any required variable is missing or invalid.
// Use this in main() for fail-fast behavior on startup.
func MustFromEnv() Config {
	cfg, err := FromEnv()
	if err != nil {
		panic(fmt.Sprintf("failed to load logger configuration from environment: %v", err))
	}
	return cfg
}

// Sync flushes any buffered log entries to the underlying writer.
//
// This should be deferred before program exit to avoid data loss.
func (l *Logger) Sync() error {
	return l.Logger.Sync()
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

// WithFields creates a derived logger with pre-attached structured fields using the global logger.
//
// Example:
//
//	log := logger.WithFields(zap.String("component", "auth"))
//	log.Info("Authentication service started")
func WithFields(fields ...zap.Field) *Logger {
	return Get().WithContext(fields...)
}
