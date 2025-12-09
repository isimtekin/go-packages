package mongoclient

import (
	"encoding/json"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Test models for transform
type TransformTestUser struct {
	BaseModel `bson:",inline"`
	Email     string `bson:"email" json:"email"`
	Name      string `bson:"name" json:"name"`
	Password  string `bson:"password" json:"password"`
	Age       int    `bson:"age" json:"age"`
	Role      string `bson:"role" json:"role"`
}

func TestTransformOptions_RenameID(t *testing.T) {
	opts := TransformOptions{
		RenameID: true,
	}

	id := primitive.NewObjectID()
	doc := map[string]interface{}{
		"_id":   id,
		"name":  "John",
		"email": "john@example.com",
	}

	result := opts.Apply(doc)

	// Check _id is renamed to id
	if _, exists := result["_id"]; exists {
		t.Error("Expected _id to be renamed")
	}

	idVal, exists := result["id"]
	if !exists {
		t.Fatal("Expected id field to exist")
	}

	// Check id is converted to string
	if idStr, ok := idVal.(string); !ok {
		t.Error("Expected id to be string")
	} else if idStr != id.Hex() {
		t.Errorf("Expected id to be %s, got %s", id.Hex(), idStr)
	}
}

func TestTransformOptions_Rename(t *testing.T) {
	opts := TransformOptions{
		Rename: map[string]string{
			"createdAt": "created_at",
			"updatedAt": "updated_at",
			"firstName": "first_name",
		},
	}

	doc := map[string]interface{}{
		"firstName": "John",
		"lastName":  "Doe",
		"createdAt": time.Now(),
		"updatedAt": time.Now(),
	}

	result := opts.Apply(doc)

	// Check renamed fields
	if _, exists := result["createdAt"]; exists {
		t.Error("Expected createdAt to be renamed")
	}
	if _, exists := result["created_at"]; !exists {
		t.Error("Expected created_at to exist")
	}

	if _, exists := result["firstName"]; exists {
		t.Error("Expected firstName to be renamed")
	}
	if _, exists := result["first_name"]; !exists {
		t.Error("Expected first_name to exist")
	}

	// lastName should remain unchanged
	if _, exists := result["lastName"]; !exists {
		t.Error("Expected lastName to remain unchanged")
	}
}

func TestTransformOptions_Omit(t *testing.T) {
	opts := TransformOptions{
		Omit: []string{"password", "secretKey", "internalId"},
	}

	doc := map[string]interface{}{
		"name":       "John",
		"email":      "john@example.com",
		"password":   "secret123",
		"secretKey":  "abc123",
		"internalId": 12345,
	}

	result := opts.Apply(doc)

	// Check omitted fields
	if _, exists := result["password"]; exists {
		t.Error("Expected password to be omitted")
	}
	if _, exists := result["secretKey"]; exists {
		t.Error("Expected secretKey to be omitted")
	}
	if _, exists := result["internalId"]; exists {
		t.Error("Expected internalId to be omitted")
	}

	// Check remaining fields
	if _, exists := result["name"]; !exists {
		t.Error("Expected name to remain")
	}
	if _, exists := result["email"]; !exists {
		t.Error("Expected email to remain")
	}
}

func TestTransformOptions_Pick(t *testing.T) {
	opts := TransformOptions{
		Pick: []string{"name", "email"},
	}

	doc := map[string]interface{}{
		"name":     "John",
		"email":    "john@example.com",
		"password": "secret123",
		"age":      30,
		"role":     "admin",
	}

	result := opts.Apply(doc)

	// Check only picked fields exist
	if _, exists := result["name"]; !exists {
		t.Error("Expected name to be picked")
	}
	if _, exists := result["email"]; !exists {
		t.Error("Expected email to be picked")
	}

	// Check other fields are excluded
	if _, exists := result["password"]; exists {
		t.Error("Expected password to be excluded")
	}
	if _, exists := result["age"]; exists {
		t.Error("Expected age to be excluded")
	}
	if _, exists := result["role"]; exists {
		t.Error("Expected role to be excluded")
	}
}

func TestTransformOptions_OmitEmpty(t *testing.T) {
	opts := TransformOptions{
		OmitEmpty: true,
	}

	doc := map[string]interface{}{
		"name":     "John",
		"email":    "",
		"age":      0,
		"active":   false,
		"tags":     []string{},
		"metadata": nil,
	}

	result := opts.Apply(doc)

	// Check empty fields are omitted
	if _, exists := result["email"]; exists {
		t.Error("Expected empty email to be omitted")
	}
	if _, exists := result["age"]; exists {
		t.Error("Expected zero age to be omitted")
	}
	if _, exists := result["active"]; exists {
		t.Error("Expected false active to be omitted")
	}
	if _, exists := result["tags"]; exists {
		t.Error("Expected empty tags to be omitted")
	}
	if _, exists := result["metadata"]; exists {
		t.Error("Expected nil metadata to be omitted")
	}

	// Check non-empty fields remain
	if _, exists := result["name"]; !exists {
		t.Error("Expected name to remain")
	}
}

func TestTransformOptions_Custom(t *testing.T) {
	opts := TransformOptions{
		Custom: func(doc map[string]interface{}) map[string]interface{} {
			// Add a computed field
			if name, ok := doc["name"].(string); ok {
				doc["nameLength"] = len(name)
			}
			// Uppercase email
			if email, ok := doc["email"].(string); ok {
				doc["email"] = email + " (verified)"
			}
			return doc
		},
	}

	doc := map[string]interface{}{
		"name":  "John",
		"email": "john@example.com",
	}

	result := opts.Apply(doc)

	// Check custom field
	if nameLen, ok := result["nameLength"].(int); !ok || nameLen != 4 {
		t.Error("Expected nameLength to be 4")
	}

	// Check modified email
	if email, ok := result["email"].(string); !ok || email != "john@example.com (verified)" {
		t.Errorf("Expected email to be modified, got %s", email)
	}
}

func TestTransformOptions_Combined(t *testing.T) {
	opts := TransformOptions{
		RenameID: true,
		Rename: map[string]string{
			"createdAt": "created_at",
		},
		Omit:      []string{"password"},
		OmitEmpty: true,
	}

	id := primitive.NewObjectID()
	doc := map[string]interface{}{
		"_id":       id,
		"name":      "John",
		"email":     "john@example.com",
		"password":  "secret",
		"bio":       "",
		"createdAt": time.Now(),
	}

	result := opts.Apply(doc)

	// Check all transformations applied
	if _, exists := result["id"]; !exists {
		t.Error("Expected _id to be renamed to id")
	}
	if _, exists := result["_id"]; exists {
		t.Error("Expected _id to not exist")
	}
	if _, exists := result["created_at"]; !exists {
		t.Error("Expected createdAt to be renamed to created_at")
	}
	if _, exists := result["password"]; exists {
		t.Error("Expected password to be omitted")
	}
	if _, exists := result["bio"]; exists {
		t.Error("Expected empty bio to be omitted")
	}
}

func TestSchema_Transform(t *testing.T) {
	schema := NewSchema().
		Field("email", TypeString).
		Field("name", TypeString).
		WithTransform(TransformOptions{
			RenameID: true,
			Omit:     []string{"password"},
		})

	user := &TransformTestUser{
		Email:    "john@example.com",
		Name:     "John",
		Password: "secret",
	}
	user.ID = primitive.NewObjectID()

	result := schema.Transform(user)

	if _, exists := result["id"]; !exists {
		t.Error("Expected id field")
	}
	if _, exists := result["password"]; exists {
		t.Error("Expected password to be omitted")
	}
}

func TestSchema_TransformMany(t *testing.T) {
	schema := NewSchema().
		Field("name", TypeString).
		WithTransform(TransformOptions{
			RenameID: true,
		})

	users := []TransformTestUser{
		{Email: "john@example.com", Name: "John"},
		{Email: "jane@example.com", Name: "Jane"},
	}
	users[0].ID = primitive.NewObjectID()
	users[1].ID = primitive.NewObjectID()

	results := schema.TransformMany(users)

	if len(results) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(results))
	}

	for i, result := range results {
		if _, exists := result["id"]; !exists {
			t.Errorf("Result %d: expected id field", i)
		}
	}
}

func TestSchema_ToJSON(t *testing.T) {
	schema := NewSchema().
		Field("email", TypeString).
		Field("name", TypeString).
		WithTransform(TransformOptions{
			RenameID: true,
			Omit:     []string{"password"},
		})

	user := &TransformTestUser{
		Email:    "john@example.com",
		Name:     "John",
		Password: "secret",
	}
	user.ID = primitive.NewObjectID()

	jsonBytes, err := schema.ToJSON(user)
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	// Parse JSON to verify
	var result map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if _, exists := result["id"]; !exists {
		t.Error("Expected id in JSON")
	}
	if _, exists := result["password"]; exists {
		t.Error("Expected password to be omitted from JSON")
	}
}

func TestSchema_ToJSONString(t *testing.T) {
	schema := NewSchema().
		Field("name", TypeString).
		WithTransform(TransformOptions{
			RenameID: true,
		})

	user := &TransformTestUser{Name: "John"}
	user.ID = primitive.NewObjectID()

	jsonStr, err := schema.ToJSONString(user)
	if err != nil {
		t.Fatalf("ToJSONString failed: %v", err)
	}

	if jsonStr == "" {
		t.Error("Expected non-empty JSON string")
	}

	// Verify it's valid JSON
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		t.Fatalf("Invalid JSON: %v", err)
	}
}

func TestSchema_ToObject(t *testing.T) {
	schema := NewSchema().
		Field("name", TypeString).
		WithTransform(TransformOptions{
			RenameID: true,
		})

	user := &TransformTestUser{Name: "John"}
	user.ID = primitive.NewObjectID()

	result := schema.ToObject(user)

	if _, exists := result["id"]; !exists {
		t.Error("Expected id field")
	}
}

func TestSchema_NoTransform(t *testing.T) {
	// Schema without transform options
	schema := NewSchema().
		Field("name", TypeString)

	user := &TransformTestUser{Name: "John"}
	user.ID = primitive.NewObjectID()

	result := schema.Transform(user)

	// Should still work, just no transformations
	if result == nil {
		t.Error("Expected non-nil result")
	}
}

func TestDefaultAPITransform(t *testing.T) {
	opts := DefaultAPITransform()

	if !opts.RenameID {
		t.Error("Expected RenameID to be true")
	}
}

func TestSecureTransform(t *testing.T) {
	opts := SecureTransform("password", "apiKey", "secretToken")

	if !opts.RenameID {
		t.Error("Expected RenameID to be true")
	}

	if len(opts.Omit) != 3 {
		t.Errorf("Expected 3 omit fields, got %d", len(opts.Omit))
	}
}

func TestSnakeCaseTransform(t *testing.T) {
	opts := SnakeCaseTransform()

	if !opts.RenameID {
		t.Error("Expected RenameID to be true")
	}

	if opts.Rename["createdAt"] != "created_at" {
		t.Error("Expected createdAt -> created_at mapping")
	}

	if opts.Rename["updatedAt"] != "updated_at" {
		t.Error("Expected updatedAt -> updated_at mapping")
	}
}

func TestJSONSchema_Transform(t *testing.T) {
	jsonStr := `{
		"name": "users",
		"fields": {
			"email": { "type": "string" },
			"name": { "type": "string" },
			"password": { "type": "string" }
		},
		"transform": {
			"renameId": true,
			"omit": ["password"],
			"rename": {
				"createdAt": "created_at"
			}
		}
	}`

	schema, err := SchemaFromJSONString(jsonStr)
	if err != nil {
		t.Fatalf("Failed to parse JSON schema: %v", err)
	}

	if schema.TransformOpts == nil {
		t.Fatal("Expected transform options")
	}

	if !schema.TransformOpts.RenameID {
		t.Error("Expected RenameID to be true")
	}

	if len(schema.TransformOpts.Omit) != 1 || schema.TransformOpts.Omit[0] != "password" {
		t.Error("Expected password in omit list")
	}

	if schema.TransformOpts.Rename["createdAt"] != "created_at" {
		t.Error("Expected createdAt -> created_at rename")
	}
}

func TestSchemaToJSON_Transform(t *testing.T) {
	schema := NewSchema().
		Field("name", TypeString).
		WithTransform(TransformOptions{
			RenameID:  true,
			Omit:      []string{"password"},
			OmitEmpty: true,
		})
	schema.CollectionName = "users"

	js := SchemaToJSON(schema)

	if js.Transform == nil {
		t.Fatal("Expected transform in JSON schema")
	}

	if !js.Transform.RenameID {
		t.Error("Expected RenameID to be true")
	}

	if len(js.Transform.Omit) != 1 {
		t.Error("Expected 1 omit field")
	}

	if !js.Transform.OmitEmpty {
		t.Error("Expected OmitEmpty to be true")
	}
}

func TestTransform_RoundTrip(t *testing.T) {
	// Create schema with transform
	original := NewSchema().
		Field("email", TypeString).
		WithTransform(TransformOptions{
			RenameID: true,
			Rename: map[string]string{
				"createdAt": "created_at",
			},
			Omit:      []string{"password"},
			OmitEmpty: true,
		})
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

	// Verify transform options
	if parsed.TransformOpts == nil {
		t.Fatal("Expected transform options after round trip")
	}

	if !parsed.TransformOpts.RenameID {
		t.Error("RenameID mismatch")
	}

	if parsed.TransformOpts.Rename["createdAt"] != "created_at" {
		t.Error("Rename mapping mismatch")
	}

	if len(parsed.TransformOpts.Omit) != 1 || parsed.TransformOpts.Omit[0] != "password" {
		t.Error("Omit list mismatch")
	}

	if !parsed.TransformOpts.OmitEmpty {
		t.Error("OmitEmpty mismatch")
	}
}

func TestTransformOptions_PickWithRenameID(t *testing.T) {
	opts := TransformOptions{
		RenameID: true,
		Pick:     []string{"id", "name"},
	}

	id := primitive.NewObjectID()
	doc := map[string]interface{}{
		"_id":      id,
		"name":     "John",
		"email":    "john@example.com",
		"password": "secret",
	}

	result := opts.Apply(doc)

	// Check id is included and renamed
	if _, exists := result["id"]; !exists {
		t.Error("Expected id to be picked")
	}

	// Check name is included
	if _, exists := result["name"]; !exists {
		t.Error("Expected name to be picked")
	}

	// Check others are excluded
	if _, exists := result["email"]; exists {
		t.Error("Expected email to be excluded")
	}
}

func TestModel_Transform(t *testing.T) {
	// We can test transform without actual MongoDB
	schema := NewSchema().
		Field("email", TypeString).
		Field("name", TypeString).
		WithTransform(TransformOptions{
			RenameID: true,
			Omit:     []string{"password"},
		})

	user := &TransformTestUser{
		Email:    "john@example.com",
		Name:     "John",
		Password: "secret",
	}
	user.ID = primitive.NewObjectID()

	// Test schema transform directly
	result := schema.Transform(user)
	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	if _, exists := result["id"]; !exists {
		t.Error("Expected id field")
	}

	if _, exists := result["password"]; exists {
		t.Error("Expected password to be omitted")
	}
}

func TestSchema_ToJSONMany(t *testing.T) {
	schema := NewSchema().
		Field("name", TypeString).
		WithTransform(TransformOptions{
			RenameID: true,
		})

	users := []TransformTestUser{
		{Name: "John"},
		{Name: "Jane"},
	}
	users[0].ID = primitive.NewObjectID()
	users[1].ID = primitive.NewObjectID()

	jsonBytes, err := schema.ToJSONMany(users)
	if err != nil {
		t.Fatalf("ToJSONMany failed: %v", err)
	}

	var results []map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &results); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	for _, result := range results {
		if _, exists := result["id"]; !exists {
			t.Error("Expected id field in result")
		}
	}
}

// Benchmark tests
func BenchmarkTransform_Simple(b *testing.B) {
	opts := TransformOptions{
		RenameID: true,
	}

	doc := map[string]interface{}{
		"_id":   primitive.NewObjectID(),
		"name":  "John",
		"email": "john@example.com",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		opts.Apply(doc)
	}
}

func BenchmarkTransform_Complex(b *testing.B) {
	opts := TransformOptions{
		RenameID: true,
		Rename: map[string]string{
			"createdAt": "created_at",
			"updatedAt": "updated_at",
		},
		Omit:      []string{"password", "secretKey"},
		OmitEmpty: true,
	}

	doc := map[string]interface{}{
		"_id":       primitive.NewObjectID(),
		"name":      "John",
		"email":     "john@example.com",
		"password":  "secret",
		"secretKey": "abc123",
		"bio":       "",
		"createdAt": time.Now(),
		"updatedAt": time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		opts.Apply(doc)
	}
}

func BenchmarkSchema_ToJSON(b *testing.B) {
	schema := NewSchema().
		Field("email", TypeString).
		Field("name", TypeString).
		WithTransform(TransformOptions{
			RenameID: true,
			Omit:     []string{"password"},
		})

	user := &TransformTestUser{
		Email:    "john@example.com",
		Name:     "John",
		Password: "secret",
	}
	user.ID = primitive.NewObjectID()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = schema.ToJSON(user)
	}
}
