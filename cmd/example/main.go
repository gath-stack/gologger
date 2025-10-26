package main

import (
	"errors"
	"fmt"
	"os"

	logger "github.com/gath-stack/gologger"

	"go.uber.org/zap"
)

func main() {
	// Initialize logger from environment variables
	// This will panic if any required env var is missing or invalid
	logger.MustInitFromEnv()
	defer func() {
		if err := logger.Sync(); err != nil {
			fmt.Fprintf(os.Stderr, "logger sync error: %v\n", err)
		}
	}()

	// Use the logger
	log := logger.Get()
	log.Info("application started",
		zap.String("version", "1.0.0"),
	)

	// Example with contextual logger
	userLogger := logger.With(
		zap.String("component", "auth"),
		zap.String("user_id", "12345"),
	)
	userLogger.Info("user authenticated successfully")

	// Example using package-level functions
	logger.Info("using package-level logger function")
	logger.Debug("debug message", zap.String("detail", "some detail"))

	// Example error logging
	err := someOperation()
	if err != nil {
		logger.Error("operation failed",
			zap.Error(err),
			zap.String("operation", "someOperation"),
		)
	}

	// Example using TryGet (non-panicking version)
	if log, err := logger.TryGet(); err == nil {
		log.Info("using TryGet successfully")
	} else {
		if errors.Is(err, logger.ErrNotInitialized) {
			fmt.Println("logger not initialized")
		}
	}
}

// Alternative: with explicit error handling instead of panic
func mainWithErrorHandling() {
	// Initialize logger with error handling
	if err := logger.InitFromEnv(); err != nil {
		// Handle error - maybe use a fallback logger or exit gracefully
		fmt.Fprintf(os.Stderr, "failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		if err := logger.Sync(); err != nil {
			fmt.Fprintf(os.Stderr, "logger sync error: %v\n", err)
		}
	}()

	logger.Info("application started with error handling")
}

// Example: Using defaults for quick setup
func mainWithDefaults() {
	// Initialize with sensible defaults (useful for development)
	if err := logger.InitWithDefaults(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		if err := logger.Sync(); err != nil {
			fmt.Fprintf(os.Stderr, "logger sync error: %v\n", err)
		}
	}()

	logger.Info("application started with defaults")
}

func someOperation() error {
	// Simulate some operation
	return nil
}
