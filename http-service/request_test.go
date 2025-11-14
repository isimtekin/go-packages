package httpservice

import (
	"context"
	"testing"

	"github.com/valyala/fasthttp"
)

func TestBindJSON(t *testing.T) {
	type TestRequest struct {
		Name  string `json:"name"`
		Email string `json:"email"`
		Age   int    `json:"age"`
	}

	tests := []struct {
		name        string
		body        string
		contentType string
		wantErr     bool
	}{
		{
			name:        "valid json",
			body:        `{"name":"John","email":"john@example.com","age":30}`,
			contentType: "application/json",
			wantErr:     false,
		},
		{
			name:        "empty body",
			body:        "",
			contentType: "application/json",
			wantErr:     true,
		},
		{
			name:        "invalid json",
			body:        `{"name":"John",invalid}`,
			contentType: "application/json",
			wantErr:     true,
		},
		{
			name:        "wrong content type",
			body:        `{"name":"John"}`,
			contentType: "text/plain",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqCtx := &fasthttp.RequestCtx{}
			reqCtx.Request.Header.SetContentType(tt.contentType)
			reqCtx.Request.SetBodyString(tt.body)

			ctx := SetRequestCtx(context.Background(), reqCtx)

			var req TestRequest
			err := BindJSON(ctx, &req)

			if (err != nil) != tt.wantErr {
				t.Errorf("BindJSON() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				if req.Name != "John" {
					t.Errorf("Expected name 'John', got %s", req.Name)
				}
			}
		})
	}
}

func TestBindAndValidate(t *testing.T) {
	type ValidatedRequest struct {
		Name  string `json:"name" validate:"required,min=3"`
		Email string `json:"email" validate:"required,email"`
		Age   int    `json:"age" validate:"required,min=18"`
	}

	validator := NewValidator()

	tests := []struct {
		name    string
		body    string
		wantErr bool
	}{
		{
			name:    "valid request",
			body:    `{"name":"John Doe","email":"john@example.com","age":25}`,
			wantErr: false,
		},
		{
			name:    "invalid email",
			body:    `{"name":"John Doe","email":"invalid","age":25}`,
			wantErr: true,
		},
		{
			name:    "age too young",
			body:    `{"name":"John Doe","email":"john@example.com","age":15}`,
			wantErr: true,
		},
		{
			name:    "name too short",
			body:    `{"name":"Jo","email":"john@example.com","age":25}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqCtx := &fasthttp.RequestCtx{}
			reqCtx.Request.Header.SetContentType("application/json")
			reqCtx.Request.SetBodyString(tt.body)

			ctx := SetRequestCtx(context.Background(), reqCtx)

			var req ValidatedRequest
			err := BindAndValidate(ctx, &req, validator)

			if (err != nil) != tt.wantErr {
				t.Errorf("BindAndValidate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParsePathParamsSuccess(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		path     string
		expected map[string]string
	}{
		{
			name:    "single parameter",
			pattern: "/users/{id}",
			path:    "/users/123",
			expected: map[string]string{
				"id": "123",
			},
		},
		{
			name:    "multiple parameters",
			pattern: "/users/{userId}/posts/{postId}",
			path:    "/users/123/posts/456",
			expected: map[string]string{
				"userId": "123",
				"postId": "456",
			},
		},
		{
			name:    "parameter with dashes",
			pattern: "/api/{version}/users/{user-id}",
			path:    "/api/v1/users/abc-123",
			expected: map[string]string{
				"version": "v1",
				"user-id": "abc-123",
			},
		},
		{
			name:     "no parameters",
			pattern:  "/users",
			path:     "/users",
			expected: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params, err := ParsePathParams(tt.pattern, tt.path)
			if err != nil {
				t.Errorf("ParsePathParams() error = %v", err)
				return
			}

			if len(params) != len(tt.expected) {
				t.Errorf("Expected %d params, got %d", len(tt.expected), len(params))
			}

			for key, expectedValue := range tt.expected {
				if params[key] != expectedValue {
					t.Errorf("Expected param %s=%s, got %s", key, expectedValue, params[key])
				}
			}
		})
	}
}

func TestParsePathParamsFailure(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		path    string
	}{
		{
			name:    "path too short",
			pattern: "/users/{id}/posts",
			path:    "/users/123",
		},
		{
			name:    "path too long",
			pattern: "/users/{id}",
			path:    "/users/123/extra",
		},
		{
			name:    "different path",
			pattern: "/users/{id}",
			path:    "/posts/123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParsePathParams(tt.pattern, tt.path)
			if err == nil {
				t.Error("Expected error for non-matching path")
			}
		})
	}
}

func TestSplitPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected []string
	}{
		{
			name:     "simple path",
			path:     "/users/123",
			expected: []string{"users", "123"},
		},
		{
			name:     "trailing slash",
			path:     "/users/123/",
			expected: []string{"users", "123"},
		},
		{
			name:     "no leading slash",
			path:     "users/123",
			expected: []string{"users", "123"},
		},
		{
			name:     "single segment",
			path:     "/users",
			expected: []string{"users"},
		},
		{
			name:     "empty path",
			path:     "",
			expected: []string{},
		},
		{
			name:     "root path",
			path:     "/",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parts := splitPath(tt.path)

			if len(parts) != len(tt.expected) {
				t.Errorf("Expected %d parts, got %d", len(tt.expected), len(parts))
			}

			for i, expected := range tt.expected {
				if i >= len(parts) || parts[i] != expected {
					t.Errorf("At position %d: expected %s, got %s", i, expected, parts[i])
				}
			}
		})
	}
}

func TestGetRequestInfo(t *testing.T) {
	reqCtx := &fasthttp.RequestCtx{}
	reqCtx.Request.SetRequestURI("/api/users?page=1&limit=20")
	reqCtx.Request.Header.SetMethod("GET")
	reqCtx.Request.Header.Set("Authorization", "Bearer token123")
	reqCtx.Request.Header.Set("User-Agent", "TestAgent/1.0")
	reqCtx.Request.Header.Set("Content-Type", "application/json")

	ctx := SetRequestCtx(context.Background(), reqCtx)
	ctx = SetRequestID(ctx, "req-123")

	info := GetRequestInfo(ctx)

	if info == nil {
		t.Fatal("Expected request info to be non-nil")
	}

	if info.Method != "GET" {
		t.Errorf("Expected method GET, got %s", info.Method)
	}

	if info.Path != "/api/users" {
		t.Errorf("Expected path /api/users, got %s", info.Path)
	}

	if info.Query == "" {
		t.Error("Expected query string to be non-empty")
	}

	if info.RequestID != "req-123" {
		t.Errorf("Expected request ID req-123, got %s", info.RequestID)
	}

	if info.Headers["Authorization"] != "Bearer token123" {
		t.Errorf("Expected Authorization header, got %s", info.Headers["Authorization"])
	}

	if info.Headers["User-Agent"] != "TestAgent/1.0" {
		t.Errorf("Expected User-Agent header, got %s", info.Headers["User-Agent"])
	}

	if info.Headers["Content-Type"] != "application/json" {
		t.Errorf("Expected Content-Type header, got %s", info.Headers["Content-Type"])
	}
}

func TestGetRequestInfoNilContext(t *testing.T) {
	ctx := context.Background()
	info := GetRequestInfo(ctx)

	if info != nil {
		t.Error("Expected nil when request context is not set")
	}
}

func TestBindJSONNilContext(t *testing.T) {
	ctx := context.Background()
	var data map[string]interface{}

	err := BindJSON(ctx, &data)
	if err == nil {
		t.Error("Expected error when request context is nil")
	}
}
