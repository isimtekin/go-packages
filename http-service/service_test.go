package httpservice

import (
	"context"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	service, err := New(
		WithTitle("Test Service"),
		WithVersion("1.0.0"),
		WithPort(9090),
	)

	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	if service == nil {
		t.Fatal("Service should not be nil")
	}

	if service.config.Title != "Test Service" {
		t.Errorf("Expected title 'Test Service', got %s", service.config.Title)
	}

	if service.config.Port != 9090 {
		t.Errorf("Expected port 9090, got %d", service.config.Port)
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.Title == "" {
		t.Error("Default title should not be empty")
	}

	if config.Port != 8080 {
		t.Errorf("Expected default port 8080, got %d", config.Port)
	}

	if config.ReadTimeout != 30*time.Second {
		t.Errorf("Expected read timeout 30s, got %v", config.ReadTimeout)
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name:    "valid config",
			config:  DefaultConfig(),
			wantErr: false,
		},
		{
			name: "empty title",
			config: &Config{
				Title:              "",
				Version:            "1.0.0",
				Host:               "localhost",
				Port:               8080,
				ReadTimeout:        30 * time.Second,
				WriteTimeout:       30 * time.Second,
				MaxRequestBodySize: 1024,
			},
			wantErr: true,
		},
		{
			name: "invalid port",
			config: &Config{
				Title:              "Test",
				Version:            "1.0.0",
				Host:               "localhost",
				Port:               0,
				ReadTimeout:        30 * time.Second,
				WriteTimeout:       30 * time.Second,
				MaxRequestBodySize: 1024,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRouteRegistration(t *testing.T) {
	service, err := New()
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	handler := func(ctx context.Context) (interface{}, error) {
		return map[string]string{"message": "Hello"}, nil
	}

	service.GET("/test", handler)

	if len(service.routes) == 0 {
		t.Error("Expected routes to be registered")
	}

	// Find the test route (skip built-in routes)
	var testRoute *Route
	for _, route := range service.routes {
		if route.Path == "/test" {
			testRoute = route
			break
		}
	}

	if testRoute == nil {
		t.Fatal("Test route not found")
	}

	if testRoute.Method != "GET" {
		t.Errorf("Expected method GET, got %s", testRoute.Method)
	}

	if testRoute.Path != "/test" {
		t.Errorf("Expected path /test, got %s", testRoute.Path)
	}
}

func TestHTTPErrors(t *testing.T) {
	tests := []struct {
		name     string
		err      *HTTPError
		wantCode int
		wantMsg  string
	}{
		{
			name:     "bad request",
			err:      BadRequest("Invalid input"),
			wantCode: 400,
			wantMsg:  "Invalid input",
		},
		{
			name:     "not found",
			err:      NotFound("Resource not found"),
			wantCode: 404,
			wantMsg:  "Resource not found",
		},
		{
			name:     "internal server error",
			err:      InternalServerError("Server error"),
			wantCode: 500,
			wantMsg:  "Server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Code != tt.wantCode {
				t.Errorf("Expected code %d, got %d", tt.wantCode, tt.err.Code)
			}

			if tt.err.Message != tt.wantMsg {
				t.Errorf("Expected message %s, got %s", tt.wantMsg, tt.err.Message)
			}
		})
	}
}

func TestValidator(t *testing.T) {
	validator := NewValidator()

	type TestStruct struct {
		Name  string `json:"name" validate:"required,min=3,max=50"`
		Email string `json:"email" validate:"required,email"`
		Age   int    `json:"age" validate:"required,min=18,max=120"`
	}

	tests := []struct {
		name    string
		input   TestStruct
		wantErr bool
	}{
		{
			name: "valid input",
			input: TestStruct{
				Name:  "John Doe",
				Email: "john@example.com",
				Age:   25,
			},
			wantErr: false,
		},
		{
			name: "invalid email",
			input: TestStruct{
				Name:  "John Doe",
				Email: "invalid-email",
				Age:   25,
			},
			wantErr: true,
		},
		{
			name: "age too young",
			input: TestStruct{
				Name:  "John Doe",
				Email: "john@example.com",
				Age:   15,
			},
			wantErr: true,
		},
		{
			name: "name too short",
			input: TestStruct{
				Name:  "JD",
				Email: "john@example.com",
				Age:   25,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validator.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPathParams(t *testing.T) {
	tests := []struct {
		name        string
		pattern     string
		path        string
		wantParams  map[string]string
		wantErr     bool
	}{
		{
			name:    "single parameter",
			pattern: "/users/{id}",
			path:    "/users/123",
			wantParams: map[string]string{
				"id": "123",
			},
			wantErr: false,
		},
		{
			name:    "multiple parameters",
			pattern: "/users/{id}/posts/{postId}",
			path:    "/users/123/posts/456",
			wantParams: map[string]string{
				"id":     "123",
				"postId": "456",
			},
			wantErr: false,
		},
		{
			name:       "no match",
			pattern:    "/users/{id}",
			path:       "/posts/123",
			wantParams: nil,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params, err := ParsePathParams(tt.pattern, tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParsePathParams() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				for k, v := range tt.wantParams {
					if params[k] != v {
						t.Errorf("Expected param %s=%s, got %s", k, v, params[k])
					}
				}
			}
		})
	}
}

func TestOpenAPIGeneration(t *testing.T) {
	config := DefaultConfig()
	config.Title = "Test API"
	config.Version = "1.0.0"

	routes := []*Route{
		{
			Method:  "GET",
			Path:    "/users",
			Tags:    []string{"users"},
			Summary: "List users",
		},
		{
			Method:  "POST",
			Path:    "/users",
			Tags:    []string{"users"},
			Summary: "Create user",
		},
	}

	spec := GenerateOpenAPISpec(config, routes)

	if spec.OpenAPI != "3.0.0" {
		t.Errorf("Expected OpenAPI version 3.0.0, got %s", spec.OpenAPI)
	}

	if spec.Info.Title != "Test API" {
		t.Errorf("Expected title 'Test API', got %s", spec.Info.Title)
	}

	if spec.Info.Version != "1.0.0" {
		t.Errorf("Expected version 1.0.0, got %s", spec.Info.Version)
	}

	if len(spec.Paths) == 0 {
		t.Error("Expected paths to be generated")
	}
}
