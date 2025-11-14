package httpservice

import (
	"context"
	"testing"

	"github.com/valyala/fasthttp"
)

func TestSetAndGetRequestCtx(t *testing.T) {
	reqCtx := &fasthttp.RequestCtx{}
	ctx := context.Background()

	// Set request context
	ctx = SetRequestCtx(ctx, reqCtx)

	// Get request context
	retrieved := GetRequestCtx(ctx)
	if retrieved == nil {
		t.Error("Expected request context to be non-nil")
	}

	if retrieved != reqCtx {
		t.Error("Retrieved context should be the same as set context")
	}
}

func TestGetRequestCtxNil(t *testing.T) {
	ctx := context.Background()

	retrieved := GetRequestCtx(ctx)
	if retrieved != nil {
		t.Error("Expected nil when request context not set")
	}
}

func TestSetAndGetPathParams(t *testing.T) {
	ctx := context.Background()

	params := map[string]string{
		"id":     "123",
		"userId": "456",
	}

	ctx = SetPathParams(ctx, params)

	retrieved := GetPathParams(ctx)
	if len(retrieved) != len(params) {
		t.Errorf("Expected %d params, got %d", len(params), len(retrieved))
	}

	if retrieved["id"] != "123" {
		t.Errorf("Expected id=123, got %s", retrieved["id"])
	}

	if retrieved["userId"] != "456" {
		t.Errorf("Expected userId=456, got %s", retrieved["userId"])
	}
}

func TestGetPathParamsEmpty(t *testing.T) {
	ctx := context.Background()

	params := GetPathParams(ctx)
	if params == nil {
		t.Error("Expected empty map, got nil")
	}

	if len(params) != 0 {
		t.Errorf("Expected empty map, got %d elements", len(params))
	}
}

func TestSetAndGetRequestID(t *testing.T) {
	ctx := context.Background()

	requestID := "req-12345"
	ctx = SetRequestID(ctx, requestID)

	retrieved := GetRequestID(ctx)
	if retrieved != requestID {
		t.Errorf("Expected request ID %s, got %s", requestID, retrieved)
	}
}

func TestGetRequestIDEmpty(t *testing.T) {
	ctx := context.Background()

	requestID := GetRequestID(ctx)
	if requestID != "" {
		t.Errorf("Expected empty request ID, got %s", requestID)
	}
}

func TestPathParam(t *testing.T) {
	ctx := context.Background()

	params := map[string]string{
		"id":       "123",
		"postId":   "456",
		"category": "tech",
	}

	ctx = SetPathParams(ctx, params)

	tests := []struct {
		name     string
		param    string
		expected string
	}{
		{"id parameter", "id", "123"},
		{"postId parameter", "postId", "456"},
		{"category parameter", "category", "tech"},
		{"non-existent parameter", "missing", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value := PathParam(ctx, tt.param)
			if value != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, value)
			}
		})
	}
}

func TestQueryParam(t *testing.T) {
	reqCtx := &fasthttp.RequestCtx{}
	reqCtx.Request.SetRequestURI("/test?page=1&limit=20&search=hello")

	ctx := SetRequestCtx(context.Background(), reqCtx)

	tests := []struct {
		name     string
		param    string
		expected string
	}{
		{"page parameter", "page", "1"},
		{"limit parameter", "limit", "20"},
		{"search parameter", "search", "hello"},
		{"non-existent parameter", "missing", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value := QueryParam(ctx, tt.param)
			if value != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, value)
			}
		})
	}
}

func TestQueryParamWithDefault(t *testing.T) {
	reqCtx := &fasthttp.RequestCtx{}
	reqCtx.Request.SetRequestURI("/test?page=1")

	ctx := SetRequestCtx(context.Background(), reqCtx)

	// Existing parameter
	value := QueryParamWithDefault(ctx, "page", "10")
	if value != "1" {
		t.Errorf("Expected 1, got %s", value)
	}

	// Non-existent parameter (should return default)
	value = QueryParamWithDefault(ctx, "limit", "20")
	if value != "20" {
		t.Errorf("Expected default 20, got %s", value)
	}
}

func TestQueryParamInt(t *testing.T) {
	reqCtx := &fasthttp.RequestCtx{}
	reqCtx.Request.SetRequestURI("/test?page=5&limit=100&invalid=abc")

	ctx := SetRequestCtx(context.Background(), reqCtx)

	tests := []struct {
		name         string
		param        string
		defaultValue int
		expected     int
	}{
		{"valid int", "page", 1, 5},
		{"valid int 2", "limit", 20, 100},
		{"missing parameter", "offset", 0, 0},
		{"invalid int", "invalid", 10, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value := QueryParamInt(ctx, tt.param, tt.defaultValue)
			if value != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, value)
			}
		})
	}
}

func TestQueryParamBool(t *testing.T) {
	reqCtx := &fasthttp.RequestCtx{}
	reqCtx.Request.SetRequestURI("/test?active=true&enabled=false&invalid=xyz")

	ctx := SetRequestCtx(context.Background(), reqCtx)

	tests := []struct {
		name         string
		param        string
		defaultValue bool
		expected     bool
	}{
		{"true value", "active", false, true},
		{"false value", "enabled", true, false},
		{"missing parameter", "missing", true, true},
		{"invalid value", "invalid", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value := QueryParamBool(ctx, tt.param, tt.defaultValue)
			if value != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, value)
			}
		})
	}
}

func TestHeader(t *testing.T) {
	reqCtx := &fasthttp.RequestCtx{}
	reqCtx.Request.Header.Set("Authorization", "Bearer token123")
	reqCtx.Request.Header.Set("Content-Type", "application/json")
	reqCtx.Request.Header.Set("User-Agent", "TestAgent/1.0")

	ctx := SetRequestCtx(context.Background(), reqCtx)

	tests := []struct {
		name     string
		header   string
		expected string
	}{
		{"Authorization header", "Authorization", "Bearer token123"},
		{"Content-Type header", "Content-Type", "application/json"},
		{"User-Agent header", "User-Agent", "TestAgent/1.0"},
		{"Non-existent header", "X-Custom", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value := Header(ctx, tt.header)
			if value != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, value)
			}
		})
	}
}

func TestSetHeader(t *testing.T) {
	reqCtx := &fasthttp.RequestCtx{}
	ctx := SetRequestCtx(context.Background(), reqCtx)

	SetHeader(ctx, "X-Custom-Header", "custom-value")

	value := string(reqCtx.Response.Header.Peek("X-Custom-Header"))
	if value != "custom-value" {
		t.Errorf("Expected custom-value, got %s", value)
	}
}

func TestSetStatus(t *testing.T) {
	reqCtx := &fasthttp.RequestCtx{}
	ctx := SetRequestCtx(context.Background(), reqCtx)

	SetStatus(ctx, 201)

	if reqCtx.Response.StatusCode() != 201 {
		t.Errorf("Expected status 201, got %d", reqCtx.Response.StatusCode())
	}
}

func TestMethod(t *testing.T) {
	tests := []struct {
		name     string
		method   string
		expected string
	}{
		{"GET method", "GET", "GET"},
		{"POST method", "POST", "POST"},
		{"PUT method", "PUT", "PUT"},
		{"DELETE method", "DELETE", "DELETE"},
		{"PATCH method", "PATCH", "PATCH"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqCtx := &fasthttp.RequestCtx{}
			reqCtx.Request.Header.SetMethod(tt.method)

			ctx := SetRequestCtx(context.Background(), reqCtx)

			method := Method(ctx)
			if method != tt.expected {
				t.Errorf("Expected method %s, got %s", tt.expected, method)
			}
		})
	}
}

func TestPath(t *testing.T) {
	reqCtx := &fasthttp.RequestCtx{}
	reqCtx.Request.SetRequestURI("/api/v1/users/123?page=1")

	ctx := SetRequestCtx(context.Background(), reqCtx)

	path := Path(ctx)
	if path != "/api/v1/users/123" {
		t.Errorf("Expected path /api/v1/users/123, got %s", path)
	}
}

func TestRemoteAddr(t *testing.T) {
	reqCtx := &fasthttp.RequestCtx{}
	ctx := SetRequestCtx(context.Background(), reqCtx)

	addr := RemoteAddr(ctx)
	// Remote addr will be empty in test context, just verify it doesn't panic
	_ = addr
}

func TestContextHelpersWithNilRequestCtx(t *testing.T) {
	ctx := context.Background()

	// These should not panic when request context is nil
	if QueryParam(ctx, "test") != "" {
		t.Error("Expected empty string for nil context")
	}

	if Header(ctx, "test") != "" {
		t.Error("Expected empty string for nil context")
	}

	if Method(ctx) != "" {
		t.Error("Expected empty string for nil context")
	}

	if Path(ctx) != "" {
		t.Error("Expected empty string for nil context")
	}

	if RemoteAddr(ctx) != "" {
		t.Error("Expected empty string for nil context")
	}

	// These should not panic
	SetHeader(ctx, "test", "value")
	SetStatus(ctx, 200)
}
