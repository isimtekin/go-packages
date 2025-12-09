package mongoclient

import (
	"errors"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Test models for schema validation
type TestUser struct {
	BaseModel `bson:",inline"`
	Email     string `bson:"email" json:"email"`
	Name      string `bson:"name" json:"name"`
	Age       int    `bson:"age" json:"age"`
	Role      string `bson:"role" json:"role"`
	IsActive  bool   `bson:"isActive" json:"isActive"`
}

type TestAddress struct {
	Street  string `bson:"street" json:"street"`
	City    string `bson:"city" json:"city"`
	ZipCode string `bson:"zipCode" json:"zipCode"`
}

type TestUserWithAddress struct {
	BaseModel `bson:",inline"`
	Name      string      `bson:"name" json:"name"`
	Address   TestAddress `bson:"address" json:"address"`
}

// TestNewSchema tests schema creation
func TestNewSchema(t *testing.T) {
	schema := NewSchema()

	if schema == nil {
		t.Fatal("Expected schema to be created")
	}

	if schema.Fields == nil {
		t.Error("Expected Fields map to be initialized")
	}

	if !schema.Timestamps {
		t.Error("Expected Timestamps to be true by default")
	}

	if !schema.Strict {
		t.Error("Expected Strict to be true by default")
	}
}

// TestSchema_Field tests adding fields to schema
func TestSchema_Field(t *testing.T) {
	schema := NewSchema().
		Field("name", TypeString, Required(), MinLen(2), MaxLen(50)).
		Field("age", TypeInt, MinValue(0), MaxValue(150)).
		Field("email", TypeString, Required(), MatchRegex(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)).
		Field("role", TypeString, Enum("admin", "user", "guest"), Default("user"))

	if len(schema.Fields) != 4 {
		t.Errorf("Expected 4 fields, got %d", len(schema.Fields))
	}

	// Check name field
	nameField := schema.Fields["name"]
	if nameField == nil {
		t.Fatal("Expected name field to exist")
	}
	if !nameField.Required {
		t.Error("Expected name field to be required")
	}
	if nameField.MinLength != 2 {
		t.Errorf("Expected MinLength to be 2, got %d", nameField.MinLength)
	}
	if nameField.MaxLength != 50 {
		t.Errorf("Expected MaxLength to be 50, got %d", nameField.MaxLength)
	}

	// Check age field
	ageField := schema.Fields["age"]
	if ageField == nil {
		t.Fatal("Expected age field to exist")
	}
	if ageField.Min != 0 {
		t.Errorf("Expected Min to be 0, got %v", ageField.Min)
	}
	if ageField.Max != 150 {
		t.Errorf("Expected Max to be 150, got %v", ageField.Max)
	}

	// Check role field
	roleField := schema.Fields["role"]
	if roleField == nil {
		t.Fatal("Expected role field to exist")
	}
	if len(roleField.Enum) != 3 {
		t.Errorf("Expected 3 enum values, got %d", len(roleField.Enum))
	}
	if roleField.Default != "user" {
		t.Errorf("Expected Default to be 'user', got %v", roleField.Default)
	}
}

// TestSchema_Validate_RequiredFields tests required field validation
func TestSchema_Validate_RequiredFields(t *testing.T) {
	schema := NewSchema().
		Field("email", TypeString, Required()).
		Field("name", TypeString, Required()).
		Field("age", TypeInt).
		Field("role", TypeString).
		Field("isActive", TypeBool)

	// Test with missing required field
	user := &TestUser{Name: "John"}
	err := schema.Validate(user)
	if err == nil {
		t.Error("Expected validation error for missing email")
	}

	// Test with all required fields
	user.Email = "john@example.com"
	err = schema.Validate(user)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

// TestSchema_Validate_StringLength tests string length validation
func TestSchema_Validate_StringLength(t *testing.T) {
	schema := NewSchema().
		Field("name", TypeString, MinLen(2), MaxLen(10)).
		Field("email", TypeString).
		Field("age", TypeInt).
		Field("role", TypeString).
		Field("isActive", TypeBool)

	tests := []struct {
		name        string
		value       string
		shouldError bool
	}{
		{"too short", "A", true},
		{"minimum", "AB", false},
		{"middle", "Hello", false},
		{"maximum", "HelloWorld", false},
		{"too long", "HelloWorldX", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &TestUser{Name: tt.value}
			err := schema.Validate(user)

			if tt.shouldError && err == nil {
				t.Error("Expected validation error")
			}
			if !tt.shouldError && err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}
		})
	}
}

// TestSchema_Validate_NumericRange tests numeric range validation
func TestSchema_Validate_NumericRange(t *testing.T) {
	schema := NewSchema().
		Field("age", TypeInt, MinValue(0), MaxValue(150)).
		Field("email", TypeString).
		Field("name", TypeString).
		Field("role", TypeString).
		Field("isActive", TypeBool)

	tests := []struct {
		name        string
		age         int
		shouldError bool
	}{
		{"below minimum", -1, true},
		{"at minimum", 0, false},
		{"middle", 30, false},
		{"at maximum", 150, false},
		{"above maximum", 151, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &TestUser{Age: tt.age}
			err := schema.Validate(user)

			if tt.shouldError && err == nil {
				t.Error("Expected validation error")
			}
			if !tt.shouldError && err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}
		})
	}
}

// TestSchema_Validate_Enum tests enum validation
func TestSchema_Validate_Enum(t *testing.T) {
	schema := NewSchema().
		Field("role", TypeString, Enum("admin", "user", "guest")).
		Field("email", TypeString).
		Field("name", TypeString).
		Field("age", TypeInt).
		Field("isActive", TypeBool)

	tests := []struct {
		name        string
		role        string
		shouldError bool
	}{
		{"valid admin", "admin", false},
		{"valid user", "user", false},
		{"valid guest", "guest", false},
		{"invalid role", "superadmin", true},
		{"empty role", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &TestUser{Role: tt.role}
			err := schema.Validate(user)

			if tt.shouldError && err == nil {
				t.Error("Expected validation error")
			}
			if !tt.shouldError && err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}
		})
	}
}

// TestSchema_Validate_Regex tests regex pattern validation
func TestSchema_Validate_Regex(t *testing.T) {
	schema := NewSchema().
		Field("email", TypeString, MatchRegex(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)).
		Field("name", TypeString).
		Field("age", TypeInt).
		Field("role", TypeString).
		Field("isActive", TypeBool)

	tests := []struct {
		name        string
		email       string
		shouldError bool
	}{
		{"valid email", "test@example.com", false},
		{"valid email with subdomain", "test@mail.example.com", false},
		{"invalid - no @", "testexample.com", true},
		{"invalid - no domain", "test@", true},
		{"invalid - no TLD", "test@example", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &TestUser{Email: tt.email}
			err := schema.Validate(user)

			if tt.shouldError && err == nil {
				t.Error("Expected validation error")
			}
			if !tt.shouldError && err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}
		})
	}
}

// TestSchema_Validate_TypeValidation tests type validation
func TestSchema_Validate_TypeValidation(t *testing.T) {
	schema := NewSchema().
		Field("count", TypeInt).
		Field("name", TypeString).
		Field("active", TypeBool)

	// Test with map (wrong types)
	doc := map[string]interface{}{
		"count":  "not a number", // wrong type
		"name":   "test",
		"active": true,
	}

	err := schema.Validate(doc)
	if err == nil {
		t.Error("Expected type validation error")
	}

	// Test with correct types
	doc["count"] = 42
	err = schema.Validate(doc)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

// TestSchema_Validate_StrictMode tests strict mode
func TestSchema_Validate_StrictMode(t *testing.T) {
	schema := NewSchema().
		Field("name", TypeString)

	// Test with unknown field in strict mode (default)
	doc := map[string]interface{}{
		"name":    "test",
		"unknown": "field",
	}

	err := schema.Validate(doc)
	if err == nil {
		t.Error("Expected validation error for unknown field in strict mode")
	}

	// Test with strict mode disabled
	schema.WithStrict(false)
	err = schema.Validate(doc)
	if err != nil {
		t.Errorf("Expected no error with strict mode disabled, got: %v", err)
	}
}

// TestSchema_Validate_CustomValidator tests custom validators
func TestSchema_Validate_CustomValidator(t *testing.T) {
	schema := NewSchema().
		Field("email", TypeString, Validator(func(value interface{}) error {
			email, ok := value.(string)
			if !ok {
				return errors.New("expected string")
			}
			if len(email) > 0 && email[0] == 'x' {
				return errors.New("email cannot start with 'x'")
			}
			return nil
		}))

	// Test valid email
	err := schema.Validate(map[string]interface{}{"email": "test@example.com"})
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Test invalid email (starts with x)
	err = schema.Validate(map[string]interface{}{"email": "xtest@example.com"})
	if err == nil {
		t.Error("Expected validation error")
	}
}

// TestSchema_ValidateUpdate tests update validation
func TestSchema_ValidateUpdate(t *testing.T) {
	schema := NewSchema().
		Field("name", TypeString, MinLen(2)).
		Field("age", TypeInt, MinValue(0))

	// Test valid update with $set
	update := M{
		"$set": M{
			"name": "John",
			"age":  30,
		},
	}

	err := schema.ValidateUpdate(update)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Test invalid update (name too short)
	update = M{
		"$set": M{
			"name": "J",
		},
	}

	err = schema.ValidateUpdate(update)
	if err == nil {
		t.Error("Expected validation error for short name")
	}

	// Test invalid update (unknown field in strict mode)
	update = M{
		"$set": M{
			"unknown": "field",
		},
	}

	err = schema.ValidateUpdate(update)
	if err == nil {
		t.Error("Expected validation error for unknown field")
	}
}

// TestSchema_ApplyDefaults tests default value application
func TestSchema_ApplyDefaults(t *testing.T) {
	schema := NewSchema().
		Field("role", TypeString, Default("user")).
		Field("isActive", TypeBool, Default(true))

	user := &TestUser{
		Name:  "John",
		Email: "john@example.com",
	}

	err := schema.ApplyDefaults(user)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Note: ApplyDefaults only works for fields that exist on the struct
	// and the field name must match
}

// TestSchema_AddIndex tests adding indexes
func TestSchema_AddIndex(t *testing.T) {
	schema := NewSchema().
		AddIndex(M{"email": 1}, true, "email_unique").
		AddIndex(M{"createdAt": -1}, false, "created_at_desc")

	if len(schema.Indexes) != 2 {
		t.Errorf("Expected 2 indexes, got %d", len(schema.Indexes))
	}

	// Check first index
	if schema.Indexes[0].Name != "email_unique" {
		t.Errorf("Expected index name 'email_unique', got '%s'", schema.Indexes[0].Name)
	}
	if !schema.Indexes[0].Unique {
		t.Error("Expected first index to be unique")
	}

	// Check second index
	if schema.Indexes[1].Name != "created_at_desc" {
		t.Errorf("Expected index name 'created_at_desc', got '%s'", schema.Indexes[1].Name)
	}
	if schema.Indexes[1].Unique {
		t.Error("Expected second index to not be unique")
	}
}

// TestSchema_WithTimestamps tests timestamp setting
func TestSchema_WithTimestamps(t *testing.T) {
	schema := NewSchema()

	if !schema.Timestamps {
		t.Error("Expected Timestamps to be true by default")
	}

	schema.WithTimestamps(false)
	if schema.Timestamps {
		t.Error("Expected Timestamps to be false after WithTimestamps(false)")
	}

	schema.WithTimestamps(true)
	if !schema.Timestamps {
		t.Error("Expected Timestamps to be true after WithTimestamps(true)")
	}
}

// TestValidationError tests validation error type
func TestValidationError(t *testing.T) {
	err := &ValidationError{
		Field:   "email",
		Message: "field 'email' is required",
	}

	if err.Error() != "field 'email' is required" {
		t.Errorf("Expected error message 'field 'email' is required', got '%s'", err.Error())
	}

	if !IsValidationError(err) {
		t.Error("Expected IsValidationError to return true")
	}

	if IsValidationError(errors.New("regular error")) {
		t.Error("Expected IsValidationError to return false for regular error")
	}
}

// TestSchema_Validate_ObjectID tests ObjectID type validation
func TestSchema_Validate_ObjectID(t *testing.T) {
	schema := NewSchema().
		Field("userId", TypeObjectID, Required())

	// Test with valid ObjectID
	doc := map[string]interface{}{
		"userId": primitive.NewObjectID(),
	}

	err := schema.Validate(doc)
	if err != nil {
		t.Errorf("Expected no error for valid ObjectID, got: %v", err)
	}

	// Test with string ObjectID (should also be valid)
	doc["userId"] = "507f1f77bcf86cd799439011"
	err = schema.Validate(doc)
	if err != nil {
		t.Errorf("Expected no error for string ObjectID, got: %v", err)
	}

	// Test with invalid type
	doc["userId"] = 12345
	err = schema.Validate(doc)
	if err == nil {
		t.Error("Expected validation error for invalid ObjectID type")
	}
}

// TestSchema_Validate_Array tests array type validation
func TestSchema_Validate_Array(t *testing.T) {
	schema := NewSchema().
		Field("tags", TypeArray, ArrayOf(TypeString))

	// Test with valid array
	doc := map[string]interface{}{
		"tags": []string{"go", "mongodb", "testing"},
	}

	err := schema.Validate(doc)
	if err != nil {
		t.Errorf("Expected no error for valid array, got: %v", err)
	}

	// Test with invalid type (not an array)
	doc["tags"] = "not an array"
	err = schema.Validate(doc)
	if err == nil {
		t.Error("Expected validation error for non-array type")
	}
}

// TestSchema_Validate_Time tests time type validation
func TestSchema_Validate_Time(t *testing.T) {
	schema := NewSchema().
		Field("birthDate", TypeTime)

	// Test with valid time.Time
	doc := map[string]interface{}{
		"birthDate": time.Now(),
	}

	err := schema.Validate(doc)
	if err != nil {
		t.Errorf("Expected no error for valid time, got: %v", err)
	}

	// Test with primitive.DateTime
	doc["birthDate"] = primitive.NewDateTimeFromTime(time.Now())
	err = schema.Validate(doc)
	if err != nil {
		t.Errorf("Expected no error for primitive.DateTime, got: %v", err)
	}

	// Test with invalid type
	doc["birthDate"] = "2023-01-01"
	err = schema.Validate(doc)
	if err == nil {
		t.Error("Expected validation error for string date")
	}
}

// TestSchema_Validate_Bool tests bool type validation
func TestSchema_Validate_Bool(t *testing.T) {
	schema := NewSchema().
		Field("isActive", TypeBool)

	// Test with valid bool
	doc := map[string]interface{}{
		"isActive": true,
	}

	err := schema.Validate(doc)
	if err != nil {
		t.Errorf("Expected no error for valid bool, got: %v", err)
	}

	// Test with invalid type
	doc["isActive"] = "true"
	err = schema.Validate(doc)
	if err == nil {
		t.Error("Expected validation error for string bool")
	}
}

// TestSchema_Validate_Float tests float type validation
func TestSchema_Validate_Float(t *testing.T) {
	schema := NewSchema().
		Field("price", TypeFloat64, MinValue(0.0), MaxValue(1000.0))

	tests := []struct {
		name        string
		price       interface{}
		shouldError bool
	}{
		{"valid float", 99.99, false},
		{"valid int coerced to float", 100, false},
		{"at minimum", 0.0, false},
		{"at maximum", 1000.0, false},
		{"below minimum", -0.01, true},
		{"above maximum", 1000.01, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := map[string]interface{}{"price": tt.price}
			err := schema.Validate(doc)

			if tt.shouldError && err == nil {
				t.Error("Expected validation error")
			}
			if !tt.shouldError && err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}
		})
	}
}

// TestSchema_BeforeValidate_Hook tests before validate hook
func TestSchema_BeforeValidate_Hook(t *testing.T) {
	hookCalled := false

	schema := NewSchema().
		Field("name", TypeString)

	schema.BeforeValidate = func(doc interface{}) error {
		hookCalled = true
		return nil
	}

	doc := map[string]interface{}{"name": "test"}
	err := schema.Validate(doc)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if !hookCalled {
		t.Error("Expected BeforeValidate hook to be called")
	}
}

// TestSchema_BeforeValidate_Error tests before validate hook with error
func TestSchema_BeforeValidate_Error(t *testing.T) {
	schema := NewSchema().
		Field("name", TypeString)

	schema.BeforeValidate = func(doc interface{}) error {
		return errors.New("before validate error")
	}

	doc := map[string]interface{}{"name": "test"}
	err := schema.Validate(doc)

	if err == nil {
		t.Error("Expected error from BeforeValidate hook")
	}

	if err.Error() != "before validate error" {
		t.Errorf("Expected 'before validate error', got '%s'", err.Error())
	}
}

// TestSchema_AfterValidate_Hook tests after validate hook
func TestSchema_AfterValidate_Hook(t *testing.T) {
	hookCalled := false

	schema := NewSchema().
		Field("name", TypeString)

	schema.AfterValidate = func(doc interface{}) error {
		hookCalled = true
		return nil
	}

	doc := map[string]interface{}{"name": "test"}
	err := schema.Validate(doc)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if !hookCalled {
		t.Error("Expected AfterValidate hook to be called")
	}
}

// TestIDSetter_Interface tests IDSetter interface
func TestIDSetter_Interface(t *testing.T) {
	// BaseModel should implement IDSetter
	user := &TestUser{}
	var _ IDSetter = &user.BaseModel

	// Test SetObjectID
	id := primitive.NewObjectID()
	user.SetObjectID(id)

	if user.GetObjectID() != id {
		t.Errorf("Expected ID to be %s, got %s", id.Hex(), user.GetObjectID().Hex())
	}
}

// TestApplyObjectID tests automatic ObjectID generation
func TestApplyObjectID(t *testing.T) {
	user := &TestUser{}

	// ID should be zero initially
	if !user.ID.IsZero() {
		t.Error("Expected ID to be zero initially")
	}

	// Apply ObjectID
	applyObjectID(user)

	// ID should now be set
	if user.ID.IsZero() {
		t.Error("Expected ID to be set after applyObjectID")
	}

	// Apply again should not change existing ID
	existingID := user.ID
	applyObjectID(user)

	if user.ID != existingID {
		t.Error("Expected existing ID to be preserved")
	}
}

// TestApplyObjectID_NonIDSetter tests applyObjectID with non-IDSetter
func TestApplyObjectID_NonIDSetter(t *testing.T) {
	// Should not panic with non-IDSetter
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("applyObjectID should not panic with non-IDSetter, got: %v", r)
		}
	}()

	doc := struct {
		Name string
	}{Name: "test"}

	applyObjectID(&doc)
}

// TestToMap tests toMap function
func TestToMap(t *testing.T) {
	// Test with struct
	user := &TestUser{
		Email: "test@example.com",
		Name:  "Test User",
		Age:   25,
	}
	user.ID = primitive.NewObjectID()

	m, err := toMap(user)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if m["email"] != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got '%v'", m["email"])
	}

	if m["name"] != "Test User" {
		t.Errorf("Expected name 'Test User', got '%v'", m["name"])
	}

	// Test with map
	inputMap := map[string]interface{}{
		"key": "value",
	}

	m, err = toMap(inputMap)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if m["key"] != "value" {
		t.Errorf("Expected key 'value', got '%v'", m["key"])
	}

	// Test with nil
	_, err = toMap(nil)
	if err == nil {
		t.Error("Expected error for nil document")
	}
}

// TestIsZeroValue tests isZeroValue function
func TestIsZeroValue(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected bool
	}{
		{"nil", nil, true},
		{"empty string", "", true},
		{"non-empty string", "test", false},
		{"zero int", 0, true},
		{"non-zero int", 42, false},
		{"false bool", false, true},
		{"true bool", true, false},
		{"empty slice", []string{}, true},
		{"non-empty slice", []string{"a"}, false},
		{"empty map", map[string]string{}, true},
		{"non-empty map", map[string]string{"a": "b"}, false},
		{"zero ObjectID", primitive.NilObjectID, true},
		{"non-zero ObjectID", primitive.NewObjectID(), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isZeroValue(tt.value)
			if result != tt.expected {
				t.Errorf("isZeroValue(%v) = %v, expected %v", tt.value, result, tt.expected)
			}
		})
	}
}
