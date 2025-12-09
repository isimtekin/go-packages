package mongoclient

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSchemaFromJSONString_Basic(t *testing.T) {
	jsonStr := `{
		"name": "users",
		"fields": {
			"email": {
				"type": "string",
				"required": true
			},
			"name": {
				"type": "string",
				"minLength": 2,
				"maxLength": 100
			}
		}
	}`

	schema, err := SchemaFromJSONString(jsonStr)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if schema.CollectionName != "users" {
		t.Errorf("Expected name 'users', got '%s'", schema.CollectionName)
	}

	if len(schema.Fields) != 2 {
		t.Errorf("Expected 2 fields, got %d", len(schema.Fields))
	}

	emailField := schema.Fields["email"]
	if emailField == nil {
		t.Fatal("Expected email field")
	}
	if !emailField.Required {
		t.Error("Expected email to be required")
	}

	nameField := schema.Fields["name"]
	if nameField == nil {
		t.Fatal("Expected name field")
	}
	if nameField.MinLength != 2 {
		t.Errorf("Expected minLength 2, got %d", nameField.MinLength)
	}
	if nameField.MaxLength != 100 {
		t.Errorf("Expected maxLength 100, got %d", nameField.MaxLength)
	}
}

func TestSchemaFromJSONString_AllFieldTypes(t *testing.T) {
	jsonStr := `{
		"name": "test",
		"fields": {
			"strField": { "type": "string" },
			"intField": { "type": "int" },
			"int64Field": { "type": "int64" },
			"floatField": { "type": "float" },
			"boolField": { "type": "bool" },
			"timeField": { "type": "time" },
			"oidField": { "type": "objectid" },
			"arrayField": { "type": "array", "items": "string" },
			"objectField": { "type": "object" },
			"anyField": { "type": "any" }
		}
	}`

	schema, err := SchemaFromJSONString(jsonStr)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	tests := []struct {
		field    string
		expected FieldType
	}{
		{"strField", TypeString},
		{"intField", TypeInt},
		{"int64Field", TypeInt64},
		{"floatField", TypeFloat64},
		{"boolField", TypeBool},
		{"timeField", TypeTime},
		{"oidField", TypeObjectID},
		{"arrayField", TypeArray},
		{"objectField", TypeObject},
		{"anyField", TypeAny},
	}

	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			field := schema.Fields[tt.field]
			if field == nil {
				t.Fatalf("Field %s not found", tt.field)
			}
			if field.Type != tt.expected {
				t.Errorf("Expected type %s, got %s", tt.expected, field.Type)
			}
		})
	}
}

func TestSchemaFromJSONString_FieldOptions(t *testing.T) {
	jsonStr := `{
		"name": "products",
		"fields": {
			"name": {
				"type": "string",
				"required": true,
				"minLength": 1,
				"maxLength": 200
			},
			"price": {
				"type": "float",
				"min": 0,
				"max": 999999.99
			},
			"status": {
				"type": "string",
				"enum": ["active", "inactive", "pending"],
				"default": "pending"
			},
			"sku": {
				"type": "string",
				"pattern": "^[A-Z]{3}-[0-9]{4}$"
			},
			"categoryId": {
				"type": "objectid",
				"ref": "categories"
			},
			"tags": {
				"type": "array",
				"items": "string"
			}
		}
	}`

	schema, err := SchemaFromJSONString(jsonStr)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Check name field
	nameField := schema.Fields["name"]
	if !nameField.Required {
		t.Error("Expected name to be required")
	}

	// Check price field
	priceField := schema.Fields["price"]
	if priceField.Min != float64(0) {
		t.Errorf("Expected min 0, got %v", priceField.Min)
	}

	// Check status field
	statusField := schema.Fields["status"]
	if len(statusField.Enum) != 3 {
		t.Errorf("Expected 3 enum values, got %d", len(statusField.Enum))
	}
	if statusField.Default != "pending" {
		t.Errorf("Expected default 'pending', got '%v'", statusField.Default)
	}

	// Check sku field
	skuField := schema.Fields["sku"]
	if skuField.Match == nil {
		t.Error("Expected pattern to be set")
	}

	// Check categoryId field
	catField := schema.Fields["categoryId"]
	if catField.Ref != "categories" {
		t.Errorf("Expected ref 'categories', got '%s'", catField.Ref)
	}

	// Check tags field
	tagsField := schema.Fields["tags"]
	if tagsField.ArrayType != TypeString {
		t.Errorf("Expected array type string, got %s", tagsField.ArrayType)
	}
}

func TestSchemaFromJSONString_NestedObject(t *testing.T) {
	jsonStr := `{
		"name": "users",
		"fields": {
			"address": {
				"type": "object",
				"fields": {
					"street": { "type": "string", "required": true },
					"city": { "type": "string", "required": true },
					"zipCode": { "type": "string", "minLength": 5 }
				}
			}
		}
	}`

	schema, err := SchemaFromJSONString(jsonStr)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	addressField := schema.Fields["address"]
	if addressField == nil {
		t.Fatal("Expected address field")
	}
	if addressField.Type != TypeObject {
		t.Errorf("Expected type object, got %s", addressField.Type)
	}
	if addressField.Nested == nil {
		t.Fatal("Expected nested schema")
	}

	streetField := addressField.Nested.Fields["street"]
	if streetField == nil {
		t.Fatal("Expected street field in nested schema")
	}
	if !streetField.Required {
		t.Error("Expected street to be required")
	}
}

func TestSchemaFromJSONString_Indexes(t *testing.T) {
	jsonStr := `{
		"name": "users",
		"fields": {
			"email": { "type": "string" }
		},
		"indexes": [
			{ "keys": { "email": 1 }, "unique": true, "name": "email_unique" },
			{ "keys": { "createdAt": -1 } }
		]
	}`

	schema, err := SchemaFromJSONString(jsonStr)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if len(schema.Indexes) != 2 {
		t.Fatalf("Expected 2 indexes, got %d", len(schema.Indexes))
	}

	// Check first index
	if !schema.Indexes[0].Unique {
		t.Error("Expected first index to be unique")
	}
	if schema.Indexes[0].Name != "email_unique" {
		t.Errorf("Expected name 'email_unique', got '%s'", schema.Indexes[0].Name)
	}

	// Check second index
	if schema.Indexes[1].Unique {
		t.Error("Expected second index to not be unique")
	}
}

func TestSchemaFromJSONString_TimestampsAndStrict(t *testing.T) {
	// Test defaults (true)
	jsonStr := `{"name": "test", "fields": {}}`
	schema, _ := SchemaFromJSONString(jsonStr)
	if !schema.Timestamps {
		t.Error("Expected timestamps to be true by default")
	}
	if !schema.Strict {
		t.Error("Expected strict to be true by default")
	}

	// Test explicit false
	jsonStr = `{"name": "test", "timestamps": false, "strict": false, "fields": {}}`
	schema, _ = SchemaFromJSONString(jsonStr)
	if schema.Timestamps {
		t.Error("Expected timestamps to be false")
	}
	if schema.Strict {
		t.Error("Expected strict to be false")
	}
}

func TestSchemaFromJSONString_InvalidJSON(t *testing.T) {
	_, err := SchemaFromJSONString("not valid json")
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestSchemaFromJSONString_InvalidFieldType(t *testing.T) {
	jsonStr := `{
		"name": "test",
		"fields": {
			"field": { "type": "unknowntype" }
		}
	}`

	_, err := SchemaFromJSONString(jsonStr)
	if err == nil {
		t.Error("Expected error for unknown field type")
	}
}

func TestSchemaFromJSON_Nil(t *testing.T) {
	_, err := SchemaFromJSON(nil)
	if err == nil {
		t.Error("Expected error for nil schema")
	}
}

func TestSchemaToJSON_Basic(t *testing.T) {
	schema := NewSchema().
		Field("email", TypeString, Required()).
		Field("age", TypeInt, MinValue(0), MaxValue(150))
	schema.CollectionName = "users"

	js := SchemaToJSON(schema)

	if js.Name != "users" {
		t.Errorf("Expected name 'users', got '%s'", js.Name)
	}

	emailField := js.Fields["email"]
	if emailField.Type != "string" {
		t.Errorf("Expected type 'string', got '%s'", emailField.Type)
	}
	if !emailField.Required {
		t.Error("Expected required to be true")
	}

	ageField := js.Fields["age"]
	if ageField.Type != "int" {
		t.Errorf("Expected type 'int', got '%s'", ageField.Type)
	}
}

func TestSchemaToJSONString(t *testing.T) {
	schema := NewSchema().
		Field("name", TypeString, Required())
	schema.CollectionName = "test"

	jsonStr, err := SchemaToJSONString(schema)
	if err != nil {
		t.Fatalf("Failed to convert to JSON: %v", err)
	}

	if jsonStr == "" {
		t.Error("Expected non-empty JSON string")
	}

	// Parse back and verify
	schema2, err := SchemaFromJSONString(jsonStr)
	if err != nil {
		t.Fatalf("Failed to parse generated JSON: %v", err)
	}

	if schema2.CollectionName != "test" {
		t.Errorf("Expected name 'test', got '%s'", schema2.CollectionName)
	}
}

func TestSchemaFromJSONFile(t *testing.T) {
	// Create temp file
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "schema.json")

	jsonContent := `{
		"name": "products",
		"fields": {
			"name": { "type": "string", "required": true }
		}
	}`

	if err := os.WriteFile(filePath, []byte(jsonContent), 0644); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	schema, err := SchemaFromJSONFile(filePath)
	if err != nil {
		t.Fatalf("Failed to load schema from file: %v", err)
	}

	if schema.CollectionName != "products" {
		t.Errorf("Expected name 'products', got '%s'", schema.CollectionName)
	}
}

func TestSchemaFromJSONFile_NotFound(t *testing.T) {
	_, err := SchemaFromJSONFile("/nonexistent/path/schema.json")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestSchemaFromJSONBytes(t *testing.T) {
	data := []byte(`{"name": "test", "fields": {"id": {"type": "objectid"}}}`)

	schema, err := SchemaFromJSONBytes(data)
	if err != nil {
		t.Fatalf("Failed to parse bytes: %v", err)
	}

	if schema.CollectionName != "test" {
		t.Errorf("Expected name 'test', got '%s'", schema.CollectionName)
	}
}

func TestSchemaToJSONBytes(t *testing.T) {
	schema := NewSchema().Field("id", TypeObjectID)
	schema.CollectionName = "test"

	data, err := SchemaToJSONBytes(schema)
	if err != nil {
		t.Fatalf("Failed to convert to bytes: %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected non-empty bytes")
	}
}

func TestSchemaToJSON_Nil(t *testing.T) {
	js := SchemaToJSON(nil)
	if js != nil {
		t.Error("Expected nil for nil schema")
	}
}

func TestRoundTrip(t *testing.T) {
	// Create a complex schema
	original := NewSchema().
		Field("email", TypeString, Required(), MatchRegex(`^.+@.+$`)).
		Field("name", TypeString, MinLen(2), MaxLen(100)).
		Field("age", TypeInt, MinValue(0), MaxValue(150)).
		Field("role", TypeString, Enum("admin", "user"), Default("user")).
		Field("tags", TypeArray, ArrayOf(TypeString)).
		AddIndex(M{"email": 1}, true, "email_unique")
	original.CollectionName = "users"

	// Convert to JSON
	jsonStr, err := SchemaToJSONString(original)
	if err != nil {
		t.Fatalf("Failed to convert to JSON: %v", err)
	}

	// Parse back
	parsed, err := SchemaFromJSONString(jsonStr)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Verify
	if parsed.CollectionName != original.CollectionName {
		t.Error("Collection name mismatch")
	}
	if len(parsed.Fields) != len(original.Fields) {
		t.Error("Field count mismatch")
	}
	if len(parsed.Indexes) != len(original.Indexes) {
		t.Error("Index count mismatch")
	}

	emailField := parsed.Fields["email"]
	if !emailField.Required {
		t.Error("Email should be required")
	}
	if emailField.Match == nil {
		t.Error("Email should have pattern")
	}
}

func TestParseFieldType_Aliases(t *testing.T) {
	tests := []struct {
		input    string
		expected FieldType
	}{
		{"string", TypeString},
		{"String", TypeString},
		{"int", TypeInt},
		{"integer", TypeInt},
		{"Integer", TypeInt},
		{"int64", TypeInt64},
		{"long", TypeInt64},
		{"float", TypeFloat64},
		{"Float64", TypeFloat64},
		{"double", TypeFloat64},
		{"number", TypeFloat64},
		{"bool", TypeBool},
		{"boolean", TypeBool},
		{"Boolean", TypeBool},
		{"time", TypeTime},
		{"date", TypeTime},
		{"datetime", TypeTime},
		{"DateTime", TypeTime},
		{"objectid", TypeObjectID},
		{"ObjectId", TypeObjectID},
		{"ObjectID", TypeObjectID},
		{"oid", TypeObjectID},
		{"array", TypeArray},
		{"list", TypeArray},
		{"object", TypeObject},
		{"nested", TypeObject},
		{"any", TypeAny},
		{"mixed", TypeAny},
		{"", TypeAny},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := parseFieldType(tt.input)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestValidateWithJSONSchema(t *testing.T) {
	jsonStr := `{
		"name": "users",
		"fields": {
			"email": { "type": "string", "required": true },
			"name": { "type": "string", "required": true, "minLength": 2 },
			"age": { "type": "int", "min": 0, "max": 150 }
		}
	}`

	schema, err := SchemaFromJSONString(jsonStr)
	if err != nil {
		t.Fatalf("Failed to parse JSON schema: %v", err)
	}

	// Valid document
	validDoc := map[string]interface{}{
		"email": "test@example.com",
		"name":  "John",
		"age":   30,
	}

	err = schema.Validate(validDoc)
	if err != nil {
		t.Errorf("Expected valid document to pass: %v", err)
	}

	// Invalid - missing required
	invalidDoc := map[string]interface{}{
		"name": "John",
	}

	err = schema.Validate(invalidDoc)
	if err == nil {
		t.Error("Expected error for missing required field")
	}

	// Invalid - age out of range
	invalidDoc2 := map[string]interface{}{
		"email": "test@example.com",
		"name":  "John",
		"age":   200,
	}

	err = schema.Validate(invalidDoc2)
	if err == nil {
		t.Error("Expected error for age out of range")
	}
}
