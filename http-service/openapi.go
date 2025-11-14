package httpservice

import (
	"encoding/json"
	"reflect"
	"strings"
)

// OpenAPISpec represents an OpenAPI 3.0 specification
type OpenAPISpec struct {
	OpenAPI string                 `json:"openapi"`
	Info    OpenAPIInfo            `json:"info"`
	Servers []OpenAPIServer        `json:"servers,omitempty"`
	Paths   map[string]interface{} `json:"paths"`
}

// OpenAPIInfo represents the info section
type OpenAPIInfo struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Version     string `json:"version"`
}

// OpenAPIServer represents a server
type OpenAPIServer struct {
	URL         string `json:"url"`
	Description string `json:"description,omitempty"`
}

// GenerateOpenAPISpec generates an OpenAPI specification
func GenerateOpenAPISpec(config *Config, routes []*Route) *OpenAPISpec {
	spec := &OpenAPISpec{
		OpenAPI: "3.0.0",
		Info: OpenAPIInfo{
			Title:       config.Title,
			Description: config.Description,
			Version:     config.Version,
		},
		Paths: make(map[string]interface{}),
	}

	// Add server
	spec.Servers = []OpenAPIServer{
		{
			URL:         "http://" + config.Addr(),
			Description: "Development server",
		},
	}

	// Process routes
	for _, route := range routes {
		addRouteToSpec(spec, route)
	}

	return spec
}

// addRouteToSpec adds a route to the OpenAPI spec
func addRouteToSpec(spec *OpenAPISpec, route *Route) {
	path := convertPathToOpenAPI(route.Path)

	// Get or create path item
	var pathItem map[string]interface{}
	if existing, ok := spec.Paths[path].(map[string]interface{}); ok {
		pathItem = existing
	} else {
		pathItem = make(map[string]interface{})
		spec.Paths[path] = pathItem
	}

	// Create operation
	operation := make(map[string]interface{})

	if route.Summary != "" {
		operation["summary"] = route.Summary
	}

	if route.Description != "" {
		operation["description"] = route.Description
	}

	if len(route.Tags) > 0 {
		operation["tags"] = route.Tags
	}

	if route.Deprecated {
		operation["deprecated"] = true
	}

	// Add request body if provided
	if route.RequestBody != nil {
		operation["requestBody"] = map[string]interface{}{
			"required": true,
			"content": map[string]interface{}{
				"application/json": map[string]interface{}{
					"schema": generateSchema(route.RequestBody),
				},
			},
		}
	}

	// Add responses
	responses := make(map[string]interface{})
	if len(route.Responses) > 0 {
		for code, response := range route.Responses {
			codeStr := string(rune(code + '0'))
			if code >= 100 {
				codeStr = string(rune(code/100+'0')) + string(rune((code/10)%10+'0')) + string(rune(code%10+'0'))
			}
			responses[codeStr] = map[string]interface{}{
				"description": "Success",
				"content": map[string]interface{}{
					"application/json": map[string]interface{}{
						"schema": generateSchema(response),
					},
				},
			}
		}
	} else {
		// Default 200 response
		responses["200"] = map[string]interface{}{
			"description": "Success",
		}
	}

	operation["responses"] = responses

	// Add operation to path
	method := strings.ToLower(route.Method)
	pathItem[method] = operation
}

// convertPathToOpenAPI converts fasthttp path to OpenAPI format
// /users/{id} -> /users/{id}
func convertPathToOpenAPI(path string) string {
	return path
}

// generateSchema generates a JSON schema from a Go type
func generateSchema(v interface{}) map[string]interface{} {
	if v == nil {
		return map[string]interface{}{
			"type": "object",
		}
	}

	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return map[string]interface{}{
			"type": getJSONType(t),
		}
	}

	schema := map[string]interface{}{
		"type":       "object",
		"properties": make(map[string]interface{}),
	}

	properties := schema["properties"].(map[string]interface{})
	var required []string

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Get JSON tag
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}

		// Parse JSON tag
		parts := strings.Split(jsonTag, ",")
		fieldName := parts[0]

		// Generate field schema
		fieldSchema := generateFieldSchema(field)
		properties[fieldName] = fieldSchema

		// Check if required
		validateTag := field.Tag.Get("validate")
		if strings.Contains(validateTag, "required") {
			required = append(required, fieldName)
		}
	}

	if len(required) > 0 {
		schema["required"] = required
	}

	return schema
}

// generateFieldSchema generates schema for a struct field
func generateFieldSchema(field reflect.StructField) map[string]interface{} {
	schema := make(map[string]interface{})

	fieldType := field.Type
	if fieldType.Kind() == reflect.Ptr {
		fieldType = fieldType.Elem()
	}

	schema["type"] = getJSONType(fieldType)

	// Add validation constraints from validate tag
	validateTag := field.Tag.Get("validate")
	if validateTag != "" {
		addValidationConstraints(schema, validateTag)
	}

	return schema
}

// getJSONType returns the JSON type for a Go type
func getJSONType(t reflect.Type) string {
	switch t.Kind() {
	case reflect.String:
		return "string"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "integer"
	case reflect.Float32, reflect.Float64:
		return "number"
	case reflect.Bool:
		return "boolean"
	case reflect.Slice, reflect.Array:
		return "array"
	case reflect.Struct, reflect.Map:
		return "object"
	default:
		return "string"
	}
}

// addValidationConstraints adds validation constraints to schema
func addValidationConstraints(schema map[string]interface{}, validateTag string) {
	parts := strings.Split(validateTag, ",")
	for _, part := range parts {
		kv := strings.Split(part, "=")
		key := kv[0]

		switch key {
		case "min":
			if len(kv) > 1 {
				schema["minLength"] = kv[1]
			}
		case "max":
			if len(kv) > 1 {
				schema["maxLength"] = kv[1]
			}
		case "email":
			schema["format"] = "email"
		case "url":
			schema["format"] = "url"
		case "uuid":
			schema["format"] = "uuid"
		}
	}
}

// MarshalOpenAPISpec marshals the spec to JSON
func MarshalOpenAPISpec(spec *OpenAPISpec) ([]byte, error) {
	return json.MarshalIndent(spec, "", "  ")
}
