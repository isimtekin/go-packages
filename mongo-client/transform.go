package mongoclient

import (
	"encoding/json"
	"reflect"
	"strings"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TransformOptions defines options for transforming documents
type TransformOptions struct {
	// RenameID converts _id to id
	RenameID bool `json:"renameId,omitempty"`

	// Rename maps field names to new names
	Rename map[string]string `json:"rename,omitempty"`

	// Omit lists fields to exclude from output
	Omit []string `json:"omit,omitempty"`

	// Pick lists fields to include (if set, only these fields are included)
	Pick []string `json:"pick,omitempty"`

	// OmitEmpty removes fields with zero/empty values
	OmitEmpty bool `json:"omitEmpty,omitempty"`

	// Custom is a custom transform function
	Custom func(doc map[string]interface{}) map[string]interface{} `json:"-"`
}

// Transform applies transformation to a document
func (s *Schema) Transform(doc interface{}) map[string]interface{} {
	if s.TransformOpts == nil {
		// No transform options, convert to map as-is
		result, err := toMap(doc)
		if err != nil {
			return nil
		}
		return result
	}

	return s.TransformOpts.Apply(doc)
}

// TransformMany applies transformation to multiple documents
func (s *Schema) TransformMany(docs interface{}) []map[string]interface{} {
	rv := reflect.ValueOf(docs)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	if rv.Kind() != reflect.Slice {
		return nil
	}

	results := make([]map[string]interface{}, rv.Len())
	for i := 0; i < rv.Len(); i++ {
		results[i] = s.Transform(rv.Index(i).Interface())
	}

	return results
}

// ToJSON transforms a document and returns JSON bytes
func (s *Schema) ToJSON(doc interface{}) ([]byte, error) {
	transformed := s.Transform(doc)
	return json.Marshal(transformed)
}

// ToJSONString transforms a document and returns JSON string
func (s *Schema) ToJSONString(doc interface{}) (string, error) {
	data, err := s.ToJSON(doc)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ToJSONMany transforms multiple documents and returns JSON bytes
func (s *Schema) ToJSONMany(docs interface{}) ([]byte, error) {
	transformed := s.TransformMany(docs)
	return json.Marshal(transformed)
}

// ToObject transforms a document to a clean map (alias for Transform)
func (s *Schema) ToObject(doc interface{}) map[string]interface{} {
	return s.Transform(doc)
}

// ToObjectMany transforms multiple documents to clean maps (alias for TransformMany)
func (s *Schema) ToObjectMany(docs interface{}) []map[string]interface{} {
	return s.TransformMany(docs)
}

// Apply applies transform options to a document
func (t *TransformOptions) Apply(doc interface{}) map[string]interface{} {
	// Convert to map
	docMap, err := toMap(doc)
	if err != nil {
		return nil
	}

	result := make(map[string]interface{})

	// If Pick is set, only include those fields
	if len(t.Pick) > 0 {
		pickSet := make(map[string]bool)
		for _, field := range t.Pick {
			pickSet[field] = true
		}

		for key, value := range docMap {
			// Check if field should be picked (considering rename)
			originalKey := key
			if t.RenameID && key == "_id" {
				if pickSet["id"] || pickSet["_id"] {
					result[key] = value
				}
			} else if pickSet[originalKey] {
				result[key] = value
			} else {
				// Check if renamed version is in pick
				for orig, renamed := range t.Rename {
					if orig == key && pickSet[renamed] {
						result[key] = value
						break
					}
				}
			}
		}
		docMap = result
		result = make(map[string]interface{})
	}

	// Build omit set
	omitSet := make(map[string]bool)
	for _, field := range t.Omit {
		omitSet[field] = true
	}

	// Process each field
	for key, value := range docMap {
		// Skip omitted fields
		if omitSet[key] {
			continue
		}

		// Skip empty values if OmitEmpty is set
		if t.OmitEmpty && isZeroValue(value) {
			continue
		}

		newKey := key

		// Handle _id -> id rename
		if t.RenameID && key == "_id" {
			newKey = "id"
			// Convert ObjectID to string
			if oid, ok := value.(primitive.ObjectID); ok {
				value = oid.Hex()
			}
		}

		// Apply custom rename
		if renamed, exists := t.Rename[key]; exists {
			newKey = renamed
		}

		result[newKey] = value
	}

	// Apply custom transform
	if t.Custom != nil {
		result = t.Custom(result)
	}

	return result
}

// WithTransform sets transform options on the schema
func (s *Schema) WithTransform(opts TransformOptions) *Schema {
	s.TransformOpts = &opts
	return s
}

// SetTransform sets transform options on the schema
func (s *Schema) SetTransform(opts *TransformOptions) *Schema {
	s.TransformOpts = opts
	return s
}

// Model transform methods

// TransformOne transforms a single document using schema's transform options
func (m *Model) TransformOne(doc interface{}) map[string]interface{} {
	return m.schema.Transform(doc)
}

// TransformMany transforms multiple documents using schema's transform options
func (m *Model) TransformMany(docs interface{}) []map[string]interface{} {
	return m.schema.TransformMany(docs)
}

// ToJSON transforms a document to JSON using schema's transform options
func (m *Model) ToJSON(doc interface{}) ([]byte, error) {
	return m.schema.ToJSON(doc)
}

// ToJSONString transforms a document to JSON string using schema's transform options
func (m *Model) ToJSONString(doc interface{}) (string, error) {
	return m.schema.ToJSONString(doc)
}

// ToJSONMany transforms multiple documents to JSON using schema's transform options
func (m *Model) ToJSONMany(docs interface{}) ([]byte, error) {
	return m.schema.ToJSONMany(docs)
}

// ToObject transforms a document to a clean map using schema's transform options
func (m *Model) ToObject(doc interface{}) map[string]interface{} {
	return m.schema.ToObject(doc)
}

// ToObjectMany transforms multiple documents to clean maps using schema's transform options
func (m *Model) ToObjectMany(docs interface{}) []map[string]interface{} {
	return m.schema.ToObjectMany(docs)
}

// Preset transform configurations

// DefaultAPITransform returns common transform options for API responses
func DefaultAPITransform() TransformOptions {
	return TransformOptions{
		RenameID:  true,
		OmitEmpty: false,
	}
}

// SecureTransform returns transform options that omit sensitive fields
func SecureTransform(sensitiveFields ...string) TransformOptions {
	return TransformOptions{
		RenameID: true,
		Omit:     sensitiveFields,
	}
}

// SnakeCaseTransform returns transform options for snake_case field names
func SnakeCaseTransform() TransformOptions {
	return TransformOptions{
		RenameID: true,
		Rename: map[string]string{
			"createdAt": "created_at",
			"updatedAt": "updated_at",
		},
	}
}

// toSnakeCase converts camelCase to snake_case
func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}
