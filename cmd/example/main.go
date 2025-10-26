package main

import (
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
		if err := logger.Get().Sync(); err != nil {
			fmt.Fprintf(os.Stderr, "logger sync error: %v\n", err)
		}
	}()

	// Use the logger
	log := logger.Get()
	log.Info("application started",
		zap.String("version", "1.0.0"),
	)

	// Example with contextual logger
	userLogger := logger.WithFields(
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
}

// Alternative: with explicit error handling instead of panic
func mainWithErrorHandling() {
	// Initialize logger with error handling
	if err := logger.InitFromEnv(); err != nil {
		// Handle error - maybe use a fallback logger or exit gracefully
		panic(err) // or handle differently
	}
	defer func() {
		if err := logger.Get().Sync(); err != nil {
			fmt.Fprintf(os.Stderr, "logger sync error: %v\n", err)
		}
	}()

	logger.Info("application started with error handling")
}

func someOperation() error {
	// Simulate some operation
	return nil
}
