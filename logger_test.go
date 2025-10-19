package logger

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// TestConfigValidate tests the Config.Validate method
func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr error
	}{
		{
			name: "valid development config",
			config: Config{
				Level:       LevelInfo,
				Environment: EnvDevelopment,
				ServiceName: "test-service",
			},
			wantErr: nil,
		},
		{
			name: "valid production config with debug level",
			config: Config{
				Level:       LevelDebug,
				Environment: EnvProduction,
				ServiceName: "prod-service",
			},
			wantErr: nil,
		},
		{
			name: "valid warn level",
			config: Config{
				Level:       LevelWarn,
				Environment: EnvProduction,
				ServiceName: "warn-service",
			},
			wantErr: nil,
		},
		{
			name: "valid error level",
			config: Config{
				Level:       LevelError,
				Environment: EnvDevelopment,
				ServiceName: "error-service",
			},
			wantErr: nil,
		},
		{
			name: "invalid log level",
			config: Config{
				Level:       "INVALID",
				Environment: EnvProduction,
				ServiceName: "test-service",
			},
			wantErr: ErrInvalidLogLevel,
		},
		{
			name: "empty log level",
			config: Config{
				Level:       "",
				Environment: EnvProduction,
				ServiceName: "test-service",
			},
			wantErr: ErrInvalidLogLevel,
		},
		{
			name: "invalid environment",
			config: Config{
				Level:       LevelInfo,
				Environment: "staging",
				ServiceName: "test-service",
			},
			wantErr: ErrInvalidEnvironment,
		},
		{
			name: "empty environment",
			config: Config{
				Level:       LevelInfo,
				Environment: "",
				ServiceName: "test-service",
			},
			wantErr: ErrInvalidEnvironment,
		},
		{
			name: "empty service name",
			config: Config{
				Level:       LevelInfo,
				Environment: EnvProduction,
				ServiceName: "",
			},
			wantErr: ErrMissingServiceName,
		},
		{
			name: "whitespace-only service name",
			config: Config{
				Level:       LevelInfo,
				Environment: EnvProduction,
				ServiceName: "   ",
			},
			wantErr: ErrMissingServiceName,
		},
		{
			name: "service name with spaces is valid",
			config: Config{
				Level:       LevelInfo,
				Environment: EnvProduction,
				ServiceName: "my service",
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("Validate() expected error %v, got nil", tt.wantErr)
					return
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("Validate() error = %v, want %v", err, tt.wantErr)
				}
			} else {
				if err != nil {
					t.Errorf("Validate() unexpected error = %v", err)
				}
			}
		})
	}
}

// TestNew tests the New function
func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid development logger",
			config: Config{
				Level:       LevelDebug,
				Environment: EnvDevelopment,
				ServiceName: "test-dev",
			},
			wantErr: false,
		},
		{
			name: "valid production logger",
			config: Config{
				Level:       LevelInfo,
				Environment: EnvProduction,
				ServiceName: "test-prod",
			},
			wantErr: false,
		},
		{
			name: "invalid config - empty service name",
			config: Config{
				Level:       LevelInfo,
				Environment: EnvProduction,
				ServiceName: "",
			},
			wantErr: true,
		},
		{
			name: "invalid config - bad log level",
			config: Config{
				Level:       "TRACE",
				Environment: EnvProduction,
				ServiceName: "test",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := New(tt.config)
			if tt.wantErr {
				if err == nil {
					t.Error("New() expected error, got nil")
				}
				if logger != nil {
					t.Error("New() expected nil logger on error")
				}
			} else {
				if err != nil {
					t.Errorf("New() unexpected error = %v", err)
				}
				if logger == nil {
					t.Error("New() returned nil logger")
				}
			}
		})
	}
}

// TestInitGlobalAndGet tests InitGlobal and Get functions
func TestInitGlobalAndGet(t *testing.T) {
	// Reset global logger before test
	globalLogger = nil

	cfg := Config{
		Level:       LevelInfo,
		Environment: EnvDevelopment,
		ServiceName: "test-global",
	}

	// Test InitGlobal
	err := InitGlobal(cfg)
	if err != nil {
		t.Fatalf("InitGlobal() unexpected error = %v", err)
	}

	// Test Get
	logger := Get()
	if logger == nil {
		t.Fatal("Get() returned nil logger")
	}

	// Verify it's the same instance
	logger2 := Get()
	if logger != logger2 {
		t.Error("Get() should return the same singleton instance")
	}

	// Clean up
	globalLogger = nil
}

// TestGetPanicsWhenNotInitialized tests that Get panics when logger is not initialized
func TestGetPanicsWhenNotInitialized(t *testing.T) {
	// Reset global logger
	globalLogger = nil

	defer func() {
		if r := recover(); r == nil {
			t.Error("Get() should panic when logger is not initialized")
		}
	}()

	Get()
}

// TestFromEnv tests the FromEnv function
func TestFromEnv(t *testing.T) {
	tests := []struct {
		name        string
		envVars     map[string]string
		wantErr     bool
		wantErrType error
		checkConfig func(t *testing.T, cfg Config)
	}{
		{
			name: "valid environment variables",
			envVars: map[string]string{
				"LOG_LEVEL": "DEBUG",
				"APP_ENV":   "production",
				"APP_NAME":  "test-app",
			},
			wantErr: false,
			checkConfig: func(t *testing.T, cfg Config) {
				if cfg.Level != LevelDebug {
					t.Errorf("Level = %v, want %v", cfg.Level, LevelDebug)
				}
				if cfg.Environment != EnvProduction {
					t.Errorf("Environment = %v, want %v", cfg.Environment, EnvProduction)
				}
				if cfg.ServiceName != "test-app" {
					t.Errorf("ServiceName = %v, want %v", cfg.ServiceName, "test-app")
				}
			},
		},
		{
			name: "lowercase log level is normalized",
			envVars: map[string]string{
				"LOG_LEVEL": "info",
				"APP_ENV":   "development",
				"APP_NAME":  "test-app",
			},
			wantErr: false,
			checkConfig: func(t *testing.T, cfg Config) {
				if cfg.Level != LevelInfo {
					t.Errorf("Level = %v, want %v", cfg.Level, LevelInfo)
				}
			},
		},
		{
			name: "uppercase environment is normalized",
			envVars: map[string]string{
				"LOG_LEVEL": "INFO",
				"APP_ENV":   "PRODUCTION",
				"APP_NAME":  "test-app",
			},
			wantErr: false,
			checkConfig: func(t *testing.T, cfg Config) {
				if cfg.Environment != EnvProduction {
					t.Errorf("Environment = %v, want %v", cfg.Environment, EnvProduction)
				}
			},
		},
		{
			name: "missing LOG_LEVEL",
			envVars: map[string]string{
				"APP_ENV":  "production",
				"APP_NAME": "test-app",
			},
			wantErr:     true,
			wantErrType: ErrMissingRequiredEnvVar,
		},
		{
			name: "missing APP_ENV",
			envVars: map[string]string{
				"LOG_LEVEL": "INFO",
				"APP_NAME":  "test-app",
			},
			wantErr:     true,
			wantErrType: ErrMissingRequiredEnvVar,
		},
		{
			name: "missing APP_NAME",
			envVars: map[string]string{
				"LOG_LEVEL": "INFO",
				"APP_ENV":   "production",
			},
			wantErr:     true,
			wantErrType: ErrMissingRequiredEnvVar,
		},
		{
			name: "invalid log level",
			envVars: map[string]string{
				"LOG_LEVEL": "TRACE",
				"APP_ENV":   "production",
				"APP_NAME":  "test-app",
			},
			wantErr:     true,
			wantErrType: ErrInvalidLogLevel,
		},
		{
			name: "invalid environment",
			envVars: map[string]string{
				"LOG_LEVEL": "INFO",
				"APP_ENV":   "staging",
				"APP_NAME":  "test-app",
			},
			wantErr:     true,
			wantErrType: ErrInvalidEnvironment,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			os.Clearenv()

			// Set test environment variables
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			cfg, err := FromEnv()

			if tt.wantErr {
				if err == nil {
					t.Error("FromEnv() expected error, got nil")
					return
				}
				if tt.wantErrType != nil && !errors.Is(err, tt.wantErrType) {
					t.Errorf("FromEnv() error = %v, want error type %v", err, tt.wantErrType)
				}
			} else {
				if err != nil {
					t.Errorf("FromEnv() unexpected error = %v", err)
					return
				}
				if tt.checkConfig != nil {
					tt.checkConfig(t, cfg)
				}
			}

			// Clean up
			os.Clearenv()
		})
	}
}

// TestMustFromEnv tests the MustFromEnv function
func TestMustFromEnv(t *testing.T) {
	t.Run("success case", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("LOG_LEVEL", "INFO")
		os.Setenv("APP_ENV", "production")
		os.Setenv("APP_NAME", "test-app")

		defer func() {
			if r := recover(); r != nil {
				t.Errorf("MustFromEnv() panicked unexpectedly: %v", r)
			}
			os.Clearenv()
		}()

		cfg := MustFromEnv()
		if cfg.Level != LevelInfo {
			t.Errorf("Level = %v, want %v", cfg.Level, LevelInfo)
		}
	})

	t.Run("panic case", func(t *testing.T) {
		os.Clearenv()
		// Missing required variables

		defer func() {
			if r := recover(); r == nil {
				t.Error("MustFromEnv() should panic when environment is invalid")
			}
			os.Clearenv()
		}()

		MustFromEnv()
	})
}

// TestWithContext tests the WithContext method
func TestWithContext(t *testing.T) {
	cfg := Config{
		Level:       LevelInfo,
		Environment: EnvDevelopment,
		ServiceName: "test-context",
	}

	logger, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	contextLogger := logger.WithContext(
		zap.String("request_id", "req-123"),
		zap.Int("user_id", 456),
	)

	if contextLogger == nil {
		t.Error("WithContext() returned nil logger")
	}

	// Verify it's a different instance
	if contextLogger == logger {
		t.Error("WithContext() should return a new logger instance")
	}
}

// TestWithFields tests the WithFields global function
func TestWithFields(t *testing.T) {
	// Initialize global logger
	globalLogger = nil
	cfg := Config{
		Level:       LevelInfo,
		Environment: EnvDevelopment,
		ServiceName: "test-fields",
	}
	err := InitGlobal(cfg)
	if err != nil {
		t.Fatalf("Failed to initialize global logger: %v", err)
	}
	defer func() { globalLogger = nil }()

	fieldsLogger := WithFields(
		zap.String("component", "auth"),
	)

	if fieldsLogger == nil {
		t.Error("WithFields() returned nil logger")
	}
}

// TestLoggerSync tests the Sync method
func TestLoggerSync(t *testing.T) {
	cfg := Config{
		Level:       LevelInfo,
		Environment: EnvDevelopment,
		ServiceName: "test-sync",
	}

	logger, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Sync should not return an error for stdout/stderr
	err = logger.Sync()
	if err != nil {
		// On some systems (Linux), syncing stdout/stderr may return an error
		// This is acceptable in tests
		t.Logf("Sync() returned error (may be expected): %v", err)
	}
}

// TestGlobalLogFunctions tests the global log functions
func TestGlobalLogFunctions(t *testing.T) {
	// Create a custom logger that writes to a buffer for testing
	globalLogger = nil

	// We can't easily capture output in these tests without significant refactoring,
	// so we just verify the functions don't panic
	cfg := Config{
		Level:       LevelDebug,
		Environment: EnvDevelopment,
		ServiceName: "test-global-funcs",
	}
	err := InitGlobal(cfg)
	if err != nil {
		t.Fatalf("Failed to initialize global logger: %v", err)
	}
	defer func() { globalLogger = nil }()

	// Test that these don't panic
	Debug("debug message", zap.String("key", "value"))
	Info("info message", zap.Int("count", 42))
	Warn("warn message", zap.Bool("flag", true))
	Error("error message", zap.Error(errors.New("test error")))

	// We don't test Fatal as it would exit the test
}

// TestProductionEncoderConfig tests the production encoder configuration
func TestProductionEncoderConfig(t *testing.T) {
	config := productionEncoderConfig()

	if config.TimeKey != "timestamp" {
		t.Errorf("TimeKey = %v, want %v", config.TimeKey, "timestamp")
	}
	if config.LevelKey != "level" {
		t.Errorf("LevelKey = %v, want %v", config.LevelKey, "level")
	}
	if config.MessageKey != "message" {
		t.Errorf("MessageKey = %v, want %v", config.MessageKey, "message")
	}
	if config.StacktraceKey != "stacktrace" {
		t.Errorf("StacktraceKey = %v, want %v", config.StacktraceKey, "stacktrace")
	}
}

// TestDevelopmentEncoderConfig tests the development encoder configuration
func TestDevelopmentEncoderConfig(t *testing.T) {
	config := developmentEncoderConfig()

	if config.TimeKey != "T" {
		t.Errorf("TimeKey = %v, want %v", config.TimeKey, "T")
	}
	if config.LevelKey != "L" {
		t.Errorf("LevelKey = %v, want %v", config.LevelKey, "L")
	}
	if config.MessageKey != "M" {
		t.Errorf("MessageKey = %v, want %v", config.MessageKey, "M")
	}
}

// TestLoggerOutputFormat tests that production logger outputs JSON
func TestLoggerOutputFormat(t *testing.T) {
	// Create a buffer to capture output
	var buf bytes.Buffer

	// Create a custom zap config that writes to our buffer
	cfg := Config{
		Level:       LevelInfo,
		Environment: EnvProduction,
		ServiceName: "test-output",
	}

	// We need to create the logger manually to capture output
	zapConfig := zap.Config{
		Level:            zap.NewAtomicLevelAt(zapcore.InfoLevel),
		Development:      false,
		Encoding:         "json",
		EncoderConfig:    productionEncoderConfig(),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(productionEncoderConfig()),
		zapcore.AddSync(&buf),
		zapConfig.Level,
	)

	zapLogger := zap.New(core).With(
		zap.String("service", cfg.ServiceName),
		zap.String("environment", string(cfg.Environment)),
	)

	logger := &Logger{Logger: zapLogger}

	// Log a message
	logger.Info("test message", zap.String("key", "value"))

	// Parse the output as JSON
	output := buf.String()
	if output == "" {
		t.Fatal("No output captured")
	}

	var logEntry map[string]interface{}
	err := json.Unmarshal([]byte(strings.TrimSpace(output)), &logEntry)
	if err != nil {
		t.Fatalf("Failed to parse JSON output: %v\nOutput: %s", err, output)
	}

	// Verify JSON structure
	if logEntry["level"] != "info" {
		t.Errorf("level = %v, want %v", logEntry["level"], "info")
	}
	if logEntry["message"] != "test message" {
		t.Errorf("message = %v, want %v", logEntry["message"], "test message")
	}
	if logEntry["service"] != "test-output" {
		t.Errorf("service = %v, want %v", logEntry["service"], "test-output")
	}
	if logEntry["environment"] != "production" {
		t.Errorf("environment = %v, want %v", logEntry["environment"], "production")
	}
	if logEntry["key"] != "value" {
		t.Errorf("key = %v, want %v", logEntry["key"], "value")
	}
}

// TestLogLevels tests that different log levels work correctly
func TestLogLevels(t *testing.T) {
	levels := []struct {
		configLevel LogLevel
		zapLevel    zapcore.Level
	}{
		{LevelDebug, zapcore.DebugLevel},
		{LevelInfo, zapcore.InfoLevel},
		{LevelWarn, zapcore.WarnLevel},
		{LevelError, zapcore.ErrorLevel},
	}

	for _, tt := range levels {
		t.Run(string(tt.configLevel), func(t *testing.T) {
			cfg := Config{
				Level:       tt.configLevel,
				Environment: EnvDevelopment,
				ServiceName: "test-level",
			}

			logger, err := New(cfg)
			if err != nil {
				t.Fatalf("Failed to create logger: %v", err)
			}

			// Verify the logger was created (we can't easily test the actual level without reflection)
			if logger == nil {
				t.Error("Logger should not be nil")
			}
		})
	}
}

// TestErrorWrapping tests that errors are properly wrapped
func TestErrorWrapping(t *testing.T) {
	cfg := Config{
		Level:       "INVALID",
		Environment: EnvProduction,
		ServiceName: "test",
	}

	_, err := New(cfg)
	if err == nil {
		t.Fatal("Expected error from New() with invalid config")
	}

	// Check that the error message contains context
	errMsg := err.Error()
	if !strings.Contains(errMsg, "invalid logger configuration") {
		t.Errorf("Error message should contain context, got: %v", errMsg)
	}
}

// BenchmarkLogger benchmarks logger creation and usage
func BenchmarkLoggerCreation(b *testing.B) {
	cfg := Config{
		Level:       LevelInfo,
		Environment: EnvProduction,
		ServiceName: "benchmark-test",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := New(cfg)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkLogInfo(b *testing.B) {
	cfg := Config{
		Level:       LevelInfo,
		Environment: EnvProduction,
		ServiceName: "benchmark-test",
	}

	logger, err := New(cfg)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("benchmark message", zap.Int("iteration", i))
	}
}

func BenchmarkLogWithFields(b *testing.B) {
	cfg := Config{
		Level:       LevelInfo,
		Environment: EnvProduction,
		ServiceName: "benchmark-test",
	}

	logger, err := New(cfg)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		contextLogger := logger.WithContext(
			zap.String("request_id", "req-123"),
			zap.Int("user_id", 456),
		)
		contextLogger.Info("benchmark message")
	}
}
