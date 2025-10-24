// Package config manages application configuration loading and validation.
//
// This package centralizes all environment variable handling and validation logic,
// providing a single source of truth for application configuration across all modules.
//
// Environment File Loading:
//   - In development (APP_ENV != "production"): Automatically loads .env file if present
//   - In production (APP_ENV == "production"): Skips .env file, uses system environment variables
//   - If .env file is missing in development, falls back to system environment variables
//
// Key features:
//   - Centralized environment variable loading with .env support
//   - Automatic .env loading in non-production environments
//   - Strict validation with descriptive error messages
//   - Type-safe configuration structs
//   - Support for multiple configuration domains (logging, database, etc.)
//
// Example usage:
//
//	func main() {
//	    cfg, err := config.Load()  // Automatically loads .env if not in production
//	    if err != nil {
//	        log.Fatalf("failed to load configuration: %v", err)
//	    }
//	    // Use cfg.Logger, cfg.Database, etc.
//	}
//
// Production deployment:
//   - Set APP_ENV=production in your deployment environment
//   - Ensure all required environment variables are set by your infrastructure
//   - Do not include .env files in production deployments
package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

var (
	// ErrMissingRequiredEnvVar is returned when a required environment variable is not set.
	ErrMissingRequiredEnvVar = errors.New("required environment variable is not set")

	// ErrInvalidValue is returned when an environment variable contains an invalid value.
	ErrInvalidValue = errors.New("invalid configuration value")
)

// LogLevel represents the verbosity level for logging.
type LogLevel string

const (
	// LogLevelDebug enables detailed debug and above level logging.
	LogLevelDebug LogLevel = "DEBUG"
	// LogLevelInfo enables informational and above level logging (default).
	LogLevelInfo LogLevel = "INFO"
	// LogLevelWarn enables warning and above level logging.
	LogLevelWarn LogLevel = "WARN"
	// LogLevelError enables error level logging only.
	LogLevelError LogLevel = "ERROR"
)

// Validate checks if the log level is valid.
func (l LogLevel) Validate() error {
	switch l {
	case LogLevelDebug, LogLevelInfo, LogLevelWarn, LogLevelError:
		return nil
	default:
		return fmt.Errorf("%w: log level must be DEBUG, INFO, WARN, or ERROR, got '%s'", ErrInvalidValue, l)
	}
}

// Environment represents the deployment environment.
type Environment string

const (
	EnvDevelopment Environment = "development"
	EnvProduction  Environment = "production"
)

// Validate checks if the environment is valid.
func (e Environment) Validate() error {
	switch e {
	case EnvDevelopment, EnvProduction:
		return nil
	default:
		return fmt.Errorf("%w: environment must be 'development' or 'production', got '%s'", ErrInvalidValue, e)
	}
}

// LoggerConfig defines the configuration for the logging subsystem.
type LoggerConfig struct {
	Level       LogLevel
	Environment Environment
	ServiceName string
}

// Validate checks if the logger configuration is valid.
func (c LoggerConfig) Validate() error {
	// Validate log level
	if err := c.Level.Validate(); err != nil {
		return err
	}

	// Validate environment
	if err := c.Environment.Validate(); err != nil {
		return err
	}

	// Validate service name
	if strings.TrimSpace(c.ServiceName) == "" {
		return fmt.Errorf("%w: service name is required and cannot be empty", ErrInvalidValue)
	}

	return nil
}

// Config holds all application configuration.
type Config struct {
	Logger LoggerConfig
	// Add more configuration domains here as needed:
	// Database DatabaseConfig
	// Server   ServerConfig
	// Cache    CacheConfig
}

// Validate checks if the entire configuration is valid.
func (c Config) Validate() error {
	if err := c.Logger.Validate(); err != nil {
		return fmt.Errorf("logger configuration error: %w", err)
	}

	// Add validation for other config domains here
	// if err := c.Database.Validate(); err != nil {
	//     return fmt.Errorf("database configuration error: %w", err)
	// }

	return nil
}

// Load reads configuration from environment variables and validates it.
//
// This function automatically loads the .env file if APP_ENV is not set to "production".
// In production, environment variables must be set by the deployment environment.
//
// Required environment variables:
//   - LOG_LEVEL: sets log level (DEBUG, INFO, WARN, ERROR)
//   - APP_ENV: defines environment ("development" or "production")
//   - APP_NAME: sets the service name field
//
// Returns an error if any required variable is missing or contains invalid values.
// The application should not start if this function returns an error.
func Load() (Config, error) {
	// Load .env file only if not in production
	if err := loadEnvFile(); err != nil {
		return Config{}, err
	}

	loggerCfg, err := loadLoggerConfig()
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		Logger: loggerCfg,
	}

	// Validate the complete configuration
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

// loadEnvFile loads the .env file if the application is not running in production.
//
// The function checks the APP_ENV environment variable:
//   - If APP_ENV is "production", the .env file is NOT loaded (assumes env vars are set by infrastructure)
//   - If APP_ENV is not set or is not "production", the .env file is loaded
//   - If the .env file doesn't exist in non-production, it's not an error (env vars might be set another way)
func loadEnvFile() error {
	// Check if we're in production BEFORE loading .env
	// This allows production to be set via actual environment variables
	appEnv := os.Getenv("APP_ENV")

	// If APP_ENV is explicitly set to production, skip .env loading
	if strings.ToLower(appEnv) == "production" {
		return nil
	}

	// Try to load .env file for non-production environments
	// It's okay if the file doesn't exist - env vars might be set another way
	err := godotenv.Load()
	if err != nil {
		// Only return error if it's not a "file not found" error
		if !os.IsNotExist(err) && err.Error() != "open .env: no such file or directory" {
			return fmt.Errorf("error loading .env file: %w", err)
		}
		// File doesn't exist, but that's okay - continue with system env vars
	}

	return nil
}

// MustLoad loads configuration from environment variables and panics on error.
//
// This is useful in main() for fail-fast behavior during application startup.
// If any required variable is missing or invalid, the application will panic
// with a descriptive error message.
func MustLoad() Config {
	cfg, err := Load()
	if err != nil {
		panic(fmt.Sprintf("failed to load configuration from environment: %v", err))
	}
	return cfg
}

// loadLoggerConfig loads and validates logger-specific configuration from environment.
func loadLoggerConfig() (LoggerConfig, error) {
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		return LoggerConfig{}, fmt.Errorf("%w: LOG_LEVEL", ErrMissingRequiredEnvVar)
	}

	appEnv := os.Getenv("APP_ENV")
	if appEnv == "" {
		return LoggerConfig{}, fmt.Errorf("%w: APP_ENV", ErrMissingRequiredEnvVar)
	}

	appName := os.Getenv("APP_NAME")
	if appName == "" {
		return LoggerConfig{}, fmt.Errorf("%w: APP_NAME", ErrMissingRequiredEnvVar)
	}

	cfg := LoggerConfig{
		Level:       LogLevel(strings.ToUpper(logLevel)),
		Environment: Environment(strings.ToLower(appEnv)),
		ServiceName: appName,
	}

	// Validate before returning
	if err := cfg.Validate(); err != nil {
		return LoggerConfig{}, err
	}

	return cfg, nil
}

// GetEnv retrieves an environment variable with a fallback default value.
//
// This is a convenience function for optional environment variables.
func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// RequireEnv retrieves a required environment variable or returns an error.
//
// This is useful for loading additional required configuration values.
func RequireEnv(key string) (string, error) {
	value := os.Getenv(key)
	if value == "" {
		return "", fmt.Errorf("%w: %s", ErrMissingRequiredEnvVar, key)
	}
	return value, nil
}
