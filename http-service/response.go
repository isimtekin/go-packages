package httpservice

import (
	"encoding/json"

	"github.com/valyala/fasthttp"
)

// Response represents an HTTP response
type Response struct {
	StatusCode int
	Body       interface{}
	Headers    map[string]string
}

// NewResponse creates a new response
func NewResponse(statusCode int, body interface{}) *Response {
	return &Response{
		StatusCode: statusCode,
		Body:       body,
		Headers:    make(map[string]string),
	}
}

// WithHeader adds a header to the response
func (r *Response) WithHeader(key, value string) *Response {
	r.Headers[key] = value
	return r
}

// WriteResponse writes a response to fasthttp.RequestCtx
func WriteResponse(ctx *fasthttp.RequestCtx, response *Response) error {
	// Set status code
	ctx.SetStatusCode(response.StatusCode)

	// Set headers
	for key, value := range response.Headers {
		ctx.Response.Header.Set(key, value)
	}

	// Set content type if not already set
	if _, exists := response.Headers["Content-Type"]; !exists {
		ctx.Response.Header.Set("Content-Type", "application/json")
	}

	// Write body
	if response.Body != nil {
		data, err := json.Marshal(response.Body)
		if err != nil {
			return err
		}
		ctx.SetBody(data)
	}

	return nil
}

// Response helpers

// OK returns a 200 OK response
func OK(body interface{}) *Response {
	return NewResponse(fasthttp.StatusOK, body)
}

// Created returns a 201 Created response
func Created(body interface{}) *Response {
	return NewResponse(fasthttp.StatusCreated, body)
}

// Accepted returns a 202 Accepted response
func Accepted(body interface{}) *Response {
	return NewResponse(fasthttp.StatusAccepted, body)
}

// NoContent returns a 204 No Content response
func NoContent() *Response {
	return NewResponse(fasthttp.StatusNoContent, nil)
}

// JSON returns a JSON response with given status code
func JSON(statusCode int, body interface{}) *Response {
	return NewResponse(statusCode, body)
}

// Error response helpers

// WriteError writes an error response
func WriteError(ctx *fasthttp.RequestCtx, err error) {
	if httpErr, ok := err.(*HTTPError); ok {
		WriteResponse(ctx, &Response{
			StatusCode: httpErr.Code,
			Body:       httpErr,
		})
		return
	}

	// Default to 500 for unknown errors
	WriteResponse(ctx, &Response{
		StatusCode: fasthttp.StatusInternalServerError,
		Body: &HTTPError{
			Code:    fasthttp.StatusInternalServerError,
			Message: "Internal server error",
		},
	})
}

// HealthResponse represents a health check response
type HealthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version,omitempty"`
}

// NewHealthResponse creates a new health response
func NewHealthResponse(status, version string) *HealthResponse {
	return &HealthResponse{
		Status:  status,
		Version: version,
	}
}
