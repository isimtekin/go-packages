package httpservice

import (
	"context"
	"encoding/json"
	"fmt"
)

// BindJSON parses JSON request body into the given struct
func BindJSON(ctx context.Context, v interface{}) error {
	reqCtx := GetRequestCtx(ctx)
	if reqCtx == nil {
		return ErrInvalidRequest
	}

	// Check content type
	contentType := string(reqCtx.Request.Header.ContentType())
	if contentType != "application/json" && contentType != "" {
		return BadRequest("Content-Type must be application/json")
	}

	// Parse JSON
	body := reqCtx.PostBody()
	if len(body) == 0 {
		return BadRequest("Request body is empty")
	}

	if err := json.Unmarshal(body, v); err != nil {
		return BadRequestf("Invalid JSON: %v", err)
	}

	return nil
}

// BindAndValidate parses and validates JSON request body
func BindAndValidate(ctx context.Context, v interface{}, validator *Validator) error {
	// Bind JSON
	if err := BindJSON(ctx, v); err != nil {
		return err
	}

	// Validate if validator is provided
	if validator != nil {
		if err := validator.Validate(v); err != nil {
			return err
		}
	}

	return nil
}

// ParsePathParams extracts path parameters from the request
// Pattern: /users/{id}/posts/{postId}
func ParsePathParams(pattern, path string) (map[string]string, error) {
	params := make(map[string]string)

	// Simple pattern matching for path parameters
	// This is a basic implementation - for production use a proper router
	patternParts := splitPath(pattern)
	pathParts := splitPath(path)

	if len(patternParts) != len(pathParts) {
		return nil, fmt.Errorf("path does not match pattern")
	}

	for i, part := range patternParts {
		if len(part) > 0 && part[0] == '{' && part[len(part)-1] == '}' {
			// Extract parameter name
			paramName := part[1 : len(part)-1]
			params[paramName] = pathParts[i]
		} else if part != pathParts[i] {
			return nil, fmt.Errorf("path does not match pattern")
		}
	}

	return params, nil
}

// splitPath splits a path into parts
func splitPath(path string) []string {
	var parts []string
	current := ""

	for i := 0; i < len(path); i++ {
		if path[i] == '/' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(path[i])
		}
	}

	if current != "" {
		parts = append(parts, current)
	}

	return parts
}

// RequestInfo contains information about the request
type RequestInfo struct {
	Method     string              `json:"method"`
	Path       string              `json:"path"`
	Query      string              `json:"query,omitempty"`
	Headers    map[string]string   `json:"headers,omitempty"`
	RemoteAddr string              `json:"remote_addr"`
	RequestID  string              `json:"request_id,omitempty"`
}

// GetRequestInfo extracts request information
func GetRequestInfo(ctx context.Context) *RequestInfo {
	reqCtx := GetRequestCtx(ctx)
	if reqCtx == nil {
		return nil
	}

	info := &RequestInfo{
		Method:     string(reqCtx.Method()),
		Path:       string(reqCtx.Path()),
		Query:      string(reqCtx.QueryArgs().QueryString()),
		RemoteAddr: reqCtx.RemoteAddr().String(),
		RequestID:  GetRequestID(ctx),
		Headers:    make(map[string]string),
	}

	// Extract common headers
	reqCtx.Request.Header.VisitAll(func(key, value []byte) {
		k := string(key)
		if k == "Authorization" || k == "Content-Type" || k == "User-Agent" {
			info.Headers[k] = string(value)
		}
	})

	return info
}
