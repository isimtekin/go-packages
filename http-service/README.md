# HTTP Service

A FastAPI-inspired HTTP service framework for Go built on `fasthttp`. Create production-ready REST APIs with automatic OpenAPI documentation, request validation, and built-in middleware.

## Features

- **=€ FastHTTP-powered**: Built on `fasthttp` for maximum performance
- **=Ý Auto OpenAPI Docs**: Automatic OpenAPI 3.0 spec generation + Swagger UI at `/docs`
- ** Request Validation**: Integrated `go-playground/validator` for automatic validation
- **<¯ Type-Safe Handlers**: Generic handler types for compile-time safety
- **=' Middleware System**: Built-in middleware (CORS, logging, recovery, rate limiting, etc.)
- **<× Builder Pattern**: Fluent API for service and route configuration
- **¡ Context Support**: Full `context.Context` integration for cancellation and timeouts
- **= Production Ready**: Graceful shutdown, panic recovery, request IDs
- **=Ê Health & Metrics**: Built-in `/health` endpoint, optional `/metrics`
- **<¨ Clean API**: Inspired by FastAPI's developer experience

## Installation

```bash
go get github.com/isimtekin/go-packages/http-service
```

## Quick Start

### Basic Service

```go
package main

import (
	"context"
	"log"

	httpservice "github.com/isimtekin/go-packages/http-service"
)

func main() {
	// Create service
	service, err := httpservice.New(
		httpservice.WithTitle("My API"),
		httpservice.WithVersion("1.0.0"),
		httpservice.WithPort(8080),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Register routes
	service.GET("/hello", HelloHandler)

	// Start server
	log.Fatal(service.Start())
}

func HelloHandler(ctx context.Context) (interface{}, error) {
	return map[string]string{
		"message": "Hello, World!",
	}, nil
}
```

Visit:
- API: http://localhost:8080/hello
- Docs: http://localhost:8080/docs
- OpenAPI: http://localhost:8080/openapi.json
- Health: http://localhost:8080/health

### CRUD Example

```go
package main

import (
	"context"
	"log"

	httpservice "github.com/isimtekin/go-packages/http-service"
)

type CreateUserRequest struct {
	Name  string `json:"name" validate:"required,min=3,max=50"`
	Email string `json:"email" validate:"required,email"`
	Age   int    `json:"age" validate:"required,min=18,max=120"`
}

type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

func main() {
	service, _ := httpservice.New(
		httpservice.WithTitle("User Service"),
		httpservice.WithVersion("1.0.0"),
	)

	// Register CRUD routes
	service.GET("/users", ListUsers,
		httpservice.WithTags("users"),
		httpservice.WithSummary("List all users"),
	)

	service.POST("/users", CreateUser,
		httpservice.WithTags("users"),
		httpservice.WithSummary("Create a new user"),
		httpservice.WithRequestBody(&CreateUserRequest{}),
	)

	service.GET("/users/{id}", GetUser,
		httpservice.WithTags("users"),
		httpservice.WithSummary("Get user by ID"),
	)

	service.PUT("/users/{id}", UpdateUser,
		httpservice.WithTags("users"),
		httpservice.WithSummary("Update user"),
	)

	service.DELETE("/users/{id}", DeleteUser,
		httpservice.WithTags("users"),
		httpservice.WithSummary("Delete user"),
	)

	log.Fatal(service.Start())
}

func ListUsers(ctx context.Context) (interface{}, error) {
	users := []User{
		{ID: "1", Name: "John Doe", Email: "john@example.com", Age: 30},
		{ID: "2", Name: "Jane Smith", Email: "jane@example.com", Age: 25},
	}
	return users, nil
}

func CreateUser(ctx context.Context, req *CreateUserRequest) (interface{}, error) {
	// Validation is automatic
	user := &User{
		ID:    "3",
		Name:  req.Name,
		Email: req.Email,
		Age:   req.Age,
	}
	return user, nil
}

func GetUser(ctx context.Context) (interface{}, error) {
	id := httpservice.PathParam(ctx, "id")
	user := &User{
		ID:    id,
		Name:  "John Doe",
		Email: "john@example.com",
		Age:   30,
	}
	return user, nil
}

func UpdateUser(ctx context.Context, req *CreateUserRequest) (interface{}, error) {
	id := httpservice.PathParam(ctx, "id")
	user := &User{
		ID:    id,
		Name:  req.Name,
		Email: req.Email,
		Age:   req.Age,
	}
	return user, nil
}

func DeleteUser(ctx context.Context) (interface{}, error) {
	return httpservice.NoContent(), nil
}
```

### Request Context Helpers

```go
func MyHandler(ctx context.Context) (interface{}, error) {
	// Path parameters
	id := httpservice.PathParam(ctx, "id")

	// Query parameters
	page := httpservice.QueryParamInt(ctx, "page", 1)
	limit := httpservice.QueryParamInt(ctx, "limit", 20)
	search := httpservice.QueryParam(ctx, "q")

	// Headers
	token := httpservice.Header(ctx, "Authorization")
	userAgent := httpservice.Header(ctx, "User-Agent")

	// Request info
	method := httpservice.Method(ctx)
	path := httpservice.Path(ctx)
	ip := httpservice.RemoteAddr(ctx)

	return map[string]interface{}{
		"id":     id,
		"page":   page,
		"limit":  limit,
		"search": search,
		"method": method,
	}, nil
}
```

### Error Handling

```go
func MyHandler(ctx context.Context) (interface{}, error) {
	// Return HTTP errors
	return nil, httpservice.BadRequest("Invalid input")
	return nil, httpservice.NotFound("User not found")
	return nil, httpservice.Unauthorized("Invalid token")
	return nil, httpservice.Forbidden("Access denied")
	return nil, httpservice.InternalServerError("Database error")

	// With formatted messages
	return nil, httpservice.BadRequestf("Invalid ID: %s", id)
	return nil, httpservice.NotFoundf("User %s not found", id)

	// With details
	err := httpservice.BadRequest("Validation failed")
	err.WithDetails(map[string]interface{}{
		"field": "email",
		"error": "invalid format",
	})
	return nil, err
}
```

### Middleware

```go
func main() {
	service, _ := httpservice.New()

	// Global middleware (applied to all routes)
	service.Use(
		httpservice.Recovery(),      // Panic recovery
		httpservice.Logger(),         // Request logging
		httpservice.RequestID(),      // Request ID generation
		httpservice.CORS(service.config), // CORS headers
	)

	// Custom middleware
	service.Use(AuthMiddleware())

	// Per-route middleware
	service.POST("/admin/users", AdminHandler,
		httpservice.WithMiddleware(RequireAdmin()),
	)

	service.Start()
}

func AuthMiddleware() httpservice.Middleware {
	return func(next httpservice.HandlerFunc) httpservice.HandlerFunc {
		return func(ctx context.Context) error {
			token := httpservice.Header(ctx, "Authorization")
			if token == "" {
				return httpservice.Unauthorized("Missing authorization token")
			}

			// Validate token...
			// Add user to context...

			return next(ctx)
		}
	}
}

func RequireAdmin() httpservice.Middleware {
	return func(next httpservice.HandlerFunc) httpservice.HandlerFunc {
		return func(ctx context.Context) error {
			// Check if user is admin...
			isAdmin := true // TODO: check from context

			if !isAdmin {
				return httpservice.Forbidden("Admin access required")
			}

			return next(ctx)
		}
	}
}
```

### Configuration

```go
service, err := httpservice.New(
	// Server info
	httpservice.WithTitle("My API"),
	httpservice.WithServiceDescription("API for my application"),
	httpservice.WithVersion("1.0.0"),

	// Network
	httpservice.WithHost("0.0.0.0"),
	httpservice.WithPort(8080),

	// Timeouts
	httpservice.WithReadTimeout(30*time.Second),
	httpservice.WithWriteTimeout(30*time.Second),
	httpservice.WithIdleTimeout(120*time.Second),

	// Limits
	httpservice.WithMaxRequestBodySize(10*1024*1024), // 10MB

	// Features
	httpservice.WithDocs(true),         // Enable /docs and /openapi.json
	httpservice.WithHealthCheck(true),  // Enable /health
	httpservice.WithMetrics(false),     // Disable /metrics
	httpservice.WithCORS(true),         // Enable CORS
	httpservice.WithRequestID(true),    // Enable request ID
	httpservice.WithLogger(true),       // Enable logging
	httpservice.WithRecovery(true),     // Enable panic recovery
	httpservice.WithCompression(true),  // Enable gzip compression
	httpservice.WithValidation(true),   // Enable request validation

	// CORS configuration
	httpservice.WithCORSOrigins("*"),

	// Rate limiting
	httpservice.WithRateLimiting(true, 100, time.Minute), // 100 req/min

	// Debug
	httpservice.WithDebug(false),
)
```

### Request Validation

```go
type CreateProductRequest struct {
	Name        string   `json:"name" validate:"required,min=3,max=100"`
	Description string   `json:"description" validate:"required,max=500"`
	Price       float64  `json:"price" validate:"required,gt=0"`
	SKU         string   `json:"sku" validate:"required,alphanum,len=8"`
	Tags        []string `json:"tags" validate:"max=10,dive,min=2,max=20"`
	Email       string   `json:"email" validate:"required,email"`
	Website     string   `json:"website" validate:"omitempty,url"`
}

// Validation happens automatically
func CreateProduct(ctx context.Context, req *CreateProductRequest) (interface{}, error) {
	// req is already validated
	product := &Product{
		Name:  req.Name,
		Price: req.Price,
		SKU:   req.SKU,
	}
	return product, nil
}
```

Validation error response:
```json
{
	"message": "Validation failed",
	"details": {
		"errors": [
			{
				"field": "email",
				"message": "email must be a valid email",
				"tag": "email"
			}
		]
	}
}
```

### Graceful Shutdown

```go
func main() {
	service, _ := httpservice.New()

	service.GET("/hello", HelloHandler)

	// Start async
	service.StartAsync()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown
	if err := service.Shutdown(); err != nil {
		log.Fatal("Forced shutdown:", err)
	}

	log.Println("Server stopped")
}
```

## API Reference

### Service Methods

#### `New(opts ...Option) (*Service, error)`
Creates a new HTTP service with functional options.

#### `(s *Service) GET(path string, handler interface{}, opts ...RouteOption)`
Registers a GET route.

#### `(s *Service) POST(path string, handler interface{}, opts ...RouteOption)`
Registers a POST route.

#### `(s *Service) PUT(path string, handler interface{}, opts ...RouteOption)`
Registers a PUT route.

#### `(s *Service) DELETE(path string, handler interface{}, opts ...RouteOption)`
Registers a DELETE route.

#### `(s *Service) PATCH(path string, handler interface{}, opts ...RouteOption)`
Registers a PATCH route.

#### `(s *Service) Use(middleware ...Middleware)`
Adds global middleware.

#### `(s *Service) Start() error`
Starts the HTTP server (blocking).

#### `(s *Service) StartAsync() error`
Starts the HTTP server asynchronously.

#### `(s *Service) Shutdown() error`
Gracefully shuts down the server.

### Handler Types

```go
// Simple handler (no request body)
type SimpleHandler func(ctx context.Context) (interface{}, error)

// Request handler (with request body)
type RequestHandler[Req any] func(ctx context.Context, req *Req) (interface{}, error)

// Typed handler (typed request and response)
type TypedHandler[Req any, Res any] func(ctx context.Context, req *Req) (*Res, error)

// No body handler (for GET/DELETE)
type NoBodyHandler func(ctx context.Context) (interface{}, error)
```

### Route Options

- `WithTags(tags ...string)` - Add OpenAPI tags
- `WithSummary(summary string)` - Add route summary
- `WithDescription(desc string)` - Add route description
- `WithDeprecated()` - Mark route as deprecated
- `WithRequestBody(body interface{})` - Set example request body
- `WithResponse(code int, response interface{})` - Set example response
- `WithMiddleware(middleware ...Middleware)` - Add route-specific middleware

### Built-in Middleware

- `Recovery()` - Panic recovery
- `Logger()` - Request logging
- `RequestID()` - Request ID generation
- `CORS(config *Config)` - CORS headers
- `Timeout(duration)` - Request timeout
- `Auth(authFunc)` - Authentication
- `RateLimit(requests, window)` - Rate limiting
- `Compress()` - Response compression

### Error Helpers

- `BadRequest(message)` - 400
- `Unauthorized(message)` - 401
- `Forbidden(message)` - 403
- `NotFound(message)` - 404
- `Conflict(message)` - 409
- `UnprocessableEntity(message)` - 422
- `InternalServerError(message)` - 500
- `ServiceUnavailable(message)` - 503

### Response Helpers

- `OK(body)` - 200
- `Created(body)` - 201
- `Accepted(body)` - 202
- `NoContent()` - 204
- `JSON(statusCode, body)` - Custom status

### Context Helpers

- `PathParam(ctx, name)` - Get path parameter
- `QueryParam(ctx, name)` - Get query parameter
- `QueryParamInt(ctx, name, default)` - Get query parameter as int
- `QueryParamBool(ctx, name, default)` - Get query parameter as bool
- `Header(ctx, name)` - Get header value
- `SetHeader(ctx, name, value)` - Set response header
- `SetStatus(ctx, code)` - Set response status
- `Method(ctx)` - Get HTTP method
- `Path(ctx)` - Get request path
- `RemoteAddr(ctx)` - Get remote address

## Testing

```bash
# Run tests
make test

# Run tests with coverage
make test-coverage

# Run tests with race detector
make test-race

# View HTML coverage report
make coverage-html

# Run benchmarks
make bench
```

## Performance

Built on `fasthttp`, one of the fastest HTTP routers for Go:
- Zero allocations for common operations
- Fast JSON encoding/decoding
- Efficient middleware chain
- Connection pooling

## Requirements

- Go 1.21 or higher
- Dependencies:
  - `github.com/valyala/fasthttp`
  - `github.com/go-playground/validator/v10`

## License

This project is licensed under the MIT License.

## Related Packages

- [slack-notifier](../slack-notifier) - Slack webhook notifier
- [kafka-client](../kafka-client) - Kafka client
- [nats-client](../nats-client) - NATS messaging client
- [mongo-client](../mongo-client) - MongoDB client wrapper
- [redis-client](../redis-client) - Redis client
- [crypto-utils](../crypto-utils) - Cryptographic utilities

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Support

For issues, questions, or contributions, please visit the [GitHub repository](https://github.com/isimtekin/go-packages).
