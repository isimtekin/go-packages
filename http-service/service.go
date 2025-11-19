package httpservice

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"strings"
	"sync"

	"github.com/valyala/fasthttp"
)

// Service represents the HTTP service
type Service struct {
	config    *Config
	routes    []*Route
	validator *Validator
	server    *fasthttp.Server

	// Middleware
	globalMiddleware []Middleware

	mu     sync.RWMutex
	closed bool
	wg     sync.WaitGroup
}

// New creates a new HTTP service
func New(opts ...Option) (*Service, error) {
	config := DefaultConfig()

	for _, opt := range opts {
		opt(config)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	service := &Service{
		config:           config,
		routes:           make([]*Route, 0),
		validator:        NewValidator(),
		globalMiddleware: make([]Middleware, 0),
		closed:           false,
	}

	// Setup fasthttp server
	service.server = &fasthttp.Server{
		Handler:            service.handler,
		ReadTimeout:        config.ReadTimeout,
		WriteTimeout:       config.WriteTimeout,
		IdleTimeout:        config.IdleTimeout,
		MaxRequestBodySize: config.MaxRequestBodySize,
	}

	// Register built-in middleware
	service.registerBuiltInMiddleware()

	// Register built-in routes
	service.registerBuiltInRoutes()

	return service, nil
}

// registerBuiltInMiddleware registers default middleware based on config
func (s *Service) registerBuiltInMiddleware() {
	if s.config.EnableRecovery {
		s.Use(Recovery())
	}

	if s.config.EnableRequestID {
		s.Use(RequestID())
	}

	if s.config.EnableLogger {
		s.Use(Logger())
	}

	if s.config.EnableCORS {
		s.Use(CORS(s.config))
	}

	if s.config.EnableRateLimiting {
		s.Use(RateLimit(s.config.RateLimitRequests, s.config.RateLimitWindow))
	}
}

// registerBuiltInRoutes registers built-in endpoints
func (s *Service) registerBuiltInRoutes() {
	if s.config.EnableHealthCheck {
		s.GET("/health", s.healthCheckHandler())
	}

	if s.config.EnableOpenAPI {
		s.GET("/openapi.json", s.openAPIHandler())
	}

	if s.config.EnableDocs {
		s.GET("/docs", s.docsHandler())
	}
}

// handler is the main fasthttp handler
func (s *Service) handler(ctx *fasthttp.RequestCtx) {
	s.mu.RLock()
	if s.closed {
		s.mu.RUnlock()
		ctx.SetStatusCode(fasthttp.StatusServiceUnavailable)
		ctx.SetBodyString("Service is closed")
		return
	}
	s.mu.RUnlock()

	// Find matching route
	method := string(ctx.Method())
	path := string(ctx.Path())

	// Handle CORS preflight requests (OPTIONS) when CORS is enabled
	if method == "OPTIONS" && s.config.EnableCORS {
		// Check if there's any route for this path (regardless of method)
		if s.hasRouteForPath(path) {
			// Handle preflight request
			s.handlePreflightRequest(ctx)
			return
		}
	}

	route := s.findRoute(method, path)
	if route == nil {
		ctx.SetStatusCode(fasthttp.StatusNotFound)
		WriteError(ctx, NotFound("Route not found"))
		return
	}

	// Create context
	reqCtx := context.Background()
	reqCtx = SetRequestCtx(reqCtx, ctx)

	// Parse path parameters
	params, _ := ParsePathParams(route.Path, path)
	reqCtx = SetPathParams(reqCtx, params)

	// Apply route-specific middleware
	handler := route.Handler
	handler = Use(handler, route.Middlewares...)

	// Apply global middleware
	handler = Use(handler, s.globalMiddleware...)

	// Execute handler
	if err := handler(reqCtx); err != nil {
		WriteError(ctx, err)
	}
}

// findRoute finds a matching route
func (s *Service) findRoute(method, path string) *Route {
	for _, route := range s.routes {
		if route.Method == method {
			// Simple exact match for now
			if route.Path == path {
				return route
			}

			// Check if path matches pattern
			if _, err := ParsePathParams(route.Path, path); err == nil {
				return route
			}
		}
	}
	return nil
}

// hasRouteForPath checks if there's any route (regardless of method) for the given path
func (s *Service) hasRouteForPath(path string) bool {
	for _, route := range s.routes {
		// Simple exact match for now
		if route.Path == path {
			return true
		}

		// Check if path matches pattern
		if _, err := ParsePathParams(route.Path, path); err == nil {
			return true
		}
	}
	return false
}

// handlePreflightRequest handles CORS preflight OPTIONS requests
func (s *Service) handlePreflightRequest(ctx *fasthttp.RequestCtx) {
	// Set CORS headers
	origin := string(ctx.Request.Header.Peek("Origin"))
	if origin != "" && isAllowedOrigin(origin, s.config.CORSAllowOrigins) {
		ctx.Response.Header.Set("Access-Control-Allow-Origin", origin)
	} else if len(s.config.CORSAllowOrigins) > 0 && s.config.CORSAllowOrigins[0] == "*" {
		ctx.Response.Header.Set("Access-Control-Allow-Origin", "*")
	}

	ctx.Response.Header.Set("Access-Control-Allow-Methods", strings.Join(s.config.CORSAllowMethods, ", "))
	ctx.Response.Header.Set("Access-Control-Allow-Headers", strings.Join(s.config.CORSAllowHeaders, ", "))

	if len(s.config.CORSExposeHeaders) > 0 {
		ctx.Response.Header.Set("Access-Control-Expose-Headers", strings.Join(s.config.CORSExposeHeaders, ", "))
	}

	if s.config.CORSAllowCredentials {
		ctx.Response.Header.Set("Access-Control-Allow-Credentials", "true")
	}

	if s.config.CORSMaxAge > 0 {
		ctx.Response.Header.Set("Access-Control-Max-Age", fmt.Sprintf("%d", s.config.CORSMaxAge))
	}

	// Return 204 No Content for preflight
	ctx.SetStatusCode(fasthttp.StatusNoContent)
}

// Use adds global middleware
func (s *Service) Use(middleware ...Middleware) {
	s.globalMiddleware = append(s.globalMiddleware, middleware...)
}

// GET registers a GET route
func (s *Service) GET(path string, handler interface{}, opts ...RouteOption) {
	s.addRoute("GET", path, handler, opts...)
}

// POST registers a POST route
func (s *Service) POST(path string, handler interface{}, opts ...RouteOption) {
	s.addRoute("POST", path, handler, opts...)
}

// PUT registers a PUT route
func (s *Service) PUT(path string, handler interface{}, opts ...RouteOption) {
	s.addRoute("PUT", path, handler, opts...)
}

// DELETE registers a DELETE route
func (s *Service) DELETE(path string, handler interface{}, opts ...RouteOption) {
	s.addRoute("DELETE", path, handler, opts...)
}

// PATCH registers a PATCH route
func (s *Service) PATCH(path string, handler interface{}, opts ...RouteOption) {
	s.addRoute("PATCH", path, handler, opts...)
}

// addRoute adds a route to the service
func (s *Service) addRoute(method, path string, handler interface{}, opts ...RouteOption) {
	route := &Route{
		Method:      method,
		Path:        path,
		Middlewares: make([]Middleware, 0),
	}

	// Convert handler to HandlerFunc
	switch h := handler.(type) {
	case HandlerFunc:
		route.Handler = h
	case func(context.Context) error:
		route.Handler = HandlerFunc(h)
	case SimpleHandler:
		route.Handler = WrapSimpleHandler(h)
	case func(context.Context) (interface{}, error):
		route.Handler = WrapSimpleHandler(h)
	case NoBodyHandler:
		route.Handler = WrapNoBodyHandler(h)
	default:
		// Try to detect RequestHandler pattern using reflection
		// func(ctx context.Context, req *T) (interface{}, error)
		handlerType := reflect.TypeOf(handler)
		if handlerType != nil && handlerType.Kind() == reflect.Func {
			if isRequestHandler(handlerType) {
				route.Handler = wrapGenericRequestHandler(handler, s.validator)
			} else {
				log.Printf("Warning: unsupported handler type for %s %s", method, path)
				return
			}
		} else {
			log.Printf("Warning: unsupported handler type for %s %s", method, path)
			return
		}
	}

	// Apply options
	for _, opt := range opts {
		opt(route)
	}

	s.routes = append(s.routes, route)
}

// Start starts the HTTP server
func (s *Service) Start() error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return ErrServiceClosed
	}
	s.mu.Unlock()

	addr := s.config.Addr()
	log.Printf("Starting HTTP service on %s", addr)
	log.Printf("OpenAPI docs: http://%s/docs", addr)
	log.Printf("Health check: http://%s/health", addr)

	return s.server.ListenAndServe(addr)
}

// StartAsync starts the server asynchronously
func (s *Service) StartAsync() error {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		if err := s.Start(); err != nil {
			log.Printf("Server error: %v", err)
		}
	}()
	return nil
}

// Shutdown gracefully shuts down the server
func (s *Service) Shutdown() error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return ErrAlreadyClosed
	}
	s.closed = true
	s.mu.Unlock()

	log.Println("Shutting down HTTP service...")

	if err := s.server.Shutdown(); err != nil {
		return fmt.Errorf("failed to shutdown server: %w", err)
	}

	s.wg.Wait()
	log.Println("HTTP service shut down successfully")

	return nil
}

// IsClosed returns true if the service is closed
func (s *Service) IsClosed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.closed
}

// Built-in handlers

// healthCheckHandler returns the health check handler
func (s *Service) healthCheckHandler() HandlerFunc {
	return func(ctx context.Context) error {
		response := NewHealthResponse("ok", s.config.Version)
		reqCtx := GetRequestCtx(ctx)
		return WriteResponse(reqCtx, OK(response))
	}
}

// openAPIHandler returns the OpenAPI spec handler
func (s *Service) openAPIHandler() HandlerFunc {
	return func(ctx context.Context) error {
		spec := GenerateOpenAPISpec(s.config, s.routes)
		reqCtx := GetRequestCtx(ctx)
		return WriteResponse(reqCtx, OK(spec))
	}
}

// docsHandler returns the Swagger UI handler
func (s *Service) docsHandler() HandlerFunc {
	return func(ctx context.Context) error {
		reqCtx := GetRequestCtx(ctx)
		html := generateSwaggerUI(s.config)
		reqCtx.SetContentType("text/html")
		reqCtx.SetBodyString(html)
		return nil
	}
}

// generateSwaggerUI generates a simple Swagger UI HTML page
func generateSwaggerUI(config *Config) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s - API Documentation</title>
    <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
    <script>
        window.onload = function() {
            SwaggerUIBundle({
                url: "/openapi.json",
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIBundle.SwaggerUIStandalonePreset
                ],
            });
        };
    </script>
</body>
</html>
`, config.Title)
}

// Validator returns the service's validator
func (s *Service) Validator() *Validator {
	return s.validator
}

// isRequestHandler checks if a function matches the RequestHandler signature:
// func(ctx context.Context, req *T) (interface{}, error)
func isRequestHandler(handlerType reflect.Type) bool {
	if handlerType.NumIn() != 2 || handlerType.NumOut() != 2 {
		return false
	}

	// First param must be context.Context
	contextType := reflect.TypeOf((*context.Context)(nil)).Elem()
	if !handlerType.In(0).Implements(contextType) {
		return false
	}

	// Second param must be a pointer
	if handlerType.In(1).Kind() != reflect.Ptr {
		return false
	}

	// First return must be interface{}
	interfaceType := reflect.TypeOf((*interface{})(nil)).Elem()
	if !handlerType.Out(0).Implements(interfaceType) && handlerType.Out(0) != interfaceType {
		return false
	}

	// Second return must be error
	errorType := reflect.TypeOf((*error)(nil)).Elem()
	if !handlerType.Out(1).Implements(errorType) {
		return false
	}

	return true
}

// wrapGenericRequestHandler wraps a RequestHandler-type function using reflection
func wrapGenericRequestHandler(handler interface{}, validator *Validator) HandlerFunc {
	handlerValue := reflect.ValueOf(handler)
	handlerType := handlerValue.Type()

	return func(ctx context.Context) error {
		// Create a new instance of the request type
		reqType := handlerType.In(1).Elem() // Get the type T from *T
		reqPtr := reflect.New(reqType)      // Create *T

		// Bind and validate the request
		if err := BindAndValidate(ctx, reqPtr.Interface(), validator); err != nil {
			return err
		}

		// Call the handler
		results := handlerValue.Call([]reflect.Value{
			reflect.ValueOf(ctx),
			reqPtr,
		})

		// Extract results
		var result interface{}
		var err error

		if !results[0].IsNil() {
			result = results[0].Interface()
		}

		if !results[1].IsNil() {
			err = results[1].Interface().(error)
		}

		if err != nil {
			return err
		}

		// Write response
		reqCtx := GetRequestCtx(ctx)
		if reqCtx != nil {
			return WriteResponse(reqCtx, OK(result))
		}

		return nil
	}
}
