# GoLogger API Reference

Complete API documentation for GoLogger v1.0.

## Table of Contents

1. [Initialization](#initialization)
2. [Basic Logging](#basic-logging)
3. [Structured Logging](#structured-logging)
4. [Contextual Loggers](#contextual-loggers)
5. [Instance Methods](#instance-methods)
6. [Error Handling](#error-handling)
7. [Advanced Features](#advanced-features)
8. [Configuration](#configuration)
9. [Best Practices](#best-practices)

---

## Initialization

### InitGlobal

Initializes the global logger with explicit configuration.

**Signature:**
```go
func InitGlobal(cfg LoggerConfig) error
```

**Parameters:**
- `cfg`: Logger configuration struct

**Returns:**
- `error`: Returns error if initialization fails or logger is already initialized

**Errors:**
- `ErrAlreadyInitialized`: Logger was already initialized
- `ErrInvalidConfig`: Configuration validation failed
- `ErrInvalidLogLevel`: Invalid log level provided
- `ErrInvalidEnvironment`: Invalid environment provided
- `ErrMissingServiceName`: Service name is empty

**Example:**
```go
package main

import (
    "log"
    "github.com/gath-stack/gologger"
)

func main() {
    cfg := logger.LoggerConfig{
        Level:       logger.LogLevelInfo,
        Environment: logger.EnvProduction,
        ServiceName: "my-api",
    }
    
    if err := logger.InitGlobal(cfg); err != nil {
        log.Fatalf("failed to initialize logger: %v", err)
    }
    defer logger.Sync()
    
    logger.Info("application started")
}
```

**When to use:**
- When you need explicit control over configuration
- When configuration comes from non-environment sources
- In tests with custom config

---

### InitFromEnv

Initializes the global logger from environment variables.

**Signature:**
```go
func InitFromEnv() error
```

**Required Environment Variables:**
- `LOG_LEVEL`: Log verbosity (DEBUG, INFO, WARN, ERROR)
- `APP_ENV`: Runtime environment (development, production)
- `APP_NAME`: Service name for log entries

**Returns:**
- `error`: Returns error if initialization fails

**Example:**
```go
package main

import (
    "log"
    "github.com/gath-stack/gologger"
)

func main() {
    // Loads from environment variables
    if err := logger.InitFromEnv(); err != nil {
        log.Fatalf("failed to initialize logger: %v", err)
    }
    defer logger.Sync()
    
    logger.Info("application started")
}
```

**Environment file (.env):**
```env
LOG_LEVEL=INFO
APP_ENV=production
APP_NAME=my-service
```

**When to use:**
- Standard application initialization (recommended)
- 12-factor app compliance
- When configuration is externalized

---

### MustInitFromEnv

Initializes the global logger from environment variables, panics on error.

**Signature:**
```go
func MustInitFromEnv()
```

**Panics:**
- Panics if initialization fails for any reason

**Example:**
```go
package main

import "github.com/gath-stack/gologger"

func main() {
    // Panic if env vars are missing or invalid
    logger.MustInitFromEnv()
    defer logger.Sync()
    
    logger.Info("application started")
}
```

**When to use:**
- In main() functions where you want fail-fast behavior
- When the application cannot run without proper logging
- Most common initialization pattern

**⚠️ Warning:** Don't use in library code or where panic recovery is needed.

---

### InitWithDefaults

Initializes the global logger with sensible defaults for development.

**Signature:**
```go
func InitWithDefaults() error
```

**Default Configuration:**
- Level: `INFO`
- Environment: `development`
- ServiceName: `"app"`

**Returns:**
- `error`: Returns error if initialization fails

**Example:**
```go
package main

import (
    "log"
    "github.com/gath-stack/gologger"
)

func main() {
    // Quick setup for development
    if err := logger.InitWithDefaults(); err != nil {
        log.Fatalf("failed to initialize logger: %v", err)
    }
    defer logger.Sync()
    
    logger.Info("development environment ready")
}
```

**When to use:**
- Quick prototyping
- Development environments
- Testing without configuration
- Demo applications

**❌ Don't use in production:** Always use explicit configuration in production.

---

## Basic Logging

### Package-Level Functions

Convenience functions that use the global logger instance.

#### Debug

Logs a message at DEBUG level.

**Signature:**
```go
func Debug(msg string, fields ...zap.Field)
```

**Example:**
```go
logger.Debug("processing request")
logger.Debug("cache lookup", 
    zap.String("key", "user:123"),
    zap.Duration("elapsed", time.Millisecond*15),
)
```

---

#### Info

Logs a message at INFO level (recommended for normal operations).

**Signature:**
```go
func Info(msg string, fields ...zap.Field)
```

**Example:**
```go
logger.Info("application started")
logger.Info("request processed", 
    zap.String("method", "POST"),
    zap.String("path", "/api/users"),
    zap.Int("status", 200),
)
```

---

#### Warn

Logs a message at WARN level.

**Signature:**
```go
func Warn(msg string, fields ...zap.Field)
```

**Example:**
```go
logger.Warn("rate limit approaching")
logger.Warn("deprecated API usage", 
    zap.String("endpoint", "/v1/users"),
    zap.String("recommended", "/v2/users"),
)
```

---

#### Error

Logs a message at ERROR level.

**Signature:**
```go
func Error(msg string, fields ...zap.Field)
```

**Example:**
```go
err := processPayment()
if err != nil {
    logger.Error("payment processing failed",
        zap.Error(err),
        zap.String("payment_id", paymentID),
        zap.String("user_id", userID),
    )
}
```

---

#### Fatal

Logs a message at FATAL level and terminates the application.

**Signature:**
```go
func Fatal(msg string, fields ...zap.Field)
```

**Example:**
```go
db, err := connectDatabase()
if err != nil {
    logger.Fatal("cannot connect to database",
        zap.Error(err),
        zap.String("host", dbHost),
    )
    // Application exits here
}
```

**⚠️ Important:**
- Calls `os.Exit(1)` after logging
- Use sparingly—prefer returning errors
- Only use for unrecoverable startup failures

---

## Structured Logging

### Using Zap Fields

GoLogger uses Zap's strongly-typed field system for structured logging.

#### Common Field Types

```go
import "go.uber.org/zap"

// String values
zap.String("key", "value")

// Numeric values
zap.Int("count", 42)
zap.Int64("id", 123456789)
zap.Float64("price", 99.99)

// Boolean
zap.Bool("active", true)

// Time and Duration
zap.Time("created_at", time.Now())
zap.Duration("elapsed", time.Second*2)

// Error
zap.Error(err)

// Complex types
zap.Any("user", userObject)

// Arrays and Objects
zap.Strings("tags", []string{"go", "logging"})
zap.Ints("scores", []int{90, 85, 95})
```

#### Example: Complete Structured Log

```go
logger.Info("order created",
    zap.String("order_id", "ORD-12345"),
    zap.String("user_id", "USR-789"),
    zap.Float64("total", 149.99),
    zap.String("currency", "USD"),
    zap.Int("item_count", 3),
    zap.Duration("processing_time", time.Millisecond*250),
    zap.Strings("items", []string{"ITEM-1", "ITEM-2", "ITEM-3"}),
    zap.Bool("gift_wrap", true),
)
```

**Output (JSON in production):**
```json
{
  "level": "info",
  "timestamp": "2025-01-15T10:30:00Z",
  "service": "my-service",
  "message": "order created",
  "order_id": "ORD-12345",
  "user_id": "USR-789",
  "total": 149.99,
  "currency": "USD",
  "item_count": 3,
  "processing_time": 0.25,
  "items": ["ITEM-1", "ITEM-2", "ITEM-3"],
  "gift_wrap": true
}
```

---

## Contextual Loggers

### With

Creates a derived logger with pre-attached fields.

**Signature:**
```go
func With(fields ...zap.Field) *Logger
```

**Returns:**
- `*Logger`: New logger instance with attached fields

**Example:**
```go
// Create component-specific logger
authLogger := logger.With(
    zap.String("component", "auth"),
    zap.String("version", "v2"),
)

// All logs from this logger include the context
authLogger.Info("authentication started")
// Output includes: component="auth" version="v2"

authLogger.Warn("rate limit exceeded")
// Output includes: component="auth" version="v2"
```

### Use Cases

#### Per-Request Logger

```go
func handleRequest(w http.ResponseWriter, r *http.Request) {
    requestID := uuid.New().String()
    
    // Create request-specific logger
    reqLogger := logger.With(
        zap.String("request_id", requestID),
        zap.String("method", r.Method),
        zap.String("path", r.URL.Path),
        zap.String("user_agent", r.UserAgent()),
    )
    
    reqLogger.Info("request received")
    
    // Pass to handlers
    processRequest(reqLogger, r)
    
    reqLogger.Info("request completed")
}
```

#### Per-Component Logger

```go
type UserService struct {
    log *logger.Logger
}

func NewUserService() *UserService {
    return &UserService{
        log: logger.With(
            zap.String("component", "user_service"),
            zap.String("version", "v1"),
        ),
    }
}

func (s *UserService) CreateUser(user User) error {
    s.log.Info("creating user", zap.String("email", user.Email))
    
    if err := s.validateUser(user); err != nil {
        s.log.Error("user validation failed",
            zap.Error(err),
            zap.String("email", user.Email),
        )
        return err
    }
    
    s.log.Info("user created successfully",
        zap.String("user_id", user.ID),
    )
    return nil
}
```

#### Nested Context

```go
// Application logger
appLogger := logger.With(zap.String("app", "my-service"))

// Component logger
authLogger := appLogger.With(zap.String("component", "auth"))

// Operation logger
loginLogger := authLogger.With(
    zap.String("operation", "login"),
    zap.String("user_id", userID),
)

loginLogger.Info("login attempt")
// Output includes: app="my-service" component="auth" operation="login" user_id="123"
```

---

## Instance Methods

### Get

Returns the global logger instance. Panics if not initialized.

**Signature:**
```go
func Get() *Logger
```

**Returns:**
- `*Logger`: Global logger instance

**Panics:**
- `ErrNotInitialized`: Logger not initialized

**Example:**
```go
log := logger.Get()
log.Info("using instance method")

// Create derived logger
requestLog := log.With(zap.String("request_id", reqID))
requestLog.Info("processing request")
```

**When to use:**
- When you need the logger instance
- To create derived loggers
- In struct methods that store logger instances

---

### TryGet

Returns the global logger instance without panicking.

**Signature:**
```go
func TryGet() (*Logger, error)
```

**Returns:**
- `*Logger`: Global logger instance or nil
- `error`: Error if logger not initialized

**Errors:**
- `ErrNotInitialized`: Logger not initialized

**Example:**
```go
log, err := logger.TryGet()
if err != nil {
    if errors.Is(err, logger.ErrNotInitialized) {
        // Handle gracefully - maybe use fallback logger
        fmt.Println("logger not ready")
        return nil
    }
    return err
}

log.Info("logger ready")
```

**When to use:**
- In library code where panicking is inappropriate
- When you want graceful degradation
- In code that might run before logger initialization

---

### Sync

Flushes buffered log entries.

**Signature:**
```go
func Sync() error
```

**Returns:**
- `error`: Error if sync fails (ignores benign OS errors)

**Example:**
```go
func main() {
    logger.MustInitFromEnv()
    defer func() {
        if err := logger.Sync(); err != nil {
            fmt.Fprintf(os.Stderr, "logger sync failed: %v\n", err)
        }
    }()
    
    // Application code
}
```

**Best Practice:** Always defer `Sync()` in main().

---

### SyncWithTimeout

Flushes buffered log entries with a timeout.

**Signature:**
```go
func SyncWithTimeout(timeout time.Duration) error
```

**Parameters:**
- `timeout`: Maximum time to wait for sync

**Returns:**
- `error`: Error if sync fails or times out

**Example:**
```go
func gracefulShutdown() {
    log.Info("shutting down")
    
    // Ensure logs are flushed within 5 seconds
    if err := logger.SyncWithTimeout(5 * time.Second); err != nil {
        fmt.Fprintf(os.Stderr, "sync timeout: %v\n", err)
    }
}
```

**When to use:**
- During graceful shutdown
- When you need guaranteed sync completion
- In critical paths where log loss is unacceptable

---

## Error Handling

### Sentinel Errors

GoLogger exports sentinel errors for programmatic error checking.

```go
var (
    ErrNotInitialized     error // Logger not initialized
    ErrAlreadyInitialized error // Logger already initialized
    ErrInvalidConfig      error // Invalid configuration
    ErrInvalidLogLevel    error // Invalid log level
    ErrInvalidEnvironment error // Invalid environment
    ErrMissingServiceName error // Service name missing
    ErrSyncFailed        error // Log sync failed
)
```

### Using errors.Is

```go
import "errors"

// Check for specific error
err := logger.InitFromEnv()
if err != nil {
    if errors.Is(err, logger.ErrNotInitialized) {
        log.Println("logger not initialized")
    } else if errors.Is(err, logger.ErrInvalidConfig) {
        log.Println("invalid configuration")
    } else {
        log.Printf("unknown error: %v", err)
    }
}
```

### Common Error Handling Patterns

#### Initialization Error Handling

```go
func initLogger() error {
    err := logger.InitFromEnv()
    if err != nil {
        if errors.Is(err, logger.ErrAlreadyInitialized) {
            // Logger already ready - this is OK
            return nil
        }
        return fmt.Errorf("logger initialization failed: %w", err)
    }
    return nil
}
```

#### Safe Get with Fallback

```go
func getLogger() *logger.Logger {
    log, err := logger.TryGet()
    if err != nil {
        // Fallback: initialize with defaults
        if err := logger.InitWithDefaults(); err != nil {
            panic("cannot initialize logger")
        }
        return logger.Get()
    }
    return log
}
```

#### Configuration Validation

```go
cfg := logger.LoggerConfig{
    Level:       logger.LogLevel(os.Getenv("LOG_LEVEL")),
    Environment: logger.Environment(os.Getenv("APP_ENV")),
    ServiceName: os.Getenv("APP_NAME"),
}

err := logger.InitGlobal(cfg)
if err != nil {
    switch {
    case errors.Is(err, logger.ErrInvalidLogLevel):
        log.Fatal("invalid LOG_LEVEL")
    case errors.Is(err, logger.ErrInvalidEnvironment):
        log.Fatal("invalid APP_ENV")
    case errors.Is(err, logger.ErrMissingServiceName):
        log.Fatal("missing APP_NAME")
    default:
        log.Fatalf("logger init failed: %v", err)
    }
}
```

---

## Advanced Features

### UnderlyingLogger

**⚠️ UNSTABLE API** - May change in future versions.

Returns the underlying zap.Logger for advanced integrations.

**Signature:**
```go
func (l *Logger) UnderlyingLogger() *zap.Logger
```

**Example:**
```go
log := logger.Get()
zapLogger := log.UnderlyingLogger()

// Use zap's advanced features
zapLogger.Core()
```

**When to use:**
- Integration with other systems requiring *zap.Logger
- Advanced zap features not exposed by GoLogger
- Custom core manipulation

---

### WithCore

**⚠️ UNSTABLE API** - May change in future versions.

Creates a new logger with a different core.

**Signature:**
```go
func (l *Logger) WithCore(core zapcore.Core) *Logger
```

**Returns:**
- `*Logger`: New logger instance with new core

**Example:**
```go
import "go.uber.org/zap/zapcore"

log := logger.Get()
currentCore := log.UnderlyingLogger().Core()

// Create custom core (e.g., write to file)
fileCore := createFileCore()

// Tee to multiple outputs
teeCore := zapcore.NewTee(currentCore, fileCore)

// Get new logger with both outputs
multiOutputLog := log.WithCore(teeCore)

multiOutputLog.Info("logged to console and file")
```

---

### WithOTELCore

**⚠️ UNSTABLE API** - May change in future versions.

Creates a new logger that sends logs to both console and OTLP.

**Signature:**
```go
func (l *Logger) WithOTELCore(otelCore zapcore.Core) *Logger
```

**Parameters:**
- `otelCore`: OpenTelemetry core for log export

**Returns:**
- `*Logger`: New logger with OTEL output

**Example:**
```go
log := logger.Get()

// Create OTEL core (implementation specific)
otelCore := createOTELCore()

// Get logger with OTEL export
otelLog := log.WithOTELCore(otelCore)

otelLog.Info("sent to console and OTLP")
```

**When to use:**
- Integrating with observability platforms (Grafana Loki, etc.)
- Sending logs to multiple destinations
- Advanced monitoring setups

---

## Configuration

### LoggerConfig

Configuration struct for logger initialization.

```go
type LoggerConfig struct {
    Level       LogLevel
    Environment Environment
    ServiceName string
}
```

#### Fields

**Level** (`LogLevel`)
- Type: String constant
- Values: `LogLevelDebug`, `LogLevelInfo`, `LogLevelWarn`, `LogLevelError`
- Description: Minimum log level to output

**Environment** (`Environment`)
- Type: String constant
- Values: `EnvDevelopment`, `EnvProduction`
- Description: Runtime environment (affects output format)

**ServiceName** (`string`)
- Type: String
- Description: Service name included in all log entries
- Validation: Cannot be empty or whitespace

### Log Levels

| Level | Use Case | Visibility |
|-------|----------|------------|
| `DEBUG` | Detailed debugging information | Development only |
| `INFO` | General informational messages | Default for production |
| `WARN` | Warning messages | Always visible |
| `ERROR` | Error messages | Always visible |
| `FATAL` | Fatal errors (terminates app) | Always visible |

### Environment Modes

#### Development Mode

```go
Environment: EnvDevelopment
```

**Output Format:** Console with colors
**Example:**
```
2025-01-15T10:30:00.123Z  INFO  my-service  application started  {"version": "1.0.0"}
```

**Features:**
- Human-readable format
- Color-coded log levels
- Easy to read during development

#### Production Mode

```go
Environment: EnvProduction
```

**Output Format:** JSON
**Example:**
```json
{
  "level": "info",
  "timestamp": "2025-01-15T10:30:00Z",
  "service": "my-service",
  "message": "application started",
  "version": "1.0.0"
}
```

**Features:**
- Machine-parseable JSON
- Optimized for log aggregation
- Compatible with ELK, Loki, etc.

---

## Best Practices

### 1. Initialize Once in main()

✅ **Good:**
```go
func main() {
    logger.MustInitFromEnv()
    defer logger.Sync()
    
    run()
}
```

❌ **Bad:**
```go
func someFunction() {
    logger.InitFromEnv() // Don't initialize in multiple places
}
```

### 2. Always Defer Sync

✅ **Good:**
```go
func main() {
    logger.MustInitFromEnv()
    defer logger.Sync() // Always defer
    
    // Application code
}
```

### 3. Use Structured Fields

✅ **Good:**
```go
logger.Info("user logged in",
    zap.String("user_id", userID),
    zap.String("ip", clientIP),
)
```

❌ **Bad:**
```go
logger.Info(fmt.Sprintf("user %s logged in from %s", userID, clientIP))
```

**Why:** Structured fields are:
- Searchable in log systems
- Type-safe
- More efficient

### 4. Create Contextual Loggers

✅ **Good:**
```go
type Handler struct {
    log *logger.Logger
}

func NewHandler() *Handler {
    return &Handler{
        log: logger.With(zap.String("component", "handler")),
    }
}
```

### 5. Log Errors with Context

✅ **Good:**
```go
if err := operation(); err != nil {
    logger.Error("operation failed",
        zap.Error(err),
        zap.String("user_id", userID),
        zap.String("operation", "payment"),
    )
}
```

❌ **Bad:**
```go
if err := operation(); err != nil {
    logger.Error(err.Error())
}
```

### 6. Use Appropriate Log Levels

```go
// DEBUG: Detailed debugging
logger.Debug("cache hit", zap.String("key", key))

// INFO: Normal operations
logger.Info("request processed", zap.Int("status", 200))

// WARN: Concerning but not critical
logger.Warn("rate limit approaching", zap.Int("remaining", 10))

// ERROR: Something failed
logger.Error("database connection failed", zap.Error(err))

// FATAL: Unrecoverable errors only
logger.Fatal("cannot start server", zap.Error(err))
```

### 7. Don't Log Sensitive Data

❌ **Bad:**
```go
logger.Info("user logged in",
    zap.String("password", password), // Never log passwords!
    zap.String("credit_card", cc),   // Never log PII!
)
```

✅ **Good:**
```go
logger.Info("user logged in",
    zap.String("user_id", userID),
    zap.String("method", "oauth"),
)
```

### 8. Use TryGet in Libraries

✅ **Good (library code):**
```go
func MyLibraryFunction() error {
    log, err := logger.TryGet()
    if err != nil {
        // Gracefully handle missing logger
        return nil
    }
    log.Debug("library operation")
    return nil
}
```

### 9. Consistent Field Naming

✅ **Good:**
```go
// Use consistent field names
logger.Info("event", zap.String("user_id", id))
logger.Info("event", zap.String("user_id", id)) // Same name
```

❌ **Bad:**
```go
logger.Info("event", zap.String("user_id", id))
logger.Info("event", zap.String("userId", id))  // Inconsistent
```

### 10. Avoid Logging in Hot Paths

❌ **Bad:**
```go
for i := 0; i < 1000000; i++ {
    logger.Debug("processing", zap.Int("index", i)) // Too much!
}
```

✅ **Good:**
```go
logger.Info("batch processing started")
for i := 0; i < 1000000; i++ {
    // Process without logging
}
logger.Info("batch processing completed", zap.Int("count", 1000000))
```