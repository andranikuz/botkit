# BotKit Middleware Package

This package contains all middleware implementations for the BotKit framework.

## Structure

- **core.go** - Core middleware implementations for all transport types
- **http.go** - HTTP-specific middleware (CORS, compression, security headers)

## Core Middleware

### Universal Middleware
These work with any transport (Telegram, HTTP, WebSocket):

- **LoggingMiddleware** - Request/response logging with timing
- **RecoveryMiddleware** - Panic recovery and graceful error handling
- **AuthMiddleware** - Authentication validation
- **RateLimitMiddleware** - Request rate limiting per user
- **ValidationMiddleware** - Custom data validation
- **ContextMiddleware** - Request context with timeout
- **MetricsMiddleware** - Performance metrics collection

### HTTP-Specific Middleware
These are designed for HTTP/REST APIs:

- **CORSMiddleware** - Cross-Origin Resource Sharing headers
- **RequestIDMiddleware** - Unique request ID generation
- **CompressionMiddleware** - Gzip response compression
- **SecurityHeadersMiddleware** - Security headers (X-Frame-Options, etc.)

## Usage

```go
import (
    "github.com/andranikuz/botkit/middleware"
    "github.com/andranikuz/botkit/routing"
)

// Create router
router := routing.NewRouter(eventBus, logger, config)

// Register middleware
router.RegisterMiddleware(middleware.NewRecoveryMiddleware(logger, 100))
router.RegisterMiddleware(middleware.NewLoggingMiddleware(logger, 90))
router.RegisterMiddleware(middleware.NewRateLimitMiddleware(limiter, 80))
```

## Creating Custom Middleware

Implement the `routing.Middleware` interface:

```go
type MyMiddleware struct {
    priority int
}

func (m *MyMiddleware) Name() string {
    return "my_middleware"
}

func (m *MyMiddleware) Priority() int {
    return m.priority
}

func (m *MyMiddleware) Process(ctx core.UniversalContext, next core.HandlerFunc) core.Response {
    // Pre-processing
    // ...
    
    // Call next handler
    response := next(ctx)
    
    // Post-processing
    // ...
    
    return response
}
```

## Middleware Chain

Create chains for specific routes:

```go
chain := middleware.Chain(
    middleware.NewLoggingMiddleware(logger, 90),
    middleware.NewAuthMiddleware(authFunc, 80),
    middleware.NewValidationMiddleware(validator, 70),
)

wrappedHandler := chain(originalHandler)
```

## Priority Guidelines

- 100: Recovery (catch panics)
- 90: Logging
- 85: Authentication
- 80: Rate limiting
- 70: Validation
- 60: Context/timeout
- 50: Metrics
- 1-49: Custom middleware