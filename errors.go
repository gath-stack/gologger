package logger

import "errors"

// Sentinel errors that can be checked with errors.Is().
//
// These errors provide a stable API for programmatic error handling.
// Users can check for specific error conditions without relying on
// error message strings.
//
// Example usage:
//
//	log, err := logger.TryGet()
//	if errors.Is(err, logger.ErrNotInitialized) {
//	    // Handle uninitialized logger
//	}

var (
	// ErrNotInitialized is returned when attempting to use the logger before initialization.
	// Call InitGlobal(), InitFromEnv(), or MustInitFromEnv() before using the logger.
	ErrNotInitialized = errors.New("logger not initialized")

	// ErrAlreadyInitialized is returned when attempting to initialize the logger multiple times.
	// The logger can only be initialized once. Subsequent calls to InitGlobal() will return this error.
	ErrAlreadyInitialized = errors.New("logger already initialized")

	// ErrInvalidConfig is returned when the provided configuration is invalid.
	// This is a wrapper error that contains the specific validation failure.
	ErrInvalidConfig = errors.New("invalid logger configuration")

	// ErrInvalidLogLevel is returned when an invalid log level is provided.
	// Valid levels are: DEBUG, INFO, WARN, ERROR.
	ErrInvalidLogLevel = errors.New("invalid log level")

	// ErrInvalidEnvironment is returned when an invalid environment is provided.
	// Valid environments are: development, production.
	ErrInvalidEnvironment = errors.New("invalid environment")

	// ErrMissingServiceName is returned when service name is empty or contains only whitespace.
	ErrMissingServiceName = errors.New("service name is required")

	// ErrSyncFailed is returned when log synchronization fails.
	// This may occur when flushing buffered log entries to the underlying writer.
	ErrSyncFailed = errors.New("failed to sync logger")
)
