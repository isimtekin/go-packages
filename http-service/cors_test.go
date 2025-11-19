package httpservice

import (
	"context"
	"testing"
	"time"

	"github.com/valyala/fasthttp"
)

func TestCORS_Enabled(t *testing.T) {
	// Create service with CORS enabled
	service, err := New(
		WithTitle("Test Service"),
		WithCORS(true),
		WithPort(9001),
	)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	// Add a test route
	service.GET("/test", func(ctx context.Context) (interface{}, error) {
		return map[string]string{"message": "hello"}, nil
	})

	// Start service async
	err = service.StartAsync()
	if err != nil {
		t.Fatalf("Failed to start service: %v", err)
	}
	defer service.Shutdown()

	// Wait for service to start
	time.Sleep(100 * time.Millisecond)

	// Test preflight request
	t.Run("Preflight Request", func(t *testing.T) {
		req := fasthttp.AcquireRequest()
		resp := fasthttp.AcquireResponse()
		defer fasthttp.ReleaseRequest(req)
		defer fasthttp.ReleaseResponse(resp)

		req.SetRequestURI("http://localhost:9001/test")
		req.Header.SetMethod("OPTIONS")
		req.Header.Set("Origin", "http://example.com")
		req.Header.Set("Access-Control-Request-Method", "GET")

		client := &fasthttp.Client{}
		if err := client.Do(req, resp); err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}

		// Check CORS headers
		if origin := string(resp.Header.Peek("Access-Control-Allow-Origin")); origin == "" {
			t.Error("Access-Control-Allow-Origin header not set")
		}

		if methods := string(resp.Header.Peek("Access-Control-Allow-Methods")); methods == "" {
			t.Error("Access-Control-Allow-Methods header not set")
		}

		if headers := string(resp.Header.Peek("Access-Control-Allow-Headers")); headers == "" {
			t.Error("Access-Control-Allow-Headers header not set")
		}

		// Check status code
		if resp.StatusCode() != fasthttp.StatusNoContent {
			t.Errorf("Expected status 204, got %d", resp.StatusCode())
		}
	})

	// Test normal request with Origin header
	t.Run("Normal Request with Origin", func(t *testing.T) {
		req := fasthttp.AcquireRequest()
		resp := fasthttp.AcquireResponse()
		defer fasthttp.ReleaseRequest(req)
		defer fasthttp.ReleaseResponse(resp)

		req.SetRequestURI("http://localhost:9001/test")
		req.Header.SetMethod("GET")
		req.Header.Set("Origin", "http://example.com")

		client := &fasthttp.Client{}
		if err := client.Do(req, resp); err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}

		// Check CORS headers
		if origin := string(resp.Header.Peek("Access-Control-Allow-Origin")); origin == "" {
			t.Error("Access-Control-Allow-Origin header not set")
		}
	})
}

func TestCORS_Disabled(t *testing.T) {
	// Create service with CORS disabled
	service, err := New(
		WithTitle("Test Service"),
		WithCORS(false),
		WithPort(9002),
	)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	// Add a test route
	service.GET("/test", func(ctx context.Context) (interface{}, error) {
		return map[string]string{"message": "hello"}, nil
	})

	// Start service async
	err = service.StartAsync()
	if err != nil {
		t.Fatalf("Failed to start service: %v", err)
	}
	defer service.Shutdown()

	// Wait for service to start
	time.Sleep(100 * time.Millisecond)

	// Test preflight request - should NOT have CORS headers
	t.Run("Preflight Request - No CORS", func(t *testing.T) {
		req := fasthttp.AcquireRequest()
		resp := fasthttp.AcquireResponse()
		defer fasthttp.ReleaseRequest(req)
		defer fasthttp.ReleaseResponse(resp)

		req.SetRequestURI("http://localhost:9002/test")
		req.Header.SetMethod("OPTIONS")
		req.Header.Set("Origin", "http://example.com")
		req.Header.Set("Access-Control-Request-Method", "GET")

		client := &fasthttp.Client{}
		if err := client.Do(req, resp); err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}

		// CORS headers should NOT be present
		if origin := string(resp.Header.Peek("Access-Control-Allow-Origin")); origin != "" {
			t.Errorf("Access-Control-Allow-Origin header should not be set when CORS is disabled, got: %s", origin)
		}
	})

	// Test normal request - should NOT have CORS headers
	t.Run("Normal Request - No CORS", func(t *testing.T) {
		req := fasthttp.AcquireRequest()
		resp := fasthttp.AcquireResponse()
		defer fasthttp.ReleaseRequest(req)
		defer fasthttp.ReleaseResponse(resp)

		req.SetRequestURI("http://localhost:9002/test")
		req.Header.SetMethod("GET")
		req.Header.Set("Origin", "http://example.com")

		client := &fasthttp.Client{}
		if err := client.Do(req, resp); err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}

		// CORS headers should NOT be present
		if origin := string(resp.Header.Peek("Access-Control-Allow-Origin")); origin != "" {
			t.Errorf("Access-Control-Allow-Origin header should not be set when CORS is disabled, got: %s", origin)
		}
	})
}

func TestCORS_CustomOrigins(t *testing.T) {
	// Create service with specific allowed origins
	service, err := New(
		WithTitle("Test Service"),
		WithCORS(true),
		WithCORSOrigins("http://localhost:3000", "https://example.com"),
		WithPort(9003),
	)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	// Add a test route
	service.GET("/test", func(ctx context.Context) (interface{}, error) {
		return map[string]string{"message": "hello"}, nil
	})

	// Start service async
	err = service.StartAsync()
	if err != nil {
		t.Fatalf("Failed to start service: %v", err)
	}
	defer service.Shutdown()

	// Wait for service to start
	time.Sleep(100 * time.Millisecond)

	// Test allowed origin
	t.Run("Allowed Origin", func(t *testing.T) {
		req := fasthttp.AcquireRequest()
		resp := fasthttp.AcquireResponse()
		defer fasthttp.ReleaseRequest(req)
		defer fasthttp.ReleaseResponse(resp)

		req.SetRequestURI("http://localhost:9003/test")
		req.Header.SetMethod("GET")
		req.Header.Set("Origin", "http://localhost:3000")

		client := &fasthttp.Client{}
		if err := client.Do(req, resp); err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}

		// Check CORS header matches the origin
		origin := string(resp.Header.Peek("Access-Control-Allow-Origin"))
		if origin != "http://localhost:3000" {
			t.Errorf("Expected Access-Control-Allow-Origin to be 'http://localhost:3000', got '%s'", origin)
		}
	})

	// Test disallowed origin
	t.Run("Disallowed Origin", func(t *testing.T) {
		req := fasthttp.AcquireRequest()
		resp := fasthttp.AcquireResponse()
		defer fasthttp.ReleaseRequest(req)
		defer fasthttp.ReleaseResponse(resp)

		req.SetRequestURI("http://localhost:9003/test")
		req.Header.SetMethod("GET")
		req.Header.Set("Origin", "http://evil.com")

		client := &fasthttp.Client{}
		if err := client.Do(req, resp); err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}

		// CORS header should NOT be set for disallowed origin
		origin := string(resp.Header.Peek("Access-Control-Allow-Origin"))
		if origin != "" {
			t.Errorf("Access-Control-Allow-Origin should not be set for disallowed origin, got '%s'", origin)
		}
	})
}
