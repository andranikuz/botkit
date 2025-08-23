# BotKit Middleware Documentation

## Overview

BotKit provides a powerful middleware system that allows you to intercept and modify requests/responses throughout the processing pipeline.

## Package Structure

All middleware implementations are located in the `middleware/` package:
- `middleware/core.go` - Core middleware implementations (logging, recovery, auth, etc.)
- `middleware/http.go` - HTTP-specific middleware (CORS, compression, security headers)

## Available Middleware

### 1. **LoggingMiddleware**
Logs all incoming requests and responses with timing information.

```go
loggingMW := middleware.NewLoggingMiddleware(logger, 90)
router.RegisterMiddleware(loggingMW)
```

### 2. **RecoveryMiddleware**
Catches panics and returns graceful error responses.

```go
recoveryMW := middleware.NewRecoveryMiddleware(logger, 100)
router.RegisterMiddleware(recoveryMW)
```

### 3. **RateLimitMiddleware**
Limits the number of requests per user within a time window.

```go
config := &routing.RateLimitConfig{
    Requests: 10,
    Window:   60, // seconds
}
rateLimiter := routing.NewRateLimiter(config, nil)
rateLimitMW := middleware.NewRateLimitMiddleware(rateLimiter, 80)
router.RegisterMiddleware(rateLimitMW)
```

### 4. **AuthMiddleware**
Validates user authentication.

```go
authMW := middleware.NewAuthMiddleware(func(ctx core.UniversalContext) bool {
    return ctx.GetUserID() > 0
}, 85)
router.RegisterMiddleware(authMW)
```

### 5. **ValidationMiddleware**
Validates request data before processing.

```go
validationMW := middleware.NewValidationMiddleware(func(ctx core.UniversalContext) error {
    if len(ctx.GetText()) > 1000 {
        return errors.New("message too long")
    }
    return nil
}, 70)
router.RegisterMiddleware(validationMW)
```

### 6. **ContextMiddleware**
Adds context with timeout to requests.

```go
contextMW := middleware.NewContextMiddleware(func(ctx core.UniversalContext) context.Context {
    reqCtx, _ := context.WithTimeout(context.Background(), 30*time.Second)
    return reqCtx
}, 60)
router.RegisterMiddleware(contextMW)
```

### 7. **MetricsMiddleware**
Collects metrics about requests and response times.

```go
metricsMW := middleware.NewMetricsMiddleware(metrics, 50)
router.RegisterMiddleware(metricsMW)
```

## HTTP-Specific Middleware

### 1. **CORSMiddleware**
Handles Cross-Origin Resource Sharing for HTTP APIs.

```go
cors := middleware.NewCORSMiddleware()
cors.AllowedOrigins = []string{"https://example.com"}
cors.AllowCredentials = true
```

### 2. **RequestIDMiddleware**
Adds unique request IDs to HTTP requests.

```go
requestID := middleware.NewRequestIDMiddleware()
```

### 3. **CompressionMiddleware**
Compresses HTTP responses using gzip.

```go
compression := middleware.NewCompressionMiddleware()
```

### 4. **SecurityHeadersMiddleware**
Adds security headers to HTTP responses.

```go
security := middleware.NewSecurityHeadersMiddleware()
```

## Creating Custom Middleware

### Simple Function Middleware

```go
customMW := middleware.NewMiddleware("custom", 50, 
    func(ctx core.UniversalContext, next core.HandlerFunc) core.Response {
        // Pre-processing
        ctx.SetParam("timestamp", time.Now())
        
        // Call next handler
        response := next(ctx)
        
        // Post-processing
        // ...
        
        return response
    })
router.RegisterMiddleware(customMW)
```

### Struct-Based Middleware

```go
type MyMiddleware struct {
    priority int
    service  MyService
}

func (m *MyMiddleware) Name() string {
    return "my_middleware"
}

func (m *MyMiddleware) Priority() int {
    return m.priority
}

func (m *MyMiddleware) Process(ctx core.UniversalContext, next core.HandlerFunc) core.Response {
    // Your logic here
    return next(ctx)
}
```

## Middleware Chain

Middleware are executed in order of priority (higher priority = executed first):

1. **Recovery** (100) - Catch panics
2. **Logging** (90) - Log requests
3. **Auth** (85) - Check authentication  
4. **RateLimit** (80) - Check rate limits
5. **Validation** (70) - Validate data
6. **Context** (60) - Add context
7. **Metrics** (50) - Collect metrics
8. **Custom** (1-49) - Your middleware

## Security Middleware

BotKit includes built-in security features through the `SecurityRule` system:

```go
route := routing.RoutePattern{
    Patterns: []string{"/admin"},
    Handler:  handleAdmin,
    Security: routing.SecurityRule{
        RequireAuth:        true,
        RequireRoles:       []string{"admin"},
        RequirePermissions: []string{"manage_users"},
        RateLimit: &routing.RateLimitConfig{
            Requests: 5,
            Window:   60,
        },
    },
}
```

## Best Practices

1. **Order Matters**: Recovery should be first, logging should be early
2. **Keep It Light**: Middleware should be fast and focused
3. **Error Handling**: Always handle errors gracefully
4. **Context Propagation**: Use context for request-scoped values
5. **Metrics**: Track performance in production
6. **Security First**: Apply security checks early in the chain

## Example: Complete Middleware Stack

```go
import "github.com/andranikuz/botkit/middleware"

func setupMiddleware(router *routing.Router) {
    // Recovery - catch panics
    router.RegisterMiddleware(
        middleware.NewRecoveryMiddleware(logger, 100))
    
    // Logging - track all requests
    router.RegisterMiddleware(
        middleware.NewLoggingMiddleware(logger, 90))
    
    // Security - rate limiting
    rateLimiter := routing.NewRateLimiter(&routing.RateLimitConfig{
        Requests: 100,
        Window:   60,
    }, storage)
    router.RegisterMiddleware(
        middleware.NewRateLimitMiddleware(rateLimiter, 80))
    
    // Validation - check message length
    router.RegisterMiddleware(
        middleware.NewValidationMiddleware(validateMessage, 70))
    
    // Metrics - collect stats
    router.RegisterMiddleware(
        middleware.NewMetricsMiddleware(metrics, 50))
}
```

## Testing Middleware

```go
// Create test context
ctx := &core.BaseContext{}
ctx.SetUserID(123)
ctx.SetText("test message")

// Create middleware chain
middlewares := []routing.Middleware{
    middleware.NewLoggingMiddleware(logger, 90),
    middleware.NewValidationMiddleware(validator, 70),
}

chain := middleware.Chain(middlewares...)

// Test handler
handler := func(ctx core.UniversalContext) core.Response {
    return core.NewMessage("Success!")
}

// Apply chain and execute
response := chain(handler)(ctx)
```