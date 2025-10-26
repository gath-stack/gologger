package logger

import (
	"errors"
	"os"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/gath-stack/gologger/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// resetGlobalLogger resets the global logger for testing purposes.
// This allows tests to run in isolation.
func resetGlobalLogger() {
	mu.Lock()
	defer mu.Unlock()
	globalLogger = nil
}

// TestBuildLogger tests the buildLogger function with various configurations.
func TestBuildLogger(t *testing.T) {
	tests := []struct {
		name      string
		config    config.LoggerConfig
		wantError bool
	}{
		{
			name: "valid debug development config",
			config: config.LoggerConfig{
				Level:       config.LogLevelDebug,
				Environment: config.EnvDevelopment,
				ServiceName: "test-service",
			},
			wantError: false,
		},
		{
			name: "valid info production config",
			config: config.LoggerConfig{
				Level:       config.LogLevelInfo,
				Environment: config.EnvProduction,
				ServiceName: "prod-service",
			},
			wantError: false,
		},
		{
			name: "valid warn config",
			config: config.LoggerConfig{
				Level:       config.LogLevelWarn,
				Environment: config.EnvDevelopment,
				ServiceName: "warn-service",
			},
			wantError: false,
		},
		{
			name: "valid error config",
			config: config.LoggerConfig{
				Level:       config.LogLevelError,
				Environment: config.EnvProduction,
				ServiceName: "error-service",
			},
			wantError: false,
		},
		{
			name: "invalid log level",
			config: config.LoggerConfig{
				Level:       config.LogLevel("INVALID"),
				Environment: config.EnvDevelopment,
				ServiceName: "test-service",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := buildLogger(tt.config)

			if tt.wantError && err == nil {
				t.Error("expected error but got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.wantError && logger == nil {
				t.Error("expected logger but got nil")
			}
			if tt.wantError && !errors.Is(err, ErrInvalidLogLevel) {
				t.Errorf("expected ErrInvalidLogLevel but got: %v", err)
			}
		})
	}
}

// TestValidateConfig tests the validateConfig function.
func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name      string
		config    config.LoggerConfig
		wantError error
	}{
		{
			name: "valid configuration",
			config: config.LoggerConfig{
				Level:       config.LogLevelInfo,
				Environment: config.EnvDevelopment,
				ServiceName: "test-service",
			},
			wantError: nil,
		},
		{
			name: "invalid log level",
			config: config.LoggerConfig{
				Level:       config.LogLevel("INVALID"),
				Environment: config.EnvDevelopment,
				ServiceName: "test-service",
			},
			wantError: ErrInvalidLogLevel,
		},
		{
			name: "invalid environment",
			config: config.LoggerConfig{
				Level:       config.LogLevelInfo,
				Environment: config.Environment("staging"),
				ServiceName: "test-service",
			},
			wantError: ErrInvalidEnvironment,
		},
		{
			name: "empty service name",
			config: config.LoggerConfig{
				Level:       config.LogLevelInfo,
				Environment: config.EnvDevelopment,
				ServiceName: "",
			},
			wantError: ErrMissingServiceName,
		},
		{
			name: "whitespace service name",
			config: config.LoggerConfig{
				Level:       config.LogLevelInfo,
				Environment: config.EnvDevelopment,
				ServiceName: "   ",
			},
			wantError: ErrMissingServiceName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(tt.config)

			if tt.wantError == nil && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if tt.wantError != nil && err == nil {
				t.Errorf("expected error %v but got nil", tt.wantError)
			}
			if tt.wantError != nil && err != nil && !errors.Is(err, tt.wantError) {
				t.Errorf("expected error %v but got: %v", tt.wantError, err)
			}
		})
	}
}

// TestInitGlobal tests the InitGlobal function.
func TestInitGlobal(t *testing.T) {
	tests := []struct {
		name          string
		config        config.LoggerConfig
		preinitialize bool
		wantError     error
	}{
		{
			name: "successful initialization",
			config: config.LoggerConfig{
				Level:       config.LogLevelInfo,
				Environment: config.EnvDevelopment,
				ServiceName: "test-service",
			},
			preinitialize: false,
			wantError:     nil,
		},
		{
			name: "already initialized",
			config: config.LoggerConfig{
				Level:       config.LogLevelInfo,
				Environment: config.EnvDevelopment,
				ServiceName: "test-service",
			},
			preinitialize: true,
			wantError:     ErrAlreadyInitialized,
		},
		{
			name: "invalid configuration",
			config: config.LoggerConfig{
				Level:       config.LogLevel("INVALID"),
				Environment: config.EnvDevelopment,
				ServiceName: "test-service",
			},
			preinitialize: false,
			wantError:     ErrInvalidConfig,
		},
		{
			name: "missing service name",
			config: config.LoggerConfig{
				Level:       config.LogLevelInfo,
				Environment: config.EnvDevelopment,
				ServiceName: "",
			},
			preinitialize: false,
			wantError:     ErrInvalidConfig,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetGlobalLogger()

			if tt.preinitialize {
				validCfg := config.LoggerConfig{
					Level:       config.LogLevelInfo,
					Environment: config.EnvDevelopment,
					ServiceName: "test-service",
				}
				if err := InitGlobal(validCfg); err != nil {
					t.Fatalf("failed to preinitialize: %v", err)
				}
			}

			err := InitGlobal(tt.config)

			if tt.wantError == nil && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if tt.wantError != nil && err == nil {
				t.Errorf("expected error %v but got nil", tt.wantError)
			}
			if tt.wantError != nil && err != nil && !errors.Is(err, tt.wantError) {
				t.Errorf("expected error %v but got: %v", tt.wantError, err)
			}

			resetGlobalLogger()
		})
	}
}

// TestGet tests the Get function.
func TestGet(t *testing.T) {
	t.Run("returns logger when initialized", func(t *testing.T) {
		resetGlobalLogger()

		cfg := config.LoggerConfig{
			Level:       config.LogLevelInfo,
			Environment: config.EnvDevelopment,
			ServiceName: "test-service",
		}
		if err := InitGlobal(cfg); err != nil {
			t.Fatalf("failed to initialize: %v", err)
		}

		logger := Get()
		if logger == nil {
			t.Error("expected logger but got nil")
		}

		resetGlobalLogger()
	})

	t.Run("panics when not initialized", func(t *testing.T) {
		resetGlobalLogger()

		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic but got none")
			}
		}()

		Get()
	})
}

// TestTryGet tests the TryGet function.
func TestTryGet(t *testing.T) {
	t.Run("returns logger when initialized", func(t *testing.T) {
		resetGlobalLogger()

		cfg := config.LoggerConfig{
			Level:       config.LogLevelInfo,
			Environment: config.EnvDevelopment,
			ServiceName: "test-service",
		}
		if err := InitGlobal(cfg); err != nil {
			t.Fatalf("failed to initialize: %v", err)
		}

		logger, err := TryGet()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if logger == nil {
			t.Error("expected logger but got nil")
		}

		resetGlobalLogger()
	})

	t.Run("returns error when not initialized", func(t *testing.T) {
		resetGlobalLogger()

		logger, err := TryGet()
		if err == nil {
			t.Error("expected error but got nil")
		}
		if !errors.Is(err, ErrNotInitialized) {
			t.Errorf("expected ErrNotInitialized but got: %v", err)
		}
		if logger != nil {
			t.Error("expected nil logger but got non-nil")
		}
	})
}

// TestLogger_With tests the With method.
func TestLogger_With(t *testing.T) {
	resetGlobalLogger()

	cfg := config.LoggerConfig{
		Level:       config.LogLevelInfo,
		Environment: config.EnvDevelopment,
		ServiceName: "test-service",
	}
	if err := InitGlobal(cfg); err != nil {
		t.Fatalf("failed to initialize: %v", err)
	}
	defer resetGlobalLogger()

	logger := Get()
	derivedLogger := logger.With(zap.String("component", "auth"))

	if derivedLogger == nil {
		t.Error("expected derived logger but got nil")
	}
	if derivedLogger == logger {
		t.Error("derived logger should be a new instance")
	}
}

// TestWith tests the package-level With function.
func TestWith(t *testing.T) {
	resetGlobalLogger()

	cfg := config.LoggerConfig{
		Level:       config.LogLevelInfo,
		Environment: config.EnvDevelopment,
		ServiceName: "test-service",
	}
	if err := InitGlobal(cfg); err != nil {
		t.Fatalf("failed to initialize: %v", err)
	}
	defer resetGlobalLogger()

	derivedLogger := With(zap.String("component", "auth"))
	if derivedLogger == nil {
		t.Error("expected derived logger but got nil")
	}
}

// TestSync tests the package-level Sync function.
func TestSync(t *testing.T) {
	t.Run("syncs when initialized", func(t *testing.T) {
		resetGlobalLogger()

		cfg := config.LoggerConfig{
			Level:       config.LogLevelInfo,
			Environment: config.EnvDevelopment,
			ServiceName: "test-service",
		}
		if err := InitGlobal(cfg); err != nil {
			t.Fatalf("failed to initialize: %v", err)
		}

		err := Sync()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		resetGlobalLogger()
	})

	t.Run("returns error when not initialized", func(t *testing.T) {
		resetGlobalLogger()

		err := Sync()
		if err == nil {
			t.Error("expected error but got nil")
		}
		if !errors.Is(err, ErrNotInitialized) {
			t.Errorf("expected ErrNotInitialized but got: %v", err)
		}
	})
}

// TestLogger_Sync tests the Sync method on Logger instance.
func TestLogger_Sync(t *testing.T) {
	resetGlobalLogger()

	cfg := config.LoggerConfig{
		Level:       config.LogLevelInfo,
		Environment: config.EnvDevelopment,
		ServiceName: "test-service",
	}
	if err := InitGlobal(cfg); err != nil {
		t.Fatalf("failed to initialize: %v", err)
	}
	defer resetGlobalLogger()

	logger := Get()
	err := logger.Sync()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestIsIgnorableSyncError tests the isIgnorableSyncError function.
func TestIsIgnorableSyncError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "EINVAL error",
			err:  syscall.EINVAL,
			want: true,
		},
		{
			name: "ENOTTY error",
			err:  syscall.ENOTTY,
			want: true,
		},
		{
			name: "EBADF error",
			err:  syscall.EBADF,
			want: true,
		},
		{
			name: "other error",
			err:  errors.New("some other error"),
			want: false,
		},
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isIgnorableSyncError(tt.err)
			if got != tt.want {
				t.Errorf("isIgnorableSyncError() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestPackageLevelLoggingFunctions tests the package-level logging functions.
func TestPackageLevelLoggingFunctions(t *testing.T) {
	resetGlobalLogger()

	cfg := config.LoggerConfig{
		Level:       config.LogLevelDebug,
		Environment: config.EnvDevelopment,
		ServiceName: "test-service",
	}
	if err := InitGlobal(cfg); err != nil {
		t.Fatalf("failed to initialize: %v", err)
	}
	defer resetGlobalLogger()

	// These should not panic when logger is initialized
	t.Run("Debug", func(t *testing.T) {
		Debug("debug message", zap.String("key", "value"))
	})

	t.Run("Info", func(t *testing.T) {
		Info("info message", zap.String("key", "value"))
	})

	t.Run("Warn", func(t *testing.T) {
		Warn("warn message", zap.String("key", "value"))
	})

	t.Run("Error", func(t *testing.T) {
		Error("error message", zap.String("key", "value"))
	})
}

// TestPackageLevelLoggingFunctions_Panic tests that logging functions panic when not initialized.
func TestPackageLevelLoggingFunctions_Panic(t *testing.T) {
	resetGlobalLogger()

	tests := []struct {
		name string
		fn   func()
	}{
		{
			name: "Debug panics",
			fn:   func() { Debug("test") },
		},
		{
			name: "Info panics",
			fn:   func() { Info("test") },
		},
		{
			name: "Warn panics",
			fn:   func() { Warn("test") },
		},
		{
			name: "Error panics",
			fn:   func() { Error("test") },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Error("expected panic but got none")
				}
			}()

			tt.fn()
		})
	}
}

// TestLogger_UnderlyingLogger tests the UnderlyingLogger method.
func TestLogger_UnderlyingLogger(t *testing.T) {
	resetGlobalLogger()

	cfg := config.LoggerConfig{
		Level:       config.LogLevelInfo,
		Environment: config.EnvDevelopment,
		ServiceName: "test-service",
	}
	if err := InitGlobal(cfg); err != nil {
		t.Fatalf("failed to initialize: %v", err)
	}
	defer resetGlobalLogger()

	logger := Get()
	zapLogger := logger.UnderlyingLogger()

	if zapLogger == nil {
		t.Error("expected zap.Logger but got nil")
	}
	if zapLogger != logger.Logger {
		t.Error("underlying logger should be the same as embedded logger")
	}
}

// TestLogger_WithCore tests the WithCore method.
func TestLogger_WithCore(t *testing.T) {
	resetGlobalLogger()

	cfg := config.LoggerConfig{
		Level:       config.LogLevelInfo,
		Environment: config.EnvDevelopment,
		ServiceName: "test-service",
	}
	if err := InitGlobal(cfg); err != nil {
		t.Fatalf("failed to initialize: %v", err)
	}
	defer resetGlobalLogger()

	logger := Get()

	// Create a new core
	encoderConfig := zapcore.EncoderConfig{
		MessageKey:  "message",
		LevelKey:    "level",
		EncodeLevel: zapcore.LowercaseLevelEncoder,
		EncodeTime:  zapcore.ISO8601TimeEncoder,
	}
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(os.Stdout),
		zapcore.InfoLevel,
	)

	newLogger := logger.WithCore(core)

	if newLogger == nil {
		t.Error("expected new logger but got nil")
	}
	if newLogger == logger {
		t.Error("new logger should be a different instance")
	}
}

// TestLogger_WithOTELCore tests the WithOTELCore method.
func TestLogger_WithOTELCore(t *testing.T) {
	resetGlobalLogger()

	cfg := config.LoggerConfig{
		Level:       config.LogLevelInfo,
		Environment: config.EnvDevelopment,
		ServiceName: "test-service",
	}
	if err := InitGlobal(cfg); err != nil {
		t.Fatalf("failed to initialize: %v", err)
	}
	defer resetGlobalLogger()

	logger := Get()

	// Create a mock OTEL core
	encoderConfig := zapcore.EncoderConfig{
		MessageKey:  "message",
		LevelKey:    "level",
		EncodeLevel: zapcore.LowercaseLevelEncoder,
		EncodeTime:  zapcore.ISO8601TimeEncoder,
	}
	otelCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(os.Stdout),
		zapcore.InfoLevel,
	)

	newLogger := logger.WithOTELCore(otelCore)

	if newLogger == nil {
		t.Error("expected new logger but got nil")
	}
	if newLogger == logger {
		t.Error("new logger should be a different instance")
	}
}

// TestInitFromEnv tests the InitFromEnv function.
func TestInitFromEnv(t *testing.T) {
	tests := []struct {
		name      string
		envVars   map[string]string
		wantError bool
	}{
		{
			name: "successful initialization",
			envVars: map[string]string{
				"LOG_LEVEL": "INFO",
				"APP_ENV":   "development",
				"APP_NAME":  "test-service",
			},
			wantError: false,
		},
		{
			name: "missing environment variables",
			envVars: map[string]string{
				"LOG_LEVEL": "INFO",
			},
			wantError: true,
		},
		{
			name: "invalid log level",
			envVars: map[string]string{
				"LOG_LEVEL": "INVALID",
				"APP_ENV":   "development",
				"APP_NAME":  "test-service",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetGlobalLogger()
			clearTestEnv()

			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			err := InitFromEnv()

			if tt.wantError && err == nil {
				t.Error("expected error but got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			resetGlobalLogger()
			clearTestEnv()
		})
	}
}

// TestMustInitFromEnv tests the MustInitFromEnv function.
func TestMustInitFromEnv(t *testing.T) {
	t.Run("panics on error", func(t *testing.T) {
		resetGlobalLogger()
		clearTestEnv()

		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic but got none")
			}
			clearTestEnv()
		}()

		MustInitFromEnv()
	})

	t.Run("succeeds with valid environment", func(t *testing.T) {
		resetGlobalLogger()
		clearTestEnv()

		os.Setenv("LOG_LEVEL", "INFO")
		os.Setenv("APP_ENV", "development")
		os.Setenv("APP_NAME", "test-service")

		defer func() {
			resetGlobalLogger()
			clearTestEnv()
		}()

		MustInitFromEnv()

		logger := Get()
		if logger == nil {
			t.Error("expected logger but got nil")
		}
	})
}

// TestInitWithDefaults tests the InitWithDefaults function.
func TestInitWithDefaults(t *testing.T) {
	resetGlobalLogger()

	err := InitWithDefaults()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	logger := Get()
	if logger == nil {
		t.Error("expected logger but got nil")
	}

	resetGlobalLogger()
}

// TestSyncWithTimeout tests the SyncWithTimeout function.
func TestSyncWithTimeout(t *testing.T) {
	t.Run("syncs successfully within timeout", func(t *testing.T) {
		resetGlobalLogger()

		cfg := config.LoggerConfig{
			Level:       config.LogLevelInfo,
			Environment: config.EnvDevelopment,
			ServiceName: "test-service",
		}
		if err := InitGlobal(cfg); err != nil {
			t.Fatalf("failed to initialize: %v", err)
		}

		err := SyncWithTimeout(5 * time.Second)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		resetGlobalLogger()
	})

	t.Run("returns error when not initialized", func(t *testing.T) {
		resetGlobalLogger()

		err := SyncWithTimeout(5 * time.Second)
		if err == nil {
			t.Error("expected error but got nil")
		}
		if !errors.Is(err, ErrNotInitialized) {
			t.Errorf("expected ErrNotInitialized but got: %v", err)
		}
	})
}

// TestLogger_SyncWithTimeout tests the SyncWithTimeout method on Logger instance.
func TestLogger_SyncWithTimeout(t *testing.T) {
	resetGlobalLogger()

	cfg := config.LoggerConfig{
		Level:       config.LogLevelInfo,
		Environment: config.EnvDevelopment,
		ServiceName: "test-service",
	}
	if err := InitGlobal(cfg); err != nil {
		t.Fatalf("failed to initialize: %v", err)
	}
	defer resetGlobalLogger()

	logger := Get()
	err := logger.SyncWithTimeout(5 * time.Second)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestConcurrentInitialization tests that concurrent initialization is thread-safe.
func TestConcurrentInitialization(t *testing.T) {
	resetGlobalLogger()

	cfg := config.LoggerConfig{
		Level:       config.LogLevelInfo,
		Environment: config.EnvDevelopment,
		ServiceName: "test-service",
	}

	var wg sync.WaitGroup
	numGoroutines := 10
	errs := make([]error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			errs[index] = InitGlobal(cfg)
		}(i)
	}

	wg.Wait()

	// Exactly one should succeed, the rest should get ErrAlreadyInitialized
	successCount := 0
	alreadyInitCount := 0

	for _, err := range errs {
		if err == nil {
			successCount++
		} else if errors.Is(err, ErrAlreadyInitialized) {
			alreadyInitCount++
		}
	}

	if successCount != 1 {
		t.Errorf("expected exactly 1 successful initialization, got %d", successCount)
	}
	if alreadyInitCount != numGoroutines-1 {
		t.Errorf("expected %d already initialized errors, got %d", numGoroutines-1, alreadyInitCount)
	}

	resetGlobalLogger()
}

// TestConcurrentGetOperations tests that concurrent Get operations are thread-safe.
func TestConcurrentGetOperations(t *testing.T) {
	resetGlobalLogger()

	cfg := config.LoggerConfig{
		Level:       config.LogLevelInfo,
		Environment: config.EnvDevelopment,
		ServiceName: "test-service",
	}
	if err := InitGlobal(cfg); err != nil {
		t.Fatalf("failed to initialize: %v", err)
	}
	defer resetGlobalLogger()

	var wg sync.WaitGroup
	numGoroutines := 100
	errorsChan := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			logger := Get()
			if logger == nil {
				errorsChan <- errors.New("got nil logger")
				return
			}
			logger.Info("test message")
		}()
	}

	wg.Wait()
	close(errorsChan)

	// Check for any errors
	for err := range errorsChan {
		t.Errorf("concurrent operation error: %v", err)
	}
}

// clearTestEnv clears test environment variables.
func clearTestEnv() {
	os.Unsetenv("LOG_LEVEL")
	os.Unsetenv("APP_ENV")
	os.Unsetenv("APP_NAME")
}
