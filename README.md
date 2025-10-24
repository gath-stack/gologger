# gologger

This package provides a unified, high-performance, and production-ready structured logging system for **gath-stack** web applications.

The package is built on top of **Uber's [zap](https://github.com/uber-go/zap)** library, providing both structured JSON logging for production environments and colorized console logging for local development.

It can be also used combined with gobservability.

---

## Overview

The `gologger` package abstracts away the complexity of configuring and managing loggers with **strict validation** to ensure your application never starts with invalid configuration.

It automatically adapts to the runtime environment, loads `.env` files in development for convenience, and provides both global and contextual logging interfaces.

### Key features

- Fast, structured, leveled logging using `zap`
- JSON output in production for ingestion by log pipelines (Loki, Elasticsearch, etc.)
- Colorized human-readable console output in development
- **Automatic `.env` file loading in development** - no manual environment setup needed
- **Production-safe** - `.env` files are ignored when `APP_ENV=production`
- **Strict configuration validation** - application fails to start if config is invalid
- **Required environment variables** - no default fallbacks in production
- Global singleton logger for easy application-wide access
- Contextual structured fields for rich, queryable logs
- Environment-aware configuration via environment variables
- Type-safe configuration with compile-time guarantees

---

## Installation

```bash
go get github.com/gath-stack/gologger
```

---

## Quick Start

### 1. Create a `.env` file for development

```bash
# .env
LOG_LEVEL=DEBUG
APP_ENV=development
APP_NAME=my-awesome-service
```

**Important:** Add `.env` to your `.gitignore` to prevent committing secrets:

```bash
echo ".env" >> .gitignore
```

### 2. Initialize the logger

```go
package main

import (
    "go.uber.org/zap"
    logger "github.com/gath-stack/gologger"
)

func main() {
    // One-line initialization with automatic .env loading
    logger.MustInitFromEnv()
    defer logger.Get().Sync()

    logger.Info("application started", zap.String("version", "1.0.0"))
}
```

That's it! In development, the `.env` file is loaded automatically. In production, environment variables are used directly.

---

## Usage

### Basic logging

The package provides convenience functions for leveled logging:

```go
logger.Debug("debugging connection", zap.String("endpoint", "/api/v1"))
logger.Info("user authenticated", zap.String("user_id", "abc123"))
logger.Warn("cache miss", zap.String("key", "session_token"))
logger.Error("database query failed", zap.Error(err))
```

Each log entry includes contextual metadata such as the service name and environment.

---

### Contextual loggers

You can derive new loggers with additional structured fields for contextual enrichment:

```go
log := logger.WithFields(
    zap.String("component", "auth"),
    zap.String("request_id", "req-42a"),
)

log.Info("authentication succeeded", zap.String("user", "alice"))
```

This produces a structured log entry:

```json
{
  "timestamp": "2025-10-24T12:34:56Z",
  "level": "info",
  "message": "authentication succeeded",
  "component": "auth",
  "request_id": "req-42a",
  "user": "alice",
  "service": "my-awesome-service",
  "environment": "development"
}
```

---

## Configuration

### Environment Variables

The logger reads its configuration from environment variables. **All variables are mandatory** - the application will not start if any are missing or invalid.

| Variable    | Description                                      | Valid Values                       | Required |
| ----------- | ------------------------------------------------ | ---------------------------------- | -------- |
| `LOG_LEVEL` | Logging level                                    | `DEBUG`, `INFO`, `WARN`, `ERROR`   | ✅ Yes   |
| `APP_ENV`   | Runtime environment                              | `development`, `production`        | ✅ Yes   |
| `APP_NAME`  | Service name used for log enrichment             | Any non-empty string               | ✅ Yes   |

### `.env` File Loading Behavior

The logger **automatically handles** `.env` file loading based on the environment:

| Environment | `.env` Loading | Use Case |
|-------------|---------------|----------|
| **Development** (`APP_ENV != production`) | ✅ **Loaded automatically** | Local development with convenience |
| **Production** (`APP_ENV = production`) | ❌ **Ignored** | Uses system environment variables |
| **Missing `.env`** | ⚠️ **Falls back to system env vars** | CI/CD or container environments |

**Security Note:** The `.env` file is **never loaded in production** to prevent accidental exposure of secrets. Always set environment variables through your deployment platform (Kubernetes secrets, Docker environment, etc.).

---

## Initialization Methods

The package provides three ways to initialize the logger, depending on your needs:

### Method 1: Simple (Recommended)

```go
// Fail-fast: panics if configuration is invalid
logger.MustInitFromEnv()
defer logger.Get().Sync()
```

**Use this when:** You want simple, one-line initialization with automatic `.env` loading.

### Method 2: With Error Handling

```go
// Returns error instead of panicking
if err := logger.InitFromEnv(); err != nil {
    log.Fatalf("failed to initialize logger: %v", err)
}
defer logger.Get().Sync()
```

**Use this when:** You want explicit error handling and custom error messages.

### Method 3: Advanced (Direct Configuration)

```go
cfg := logger.Config{
    Level:       logger.LevelInfo,
    Environment: logger.EnvProduction,
    ServiceName: "my-service",
}

if err := logger.InitGlobal(cfg); err != nil {
    panic(err)
}
defer logger.Get().Sync()
```

**Use this when:** You need programmatic configuration or testing scenarios.

---

## Development Workflow

### Local Development

1. **Create `.env` file:**
   ```bash
   LOG_LEVEL=DEBUG
   APP_ENV=development
   APP_NAME=my-service
   ```

2. **Run your application:**
   ```bash
   go run main.go
   ```
   
   The `.env` file is loaded automatically, and logs appear in colorized console format.

### Testing

```go
func TestMyFunction(t *testing.T) {
    // Setup logger for tests
    cfg := logger.Config{
        Level:       logger.LevelDebug,
        Environment: logger.EnvDevelopment,
        ServiceName: "test-service",
    }
    logger.InitGlobal(cfg)
    
    // Your test code here
    logger.Info("test running")
}
```

### Production Deployment

Set environment variables in your deployment platform:

**Docker:**
```dockerfile
ENV APP_ENV=production
ENV APP_NAME=my-service
ENV LOG_LEVEL=INFO
```

**Kubernetes:**
```yaml
env:
  - name: APP_ENV
    value: "production"
  - name: APP_NAME
    value: "my-service"
  - name: LOG_LEVEL
    value: "INFO"
```

**Docker Compose:**
```yaml
environment:
  APP_ENV: production
  APP_NAME: my-service
  LOG_LEVEL: INFO
```

---

## Example Integration

### HTTP Server

```go
package server

import (
    "net/http"
    "go.uber.org/zap"
    logger "github.com/gath-stack/gologger"
)

func HandleRequest(w http.ResponseWriter, r *http.Request) {
    // Create contextual logger with request metadata
    log := logger.WithFields(
        zap.String("component", "server"),
        zap.String("request_id", r.Header.Get("X-Request-ID")),
        zap.String("method", r.Method),
        zap.String("path", r.URL.Path),
    )
    
    log.Info("processing request")
    
    // ... handle request ...
    
    log.Info("request completed", zap.Int("status", http.StatusOK))
}
```

### Background Worker

```go
package worker

import (
    "go.uber.org/zap"
    logger "github.com/gath-stack/gologger"
)

func ProcessJob(jobID string) error {
    log := logger.WithFields(
        zap.String("component", "worker"),
        zap.String("job_id", jobID),
    )
    
    log.Info("job started")
    
    // ... process job ...
    
    if err != nil {
        log.Error("job failed", zap.Error(err))
        return err
    }
    
    log.Info("job completed successfully")
    return nil
}
```

---

## Configuration Validation

The package validates all configuration before creating a logger:

```go
cfg := logger.Config{
    Level:       logger.LevelInfo,
    Environment: logger.EnvProduction,
    ServiceName: "my-service",
}

// Validate returns an error if config is invalid
if err := cfg.Validate(); err != nil {
    panic(err)
}

l, err := logger.New(cfg)
if err != nil {
    panic(err)
}
```

Validation checks:
- Log level must be one of: `DEBUG`, `INFO`, `WARN`, `ERROR`
- Environment must be one of: `development`, `production`
- Service name must not be empty or whitespace-only

---

## Error Handling

The package defines specific error types for different validation failures:

```go
var (
    ErrInvalidLogLevel        = errors.New("invalid log level")
    ErrInvalidEnvironment     = errors.New("invalid environment")
    ErrMissingServiceName     = errors.New("service name is required")
    ErrMissingRequiredEnvVar  = errors.New("required environment variable is not set")
)
```

You can check for specific errors:

```go
if err := logger.InitFromEnv(); err != nil {
    if errors.Is(err, logger.ErrMissingRequiredEnvVar) {
        fmt.Println("Missing required environment variable")
    }
    os.Exit(1)
}
```

---

## Integration Guidelines

* **Always** initialize the global logger at the start of your application
* The application will panic if you call `Get()` before initialization - this is intentional
* In development, create a `.env` file in the root of your project
* In production, set environment variables through your deployment platform
* **Never commit** `.env` files to version control - add them to `.gitignore`
* Prefer JSON output in production for ingestion by observability pipelines (Loki, Elasticsearch)
* Avoid using `logger.Fatal` except during startup or non-recoverable failures
* Use structured fields (`zap.Field`) for all dynamic data instead of string concatenation
* When integrating across packages, **never create new logger instances directly**; use the global logger or derive contextual loggers with `WithFields`

---

## Production Deployment Checklist

Before deploying to production, ensure:

- [ ] `APP_ENV=production` is set in your deployment environment
- [ ] `LOG_LEVEL` and `APP_NAME` environment variables are set
- [ ] `.env` files are **not** included in production builds or containers
- [ ] `.env` is added to `.gitignore`
- [ ] Logger is initialized in `main()` before any other operations
- [ ] `defer logger.Get().Sync()` is called to flush logs on exit
- [ ] Your deployment fails if required environment variables are missing (fail-fast)

---

## Troubleshooting

### Error: "required environment variable is not set"

**In Development:**
- Create a `.env` file in the root of your project with all required variables
- Ensure the file is in the correct location (same directory where you run the app)

**In Production:**
- Set environment variables through your deployment platform
- Verify `APP_ENV=production` is set
- Check that all required variables are present in your deployment configuration

### Error: ".env file loading failed"

- Check that the `.env` file format is correct (`KEY=value`, no spaces around `=`)
- Verify the file is readable and not corrupted
- Ensure you have read permissions on the file

### Logs not appearing

- Call `defer logger.Get().Sync()` in your `main()` to flush buffered logs
- Check that `LOG_LEVEL` is set appropriately (e.g., `DEBUG` to see all logs)

### `.env` file ignored in production

This is **correct behavior**. When `APP_ENV=production`, the `.env` file is intentionally ignored for security. Set environment variables through your deployment platform.

---

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

---

## License

MIT License

Copyright (c) 2025 gath-stack

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.