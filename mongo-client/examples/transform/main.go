// Package main demonstrates the transform feature of mongo-client
//
// Transform allows you to transform documents when converting them to JSON or objects.
// Common use cases:
// - Rename _id to id for API responses
// - Omit sensitive fields like passwords
// - Convert camelCase to snake_case
// - Pick only specific fields
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	mongoclient "github.com/isimtekin/go-packages/mongo-client"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User represents a user document
type User struct {
	mongoclient.BaseModel `bson:",inline"`
	Email                 string   `bson:"email" json:"email"`
	Name                  string   `bson:"name" json:"name"`
	Password              string   `bson:"password" json:"password"`
	APIKey                string   `bson:"apiKey" json:"apiKey"`
	Age                   int      `bson:"age" json:"age"`
	Role                  string   `bson:"role" json:"role"`
	Tags                  []string `bson:"tags" json:"tags"`
	Bio                   string   `bson:"bio" json:"bio"`
}

func main() {
	fmt.Println("=== Transform Examples ===")
	fmt.Println()

	// Create a sample user
	user := createSampleUser()

	// Example 1: Basic transform - Rename _id to id
	example1BasicTransform(user)

	// Example 2: Omit sensitive fields
	example2OmitSensitive(user)

	// Example 3: Snake case transform
	example3SnakeCase(user)

	// Example 4: Pick specific fields
	example4PickFields(user)

	// Example 5: Combined transforms
	example5Combined(user)

	// Example 6: Custom transform function
	example6CustomTransform(user)

	// Example 7: Transform from JSON schema
	example7JSONSchema(user)

	// Example 8: Transform multiple documents
	example8TransformMany()

	// Example 9: Preset transforms
	example9Presets(user)
}

func createSampleUser() *User {
	user := &User{
		Email:    "john.doe@example.com",
		Name:     "John Doe",
		Password: "hashed_password_12345",
		APIKey:   "sk-abc123xyz789",
		Age:      30,
		Role:     "admin",
		Tags:     []string{"developer", "golang"},
		Bio:      "",
	}
	user.ID = primitive.NewObjectID()
	user.CreatedAt = time.Now().Add(-24 * time.Hour)
	user.UpdatedAt = time.Now()
	return user
}

func example1BasicTransform(user *User) {
	fmt.Println("--- Example 1: Basic Transform (Rename _id to id) ---")

	schema := mongoclient.NewSchema().
		Field("email", mongoclient.TypeString).
		Field("name", mongoclient.TypeString).
		WithTransform(mongoclient.TransformOptions{
			RenameID: true,
		})

	result := schema.Transform(user)
	printJSON("Transformed", result)
	fmt.Println()
}

func example2OmitSensitive(user *User) {
	fmt.Println("--- Example 2: Omit Sensitive Fields ---")

	schema := mongoclient.NewSchema().
		Field("email", mongoclient.TypeString).
		Field("name", mongoclient.TypeString).
		Field("password", mongoclient.TypeString).
		Field("apiKey", mongoclient.TypeString).
		WithTransform(mongoclient.TransformOptions{
			RenameID: true,
			Omit:     []string{"password", "apiKey"},
		})

	result := schema.Transform(user)
	printJSON("Without sensitive fields", result)
	fmt.Println()
}

func example3SnakeCase(user *User) {
	fmt.Println("--- Example 3: Snake Case Transform ---")

	schema := mongoclient.NewSchema().
		Field("email", mongoclient.TypeString).
		Field("name", mongoclient.TypeString).
		WithTransform(mongoclient.TransformOptions{
			RenameID: true,
			Rename: map[string]string{
				"createdAt": "created_at",
				"updatedAt": "updated_at",
				"apiKey":    "api_key",
			},
			Omit: []string{"password"},
		})

	result := schema.Transform(user)
	printJSON("Snake case", result)
	fmt.Println()
}

func example4PickFields(user *User) {
	fmt.Println("--- Example 4: Pick Specific Fields ---")

	schema := mongoclient.NewSchema().
		Field("email", mongoclient.TypeString).
		Field("name", mongoclient.TypeString).
		WithTransform(mongoclient.TransformOptions{
			RenameID: true,
			Pick:     []string{"id", "email", "name", "role"},
		})

	result := schema.Transform(user)
	printJSON("Picked fields only", result)
	fmt.Println()
}

func example5Combined(user *User) {
	fmt.Println("--- Example 5: Combined Transforms ---")

	schema := mongoclient.NewSchema().
		Field("email", mongoclient.TypeString).
		Field("name", mongoclient.TypeString).
		WithTransform(mongoclient.TransformOptions{
			RenameID: true,
			Rename: map[string]string{
				"createdAt": "created_at",
				"updatedAt": "updated_at",
			},
			Omit:      []string{"password", "apiKey"},
			OmitEmpty: true, // Remove empty bio
		})

	result := schema.Transform(user)
	printJSON("Combined transforms", result)
	fmt.Println()
}

func example6CustomTransform(user *User) {
	fmt.Println("--- Example 6: Custom Transform Function ---")

	schema := mongoclient.NewSchema().
		Field("email", mongoclient.TypeString).
		Field("name", mongoclient.TypeString).
		WithTransform(mongoclient.TransformOptions{
			RenameID: true,
			Omit:     []string{"password", "apiKey"},
			Custom: func(doc map[string]interface{}) map[string]interface{} {
				// Add computed fields
				if name, ok := doc["name"].(string); ok {
					doc["displayName"] = fmt.Sprintf("@%s", name)
				}

				// Add metadata
				doc["_meta"] = map[string]interface{}{
					"version":   "v1",
					"generated": time.Now().Format(time.RFC3339),
				}

				return doc
			},
		})

	result := schema.Transform(user)
	printJSON("With custom transform", result)
	fmt.Println()
}

func example7JSONSchema(user *User) {
	fmt.Println("--- Example 7: Transform from JSON Schema ---")

	jsonSchema := `{
		"name": "users",
		"fields": {
			"email": { "type": "string", "required": true },
			"name": { "type": "string", "required": true },
			"password": { "type": "string" },
			"age": { "type": "int" },
			"role": { "type": "string" }
		},
		"transform": {
			"renameId": true,
			"omit": ["password", "apiKey"],
			"rename": {
				"createdAt": "created_at",
				"updatedAt": "updated_at"
			},
			"omitEmpty": true
		}
	}`

	schema, err := mongoclient.SchemaFromJSONString(jsonSchema)
	if err != nil {
		log.Fatalf("Failed to parse JSON schema: %v", err)
	}

	result := schema.Transform(user)
	printJSON("From JSON schema", result)
	fmt.Println()
}

func example8TransformMany() {
	fmt.Println("--- Example 8: Transform Multiple Documents ---")

	users := []User{
		{
			BaseModel: mongoclient.BaseModel{ID: primitive.NewObjectID()},
			Email:     "john@example.com",
			Name:      "John",
			Password:  "secret1",
			Role:      "admin",
		},
		{
			BaseModel: mongoclient.BaseModel{ID: primitive.NewObjectID()},
			Email:     "jane@example.com",
			Name:      "Jane",
			Password:  "secret2",
			Role:      "user",
		},
		{
			BaseModel: mongoclient.BaseModel{ID: primitive.NewObjectID()},
			Email:     "bob@example.com",
			Name:      "Bob",
			Password:  "secret3",
			Role:      "user",
		},
	}

	schema := mongoclient.NewSchema().
		Field("email", mongoclient.TypeString).
		Field("name", mongoclient.TypeString).
		WithTransform(mongoclient.TransformOptions{
			RenameID: true,
			Omit:     []string{"password"},
		})

	// Transform all at once
	results := schema.TransformMany(users)
	printJSON("Multiple users", results)

	// Or get as JSON directly
	jsonBytes, _ := schema.ToJSONMany(users)
	fmt.Println("As JSON array:", string(jsonBytes))
	fmt.Println()
}

func example9Presets(user *User) {
	fmt.Println("--- Example 9: Using Transform Presets ---")

	// Preset 1: Default API Transform
	schema1 := mongoclient.NewSchema().
		Field("email", mongoclient.TypeString).
		Field("name", mongoclient.TypeString).
		WithTransform(mongoclient.DefaultAPITransform())

	result1 := schema1.Transform(user)
	fmt.Println("DefaultAPITransform:")
	printJSON("", result1)

	// Preset 2: Secure Transform (omits sensitive fields)
	schema2 := mongoclient.NewSchema().
		Field("email", mongoclient.TypeString).
		Field("name", mongoclient.TypeString).
		WithTransform(mongoclient.SecureTransform("password", "apiKey", "secretToken"))

	result2 := schema2.Transform(user)
	fmt.Println("SecureTransform:")
	printJSON("", result2)

	// Preset 3: Snake Case Transform
	schema3 := mongoclient.NewSchema().
		Field("email", mongoclient.TypeString).
		Field("name", mongoclient.TypeString).
		WithTransform(mongoclient.SnakeCaseTransform())

	result3 := schema3.Transform(user)
	fmt.Println("SnakeCaseTransform:")
	printJSON("", result3)
}

func printJSON(label string, v interface{}) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		log.Printf("Failed to marshal JSON: %v", err)
		return
	}
	if label != "" {
		fmt.Printf("%s:\n", label)
	}
	fmt.Println(string(data))
}
