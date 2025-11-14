package httpservice

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/valyala/fasthttp"
)

// Middleware is a function that wraps a handler
type Middleware func(next HandlerFunc) HandlerFunc

// HandlerFunc is the signature for HTTP handlers
type HandlerFunc func(ctx context.Context) error

// Use applies middleware to a handler
func Use(handler HandlerFunc, middleware ...Middleware) HandlerFunc {
	for i := len(middleware) - 1; i >= 0; i-- {
		handler = middleware[i](handler)
	}
	return handler
}

// Recovery middleware recovers from panics
func Recovery() Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx context.Context) (err error) {
			defer func() {
				if r := recover(); r != nil {
					err = InternalServerErrorf("panic recovered: %v", r)
					log.Printf("PANIC: %v", r)
				}
			}()
			return next(ctx)
		}
	}
}

// Logger middleware logs HTTP requests
func Logger() Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx context.Context) error {
			start := time.Now()
			reqCtx := GetRequestCtx(ctx)

			// Log request
			method := string(reqCtx.Method())
			path := string(reqCtx.Path())
			requestID := GetRequestID(ctx)

			// Execute handler
			err := next(ctx)

			// Log response
			duration := time.Since(start)
			status := reqCtx.Response.StatusCode()

			logMsg := fmt.Sprintf("[%s] %s %s - %d - %v",
				requestID,
				method,
				path,
				status,
				duration,
			)

			if err != nil {
				log.Printf("%s - ERROR: %v", logMsg, err)
			} else {
				log.Println(logMsg)
			}

			return err
		}
	}
}

// RequestID middleware generates a unique request ID
func RequestID() Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx context.Context) error {
			reqCtx := GetRequestCtx(ctx)

			// Check for existing request ID in headers
			requestID := string(reqCtx.Request.Header.Peek("X-Request-ID"))
			if requestID == "" {
				// Generate new request ID
				requestID = generateRequestID()
			}

			// Set request ID in context and response header
			ctx = SetRequestID(ctx, requestID)
			reqCtx.Response.Header.Set("X-Request-ID", requestID)

			return next(ctx)
		}
	}
}

// generateRequestID generates a simple request ID
func generateRequestID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// CORS middleware adds CORS headers
func CORS(config *Config) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx context.Context) error {
			reqCtx := GetRequestCtx(ctx)

			// Set CORS headers
			origin := string(reqCtx.Request.Header.Peek("Origin"))
			if origin != "" && isAllowedOrigin(origin, config.CORSAllowOrigins) {
				reqCtx.Response.Header.Set("Access-Control-Allow-Origin", origin)
			} else if len(config.CORSAllowOrigins) > 0 && config.CORSAllowOrigins[0] == "*" {
				reqCtx.Response.Header.Set("Access-Control-Allow-Origin", "*")
			}

			reqCtx.Response.Header.Set("Access-Control-Allow-Methods", strings.Join(config.CORSAllowMethods, ", "))
			reqCtx.Response.Header.Set("Access-Control-Allow-Headers", strings.Join(config.CORSAllowHeaders, ", "))

			if len(config.CORSExposeHeaders) > 0 {
				reqCtx.Response.Header.Set("Access-Control-Expose-Headers", strings.Join(config.CORSExposeHeaders, ", "))
			}

			if config.CORSAllowCredentials {
				reqCtx.Response.Header.Set("Access-Control-Allow-Credentials", "true")
			}

			if config.CORSMaxAge > 0 {
				reqCtx.Response.Header.Set("Access-Control-Max-Age", fmt.Sprintf("%d", config.CORSMaxAge))
			}

			// Handle preflight requests
			if string(reqCtx.Method()) == "OPTIONS" {
				reqCtx.SetStatusCode(fasthttp.StatusNoContent)
				return nil
			}

			return next(ctx)
		}
	}
}

// isAllowedOrigin checks if origin is allowed
func isAllowedOrigin(origin string, allowedOrigins []string) bool {
	for _, allowed := range allowedOrigins {
		if allowed == "*" || allowed == origin {
			return true
		}
	}
	return false
}

// Timeout middleware adds a timeout to requests
func Timeout(duration time.Duration) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx context.Context) error {
			ctx, cancel := context.WithTimeout(ctx, duration)
			defer cancel()

			done := make(chan error, 1)

			go func() {
				done <- next(ctx)
			}()

			select {
			case err := <-done:
				return err
			case <-ctx.Done():
				return ServiceUnavailable("Request timeout")
			}
		}
	}
}

// Auth middleware for authentication
func Auth(authFunc func(ctx context.Context) error) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx context.Context) error {
			if err := authFunc(ctx); err != nil {
				return err
			}
			return next(ctx)
		}
	}
}

// RateLimit middleware implements basic rate limiting
func RateLimit(requests int, window time.Duration) Middleware {
	// Simple in-memory rate limiter (use Redis for production)
	limiter := make(map[string]*rateLimiter)

	return func(next HandlerFunc) HandlerFunc {
		return func(ctx context.Context) error {
			reqCtx := GetRequestCtx(ctx)
			ip := reqCtx.RemoteAddr().String()

			// Get or create limiter for this IP
			if limiter[ip] == nil {
				limiter[ip] = &rateLimiter{
					requests: 0,
					window:   window,
					resetAt:  time.Now().Add(window),
				}
			}

			rl := limiter[ip]

			// Reset if window expired
			if time.Now().After(rl.resetAt) {
				rl.requests = 0
				rl.resetAt = time.Now().Add(window)
			}

			// Check limit
			if rl.requests >= requests {
				return &HTTPError{
					Code:    429,
					Message: "Too many requests",
				}
			}

			rl.requests++
			return next(ctx)
		}
	}
}

type rateLimiter struct {
	requests int
	window   time.Duration
	resetAt  time.Time
}

// Compress middleware compresses responses
func Compress() Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx context.Context) error {
			reqCtx := GetRequestCtx(ctx)

			// Check if client accepts gzip
			acceptEncoding := string(reqCtx.Request.Header.Peek("Accept-Encoding"))
			if !strings.Contains(acceptEncoding, "gzip") {
				return next(ctx)
			}

			// Execute handler
			err := next(ctx)
			if err != nil {
				return err
			}

			// Compress response
			reqCtx.Response.Header.Set("Content-Encoding", "gzip")
			fasthttp.CompressHandler(func(ctx *fasthttp.RequestCtx) {
				// Response already set by handler
			})(reqCtx)

			return nil
		}
	}
}
