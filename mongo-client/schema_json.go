package mongoclient

import (
	"encoding/json"
	"fmt"
	"os"
)

// JSONSchema represents a schema definition in JSON format (meta-model)
type JSONSchema struct {
	Name       string               `json:"name"`                 // Collection name
	Timestamps *bool                `json:"timestamps,omitempty"` // Auto timestamps (default: true)
	Strict     *bool                `json:"strict,omitempty"`     // Strict mode (default: true)
	Fields     map[string]JSONField `json:"fields"`               // Field definitions
	Indexes    []JSONIndex          `json:"indexes,omitempty"`    // Index definitions
	Transform  *JSONTransform       `json:"transform,omitempty"`  // Transform options
}

// JSONTransform represents transform options in JSON format
type JSONTransform struct {
	RenameID  bool              `json:"renameId,omitempty"`  // _id -> id
	Rename    map[string]string `json:"rename,omitempty"`    // Field renames
	Omit      []string          `json:"omit,omitempty"`      // Fields to exclude
	Pick      []string          `json:"pick,omitempty"`      // Fields to include
	OmitEmpty bool              `json:"omitEmpty,omitempty"` // Remove empty fields
}

// JSONField represents a field definition in JSON format
type JSONField struct {
	Type      string               `json:"type"`                // Field type
	Required  bool                 `json:"required,omitempty"`  // Is required
	Default   interface{}          `json:"default,omitempty"`   // Default value
	MinLength int                  `json:"minLength,omitempty"` // Min string length
	MaxLength int                  `json:"maxLength,omitempty"` // Max string length
	Min       interface{}          `json:"min,omitempty"`       // Min numeric value
	Max       interface{}          `json:"max,omitempty"`       // Max numeric value
	Enum      []interface{}        `json:"enum,omitempty"`      // Allowed values
	Pattern   string               `json:"pattern,omitempty"`   // Regex pattern
	Ref       string               `json:"ref,omitempty"`       // Reference collection
	Items     string               `json:"items,omitempty"`     // Array element type
	Fields    map[string]JSONField `json:"fields,omitempty"`    // Nested object fields
}

// JSONIndex represents an index definition in JSON format
type JSONIndex struct {
	Keys   map[string]int `json:"keys"`             // Field:order (1=asc, -1=desc)
	Unique bool           `json:"unique,omitempty"` // Unique index
	Name   string         `json:"name,omitempty"`   // Index name
}

// SchemaFromJSON converts a JSONSchema to a Schema
func SchemaFromJSON(js *JSONSchema) (*Schema, error) {
	if js == nil {
		return nil, fmt.Errorf("json schema is nil")
	}

	schema := NewSchema()
	schema.CollectionName = js.Name

	// Set timestamps (default true)
	if js.Timestamps != nil {
		schema.Timestamps = *js.Timestamps
	}

	// Set strict mode (default true)
	if js.Strict != nil {
		schema.Strict = *js.Strict
	}

	// Convert fields
	for fieldName, jf := range js.Fields {
		fieldType, err := parseFieldType(jf.Type)
		if err != nil {
			return nil, fmt.Errorf("field '%s': %w", fieldName, err)
		}

		opts := buildFieldOptions(jf)

		// Handle nested object
		if fieldType == TypeObject && len(jf.Fields) > 0 {
			nestedJS := &JSONSchema{
				Fields: jf.Fields,
			}
			nestedSchema, err := SchemaFromJSON(nestedJS)
			if err != nil {
				return nil, fmt.Errorf("field '%s' nested: %w", fieldName, err)
			}
			opts = append(opts, NestedSchema(nestedSchema))
		}

		schema.Field(fieldName, fieldType, opts...)
	}

	// Convert indexes
	for _, ji := range js.Indexes {
		keys := make(M)
		for k, v := range ji.Keys {
			keys[k] = v
		}
		schema.AddIndex(keys, ji.Unique, ji.Name)
	}

	// Convert transform options
	if js.Transform != nil {
		schema.TransformOpts = &TransformOptions{
			RenameID:  js.Transform.RenameID,
			Rename:    js.Transform.Rename,
			Omit:      js.Transform.Omit,
			Pick:      js.Transform.Pick,
			OmitEmpty: js.Transform.OmitEmpty,
		}
	}

	return schema, nil
}

// SchemaFromJSONString parses a JSON string and converts to Schema
func SchemaFromJSONString(jsonStr string) (*Schema, error) {
	var js JSONSchema
	if err := json.Unmarshal([]byte(jsonStr), &js); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}
	return SchemaFromJSON(&js)
}

// SchemaFromJSONFile reads a JSON file and converts to Schema
func SchemaFromJSONFile(filePath string) (*Schema, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	return SchemaFromJSONString(string(data))
}

// SchemaFromJSONBytes parses JSON bytes and converts to Schema
func SchemaFromJSONBytes(data []byte) (*Schema, error) {
	var js JSONSchema
	if err := json.Unmarshal(data, &js); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}
	return SchemaFromJSON(&js)
}

// SchemaToJSON converts a Schema to JSONSchema
func SchemaToJSON(s *Schema) *JSONSchema {
	if s == nil {
		return nil
	}

	js := &JSONSchema{
		Name:    s.CollectionName,
		Fields:  make(map[string]JSONField),
		Indexes: make([]JSONIndex, 0, len(s.Indexes)),
	}

	// Set timestamps and strict only if not default
	if !s.Timestamps {
		f := false
		js.Timestamps = &f
	}
	if !s.Strict {
		f := false
		js.Strict = &f
	}

	// Convert transform options
	if s.TransformOpts != nil {
		js.Transform = &JSONTransform{
			RenameID:  s.TransformOpts.RenameID,
			Rename:    s.TransformOpts.Rename,
			Omit:      s.TransformOpts.Omit,
			Pick:      s.TransformOpts.Pick,
			OmitEmpty: s.TransformOpts.OmitEmpty,
		}
	}

	// Convert fields
	for fieldName, fd := range s.Fields {
		jf := JSONField{
			Type:      fieldTypeToString(fd.Type),
			Required:  fd.Required,
			Default:   fd.Default,
			MinLength: fd.MinLength,
			MaxLength: fd.MaxLength,
			Min:       fd.Min,
			Max:       fd.Max,
			Enum:      fd.Enum,
			Ref:       fd.Ref,
			Items:     fieldTypeToString(fd.ArrayType),
		}

		if fd.Match != nil {
			jf.Pattern = fd.Match.String()
		}

		if fd.Nested != nil {
			nestedJS := SchemaToJSON(fd.Nested)
			if nestedJS != nil {
				jf.Fields = nestedJS.Fields
			}
		}

		js.Fields[fieldName] = jf
	}

	// Convert indexes
	for _, idx := range s.Indexes {
		ji := JSONIndex{
			Keys:   make(map[string]int),
			Unique: idx.Unique,
			Name:   idx.Name,
		}
		for k, v := range idx.Keys {
			switch val := v.(type) {
			case int:
				ji.Keys[k] = val
			case int32:
				ji.Keys[k] = int(val)
			case int64:
				ji.Keys[k] = int(val)
			case float64:
				ji.Keys[k] = int(val)
			}
		}
		js.Indexes = append(js.Indexes, ji)
	}

	return js
}

// SchemaToJSONString converts a Schema to JSON string
func SchemaToJSONString(s *Schema) (string, error) {
	js := SchemaToJSON(s)
	data, err := json.MarshalIndent(js, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal: %w", err)
	}
	return string(data), nil
}

// SchemaToJSONBytes converts a Schema to JSON bytes
func SchemaToJSONBytes(s *Schema) ([]byte, error) {
	js := SchemaToJSON(s)
	return json.MarshalIndent(js, "", "  ")
}

// parseFieldType converts string type to FieldType
func parseFieldType(typeStr string) (FieldType, error) {
	switch typeStr {
	case "string", "String":
		return TypeString, nil
	case "int", "Int", "integer", "Integer":
		return TypeInt, nil
	case "int64", "Int64", "long", "Long":
		return TypeInt64, nil
	case "float", "Float", "float64", "Float64", "double", "Double", "number", "Number":
		return TypeFloat64, nil
	case "bool", "Bool", "boolean", "Boolean":
		return TypeBool, nil
	case "time", "Time", "date", "Date", "datetime", "DateTime":
		return TypeTime, nil
	case "objectid", "ObjectId", "ObjectID", "oid":
		return TypeObjectID, nil
	case "array", "Array", "list", "List":
		return TypeArray, nil
	case "object", "Object", "nested", "Nested":
		return TypeObject, nil
	case "any", "Any", "mixed", "Mixed", "":
		return TypeAny, nil
	default:
		return "", fmt.Errorf("unknown type: %s", typeStr)
	}
}

// fieldTypeToString converts FieldType to string
func fieldTypeToString(ft FieldType) string {
	switch ft {
	case TypeString:
		return "string"
	case TypeInt:
		return "int"
	case TypeInt64:
		return "int64"
	case TypeFloat64:
		return "float"
	case TypeBool:
		return "bool"
	case TypeTime:
		return "time"
	case TypeObjectID:
		return "objectid"
	case TypeArray:
		return "array"
	case TypeObject:
		return "object"
	case TypeAny:
		return "any"
	default:
		return string(ft)
	}
}

// buildFieldOptions builds FieldOption slice from JSONField
func buildFieldOptions(jf JSONField) []FieldOption {
	var opts []FieldOption

	if jf.Required {
		opts = append(opts, Required())
	}
	if jf.Default != nil {
		opts = append(opts, Default(jf.Default))
	}
	if jf.MinLength > 0 {
		opts = append(opts, MinLen(jf.MinLength))
	}
	if jf.MaxLength > 0 {
		opts = append(opts, MaxLen(jf.MaxLength))
	}
	if jf.Min != nil {
		opts = append(opts, MinValue(jf.Min))
	}
	if jf.Max != nil {
		opts = append(opts, MaxValue(jf.Max))
	}
	if len(jf.Enum) > 0 {
		opts = append(opts, Enum(jf.Enum...))
	}
	if jf.Pattern != "" {
		opts = append(opts, MatchRegex(jf.Pattern))
	}
	if jf.Ref != "" {
		opts = append(opts, Ref(jf.Ref))
	}
	if jf.Items != "" {
		if itemType, err := parseFieldType(jf.Items); err == nil {
			opts = append(opts, ArrayOf(itemType))
		}
	}

	return opts
}
