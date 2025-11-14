package httpservice

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"
)

// TestIntegrationBasicService tests a complete service lifecycle
func TestIntegrationBasicService(t *testing.T) {
	// Create service
	service, err := New(
		WithTitle("Test API"),
		WithVersion("1.0.0"),
		WithPort(getFreePort()),
	)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	// Register a simple handler
	service.GET("/hello", func(ctx context.Context) (interface{}, error) {
		return map[string]string{"message": "Hello, World!"}, nil
	})

	// Start service asynchronously
	err = service.StartAsync()
	if err != nil {
		t.Fatalf("Failed to start service: %v", err)
	}

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Make HTTP request
	resp, err := http.Get("http://" + service.config.Addr() + "/hello")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var result map[string]string
	json.Unmarshal(body, &result)

	if result["message"] != "Hello, World!" {
		t.Errorf("Expected message 'Hello, World!', got %s", result["message"])
	}

	// Shutdown service
	err = service.Shutdown()
	if err != nil {
		t.Errorf("Failed to shutdown service: %v", err)
	}
}

// TestIntegrationHealthCheck tests the built-in health check
func TestIntegrationHealthCheck(t *testing.T) {
	service, err := New(
		WithTitle("Test API"),
		WithVersion("2.0.0"),
		WithPort(getFreePort()),
		WithHealthCheck(true),
	)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	err = service.StartAsync()
	if err != nil {
		t.Fatalf("Failed to start service: %v", err)
	}
	defer service.Shutdown()

	time.Sleep(100 * time.Millisecond)

	resp, err := http.Get("http://" + service.config.Addr() + "/health")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var health HealthResponse
	json.Unmarshal(body, &health)

	if health.Status != "ok" {
		t.Errorf("Expected status 'ok', got %s", health.Status)
	}

	if health.Version != "2.0.0" {
		t.Errorf("Expected version '2.0.0', got %s", health.Version)
	}
}

// TestIntegrationOpenAPI tests OpenAPI spec generation
func TestIntegrationOpenAPI(t *testing.T) {
	service, err := New(
		WithTitle("Test API"),
		WithVersion("1.0.0"),
		WithPort(getFreePort()),
		WithDocs(true),
	)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	// Register routes with metadata
	service.GET("/users", func(ctx context.Context) (interface{}, error) {
		return []map[string]string{}, nil
	}, WithTags("users"), WithSummary("List users"))

	service.POST("/users", func(ctx context.Context, req *struct {
		Name string `json:"name"`
	}) (interface{}, error) {
		return map[string]string{"id": "123"}, nil
	}, WithTags("users"), WithSummary("Create user"))

	err = service.StartAsync()
	if err != nil {
		t.Fatalf("Failed to start service: %v", err)
	}
	defer service.Shutdown()

	time.Sleep(100 * time.Millisecond)

	// Get OpenAPI spec
	resp, err := http.Get("http://" + service.config.Addr() + "/openapi.json")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var spec OpenAPISpec
	err = json.Unmarshal(body, &spec)
	if err != nil {
		t.Fatalf("Failed to unmarshal OpenAPI spec: %v", err)
	}

	if spec.OpenAPI != "3.0.0" {
		t.Errorf("Expected OpenAPI 3.0.0, got %s", spec.OpenAPI)
	}

	if spec.Info.Title != "Test API" {
		t.Errorf("Expected title 'Test API', got %s", spec.Info.Title)
	}

	if len(spec.Paths) == 0 {
		t.Error("Expected paths to be generated")
	}
}

// TestIntegrationRequestValidation tests automatic request validation
func TestIntegrationRequestValidation(t *testing.T) {
	service, err := New(
		WithTitle("Test API"),
		WithPort(getFreePort()),
		WithValidation(true),
	)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	type CreateUserRequest struct {
		Name  string `json:"name" validate:"required,min=3"`
		Email string `json:"email" validate:"required,email"`
		Age   int    `json:"age" validate:"required,min=18"`
	}

	service.POST("/users", func(ctx context.Context, req *CreateUserRequest) (interface{}, error) {
		return map[string]string{"id": "123"}, nil
	})

	err = service.StartAsync()
	if err != nil {
		t.Fatalf("Failed to start service: %v", err)
	}
	defer service.Shutdown()

	time.Sleep(100 * time.Millisecond)

	// Test with invalid data
	invalidJSON := `{"name":"Jo","email":"invalid","age":15}`
	resp, err := http.Post(
		"http://"+service.config.Addr()+"/users",
		"application/json",
		strings.NewReader(invalidJSON),
	)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 422 {
		t.Errorf("Expected status 422 for validation error, got %d", resp.StatusCode)
	}
}

// TestIntegrationCORS tests CORS headers
func TestIntegrationCORS(t *testing.T) {
	service, err := New(
		WithTitle("Test API"),
		WithPort(getFreePort()),
		WithCORS(true),
		WithCORSOrigins("http://example.com"),
	)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	service.GET("/test", func(ctx context.Context) (interface{}, error) {
		return map[string]string{"test": "ok"}, nil
	})

	err = service.StartAsync()
	if err != nil {
		t.Fatalf("Failed to start service: %v", err)
	}
	defer service.Shutdown()

	time.Sleep(100 * time.Millisecond)

	// Make request with Origin header
	req, _ := http.NewRequest("GET", "http://"+service.config.Addr()+"/test", nil)
	req.Header.Set("Origin", "http://example.com")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Check CORS header
	allowOrigin := resp.Header.Get("Access-Control-Allow-Origin")
	if allowOrigin != "http://example.com" {
		t.Errorf("Expected CORS origin 'http://example.com', got %s", allowOrigin)
	}
}

// TestIntegrationPathParameters tests path parameter extraction
func TestIntegrationPathParameters(t *testing.T) {
	service, err := New(
		WithTitle("Test API"),
		WithPort(getFreePort()),
	)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	service.GET("/users/{id}", func(ctx context.Context) (interface{}, error) {
		id := PathParam(ctx, "id")
		return map[string]string{"id": id}, nil
	})

	err = service.StartAsync()
	if err != nil {
		t.Fatalf("Failed to start service: %v", err)
	}
	defer service.Shutdown()

	time.Sleep(100 * time.Millisecond)

	resp, err := http.Get("http://" + service.config.Addr() + "/users/123")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result map[string]string
	json.Unmarshal(body, &result)

	if result["id"] != "123" {
		t.Errorf("Expected id '123', got %s", result["id"])
	}
}

// TestIntegrationNotFound tests 404 handling
func TestIntegrationNotFound(t *testing.T) {
	service, err := New(
		WithTitle("Test API"),
		WithPort(getFreePort()),
	)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	service.GET("/exists", func(ctx context.Context) (interface{}, error) {
		return map[string]string{"status": "ok"}, nil
	})

	err = service.StartAsync()
	if err != nil {
		t.Fatalf("Failed to start service: %v", err)
	}
	defer service.Shutdown()

	time.Sleep(100 * time.Millisecond)

	resp, err := http.Get("http://" + service.config.Addr() + "/does-not-exist")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 404 {
		t.Errorf("Expected status 404, got %d", resp.StatusCode)
	}
}

// getFreePort returns an available port on the localhost
func getFreePort() int {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 9999 // Fallback port
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port
}
