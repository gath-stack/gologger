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
//   - Configuration managed by the config package for centralized validation.
//
// Example usage:
//
//	func main() {
//	    logger.MustInitFromEnv()
//	    defer logger.Get().Sync()
//
//	    log := logger.Get()
//	    log.Info("application started", zap.String("version", "1.0.0"))
//	}
//
// Alternative with error handling:
//
//	func main() {
//	    if err := logger.InitFromEnv(); err != nil {
//	        log.Fatalf("failed to initialize logger: %v", err)
//	    }
//	    defer logger.Get().Sync()
//
//	    logger.Info("application started")
//	}
//
// Production recommendations:
//   - Always initialize the global logger early in application startup using MustInitFromEnv().
//   - In production, prefer JSON encoding for structured log ingestion (set APP_ENV=production).
//   - Call `Sync()` before process exit to flush any buffered log entries.
//   - Ensure all required environment variables (LOG_LEVEL, APP_ENV, APP_NAME) are set.
package logger

import (
	"fmt"

	"github.com/gath-stack/gologger/internal/config"

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

var (
	// globalLogger is the shared singleton logger instance for the application.
	globalLogger *Logger
)

// New creates a new logger instance according to the given configuration.
//
// The configuration should come from the config package and will be validated
// by that package before being passed here.
//
// In production mode, logs are formatted as structured JSON suitable for ingestion by Loki,
// FluentBit, or Elasticsearch. In development mode, logs use a colorized console encoder.
//
// Example:
//
//	cfg := config.MustLoad()
//	logger, err := logger.New(cfg.Logger)
//	if err != nil {
//	    panic(err)
//	}
func New(cfg LoggerConfig) (*Logger, error) {
	// Note: validation is handled by the config package
	// We can assume cfg is already validated when it reaches here

	var zapLevel zapcore.Level

	// Map config log levels to zap internal levels
	switch cfg.Level {
	case LogLevelDebug:
		zapLevel = zapcore.DebugLevel
	case LogLevelInfo:
		zapLevel = zapcore.InfoLevel
	case LogLevelWarn:
		zapLevel = zapcore.WarnLevel
	case LogLevelError:
		zapLevel = zapcore.ErrorLevel
	default:
		// Default to info level as a safe fallback
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
// The configuration should come from the config package, which handles all validation.
func InitGlobal(cfg LoggerConfig) error {
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

// UnderlyingLogger returns the underlying zap.Logger for advanced integrations.
//
// This is useful when you need direct access to the zap.Logger for
// advanced features like OTLP log export or custom cores.
//
// Example:
//
//	log := logger.Get()
//	zapLogger := log.UnderlyingLogger()
//	// Use zapLogger for advanced operations
func (l *Logger) UnderlyingLogger() *zap.Logger {
	return l.Logger
}

// ReplaceCore replaces the logger's core with a new one.
//
// This is useful for adding additional outputs (like OTLP) while
// maintaining the existing logger configuration.
//
// Example:
//
//	log := logger.Get()
//	currentCore := log.UnderlyingLogger().Core()
//	otelCore := createOTELCore()
//	teeCore := zapcore.NewTee(currentCore, otelCore)
//	log.ReplaceCore(teeCore)
func (l *Logger) ReplaceCore(core zapcore.Core) {
	// Create new logger with the new core, preserving options
	newLogger := zap.New(core,
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)

	// Replace the underlying logger
	l.Logger = newLogger
}

// WithOTELCore creates a new logger that sends logs to both console and OTLP.
//
// This is a convenience method for adding OTLP export to the logger.
// The returned logger will write to both the original output and OTLP.
//
// Example:
//
//	log := logger.Get()
//	otelCore := createOTELCore()
//	log.WithOTELCore(otelCore)
//	log.Info("This goes to both console and Loki")
func (l *Logger) WithOTELCore(otelCore zapcore.Core) {
	currentCore := l.Logger.Core()
	teeCore := zapcore.NewTee(currentCore, otelCore)
	l.ReplaceCore(teeCore)
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
//	defer logger.Get().Sync()
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
//	    defer logger.Get().Sync()
//
//	    logger.Info("application started")
//	}
func MustInitFromEnv() {
	if err := InitFromEnv(); err != nil {
		panic(fmt.Sprintf("failed to initialize logger from environment: %v", err))
	}
}
