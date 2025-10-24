package logger

import "github.com/gath-stack/gologger/internal/config"

// Re-export config types for convenience and backward compatibility.
// This allows users to reference logger.LogLevel instead of config.LogLevel
// when working directly with the logger package.

// LogLevel represents the verbosity level for logging.
type LogLevel = config.LogLevel

const (
	// LogLevelDebug enables detailed debug and above level logging.
	LogLevelDebug = config.LogLevelDebug
	// LogLevelInfo enables informational and above level logging (default).
	LogLevelInfo = config.LogLevelInfo
	// LogLevelWarn enables warning and above level logging.
	LogLevelWarn = config.LogLevelWarn
	// LogLevelError enables error level logging only.
	LogLevelError = config.LogLevelError
)

// Environment represents the deployment environment.
type Environment = config.Environment

const (
	// EnvDevelopment represents the development environment.
	EnvDevelopment = config.EnvDevelopment
	// EnvProduction represents the production environment.
	EnvProduction = config.EnvProduction
)

// LoggerConfig defines the configuration parameters for the logger.
// This type is defined in the config package and re-exported here.
type LoggerConfig = config.LoggerConfig
