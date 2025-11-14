package httpservice

import (
	"encoding/json"
	"testing"

	"github.com/valyala/fasthttp"
)

func TestNewResponse(t *testing.T) {
	body := map[string]string{"message": "hello"}
	response := NewResponse(200, body)

	if response.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", response.StatusCode)
	}

	if response.Body == nil {
		t.Error("Expected body to be non-nil")
	}

	if response.Headers == nil {
		t.Error("Expected headers map to be non-nil")
	}
}

func TestResponseWithHeader(t *testing.T) {
	response := NewResponse(200, nil)
	response.WithHeader("X-Custom-Header", "test-value")

	if response.Headers["X-Custom-Header"] != "test-value" {
		t.Errorf("Expected header value 'test-value', got %s", response.Headers["X-Custom-Header"])
	}

	// Test chaining
	response.WithHeader("X-Another-Header", "another-value")

	if len(response.Headers) != 2 {
		t.Errorf("Expected 2 headers, got %d", len(response.Headers))
	}
}

func TestWriteResponse(t *testing.T) {
	reqCtx := &fasthttp.RequestCtx{}

	body := map[string]string{
		"message": "hello",
		"status":  "ok",
	}

	response := NewResponse(200, body)
	response.WithHeader("X-Custom", "value")

	err := WriteResponse(reqCtx, response)
	if err != nil {
		t.Errorf("WriteResponse() error = %v", err)
	}

	// Check status code
	if reqCtx.Response.StatusCode() != 200 {
		t.Errorf("Expected status 200, got %d", reqCtx.Response.StatusCode())
	}

	// Check custom header
	customHeader := string(reqCtx.Response.Header.Peek("X-Custom"))
	if customHeader != "value" {
		t.Errorf("Expected custom header 'value', got %s", customHeader)
	}

	// Check content type
	contentType := string(reqCtx.Response.Header.ContentType())
	if contentType != "application/json" {
		t.Errorf("Expected content type 'application/json', got %s", contentType)
	}

	// Check body
	var responseBody map[string]string
	err = json.Unmarshal(reqCtx.Response.Body(), &responseBody)
	if err != nil {
		t.Errorf("Failed to unmarshal response body: %v", err)
	}

	if responseBody["message"] != "hello" {
		t.Errorf("Expected message 'hello', got %s", responseBody["message"])
	}
}

func TestResponseHelpers(t *testing.T) {
	tests := []struct {
		name           string
		response       *Response
		expectedStatus int
	}{
		{
			name:           "OK",
			response:       OK(map[string]string{"test": "data"}),
			expectedStatus: 200,
		},
		{
			name:           "Created",
			response:       Created(map[string]string{"id": "123"}),
			expectedStatus: 201,
		},
		{
			name:           "Accepted",
			response:       Accepted(map[string]string{"job": "queued"}),
			expectedStatus: 202,
		},
		{
			name:           "NoContent",
			response:       NoContent(),
			expectedStatus: 204,
		},
		{
			name:           "JSON custom status",
			response:       JSON(418, map[string]string{"error": "I'm a teapot"}),
			expectedStatus: 418,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.response.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, tt.response.StatusCode)
			}
		})
	}
}

func TestNoContentResponse(t *testing.T) {
	response := NoContent()

	if response.StatusCode != 204 {
		t.Errorf("Expected status 204, got %d", response.StatusCode)
	}

	if response.Body != nil {
		t.Error("Expected body to be nil for NoContent")
	}
}

func TestWriteError(t *testing.T) {
	reqCtx := &fasthttp.RequestCtx{}

	httpErr := BadRequest("Invalid input")
	httpErr.WithDetails(map[string]interface{}{
		"field": "email",
		"error": "invalid format",
	})

	WriteError(reqCtx, httpErr)

	// Check status code
	if reqCtx.Response.StatusCode() != 400 {
		t.Errorf("Expected status 400, got %d", reqCtx.Response.StatusCode())
	}

	// Check body
	var errorResponse HTTPError
	err := json.Unmarshal(reqCtx.Response.Body(), &errorResponse)
	if err != nil {
		t.Errorf("Failed to unmarshal error response: %v", err)
	}

	if errorResponse.Message != "Invalid input" {
		t.Errorf("Expected message 'Invalid input', got %s", errorResponse.Message)
	}

	if errorResponse.Details == nil {
		t.Error("Expected details to be non-nil")
	}
}

func TestWriteErrorUnknown(t *testing.T) {
	reqCtx := &fasthttp.RequestCtx{}

	// Non-HTTP error
	unknownErr := ErrInvalidHandler

	WriteError(reqCtx, unknownErr)

	// Should default to 500
	if reqCtx.Response.StatusCode() != 500 {
		t.Errorf("Expected status 500 for unknown error, got %d", reqCtx.Response.StatusCode())
	}
}

func TestNewHealthResponse(t *testing.T) {
	health := NewHealthResponse("ok", "1.0.0")

	if health.Status != "ok" {
		t.Errorf("Expected status 'ok', got %s", health.Status)
	}

	if health.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got %s", health.Version)
	}
}

func TestWriteResponseWithNilBody(t *testing.T) {
	reqCtx := &fasthttp.RequestCtx{}

	response := NewResponse(204, nil)

	err := WriteResponse(reqCtx, response)
	if err != nil {
		t.Errorf("WriteResponse() error = %v", err)
	}

	if len(reqCtx.Response.Body()) != 0 {
		t.Error("Expected empty body for nil response body")
	}
}

func TestWriteResponseMarshalError(t *testing.T) {
	reqCtx := &fasthttp.RequestCtx{}

	// Create an unmarshalable body (channel cannot be marshaled to JSON)
	response := NewResponse(200, make(chan int))

	err := WriteResponse(reqCtx, response)
	if err == nil {
		t.Error("Expected error when marshaling invalid body")
	}
}

func TestResponseChaining(t *testing.T) {
	response := NewResponse(200, map[string]string{"test": "data"}).
		WithHeader("X-Header-1", "value1").
		WithHeader("X-Header-2", "value2").
		WithHeader("X-Header-3", "value3")

	if len(response.Headers) != 3 {
		t.Errorf("Expected 3 headers, got %d", len(response.Headers))
	}

	reqCtx := &fasthttp.RequestCtx{}
	err := WriteResponse(reqCtx, response)
	if err != nil {
		t.Errorf("WriteResponse() error = %v", err)
	}

	for i := 1; i <= 3; i++ {
		headerName := "X-Header-" + string(rune('0'+i))
		expectedValue := "value" + string(rune('0'+i))
		actualValue := string(reqCtx.Response.Header.Peek(headerName))
		if actualValue != expectedValue {
			t.Errorf("Expected header %s=%s, got %s", headerName, expectedValue, actualValue)
		}
	}
}
