```markdown
# gologger

Production-ready structured logging for **gath-stack** applications, built on [Uber's zap](https://github.com/uber-go/zap).

## Features

- Fast structured logging with JSON output (production) and colorized console (development)
- **Automatic `.env` loading in development** - ignored in production for security
- **Strict validation** - fails fast if configuration is invalid
- Global and contextual logging interfaces
- Zero configuration needed for common use cases

## Installation

```bash
go get github.com/gath-stack/gologger
```

## Quick Start

**1. Create `.env` file (development only):**

```bash
LOG_LEVEL=DEBUG
APP_ENV=development
APP_NAME=my-service
```

**2. Initialize logger:**

```go
package main

import (
    "go.uber.org/zap"
    logger "github.com/gath-stack/gologger"
)

func main() {
    logger.MustInitFromEnv()
    defer logger.Get().Sync()

    logger.Info("app started", zap.String("version", "1.0.0"))
}
```

## Basic Usage

```go
// Simple logging
logger.Debug("debugging message")
logger.Info("user authenticated", zap.String("user_id", "abc123"))
logger.Warn("cache miss", zap.String("key", "session"))
logger.Error("query failed", zap.Error(err))

// Contextual logging
log := logger.With(
    zap.String("component", "auth"),
    zap.String("request_id", "req-42"),
)
log.Info("processing request")
```

## Configuration

### Environment Variables (Required)

| Variable    | Valid Values                    | Description           |
|-------------|--------------------------------|----------------------|
| `LOG_LEVEL` | `DEBUG`, `INFO`, `WARN`, `ERROR` | Logging verbosity    |
| `APP_ENV`   | `development`, `production`     | Runtime environment  |
| `APP_NAME`  | Any non-empty string           | Service name         |

### `.env` File Behavior

- **Development**: `.env` loaded automatically
- **Production**: `.env` ignored, uses system environment variables
- **Security**: Always add `.env` to `.gitignore`

## Production Deployment

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

## Documentation

📖 **For complete API documentation, examples, and advanced usage, see [docs/api-reference.md](docs/api-reference.md)**

Topics covered:
- All initialization methods
- Structured logging with Zap fields
- Contextual loggers
- Error handling
- Production deployment patterns
- Testing strategies
- Troubleshooting

## License

MIT License - Copyright (c) 2025 gath-stack
```