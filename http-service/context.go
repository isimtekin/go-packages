package httpservice

import (
	"context"
	"strconv"

	"github.com/valyala/fasthttp"
)

// Context keys
type contextKey string

const (
	contextKeyRequestCtx contextKey = "request_ctx"
	contextKeyPathParams contextKey = "path_params"
	contextKeyRequestID  contextKey = "request_id"
)

// GetRequestCtx retrieves the fasthttp.RequestCtx from context
func GetRequestCtx(ctx context.Context) *fasthttp.RequestCtx {
	if reqCtx, ok := ctx.Value(contextKeyRequestCtx).(*fasthttp.RequestCtx); ok {
		return reqCtx
	}
	return nil
}

// SetRequestCtx sets the fasthttp.RequestCtx in context
func SetRequestCtx(ctx context.Context, reqCtx *fasthttp.RequestCtx) context.Context {
	return context.WithValue(ctx, contextKeyRequestCtx, reqCtx)
}

// GetPathParams retrieves path parameters from context
func GetPathParams(ctx context.Context) map[string]string {
	if params, ok := ctx.Value(contextKeyPathParams).(map[string]string); ok {
		return params
	}
	return make(map[string]string)
}

// SetPathParams sets path parameters in context
func SetPathParams(ctx context.Context, params map[string]string) context.Context {
	return context.WithValue(ctx, contextKeyPathParams, params)
}

// GetRequestID retrieves the request ID from context
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(contextKeyRequestID).(string); ok {
		return id
	}
	return ""
}

// SetRequestID sets the request ID in context
func SetRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, contextKeyRequestID, id)
}

// PathParam retrieves a path parameter by name
func PathParam(ctx context.Context, name string) string {
	params := GetPathParams(ctx)
	return params[name]
}

// QueryParam retrieves a query parameter by name
func QueryParam(ctx context.Context, name string) string {
	reqCtx := GetRequestCtx(ctx)
	if reqCtx == nil {
		return ""
	}
	return string(reqCtx.QueryArgs().Peek(name))
}

// QueryParamWithDefault retrieves a query parameter with default value
func QueryParamWithDefault(ctx context.Context, name, defaultValue string) string {
	value := QueryParam(ctx, name)
	if value == "" {
		return defaultValue
	}
	return value
}

// QueryParamInt retrieves a query parameter as int
func QueryParamInt(ctx context.Context, name string, defaultValue int) int {
	value := QueryParam(ctx, name)
	if value == "" {
		return defaultValue
	}
	if i, err := strconv.Atoi(value); err == nil {
		return i
	}
	return defaultValue
}

// QueryParamBool retrieves a query parameter as bool
func QueryParamBool(ctx context.Context, name string, defaultValue bool) bool {
	value := QueryParam(ctx, name)
	if value == "" {
		return defaultValue
	}
	if b, err := strconv.ParseBool(value); err == nil {
		return b
	}
	return defaultValue
}

// Header retrieves a header value
func Header(ctx context.Context, name string) string {
	reqCtx := GetRequestCtx(ctx)
	if reqCtx == nil {
		return ""
	}
	return string(reqCtx.Request.Header.Peek(name))
}

// SetHeader sets a response header
func SetHeader(ctx context.Context, name, value string) {
	reqCtx := GetRequestCtx(ctx)
	if reqCtx != nil {
		reqCtx.Response.Header.Set(name, value)
	}
}

// SetStatus sets the response status code
func SetStatus(ctx context.Context, code int) {
	reqCtx := GetRequestCtx(ctx)
	if reqCtx != nil {
		reqCtx.SetStatusCode(code)
	}
}

// Method retrieves the HTTP method
func Method(ctx context.Context) string {
	reqCtx := GetRequestCtx(ctx)
	if reqCtx == nil {
		return ""
	}
	return string(reqCtx.Method())
}

// Path retrieves the request path
func Path(ctx context.Context) string {
	reqCtx := GetRequestCtx(ctx)
	if reqCtx == nil {
		return ""
	}
	return string(reqCtx.Path())
}

// RemoteAddr retrieves the remote address
func RemoteAddr(ctx context.Context) string {
	reqCtx := GetRequestCtx(ctx)
	if reqCtx == nil {
		return ""
	}
	return reqCtx.RemoteAddr().String()
}
