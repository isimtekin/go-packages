package httpservice

import (
	"context"
	"testing"

	"github.com/valyala/fasthttp"
)

func TestWrapSimpleHandler(t *testing.T) {
	handler := func(ctx context.Context) (interface{}, error) {
		return map[string]string{"message": "success"}, nil
	}

	wrappedHandler := WrapSimpleHandler(handler)

	reqCtx := &fasthttp.RequestCtx{}
	ctx := SetRequestCtx(context.Background(), reqCtx)

	err := wrappedHandler(ctx)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if reqCtx.Response.StatusCode() != 200 {
		t.Errorf("Expected status 200, got %d", reqCtx.Response.StatusCode())
	}
}

func TestWrapSimpleHandlerError(t *testing.T) {
	handler := func(ctx context.Context) (interface{}, error) {
		return nil, BadRequest("Invalid input")
	}

	wrappedHandler := WrapSimpleHandler(handler)

	reqCtx := &fasthttp.RequestCtx{}
	ctx := SetRequestCtx(context.Background(), reqCtx)

	err := wrappedHandler(ctx)
	if err == nil {
		t.Error("Expected error")
	}

	httpErr, ok := err.(*HTTPError)
	if !ok {
		t.Error("Expected HTTPError")
	}

	if httpErr.Code != 400 {
		t.Errorf("Expected status 400, got %d", httpErr.Code)
	}
}

func TestWrapRequestHandler(t *testing.T) {
	type TestRequest struct {
		Name string `json:"name" validate:"required"`
		Age  int    `json:"age" validate:"required,min=18"`
	}

	handler := func(ctx context.Context, req *TestRequest) (interface{}, error) {
		return map[string]interface{}{
			"received_name": req.Name,
			"received_age":  req.Age,
		}, nil
	}

	validator := NewValidator()
	wrappedHandler := WrapRequestHandler(handler, validator)

	reqCtx := &fasthttp.RequestCtx{}
	reqCtx.Request.Header.SetContentType("application/json")
	reqCtx.Request.SetBodyString(`{"name":"John","age":25}`)

	ctx := SetRequestCtx(context.Background(), reqCtx)

	err := wrappedHandler(ctx)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if reqCtx.Response.StatusCode() != 200 {
		t.Errorf("Expected status 200, got %d", reqCtx.Response.StatusCode())
	}
}

func TestWrapRequestHandlerValidationError(t *testing.T) {
	type TestRequest struct {
		Name string `json:"name" validate:"required"`
		Age  int    `json:"age" validate:"required,min=18"`
	}

	handler := func(ctx context.Context, req *TestRequest) (interface{}, error) {
		return map[string]string{"status": "ok"}, nil
	}

	validator := NewValidator()
	wrappedHandler := WrapRequestHandler(handler, validator)

	reqCtx := &fasthttp.RequestCtx{}
	reqCtx.Request.Header.SetContentType("application/json")
	reqCtx.Request.SetBodyString(`{"name":"John","age":15}`) // Age too young

	ctx := SetRequestCtx(context.Background(), reqCtx)

	err := wrappedHandler(ctx)
	if err == nil {
		t.Error("Expected validation error")
	}

	httpErr, ok := err.(*HTTPError)
	if !ok {
		t.Error("Expected HTTPError")
	}

	if httpErr.Code != 422 {
		t.Errorf("Expected status 422, got %d", httpErr.Code)
	}
}

func TestWrapTypedHandler(t *testing.T) {
	type Request struct {
		Name string `json:"name" validate:"required"`
	}

	type Response struct {
		Greeting string `json:"greeting"`
	}

	handler := func(ctx context.Context, req *Request) (*Response, error) {
		return &Response{
			Greeting: "Hello, " + req.Name,
		}, nil
	}

	validator := NewValidator()
	wrappedHandler := WrapTypedHandler(handler, validator)

	reqCtx := &fasthttp.RequestCtx{}
	reqCtx.Request.Header.SetContentType("application/json")
	reqCtx.Request.SetBodyString(`{"name":"Alice"}`)

	ctx := SetRequestCtx(context.Background(), reqCtx)

	err := wrappedHandler(ctx)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if reqCtx.Response.StatusCode() != 200 {
		t.Errorf("Expected status 200, got %d", reqCtx.Response.StatusCode())
	}
}

func TestWrapNoBodyHandler(t *testing.T) {
	handler := func(ctx context.Context) (interface{}, error) {
		id := PathParam(ctx, "id")
		return map[string]string{"id": id}, nil
	}

	wrappedHandler := WrapNoBodyHandler(handler)

	reqCtx := &fasthttp.RequestCtx{}
	ctx := SetRequestCtx(context.Background(), reqCtx)
	ctx = SetPathParams(ctx, map[string]string{"id": "123"})

	err := wrappedHandler(ctx)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if reqCtx.Response.StatusCode() != 200 {
		t.Errorf("Expected status 200, got %d", reqCtx.Response.StatusCode())
	}
}

func TestWrapNoBodyHandlerNilResponse(t *testing.T) {
	handler := func(ctx context.Context) (interface{}, error) {
		return nil, nil // No content
	}

	wrappedHandler := WrapNoBodyHandler(handler)

	reqCtx := &fasthttp.RequestCtx{}
	ctx := SetRequestCtx(context.Background(), reqCtx)

	err := wrappedHandler(ctx)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if reqCtx.Response.StatusCode() != 204 {
		t.Errorf("Expected status 204, got %d", reqCtx.Response.StatusCode())
	}
}

func TestRouteOptions(t *testing.T) {
	route := &Route{}

	// Test WithTags
	WithTags("users", "admin")(route)
	if len(route.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(route.Tags))
	}

	// Test WithSummary
	WithSummary("Create a user")(route)
	if route.Summary != "Create a user" {
		t.Errorf("Expected summary 'Create a user', got %s", route.Summary)
	}

	// Test WithDescription
	WithDescription("This endpoint creates a new user")(route)
	if route.Description != "This endpoint creates a new user" {
		t.Errorf("Expected description to be set")
	}

	// Test WithDeprecated
	WithDeprecated()(route)
	if !route.Deprecated {
		t.Error("Expected route to be deprecated")
	}

	// Test WithRequestBody
	type TestRequest struct {
		Name string `json:"name"`
	}
	WithRequestBody(&TestRequest{})(route)
	if route.RequestBody == nil {
		t.Error("Expected request body to be set")
	}

	// Test WithResponse
	type TestResponse struct {
		ID string `json:"id"`
	}
	WithResponse(200, &TestResponse{})(route)
	if route.Responses == nil {
		t.Error("Expected responses to be initialized")
	}
	if route.Responses[200] == nil {
		t.Error("Expected response for status 200")
	}

	// Test WithMiddleware
	testMiddleware := func(next HandlerFunc) HandlerFunc {
		return next
	}
	WithMiddleware(testMiddleware)(route)
	if len(route.Middlewares) != 1 {
		t.Errorf("Expected 1 middleware, got %d", len(route.Middlewares))
	}
}

func TestRouteOptionsChaining(t *testing.T) {
	route := &Route{}

	// Apply multiple options
	options := []RouteOption{
		WithTags("users"),
		WithSummary("Test endpoint"),
		WithDescription("Test description"),
		WithDeprecated(),
	}

	for _, opt := range options {
		opt(route)
	}

	if len(route.Tags) != 1 || route.Tags[0] != "users" {
		t.Error("Tags not applied correctly")
	}

	if route.Summary != "Test endpoint" {
		t.Error("Summary not applied correctly")
	}

	if route.Description != "Test description" {
		t.Error("Description not applied correctly")
	}

	if !route.Deprecated {
		t.Error("Deprecated flag not applied correctly")
	}
}

func TestMultipleResponses(t *testing.T) {
	route := &Route{}

	type SuccessResponse struct {
		Data string `json:"data"`
	}

	type ErrorResponse struct {
		Error string `json:"error"`
	}

	WithResponse(200, &SuccessResponse{})(route)
	WithResponse(400, &ErrorResponse{})(route)
	WithResponse(404, &ErrorResponse{})(route)

	if len(route.Responses) != 3 {
		t.Errorf("Expected 3 responses, got %d", len(route.Responses))
	}

	if route.Responses[200] == nil {
		t.Error("Expected response for status 200")
	}

	if route.Responses[400] == nil {
		t.Error("Expected response for status 400")
	}

	if route.Responses[404] == nil {
		t.Error("Expected response for status 404")
	}
}
