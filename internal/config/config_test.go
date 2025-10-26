// Package config manages application configuration loading and validation.
package config

import (
	"errors"
	"os"
	"testing"
)

// TestLogLevel_Validate tests the validation of log levels.
func TestLogLevel_Validate(t *testing.T) {
	tests := []struct {
		name      string
		level     LogLevel
		wantError bool
	}{
		{
			name:      "valid debug level",
			level:     LogLevelDebug,
			wantError: false,
		},
		{
			name:      "valid info level",
			level:     LogLevelInfo,
			wantError: false,
		},
		{
			name:      "valid warn level",
			level:     LogLevelWarn,
			wantError: false,
		},
		{
			name:      "valid error level",
			level:     LogLevelError,
			wantError: false,
		},
		{
			name:      "invalid log level",
			level:     LogLevel("INVALID"),
			wantError: true,
		},
		{
			name:      "empty log level",
			level:     LogLevel(""),
			wantError: true,
		},
		{
			name:      "lowercase log level",
			level:     LogLevel("debug"),
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.level.Validate()
			if tt.wantError && err == nil {
				t.Error("expected error but got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if tt.wantError && err != nil && !errors.Is(err, ErrInvalidValue) {
				t.Errorf("expected ErrInvalidValue but got: %v", err)
			}
		})
	}
}

// TestEnvironment_Validate tests the validation of environment values.
func TestEnvironment_Validate(t *testing.T) {
	tests := []struct {
		name      string
		env       Environment
		wantError bool
	}{
		{
			name:      "valid development environment",
			env:       EnvDevelopment,
			wantError: false,
		},
		{
			name:      "valid production environment",
			env:       EnvProduction,
			wantError: false,
		},
		{
			name:      "invalid environment",
			env:       Environment("staging"),
			wantError: true,
		},
		{
			name:      "empty environment",
			env:       Environment(""),
			wantError: true,
		},
		{
			name:      "uppercase environment",
			env:       Environment("PRODUCTION"),
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.env.Validate()
			if tt.wantError && err == nil {
				t.Error("expected error but got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if tt.wantError && err != nil && !errors.Is(err, ErrInvalidValue) {
				t.Errorf("expected ErrInvalidValue but got: %v", err)
			}
		})
	}
}

// TestLoggerConfig_Validate tests the validation of logger configuration.
func TestLoggerConfig_Validate(t *testing.T) {
	tests := []struct {
		name      string
		config    LoggerConfig
		wantError bool
	}{
		{
			name: "valid configuration",
			config: LoggerConfig{
				Level:       LogLevelInfo,
				Environment: EnvDevelopment,
				ServiceName: "test-service",
			},
			wantError: false,
		},
		{
			name: "invalid log level",
			config: LoggerConfig{
				Level:       LogLevel("INVALID"),
				Environment: EnvDevelopment,
				ServiceName: "test-service",
			},
			wantError: true,
		},
		{
			name: "invalid environment",
			config: LoggerConfig{
				Level:       LogLevelInfo,
				Environment: Environment("staging"),
				ServiceName: "test-service",
			},
			wantError: true,
		},
		{
			name: "empty service name",
			config: LoggerConfig{
				Level:       LogLevelInfo,
				Environment: EnvDevelopment,
				ServiceName: "",
			},
			wantError: true,
		},
		{
			name: "whitespace only service name",
			config: LoggerConfig{
				Level:       LogLevelInfo,
				Environment: EnvDevelopment,
				ServiceName: "   ",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantError && err == nil {
				t.Error("expected error but got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// TestConfig_Validate tests the validation of the complete configuration.
func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name      string
		config    Config
		wantError bool
	}{
		{
			name: "valid configuration",
			config: Config{
				Logger: LoggerConfig{
					Level:       LogLevelInfo,
					Environment: EnvProduction,
					ServiceName: "test-service",
				},
			},
			wantError: false,
		},
		{
			name: "invalid logger configuration",
			config: Config{
				Logger: LoggerConfig{
					Level:       LogLevel("INVALID"),
					Environment: EnvProduction,
					ServiceName: "test-service",
				},
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantError && err == nil {
				t.Error("expected error but got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// TestLoad tests the Load function with various environment configurations.
func TestLoad(t *testing.T) {
	tests := []struct {
		name      string
		envVars   map[string]string
		wantError bool
	}{
		{
			name: "valid configuration",
			envVars: map[string]string{
				"LOG_LEVEL": "INFO",
				"APP_ENV":   "development",
				"APP_NAME":  "test-service",
			},
			wantError: false,
		},
		{
			name: "valid production configuration",
			envVars: map[string]string{
				"LOG_LEVEL": "ERROR",
				"APP_ENV":   "production",
				"APP_NAME":  "prod-service",
			},
			wantError: false,
		},
		{
			name: "missing LOG_LEVEL",
			envVars: map[string]string{
				"APP_ENV":  "development",
				"APP_NAME": "test-service",
			},
			wantError: true,
		},
		{
			name: "missing APP_ENV",
			envVars: map[string]string{
				"LOG_LEVEL": "INFO",
				"APP_NAME":  "test-service",
			},
			wantError: true,
		},
		{
			name: "missing APP_NAME",
			envVars: map[string]string{
				"LOG_LEVEL": "INFO",
				"APP_ENV":   "development",
			},
			wantError: true,
		},
		{
			name: "invalid LOG_LEVEL",
			envVars: map[string]string{
				"LOG_LEVEL": "INVALID",
				"APP_ENV":   "development",
				"APP_NAME":  "test-service",
			},
			wantError: true,
		},
		{
			name: "invalid APP_ENV",
			envVars: map[string]string{
				"LOG_LEVEL": "INFO",
				"APP_ENV":   "staging",
				"APP_NAME":  "test-service",
			},
			wantError: true,
		},
		{
			name: "lowercase log level gets normalized",
			envVars: map[string]string{
				"LOG_LEVEL": "debug",
				"APP_ENV":   "development",
				"APP_NAME":  "test-service",
			},
			wantError: false,
		},
		{
			name: "uppercase app env gets normalized",
			envVars: map[string]string{
				"LOG_LEVEL": "INFO",
				"APP_ENV":   "DEVELOPMENT",
				"APP_NAME":  "test-service",
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment before each test
			clearEnv()

			// Set environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			// Run Load
			cfg, err := Load()

			// Check error expectation
			if tt.wantError && err == nil {
				t.Error("expected error but got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			// Verify configuration values when no error expected
			if !tt.wantError && err == nil {
				if cfg.Logger.ServiceName != tt.envVars["APP_NAME"] {
					t.Errorf("expected service name %q, got %q", tt.envVars["APP_NAME"], cfg.Logger.ServiceName)
				}
			}

			// Clean up
			clearEnv()
		})
	}
}

// TestMustLoad tests the MustLoad function.
func TestMustLoad(t *testing.T) {
	t.Run("panics on missing environment variables", func(t *testing.T) {
		clearEnv()

		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic but got none")
			}
		}()

		MustLoad()
	})

	t.Run("returns config on valid environment", func(t *testing.T) {
		clearEnv()
		os.Setenv("LOG_LEVEL", "INFO")
		os.Setenv("APP_ENV", "development")
		os.Setenv("APP_NAME", "test-service")

		defer clearEnv()

		cfg := MustLoad()
		if cfg.Logger.ServiceName != "test-service" {
			t.Errorf("expected service name %q, got %q", "test-service", cfg.Logger.ServiceName)
		}
	})
}

// TestGetEnv tests the GetEnv function.
func TestGetEnv(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		envValue     string
		setEnv       bool
		want         string
	}{
		{
			name:         "returns environment value when set",
			key:          "TEST_VAR",
			defaultValue: "default",
			envValue:     "custom",
			setEnv:       true,
			want:         "custom",
		},
		{
			name:         "returns default value when not set",
			key:          "TEST_VAR",
			defaultValue: "default",
			envValue:     "",
			setEnv:       false,
			want:         "default",
		},
		{
			name:         "returns empty string default when not set",
			key:          "TEST_VAR",
			defaultValue: "",
			envValue:     "",
			setEnv:       false,
			want:         "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearEnv()

			if tt.setEnv {
				os.Setenv(tt.key, tt.envValue)
			}

			got := GetEnv(tt.key, tt.defaultValue)
			if got != tt.want {
				t.Errorf("expected %q, got %q", tt.want, got)
			}

			clearEnv()
		})
	}
}

// TestRequireEnv tests the RequireEnv function.
func TestRequireEnv(t *testing.T) {
	tests := []struct {
		name      string
		key       string
		value     string
		setEnv    bool
		wantError bool
		wantValue string
	}{
		{
			name:      "returns value when set",
			key:       "REQUIRED_VAR",
			value:     "test-value",
			setEnv:    true,
			wantError: false,
			wantValue: "test-value",
		},
		{
			name:      "returns error when not set",
			key:       "REQUIRED_VAR",
			value:     "",
			setEnv:    false,
			wantError: true,
			wantValue: "",
		},
		{
			name:      "returns error when set to empty string",
			key:       "REQUIRED_VAR",
			value:     "",
			setEnv:    true,
			wantError: true,
			wantValue: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearEnv()

			if tt.setEnv {
				os.Setenv(tt.key, tt.value)
			}

			got, err := RequireEnv(tt.key)

			if tt.wantError && err == nil {
				t.Error("expected error but got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if tt.wantError && err != nil && !errors.Is(err, ErrMissingRequiredEnvVar) {
				t.Errorf("expected ErrMissingRequiredEnvVar but got: %v", err)
			}
			if got != tt.wantValue {
				t.Errorf("expected %q, got %q", tt.wantValue, got)
			}

			clearEnv()
		})
	}
}

// TestLoadEnvFile tests the loadEnvFile function behavior.
func TestLoadEnvFile(t *testing.T) {
	tests := []struct {
		name      string
		appEnv    string
		setAppEnv bool
		wantError bool
	}{
		{
			name:      "skips loading in production",
			appEnv:    "production",
			setAppEnv: true,
			wantError: false,
		},
		{
			name:      "skips loading in PRODUCTION (uppercase)",
			appEnv:    "PRODUCTION",
			setAppEnv: true,
			wantError: false,
		},
		{
			name:      "attempts loading in development",
			appEnv:    "development",
			setAppEnv: true,
			wantError: false, // No error even if .env doesn't exist
		},
		{
			name:      "attempts loading when APP_ENV not set",
			appEnv:    "",
			setAppEnv: false,
			wantError: false, // No error even if .env doesn't exist
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearEnv()

			if tt.setAppEnv {
				os.Setenv("APP_ENV", tt.appEnv)
			}

			err := loadEnvFile()

			if tt.wantError && err == nil {
				t.Error("expected error but got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			clearEnv()
		})
	}
}

// clearEnv clears all test-related environment variables.
func clearEnv() {
	os.Unsetenv("LOG_LEVEL")
	os.Unsetenv("APP_ENV")
	os.Unsetenv("APP_NAME")
	os.Unsetenv("TEST_VAR")
	os.Unsetenv("REQUIRED_VAR")
}
