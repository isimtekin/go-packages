package mongoclient

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Timestamped interface - implement this to enable auto-timestamps
type Timestamped interface {
	SetCreatedAt(time.Time)
	SetUpdatedAt(time.Time)
	GetCreatedAt() time.Time
	GetUpdatedAt() time.Time
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

// Schema validation helpers

// ValidationRule represents a validation rule for a field
type ValidationRule struct {
	Field    string
	Required bool
	MinLen   int
	MaxLen   int
	Min      interface{}
	Max      interface{}
	Enum     []interface{}
	Match    string // regex pattern
}

// SchemaDefinition defines the structure and validation rules for a collection
type SchemaDefinition struct {
	Collection string
	Indexes    []IndexDefinition
	Rules      []ValidationRule
}

// IndexDefinition defines an index for a collection
type IndexDefinition struct {
	Keys   M
	Unique bool
	Name   string
}

// Example usage patterns:
//
// type User struct {
//     BaseModel `bson:",inline"`
//     Email     string `bson:"email" json:"email"`
//     Name      string `bson:"name" json:"name"`
//     Age       int    `bson:"age" json:"age"`
// }
//
// func (u *User) BeforeInsert() {
//     u.BaseModel.BeforeInsert()
//     // Additional user-specific logic
// }
