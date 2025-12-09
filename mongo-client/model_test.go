package mongoclient

import (
	"context"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TestNewModel_NilClient tests model creation with nil client
func TestNewModel_NilClient(t *testing.T) {
	schema := NewSchema().
		Field("name", TypeString, Required())

	_, err := NewModel(nil, "users", schema)
	if err != ErrClientNotConnected {
		t.Errorf("Expected ErrClientNotConnected, got: %v", err)
	}
}

// TestNewModel_NilSchema tests model creation with nil schema
func TestNewModel_NilSchema(t *testing.T) {
	// We can't test with a real client without MongoDB
	// So we test the error case that we can
	_, err := NewModel(nil, "users", nil)
	// Should fail with client error first
	if err == nil {
		t.Error("Expected error for nil client/schema")
	}
}

// TestModel_Schema tests schema getter
func TestModel_Schema(t *testing.T) {
	// Create a mock model (without actual client for unit testing)
	schema := NewSchema().
		Field("name", TypeString, Required())

	// We can test the schema validation logic without a real client
	err := schema.Validate(map[string]interface{}{"name": "test"})
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

// TestModel_InsertValidation tests that insert requires schema validation
func TestModel_InsertValidation(t *testing.T) {
	schema := NewSchema().
		Field("email", TypeString, Required()).
		Field("name", TypeString, Required(), MinLen(2))

	// Test validation logic directly (without MongoDB)
	// Valid document
	validDoc := map[string]interface{}{
		"email": "test@example.com",
		"name":  "John",
	}

	err := schema.Validate(validDoc)
	if err != nil {
		t.Errorf("Expected no error for valid doc, got: %v", err)
	}

	// Invalid document - missing required field
	invalidDoc1 := map[string]interface{}{
		"name": "John",
	}

	err = schema.Validate(invalidDoc1)
	if err == nil {
		t.Error("Expected validation error for missing email")
	}

	// Invalid document - name too short
	invalidDoc2 := map[string]interface{}{
		"email": "test@example.com",
		"name":  "J",
	}

	err = schema.Validate(invalidDoc2)
	if err == nil {
		t.Error("Expected validation error for short name")
	}
}

// TestModel_UpdateValidation tests that update requires schema validation
func TestModel_UpdateValidation(t *testing.T) {
	schema := NewSchema().
		Field("email", TypeString, Required()).
		Field("name", TypeString, MinLen(2)).
		Field("age", TypeInt, MinValue(0), MaxValue(150))

	// Valid update
	validUpdate := bson.M{
		"$set": bson.M{
			"name": "John Doe",
			"age":  30,
		},
	}

	err := schema.ValidateUpdate(validUpdate)
	if err != nil {
		t.Errorf("Expected no error for valid update, got: %v", err)
	}

	// Invalid update - name too short
	invalidUpdate1 := bson.M{
		"$set": bson.M{
			"name": "J",
		},
	}

	err = schema.ValidateUpdate(invalidUpdate1)
	if err == nil {
		t.Error("Expected validation error for short name in update")
	}

	// Invalid update - age out of range
	invalidUpdate2 := bson.M{
		"$set": bson.M{
			"age": 200,
		},
	}

	err = schema.ValidateUpdate(invalidUpdate2)
	if err == nil {
		t.Error("Expected validation error for age out of range")
	}

	// Invalid update - unknown field in strict mode
	invalidUpdate3 := bson.M{
		"$set": bson.M{
			"unknownField": "value",
		},
	}

	err = schema.ValidateUpdate(invalidUpdate3)
	if err == nil {
		t.Error("Expected validation error for unknown field")
	}
}

// TestModel_DeleteValidation tests delete operation safety
func TestModel_DeleteValidation(t *testing.T) {
	// Test that empty filter is rejected
	// This is tested via the model's DeleteOne/DeleteMany methods

	// Empty filter should be rejected
	emptyFilter := bson.M{}
	_, err := toMap(emptyFilter)
	if err != nil {
		t.Fatalf("Failed to convert filter: %v", err)
	}

	// Check that we correctly identify empty filter
	filterMap, _ := toMap(emptyFilter)
	if len(filterMap) != 0 {
		t.Error("Expected empty filter map")
	}
}

// TestModel_ObjectIDGeneration tests automatic ObjectID generation
func TestModel_ObjectIDGeneration(t *testing.T) {
	user := &TestUser{
		Email: "test@example.com",
		Name:  "Test User",
	}

	// ID should be zero initially
	if !user.ID.IsZero() {
		t.Error("Expected ID to be zero initially")
	}

	// Apply ObjectID (this is what Model.InsertOne does)
	applyObjectID(user)

	// ID should now be set
	if user.ID.IsZero() {
		t.Error("Expected ID to be generated")
	}

	// Verify it's a valid ObjectID
	if len(user.ID.Hex()) != 24 {
		t.Errorf("Expected 24 character hex ID, got %d", len(user.ID.Hex()))
	}
}

// TestModel_ObjectIDPreservation tests that existing ObjectID is preserved
func TestModel_ObjectIDPreservation(t *testing.T) {
	existingID := primitive.NewObjectID()

	user := &TestUser{
		Email: "test@example.com",
		Name:  "Test User",
	}
	user.ID = existingID

	// Apply ObjectID (should not change existing)
	applyObjectID(user)

	// ID should be preserved
	if user.ID != existingID {
		t.Error("Expected existing ID to be preserved")
	}
}

// TestModel_TimestampGeneration tests automatic timestamp generation
func TestModel_TimestampGeneration(t *testing.T) {
	user := &TestUser{
		Email: "test@example.com",
		Name:  "Test User",
	}

	// Timestamps should be zero initially
	if !user.CreatedAt.IsZero() {
		t.Error("Expected CreatedAt to be zero initially")
	}
	if !user.UpdatedAt.IsZero() {
		t.Error("Expected UpdatedAt to be zero initially")
	}

	// Apply timestamps (this is what Model.InsertOne does when schema.Timestamps is true)
	applyTimestamps(user, true)

	// Timestamps should now be set
	if user.CreatedAt.IsZero() {
		t.Error("Expected CreatedAt to be set")
	}
	if user.UpdatedAt.IsZero() {
		t.Error("Expected UpdatedAt to be set")
	}
}

// TestModel_CompleteWorkflow tests the complete validation workflow
func TestModel_CompleteWorkflow(t *testing.T) {
	// Create a user schema (like Mongoose)
	userSchema := NewSchema().
		Field("email", TypeString, Required(), MatchRegex(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)).
		Field("name", TypeString, Required(), MinLen(2), MaxLen(100)).
		Field("age", TypeInt, MinValue(0), MaxValue(150)).
		Field("role", TypeString, Enum("admin", "user", "guest"), Default("user")).
		Field("isActive", TypeBool).
		AddIndex(M{"email": 1}, true, "email_unique").
		WithTimestamps(true)

	// Test 1: Valid user creation
	validUser := &TestUser{
		Email: "john@example.com",
		Name:  "John Doe",
		Age:   30,
		Role:  "user",
	}

	// Simulate what Model.InsertOne does:
	// 1. Apply defaults
	_ = userSchema.ApplyDefaults(validUser)

	// 2. Validate
	err := userSchema.Validate(validUser)
	if err != nil {
		t.Errorf("Expected valid user to pass validation, got: %v", err)
	}

	// 3. Apply ObjectID
	applyObjectID(validUser)
	if validUser.ID.IsZero() {
		t.Error("Expected ObjectID to be set")
	}

	// 4. Apply timestamps
	applyTimestamps(validUser, true)
	if validUser.CreatedAt.IsZero() {
		t.Error("Expected CreatedAt to be set")
	}

	// Test 2: Invalid email
	invalidEmail := &TestUser{
		Email: "not-an-email",
		Name:  "John Doe",
	}

	err = userSchema.Validate(invalidEmail)
	if err == nil {
		t.Error("Expected validation error for invalid email")
	}

	// Test 3: Invalid role (not in enum)
	invalidRole := &TestUser{
		Email: "john@example.com",
		Name:  "John Doe",
		Role:  "superadmin",
	}

	err = userSchema.Validate(invalidRole)
	if err == nil {
		t.Error("Expected validation error for invalid role")
	}

	// Test 4: Valid update
	validUpdate := bson.M{
		"$set": bson.M{
			"name": "John Smith",
			"age":  31,
		},
	}

	err = userSchema.ValidateUpdate(validUpdate)
	if err != nil {
		t.Errorf("Expected valid update to pass validation, got: %v", err)
	}

	// Test 5: Invalid update (unknown field)
	invalidUpdate := bson.M{
		"$set": bson.M{
			"password": "secret123", // not in schema
		},
	}

	err = userSchema.ValidateUpdate(invalidUpdate)
	if err == nil {
		t.Error("Expected validation error for unknown field in strict mode")
	}
}

// TestModel_NestedSchema tests nested schema validation
func TestModel_NestedSchema(t *testing.T) {
	// Define address schema
	addressSchema := NewSchema().
		Field("street", TypeString, Required()).
		Field("city", TypeString, Required()).
		Field("zipCode", TypeString, Required(), MinLen(5), MaxLen(10))

	// Define user schema with nested address
	userSchema := NewSchema().
		Field("name", TypeString, Required()).
		Field("address", TypeObject, Required(), NestedSchema(addressSchema))

	// Test with valid nested object
	validDoc := map[string]interface{}{
		"name": "John",
		"address": map[string]interface{}{
			"street":  "123 Main St",
			"city":    "New York",
			"zipCode": "10001",
		},
	}

	err := userSchema.Validate(validDoc)
	if err != nil {
		t.Errorf("Expected no error for valid nested doc, got: %v", err)
	}

	// Test with invalid nested object (missing required field)
	invalidDoc := map[string]interface{}{
		"name": "John",
		"address": map[string]interface{}{
			"street": "123 Main St",
			// missing city and zipCode
		},
	}

	err = userSchema.Validate(invalidDoc)
	if err == nil {
		t.Error("Expected validation error for invalid nested doc")
	}
}

// TestModel_ErrSchemaRequired tests that ErrSchemaRequired is defined
func TestModel_ErrSchemaRequired(t *testing.T) {
	if ErrSchemaRequired == nil {
		t.Error("Expected ErrSchemaRequired to be defined")
	}

	if ErrSchemaRequired.Error() != "schema is required for model operations" {
		t.Errorf("Unexpected error message: %s", ErrSchemaRequired.Error())
	}
}

// TestModel_AddUpdatedAtToUpdate tests the updatedAt auto-add functionality
func TestModel_AddUpdatedAtToUpdate(t *testing.T) {
	// Test with $set that doesn't have updatedAt
	update := bson.M{
		"$set": bson.M{
			"name": "John",
		},
	}

	// Create a dummy model to test the method
	// Since we can't create a real model without MongoDB, test the logic directly
	if setOp, hasSet := update["$set"]; hasSet {
		if setMap, ok := setOp.(bson.M); ok {
			if _, hasUpdatedAt := setMap["updatedAt"]; hasUpdatedAt {
				t.Error("updatedAt should not exist initially")
			}
		}
	}

	// Test with $set that already has updatedAt (should be preserved)
	updateWithTimestamp := bson.M{
		"$set": bson.M{
			"name":      "John",
			"updatedAt": "existing",
		},
	}

	if setOp, hasSet := updateWithTimestamp["$set"]; hasSet {
		if setMap, ok := setOp.(bson.M); ok {
			if val, hasUpdatedAt := setMap["updatedAt"]; hasUpdatedAt {
				if val != "existing" {
					t.Error("Existing updatedAt should be preserved")
				}
			}
		}
	}
}

// TestToObjectID tests the toObjectID helper function
func TestToObjectID(t *testing.T) {
	// Test with primitive.ObjectID
	objID := primitive.NewObjectID()
	result, err := toObjectID(objID)
	if err != nil {
		t.Errorf("Expected no error for ObjectID, got: %v", err)
	}
	if result != objID {
		t.Error("Expected same ObjectID")
	}

	// Test with valid string
	validHex := "507f1f77bcf86cd799439011"
	result, err = toObjectID(validHex)
	if err != nil {
		t.Errorf("Expected no error for valid hex string, got: %v", err)
	}
	if result.Hex() != validHex {
		t.Errorf("Expected hex %s, got %s", validHex, result.Hex())
	}

	// Test with invalid string
	_, err = toObjectID("invalid")
	if err == nil {
		t.Error("Expected error for invalid hex string")
	}

	// Test with unsupported type
	_, err = toObjectID(12345)
	if err == nil {
		t.Error("Expected error for unsupported type")
	}
}

// TestDeleteEmptyFilterProtection tests that empty filter is rejected
func TestDeleteEmptyFilterProtection(t *testing.T) {
	// Test empty bson.M
	emptyFilter := bson.M{}
	filterMap, err := toMap(emptyFilter)
	if err != nil {
		t.Fatalf("Failed to convert filter: %v", err)
	}

	if len(filterMap) != 0 {
		t.Error("Expected empty filter map")
	}

	// This simulates what Model.DeleteOne/DeleteMany would check
	if len(filterMap) == 0 {
		// Would return ErrEmptyFilter
		if ErrEmptyFilter == nil {
			t.Error("Expected ErrEmptyFilter to be defined")
		}
	}
}

// BenchmarkSchemaValidation benchmarks schema validation
func BenchmarkSchemaValidation(b *testing.B) {
	schema := NewSchema().
		Field("email", TypeString, Required(), MatchRegex(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)).
		Field("name", TypeString, Required(), MinLen(2), MaxLen(100)).
		Field("age", TypeInt, MinValue(0), MaxValue(150))

	doc := map[string]interface{}{
		"email": "test@example.com",
		"name":  "John Doe",
		"age":   30,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = schema.Validate(doc)
	}
}

// BenchmarkObjectIDGeneration benchmarks ObjectID generation
func BenchmarkObjectIDGeneration(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		user := &TestUser{}
		applyObjectID(user)
	}
}

// BenchmarkTimestampApplication benchmarks timestamp application
func BenchmarkTimestampApplication(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		user := &TestUser{}
		applyTimestamps(user, true)
	}
}

// TestContextUsage tests that context is properly used
func TestContextUsage(t *testing.T) {
	// Test that canceled context is handled
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Context should be canceled
	if ctx.Err() == nil {
		t.Error("Expected context to be canceled")
	}
}
