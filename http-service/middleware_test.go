package httpservice

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/valyala/fasthttp"
)

func TestRecoveryMiddleware(t *testing.T) {
	middleware := Recovery()

	handler := func(ctx context.Context) error {
		panic("test panic")
	}

	wrappedHandler := middleware(handler)

	// Create test context
	reqCtx := &fasthttp.RequestCtx{}
	ctx := SetRequestCtx(context.Background(), reqCtx)

	err := wrappedHandler(ctx)
	if err == nil {
		t.Error("Expected error from panic recovery")
	}

	httpErr, ok := err.(*HTTPError)
	if !ok {
		t.Error("Expected HTTPError")
	}

	if httpErr.Code != 500 {
		t.Errorf("Expected status 500, got %d", httpErr.Code)
	}
}

func TestLoggerMiddleware(t *testing.T) {
	middleware := Logger()

	callCount := 0
	handler := func(ctx context.Context) error {
		callCount++
		return nil
	}

	wrappedHandler := middleware(handler)

	// Create test context
	reqCtx := &fasthttp.RequestCtx{}
	reqCtx.Request.SetRequestURI("/test")
	ctx := SetRequestCtx(context.Background(), reqCtx)
	ctx = SetRequestID(ctx, "test-123")

	err := wrappedHandler(ctx)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if callCount != 1 {
		t.Errorf("Expected handler to be called once, got %d", callCount)
	}
}

func TestRequestIDMiddleware(t *testing.T) {
	middleware := RequestID()

	handler := func(ctx context.Context) error {
		requestID := GetRequestID(ctx)
		if requestID == "" {
			t.Error("Expected request ID to be set")
		}
		return nil
	}

	wrappedHandler := middleware(handler)

	// Test without existing request ID
	reqCtx := &fasthttp.RequestCtx{}
	ctx := SetRequestCtx(context.Background(), reqCtx)

	err := wrappedHandler(ctx)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Check response header
	responseID := string(reqCtx.Response.Header.Peek("X-Request-ID"))
	if responseID == "" {
		t.Error("Expected X-Request-ID header in response")
	}

	// Test with existing request ID
	reqCtx2 := &fasthttp.RequestCtx{}
	reqCtx2.Request.Header.Set("X-Request-ID", "existing-123")
	ctx2 := SetRequestCtx(context.Background(), reqCtx2)

	err = wrappedHandler(ctx2)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	responseID2 := string(reqCtx2.Response.Header.Peek("X-Request-ID"))
	if responseID2 != "existing-123" {
		t.Errorf("Expected existing request ID, got %s", responseID2)
	}
}

func TestCORSMiddleware(t *testing.T) {
	config := DefaultConfig()
	config.CORSAllowOrigins = []string{"http://example.com"}
	config.CORSAllowMethods = []string{"GET", "POST"}
	config.CORSAllowHeaders = []string{"Content-Type"}

	middleware := CORS(config)

	handler := func(ctx context.Context) error {
		return nil
	}

	wrappedHandler := middleware(handler)

	// Test with allowed origin
	reqCtx := &fasthttp.RequestCtx{}
	reqCtx.Request.Header.Set("Origin", "http://example.com")
	ctx := SetRequestCtx(context.Background(), reqCtx)

	err := wrappedHandler(ctx)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	allowOrigin := string(reqCtx.Response.Header.Peek("Access-Control-Allow-Origin"))
	if allowOrigin != "http://example.com" {
		t.Errorf("Expected CORS origin header, got %s", allowOrigin)
	}

	// Test OPTIONS preflight
	reqCtx2 := &fasthttp.RequestCtx{}
	reqCtx2.Request.Header.SetMethod("OPTIONS")
	ctx2 := SetRequestCtx(context.Background(), reqCtx2)

	err = wrappedHandler(ctx2)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if reqCtx2.Response.StatusCode() != fasthttp.StatusNoContent {
		t.Errorf("Expected status 204 for OPTIONS, got %d", reqCtx2.Response.StatusCode())
	}
}

func TestTimeoutMiddleware(t *testing.T) {
	middleware := Timeout(50 * time.Millisecond)

	// Handler that takes too long
	handler := func(ctx context.Context) error {
		time.Sleep(100 * time.Millisecond)
		return nil
	}

	wrappedHandler := middleware(handler)

	reqCtx := &fasthttp.RequestCtx{}
	ctx := SetRequestCtx(context.Background(), reqCtx)

	err := wrappedHandler(ctx)
	if err == nil {
		t.Error("Expected timeout error")
	}

	httpErr, ok := err.(*HTTPError)
	if !ok {
		t.Error("Expected HTTPError")
	}

	if httpErr.Code != 503 {
		t.Errorf("Expected status 503, got %d", httpErr.Code)
	}
}

func TestAuthMiddleware(t *testing.T) {
	authFunc := func(ctx context.Context) error {
		token := Header(ctx, "Authorization")
		if token != "Bearer valid-token" {
			return Unauthorized("Invalid token")
		}
		return nil
	}

	middleware := Auth(authFunc)

	handler := func(ctx context.Context) error {
		return nil
	}

	wrappedHandler := middleware(handler)

	// Test with valid token
	reqCtx := &fasthttp.RequestCtx{}
	reqCtx.Request.Header.Set("Authorization", "Bearer valid-token")
	ctx := SetRequestCtx(context.Background(), reqCtx)

	err := wrappedHandler(ctx)
	if err != nil {
		t.Errorf("Unexpected error with valid token: %v", err)
	}

	// Test with invalid token
	reqCtx2 := &fasthttp.RequestCtx{}
	reqCtx2.Request.Header.Set("Authorization", "Bearer invalid-token")
	ctx2 := SetRequestCtx(context.Background(), reqCtx2)

	err = wrappedHandler(ctx2)
	if err == nil {
		t.Error("Expected error with invalid token")
	}

	httpErr, ok := err.(*HTTPError)
	if !ok {
		t.Error("Expected HTTPError")
	}

	if httpErr.Code != 401 {
		t.Errorf("Expected status 401, got %d", httpErr.Code)
	}
}

func TestRateLimitMiddleware(t *testing.T) {
	middleware := RateLimit(2, time.Minute)

	handler := func(ctx context.Context) error {
		return nil
	}

	wrappedHandler := middleware(handler)

	reqCtx := &fasthttp.RequestCtx{}
	ctx := SetRequestCtx(context.Background(), reqCtx)

	// First request - should pass
	err := wrappedHandler(ctx)
	if err != nil {
		t.Errorf("First request failed: %v", err)
	}

	// Second request - should pass
	err = wrappedHandler(ctx)
	if err != nil {
		t.Errorf("Second request failed: %v", err)
	}

	// Third request - should be rate limited
	err = wrappedHandler(ctx)
	if err == nil {
		t.Error("Expected rate limit error")
	}

	httpErr, ok := err.(*HTTPError)
	if !ok {
		t.Error("Expected HTTPError")
	}

	if httpErr.Code != 429 {
		t.Errorf("Expected status 429, got %d", httpErr.Code)
	}
}

func TestMiddlewareChaining(t *testing.T) {
	called := []string{}

	middleware1 := func(next HandlerFunc) HandlerFunc {
		return func(ctx context.Context) error {
			called = append(called, "before1")
			err := next(ctx)
			called = append(called, "after1")
			return err
		}
	}

	middleware2 := func(next HandlerFunc) HandlerFunc {
		return func(ctx context.Context) error {
			called = append(called, "before2")
			err := next(ctx)
			called = append(called, "after2")
			return err
		}
	}

	handler := func(ctx context.Context) error {
		called = append(called, "handler")
		return nil
	}

	wrappedHandler := Use(handler, middleware1, middleware2)

	reqCtx := &fasthttp.RequestCtx{}
	ctx := SetRequestCtx(context.Background(), reqCtx)

	err := wrappedHandler(ctx)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := []string{"before1", "before2", "handler", "after2", "after1"}
	if len(called) != len(expected) {
		t.Errorf("Expected %d calls, got %d", len(expected), len(called))
	}

	for i, call := range expected {
		if i >= len(called) || called[i] != call {
			t.Errorf("At position %d: expected %s, got %s", i, call, called[i])
		}
	}
}

func TestMiddlewareErrorPropagation(t *testing.T) {
	testErr := errors.New("test error")

	middleware := func(next HandlerFunc) HandlerFunc {
		return func(ctx context.Context) error {
			return testErr
		}
	}

	handler := func(ctx context.Context) error {
		t.Error("Handler should not be called")
		return nil
	}

	wrappedHandler := middleware(handler)

	reqCtx := &fasthttp.RequestCtx{}
	ctx := SetRequestCtx(context.Background(), reqCtx)

	err := wrappedHandler(ctx)
	if err != testErr {
		t.Errorf("Expected error to propagate, got %v", err)
	}
}
