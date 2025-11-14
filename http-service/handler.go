package httpservice

import (
	"context"
)

// Route represents an HTTP route
type Route struct {
	Method      string
	Path        string
	Handler     HandlerFunc
	Middlewares []Middleware

	// OpenAPI documentation
	Tags        []string
	Summary     string
	Description string
	Deprecated  bool
	RequestBody interface{} // Example request body for docs
	Responses   map[int]interface{} // Status code -> example response
}

// RouteOption is a function option for configuring routes
type RouteOption func(*Route)

// WithTags sets the tags for OpenAPI documentation
func WithTags(tags ...string) RouteOption {
	return func(r *Route) {
		r.Tags = tags
	}
}

// WithSummary sets the summary for OpenAPI documentation
func WithSummary(summary string) RouteOption {
	return func(r *Route) {
		r.Summary = summary
	}
}

// WithDescription sets the description for OpenAPI documentation
func WithDescription(desc string) RouteOption {
	return func(r *Route) {
		r.Description = desc
	}
}

// WithDeprecated marks the route as deprecated
func WithDeprecated() RouteOption {
	return func(r *Route) {
		r.Deprecated = true
	}
}

// WithRequestBody sets the example request body for documentation
func WithRequestBody(body interface{}) RouteOption {
	return func(r *Route) {
		r.RequestBody = body
	}
}

// WithResponse sets an example response for documentation
func WithResponse(statusCode int, response interface{}) RouteOption {
	return func(r *Route) {
		if r.Responses == nil {
			r.Responses = make(map[int]interface{})
		}
		r.Responses[statusCode] = response
	}
}

// WithMiddleware adds middleware to a specific route
func WithMiddleware(middleware ...Middleware) RouteOption {
	return func(r *Route) {
		r.Middlewares = append(r.Middlewares, middleware...)
	}
}

// Handler types for different use cases

// SimpleHandler is a handler that doesn't need request body
type SimpleHandler func(ctx context.Context) (interface{}, error)

// RequestHandler is a handler that accepts a request body
type RequestHandler[Req any] func(ctx context.Context, req *Req) (interface{}, error)

// TypedHandler is a fully typed handler with request and response types
type TypedHandler[Req any, Res any] func(ctx context.Context, req *Req) (*Res, error)

// WrapSimpleHandler wraps a SimpleHandler to HandlerFunc
func WrapSimpleHandler(handler SimpleHandler) HandlerFunc {
	return func(ctx context.Context) error {
		result, err := handler(ctx)
		if err != nil {
			return err
		}

		reqCtx := GetRequestCtx(ctx)
		if reqCtx != nil {
			return WriteResponse(reqCtx, OK(result))
		}

		return nil
	}
}

// WrapRequestHandler wraps a RequestHandler to HandlerFunc
func WrapRequestHandler[Req any](handler RequestHandler[Req], validator *Validator) HandlerFunc {
	return func(ctx context.Context) error {
		var req Req

		// Parse and validate request
		if err := BindAndValidate(ctx, &req, validator); err != nil {
			return err
		}

		// Execute handler
		result, err := handler(ctx, &req)
		if err != nil {
			return err
		}

		reqCtx := GetRequestCtx(ctx)
		if reqCtx != nil {
			return WriteResponse(reqCtx, OK(result))
		}

		return nil
	}
}

// WrapTypedHandler wraps a TypedHandler to HandlerFunc
func WrapTypedHandler[Req any, Res any](handler TypedHandler[Req, Res], validator *Validator) HandlerFunc {
	return func(ctx context.Context) error {
		var req Req

		// Parse and validate request
		if err := BindAndValidate(ctx, &req, validator); err != nil {
			return err
		}

		// Execute handler
		result, err := handler(ctx, &req)
		if err != nil {
			return err
		}

		reqCtx := GetRequestCtx(ctx)
		if reqCtx != nil {
			return WriteResponse(reqCtx, OK(result))
		}

		return nil
	}
}

// NoBodyHandler is for GET/DELETE requests without request body
type NoBodyHandler func(ctx context.Context) (interface{}, error)

// WrapNoBodyHandler wraps a NoBodyHandler to HandlerFunc
func WrapNoBodyHandler(handler NoBodyHandler) HandlerFunc {
	return func(ctx context.Context) error {
		result, err := handler(ctx)
		if err != nil {
			return err
		}

		if result == nil {
			// No content response
			reqCtx := GetRequestCtx(ctx)
			if reqCtx != nil {
				return WriteResponse(reqCtx, NoContent())
			}
			return nil
		}

		reqCtx := GetRequestCtx(ctx)
		if reqCtx != nil {
			return WriteResponse(reqCtx, OK(result))
		}

		return nil
	}
}
