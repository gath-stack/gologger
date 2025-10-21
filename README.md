# gologger

This package provides a unified, high-performance, and production-ready structured logging system for **gath-stack** web applications.

The package is built on top of **Uber's [zap](https://github.com/uber-go/zap)** library, providing both structured JSON logging for production environments and colorized console logging for local development.

It can be also used combined with gobservability.

---

## Overview

The `gologger` package abstracts away the complexity of configuring and managing loggers with **strict validation** to ensure your application never starts with invalid configuration.

It automatically adapts to the runtime environment and provides both global and contextual logging interfaces.

### Key features

- Fast, structured, leveled logging using `zap`
- JSON output in production for ingestion by log pipelines (Loki, Elasticsearch, etc.)
- Colorized human-readable console output in development
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

## Usage

### 1. Initialize the logger at startup

You **must** initialize the global logger as early as possible in your application, typically in `main()`.

The application will **fail to start** if required environment variables are missing or invalid.

```go
package main

import (
    "fmt"
    "os"
    "go.uber.org/zap"
    logger "github.com/gath-stack/gologger"
)

func main() {
    // Option 1: Explicit error handling
    cfg, err := logger.FromEnv()
    if err != nil {
        fmt.Fprintf(os.Stderr, "FATAL: invalid logger configuration: %v\n", err)
        os.Exit(1)
    }
    
    if err := logger.InitGlobal(cfg); err != nil {
        fmt.Fprintf(os.Stderr, "FATAL: failed to initialize logger: %v\n", err)
        os.Exit(1)
    }
    defer logger.Get().Sync()

    // Option 2: Fail-fast with panic (recommended for main)
    cfg := logger.MustFromEnv()
    if err := logger.InitGlobal(cfg); err != nil {
        panic(err)
    }
    defer logger.Get().Sync()

    logger.Info("application started", zap.String("version", "1.0.0"))
}
```

---

### 2. Using the global logger

The package provides convenience functions for leveled logging:

```go
logger.Debug("debugging connection", zap.String("endpoint", "/api/v1"))
logger.Info("user authenticated", zap.String("user_id", "abc123"))
logger.Warn("cache miss", zap.String("key", "session_token"))
logger.Error("database query failed", zap.Error(err))
```

Each log entry includes contextual metadata such as the service name and environment.

---

### 3. Using contextual loggers

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
  "timestamp": "2025-10-19T12:34:56Z",
  "level": "info",
  "message": "authentication succeeded",
  "component": "auth",
  "request_id": "req-42a",
  "user": "alice",
  "service": "api-service",
  "environment": "production"
}
```

---

### 4. Configuration via environment variables

The logger reads its configuration from **required** environment variables when using `gologger.FromEnv()`.

**All variables are mandatory** - the application will not start if any are missing or invalid.

| Variable    | Description                                      | Valid Values                       | Required |
| ----------- | ------------------------------------------------ | ---------------------------------- | -------- |
| `LOG_LEVEL` | Logging level                                    | `DEBUG`, `INFO`, `WARN`, `ERROR`   | ✅ Yes   |
| `APP_ENV`   | Runtime environment                              | `development`, `production`        | ✅ Yes   |
| `APP_NAME`  | Service name used for log enrichment             | Any non-empty string               | ✅ Yes   |

Example:

```bash
export LOG_LEVEL=INFO
export APP_ENV=production
export APP_NAME=api-service
```

**Note:** If any variable is missing, `FromEnv()` returns an error and the application must terminate.

---

### 5. Configuration validation

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

### 6. Flushing logs

Always ensure that buffered logs are flushed before the process exits:

```go
defer logger.Get().Sync()
```

This is especially important in production to prevent loss of pending log entries.

---

## Integration guidelines

* **Always** initialize the global logger at the start of your application
* The application will panic if you call `Get()` before initialization - this is intentional
* In production environments, prefer JSON output for ingestion by observability pipelines such as Loki or Elasticsearch
* Avoid using `gologger.Fatal` except during startup or non-recoverable failures
* Use structured fields (`zap.Field`) for all dynamic data instead of string concatenation
* When integrating across packages, **never create new logger instances directly**; use the global logger or derive contextual loggers with `WithFields`
* Set all required environment variables in your deployment configuration (Kubernetes secrets, Docker Compose, systemd, etc.)

---

## Example integration

```go
package server

import (
    "go.uber.org/zap"
    logger "github.com/gath-stack/gologger"
)

func HandleRequest(requestID string) {
    log := logger.WithFields(
        zap.String("component", "server"),
        zap.String("request_id", requestID),
    )
    
    log.Info("processing request")
    
    // ... handle request ...
    
    log.Info("request processed successfully")
}
```

---

## Error handling

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
cfg, err := logger.FromEnv()
if err != nil {
    if errors.Is(err, logger.ErrMissingRequiredEnvVar) {
        fmt.Println("Missing required environment variable")
    }
    os.Exit(1)
}
```

---

## Production deployment checklist

Before deploying to production, ensure:

- `LOG_LEVEL` environment variable is set
- `APP_ENV` is set to `production`
- `APP_NAME` is set to your service identifier
- Logger is initialized in `main()` before any other operations
- `defer logger.Get().Sync()` is called to flush logs on exit
- Your deployment fails if environment variables are missing (fail-fast)

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