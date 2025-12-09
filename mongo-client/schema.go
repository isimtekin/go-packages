package mongoclient

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// FieldType represents the type of a schema field
type FieldType string

const (
	TypeString   FieldType = "string"
	TypeInt      FieldType = "int"
	TypeInt64    FieldType = "int64"
	TypeFloat64  FieldType = "float64"
	TypeBool     FieldType = "bool"
	TypeTime     FieldType = "time"
	TypeObjectID FieldType = "objectid"
	TypeArray    FieldType = "array"
	TypeObject   FieldType = "object"
	TypeAny      FieldType = "any"
)

// FieldDefinition defines a single field in a schema
type FieldDefinition struct {
	Name       string
	Type       FieldType
	Required   bool
	Default    interface{}
	MinLength  int // for strings
	MaxLength  int // for strings
	Min        interface{}
	Max        interface{}
	Enum       []interface{}
	Match      *regexp.Regexp // regex pattern for strings
	Ref        string         // reference to another collection (for ObjectID fields)
	ArrayType  FieldType      // type of array elements
	Nested     *Schema        // for nested objects
	Validators []ValidatorFunc
}

// ValidatorFunc is a custom validation function
type ValidatorFunc func(value interface{}) error

// Schema defines the structure and validation rules for a collection
type Schema struct {
	Fields          map[string]*FieldDefinition
	Timestamps      bool // auto-add createdAt and updatedAt
	Strict          bool // reject fields not in schema
	CollectionName  string
	Indexes         []IndexDefinition
	fieldOrder      []string // to maintain field order
	BeforeValidate  func(doc interface{}) error
	AfterValidate   func(doc interface{}) error
	TransformOpts   *TransformOptions // transform options for ToJSON/ToObject
}

// NewSchema creates a new schema
func NewSchema() *Schema {
	return &Schema{
		Fields:     make(map[string]*FieldDefinition),
		Timestamps: true, // default to true like Mongoose
		Strict:     true, // default to strict mode
		fieldOrder: make([]string, 0),
	}
}

// Field adds a field definition to the schema
func (s *Schema) Field(name string, fieldType FieldType, opts ...FieldOption) *Schema {
	field := &FieldDefinition{
		Name: name,
		Type: fieldType,
	}

	for _, opt := range opts {
		opt(field)
	}

	s.Fields[name] = field
	s.fieldOrder = append(s.fieldOrder, name)
	return s
}

// AddIndex adds an index definition to the schema
func (s *Schema) AddIndex(keys M, unique bool, name string) *Schema {
	s.Indexes = append(s.Indexes, IndexDefinition{
		Keys:   keys,
		Unique: unique,
		Name:   name,
	})
	return s
}

// WithTimestamps enables or disables automatic timestamps
func (s *Schema) WithTimestamps(enabled bool) *Schema {
	s.Timestamps = enabled
	return s
}

// WithStrict enables or disables strict mode
func (s *Schema) WithStrict(enabled bool) *Schema {
	s.Strict = enabled
	return s
}

// FieldOption is a functional option for field definitions
type FieldOption func(*FieldDefinition)

// Required marks a field as required
func Required() FieldOption {
	return func(f *FieldDefinition) {
		f.Required = true
	}
}

// Default sets a default value for the field
func Default(value interface{}) FieldOption {
	return func(f *FieldDefinition) {
		f.Default = value
	}
}

// MinLen sets minimum length for string fields
func MinLen(length int) FieldOption {
	return func(f *FieldDefinition) {
		f.MinLength = length
	}
}

// MaxLen sets maximum length for string fields
func MaxLen(length int) FieldOption {
	return func(f *FieldDefinition) {
		f.MaxLength = length
	}
}

// MinValue sets minimum value for numeric fields
func MinValue(value interface{}) FieldOption {
	return func(f *FieldDefinition) {
		f.Min = value
	}
}

// MaxValue sets maximum value for numeric fields
func MaxValue(value interface{}) FieldOption {
	return func(f *FieldDefinition) {
		f.Max = value
	}
}

// Enum restricts field to specific values
func Enum(values ...interface{}) FieldOption {
	return func(f *FieldDefinition) {
		f.Enum = values
	}
}

// MatchRegex sets a regex pattern for string validation
func MatchRegex(pattern string) FieldOption {
	return func(f *FieldDefinition) {
		f.Match = regexp.MustCompile(pattern)
	}
}

// Ref sets a reference to another collection (for ObjectID fields)
func Ref(collectionName string) FieldOption {
	return func(f *FieldDefinition) {
		f.Ref = collectionName
	}
}

// ArrayOf sets the type of array elements
func ArrayOf(elementType FieldType) FieldOption {
	return func(f *FieldDefinition) {
		f.ArrayType = elementType
	}
}

// NestedSchema sets a nested schema for object fields
func NestedSchema(schema *Schema) FieldOption {
	return func(f *FieldDefinition) {
		f.Nested = schema
	}
}

// Validator adds a custom validation function
func Validator(fn ValidatorFunc) FieldOption {
	return func(f *FieldDefinition) {
		f.Validators = append(f.Validators, fn)
	}
}

// Validate validates a document against the schema
func (s *Schema) Validate(doc interface{}) error {
	// Call BeforeValidate hook if set
	if s.BeforeValidate != nil {
		if err := s.BeforeValidate(doc); err != nil {
			return err
		}
	}

	// Convert document to map for validation
	docMap, err := toMap(doc)
	if err != nil {
		return fmt.Errorf("failed to convert document to map: %w", err)
	}

	// Validate each field in schema
	for fieldName, fieldDef := range s.Fields {
		value, exists := docMap[fieldName]

		// Check required fields
		if fieldDef.Required && (!exists || isZeroValue(value)) {
			return &ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("field '%s' is required", fieldName),
			}
		}

		// Skip validation if field doesn't exist and not required
		if !exists {
			continue
		}

		// Validate field type and constraints
		if err := s.validateField(fieldName, fieldDef, value); err != nil {
			return err
		}
	}

	// In strict mode, reject unknown fields
	if s.Strict {
		for fieldName := range docMap {
			// Skip internal fields
			if fieldName == "_id" || fieldName == "createdAt" || fieldName == "updatedAt" {
				continue
			}
			if _, exists := s.Fields[fieldName]; !exists {
				return &ValidationError{
					Field:   fieldName,
					Message: fmt.Sprintf("field '%s' is not defined in schema", fieldName),
				}
			}
		}
	}

	// Call AfterValidate hook if set
	if s.AfterValidate != nil {
		if err := s.AfterValidate(doc); err != nil {
			return err
		}
	}

	return nil
}

// ValidateUpdate validates update data (only validates fields that are being updated)
func (s *Schema) ValidateUpdate(update interface{}) error {
	updateMap, err := toMap(update)
	if err != nil {
		return fmt.Errorf("failed to convert update to map: %w", err)
	}

	// If update uses MongoDB operators like $set, validate the values inside
	if setData, hasSet := updateMap["$set"]; hasSet {
		if setMap, ok := setData.(map[string]interface{}); ok {
			return s.validateUpdateFields(setMap)
		}
		if setMap, ok := setData.(M); ok {
			return s.validateUpdateFields(map[string]interface{}(setMap))
		}
	}

	// If no operators, treat as direct field update
	return s.validateUpdateFields(updateMap)
}

func (s *Schema) validateUpdateFields(fields map[string]interface{}) error {
	for fieldName, value := range fields {
		// Skip internal fields
		if fieldName == "_id" || fieldName == "createdAt" || fieldName == "updatedAt" {
			continue
		}

		// In strict mode, reject unknown fields
		fieldDef, exists := s.Fields[fieldName]
		if !exists {
			if s.Strict {
				return &ValidationError{
					Field:   fieldName,
					Message: fmt.Sprintf("field '%s' is not defined in schema", fieldName),
				}
			}
			continue
		}

		// Validate field constraints
		if err := s.validateField(fieldName, fieldDef, value); err != nil {
			return err
		}
	}

	return nil
}

func (s *Schema) validateField(fieldName string, fieldDef *FieldDefinition, value interface{}) error {
	// Skip nil values (they pass type check if not required)
	if value == nil {
		return nil
	}

	// Type validation
	if err := validateType(fieldDef.Type, value); err != nil {
		return &ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("field '%s': %s", fieldName, err.Error()),
		}
	}

	// String validations
	if fieldDef.Type == TypeString {
		str, ok := value.(string)
		if ok {
			if fieldDef.MinLength > 0 && len(str) < fieldDef.MinLength {
				return &ValidationError{
					Field:   fieldName,
					Message: fmt.Sprintf("field '%s' must be at least %d characters", fieldName, fieldDef.MinLength),
				}
			}
			if fieldDef.MaxLength > 0 && len(str) > fieldDef.MaxLength {
				return &ValidationError{
					Field:   fieldName,
					Message: fmt.Sprintf("field '%s' must be at most %d characters", fieldName, fieldDef.MaxLength),
				}
			}
			if fieldDef.Match != nil && !fieldDef.Match.MatchString(str) {
				return &ValidationError{
					Field:   fieldName,
					Message: fmt.Sprintf("field '%s' does not match required pattern", fieldName),
				}
			}
		}
	}

	// Numeric validations
	if fieldDef.Min != nil || fieldDef.Max != nil {
		numValue, err := toFloat64(value)
		if err == nil {
			if fieldDef.Min != nil {
				minVal, _ := toFloat64(fieldDef.Min)
				if numValue < minVal {
					return &ValidationError{
						Field:   fieldName,
						Message: fmt.Sprintf("field '%s' must be at least %v", fieldName, fieldDef.Min),
					}
				}
			}
			if fieldDef.Max != nil {
				maxVal, _ := toFloat64(fieldDef.Max)
				if numValue > maxVal {
					return &ValidationError{
						Field:   fieldName,
						Message: fmt.Sprintf("field '%s' must be at most %v", fieldName, fieldDef.Max),
					}
				}
			}
		}
	}

	// Enum validation
	if len(fieldDef.Enum) > 0 {
		found := false
		for _, enumVal := range fieldDef.Enum {
			if reflect.DeepEqual(value, enumVal) {
				found = true
				break
			}
		}
		if !found {
			return &ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("field '%s' must be one of: %v", fieldName, fieldDef.Enum),
			}
		}
	}

	// Custom validators
	for _, validator := range fieldDef.Validators {
		if err := validator(value); err != nil {
			return &ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("field '%s': %s", fieldName, err.Error()),
			}
		}
	}

	// Nested schema validation
	if fieldDef.Nested != nil && fieldDef.Type == TypeObject {
		if err := fieldDef.Nested.Validate(value); err != nil {
			return err
		}
	}

	return nil
}

// ApplyDefaults applies default values to a document
func (s *Schema) ApplyDefaults(doc interface{}) error {
	docMap, err := toMap(doc)
	if err != nil {
		return err
	}

	// Apply defaults for missing fields
	for fieldName, fieldDef := range s.Fields {
		if fieldDef.Default != nil {
			if _, exists := docMap[fieldName]; !exists {
				if err := setField(doc, fieldName, fieldDef.Default); err != nil {
					// Field might not exist on struct, that's okay
					continue
				}
			}
		}
	}

	return nil
}

// validateType checks if a value matches the expected type
func validateType(expected FieldType, value interface{}) error {
	if value == nil {
		return nil
	}

	switch expected {
	case TypeString:
		if _, ok := value.(string); !ok {
			return fmt.Errorf("expected string, got %T", value)
		}
	case TypeInt:
		switch value.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32:
			// OK
		default:
			return fmt.Errorf("expected int, got %T", value)
		}
	case TypeInt64:
		switch value.(type) {
		case int64, int, int32:
			// OK
		default:
			return fmt.Errorf("expected int64, got %T", value)
		}
	case TypeFloat64:
		switch value.(type) {
		case float32, float64, int, int32, int64:
			// OK - allow int types to be coerced to float
		default:
			return fmt.Errorf("expected float64, got %T", value)
		}
	case TypeBool:
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("expected bool, got %T", value)
		}
	case TypeTime:
		switch value.(type) {
		case time.Time, primitive.DateTime:
			// OK
		default:
			return fmt.Errorf("expected time.Time, got %T", value)
		}
	case TypeObjectID:
		switch value.(type) {
		case primitive.ObjectID, string:
			// OK - allow string for ObjectID
		default:
			return fmt.Errorf("expected ObjectID, got %T", value)
		}
	case TypeArray:
		rv := reflect.ValueOf(value)
		if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
			return fmt.Errorf("expected array, got %T", value)
		}
	case TypeObject:
		rv := reflect.ValueOf(value)
		if rv.Kind() != reflect.Map && rv.Kind() != reflect.Struct && rv.Kind() != reflect.Ptr {
			return fmt.Errorf("expected object, got %T", value)
		}
	case TypeAny:
		// Accept anything
	}

	return nil
}

// ValidationError represents a schema validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

// IsValidationError checks if an error is a validation error
func IsValidationError(err error) bool {
	_, ok := err.(*ValidationError)
	return ok
}

// toMap converts a struct or map to map[string]interface{}
func toMap(doc interface{}) (map[string]interface{}, error) {
	if doc == nil {
		return nil, fmt.Errorf("document is nil")
	}

	// If already a map
	if m, ok := doc.(map[string]interface{}); ok {
		return m, nil
	}
	if m, ok := doc.(M); ok {
		return map[string]interface{}(m), nil
	}

	// Use reflection for structs
	v := reflect.ValueOf(doc)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected struct or map, got %T", doc)
	}

	result := make(map[string]interface{})
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)

		// Get bson tag or use field name
		tag := field.Tag.Get("bson")
		if tag == "-" {
			continue
		}

		name := tag
		if name == "" || name == ",inline" {
			name = field.Name
		} else {
			// Handle tags like "fieldname,omitempty"
			if idx := strings.Index(name, ","); idx != -1 {
				name = name[:idx]
			}
		}

		// Handle inline embedded structs
		if tag == ",inline" || (field.Anonymous && tag == "") {
			embeddedMap, err := toMap(v.Field(i).Interface())
			if err == nil {
				for k, val := range embeddedMap {
					result[k] = val
				}
			}
			continue
		}

		fieldValue := v.Field(i).Interface()

		// Skip zero values with omitempty
		if strings.Contains(tag, "omitempty") && isZeroValue(fieldValue) {
			continue
		}

		result[name] = fieldValue
	}

	return result, nil
}

// isZeroValue checks if a value is the zero value for its type
func isZeroValue(v interface{}) bool {
	if v == nil {
		return true
	}

	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Ptr, reflect.Interface:
		return rv.IsNil()
	case reflect.Slice, reflect.Map:
		return rv.IsNil() || rv.Len() == 0
	default:
		return rv.IsZero()
	}
}

// toFloat64 converts numeric types to float64
func toFloat64(v interface{}) (float64, error) {
	switch val := v.(type) {
	case int:
		return float64(val), nil
	case int8:
		return float64(val), nil
	case int16:
		return float64(val), nil
	case int32:
		return float64(val), nil
	case int64:
		return float64(val), nil
	case uint:
		return float64(val), nil
	case uint8:
		return float64(val), nil
	case uint16:
		return float64(val), nil
	case uint32:
		return float64(val), nil
	case uint64:
		return float64(val), nil
	case float32:
		return float64(val), nil
	case float64:
		return val, nil
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", v)
	}
}

// setField sets a field value on a struct using reflection
func setField(doc interface{}, fieldName string, value interface{}) error {
	v := reflect.ValueOf(doc)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return fmt.Errorf("document is not a struct")
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("bson")
		name := tag
		if name == "" {
			name = field.Name
		} else if idx := strings.Index(name, ","); idx != -1 {
			name = name[:idx]
		}

		if name == fieldName {
			fieldVal := v.Field(i)
			if fieldVal.CanSet() {
				fieldVal.Set(reflect.ValueOf(value))
				return nil
			}
		}
	}

	return fmt.Errorf("field %s not found", fieldName)
}

// Timestamped interface - implement this to enable auto-timestamps
type Timestamped interface {
	SetCreatedAt(time.Time)
	SetUpdatedAt(time.Time)
	GetCreatedAt() time.Time
	GetUpdatedAt() time.Time
}

// IDSetter interface - implement this to allow automatic ID setting
type IDSetter interface {
	SetObjectID(primitive.ObjectID)
	GetObjectID() primitive.ObjectID
}

// BaseModel provides common fields for all documents (like Mongoose)
// Use this for models that need auto-timestamps
type BaseModel struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	CreatedAt time.Time          `bson:"createdAt,omitempty" json:"createdAt,omitempty"`
	UpdatedAt time.Time          `bson:"updatedAt,omitempty" json:"updatedAt,omitempty"`
}

// SetCreatedAt sets the created timestamp
func (m *BaseModel) SetCreatedAt(t time.Time) {
	m.CreatedAt = t
}

// SetUpdatedAt sets the updated timestamp
func (m *BaseModel) SetUpdatedAt(t time.Time) {
	m.UpdatedAt = t
}

// GetCreatedAt returns the created timestamp
func (m *BaseModel) GetCreatedAt() time.Time {
	return m.CreatedAt
}

// GetUpdatedAt returns the updated timestamp
func (m *BaseModel) GetUpdatedAt() time.Time {
	return m.UpdatedAt
}

// GetID returns the document ID as a string
func (m *BaseModel) GetID() string {
	return m.ID.Hex()
}

// SetID sets the document ID from a string
func (m *BaseModel) SetID(id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	m.ID = objID
	return nil
}

// SetObjectID sets the document ID
func (m *BaseModel) SetObjectID(id primitive.ObjectID) {
	m.ID = id
}

// GetObjectID returns the document ID
func (m *BaseModel) GetObjectID() primitive.ObjectID {
	return m.ID
}

// SimpleModel provides just an ID field without timestamps
// Use this for models that don't need auto-timestamps
type SimpleModel struct {
	ID primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
}

// GetID returns the document ID as a string
func (m *SimpleModel) GetID() string {
	return m.ID.Hex()
}

// SetID sets the document ID from a string
func (m *SimpleModel) SetID(id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	m.ID = objID
	return nil
}

// SetObjectID sets the document ID
func (m *SimpleModel) SetObjectID(id primitive.ObjectID) {
	m.ID = id
}

// GetObjectID returns the document ID
func (m *SimpleModel) GetObjectID() primitive.ObjectID {
	return m.ID
}

// applyTimestamps applies timestamps to a document if it implements Timestamped
func applyTimestamps(doc interface{}, isInsert bool) {
	if timestamped, ok := doc.(Timestamped); ok {
		now := time.Now()
		if isInsert {
			// On insert: set both createdAt and updatedAt
			if timestamped.GetCreatedAt().IsZero() {
				timestamped.SetCreatedAt(now)
			}
			if timestamped.GetUpdatedAt().IsZero() {
				timestamped.SetUpdatedAt(now)
			}
		} else {
			// On update: only set updatedAt
			timestamped.SetUpdatedAt(now)
		}
	}
}

// applyObjectID ensures the document has an ObjectID set
func applyObjectID(doc interface{}) {
	if setter, ok := doc.(IDSetter); ok {
		if setter.GetObjectID().IsZero() {
			setter.SetObjectID(primitive.NewObjectID())
		}
	}
}

// IndexDefinition defines an index for a collection
type IndexDefinition struct {
	Keys   M
	Unique bool
	Name   string
}
