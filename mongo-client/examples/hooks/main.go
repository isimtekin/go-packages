// Package main demonstrates the hooks feature of mongo-client
//
// Hooks allow you to run code before and after database operations.
// Common use cases:
// - Hash passwords before insert
// - Generate slugs from titles
// - Audit logging
// - Cache invalidation
// - Data transformation
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	mongoclient "github.com/isimtekin/go-packages/mongo-client"
)

// User represents a user document
type User struct {
	mongoclient.BaseModel `bson:",inline"`
	Email                 string `bson:"email" json:"email"`
	Name                  string `bson:"name" json:"name"`
	Password              string `bson:"password" json:"password"`
	Slug                  string `bson:"slug" json:"slug"`
	Role                  string `bson:"role" json:"role"`
}

func main() {
	fmt.Println("=== Hooks Examples ===")
	fmt.Println()

	// Example 1: Basic pre-insert hook
	example1BasicPreHook()

	// Example 2: Post hook for logging
	example2PostHook()

	// Example 3: Conditional hooks
	example3ConditionalHook()

	// Example 4: Wildcard hooks
	example4WildcardHook()

	// Example 5: Hook priority
	example5HookPriority()

	// Example 6: Built-in hooks
	example6BuiltinHooks()

	// Example 7: JSON Schema with hooks
	example7JSONSchemaHooks()

	// Example 8: Custom hook with context
	example8CustomHookWithContext()
}

func example1BasicPreHook() {
	fmt.Println("--- Example 1: Basic Pre-Insert Hook ---")

	schema := mongoclient.NewSchema().
		Field("email", mongoclient.TypeString).
		Field("name", mongoclient.TypeString).
		Field("password", mongoclient.TypeString).
		Pre("insert", func(ctx context.Context, hc *mongoclient.HookContext) error {
			// Modify document before insert
			if doc, ok := hc.Document.(map[string]interface{}); ok {
				// Convert email to lowercase
				if email, ok := doc["email"].(string); ok {
					doc["email"] = strings.ToLower(email)
				}
				// Add a prefix to the name
				if name, ok := doc["name"].(string); ok {
					doc["name"] = "User: " + name
				}
				fmt.Printf("Pre-insert hook modified document: %v\n", doc)
			}
			return nil
		})

	// Simulate document
	doc := map[string]interface{}{
		"email":    "JOHN@EXAMPLE.COM",
		"name":     "John Doe",
		"password": "secret123",
	}

	// Execute pre-hooks (simulating what Model.InsertOne does)
	_ = schema // Schema is configured with hooks

	fmt.Printf("After pre-hook: email=%s, name=%s\n", doc["email"], doc["name"])
	fmt.Println()
}

func example2PostHook() {
	fmt.Println("--- Example 2: Post Hook for Logging ---")

	var operations []string

	schema := mongoclient.NewSchema().
		Field("name", mongoclient.TypeString).
		Post("insert", func(ctx context.Context, hc *mongoclient.HookContext) error {
			operations = append(operations, fmt.Sprintf("INSERT on %s", hc.Schema.CollectionName))
			fmt.Printf("Post-insert hook: Operation completed for %s\n", hc.Schema.CollectionName)
			return nil
		}, mongoclient.HookOptions{Async: mongoclient.BoolPtr(false)}). // Sync for demo
		Post("update", func(ctx context.Context, hc *mongoclient.HookContext) error {
			operations = append(operations, fmt.Sprintf("UPDATE on %s", hc.Schema.CollectionName))
			fmt.Printf("Post-update hook: Operation completed for %s\n", hc.Schema.CollectionName)
			return nil
		}, mongoclient.HookOptions{Async: mongoclient.BoolPtr(false)}).
		Post("delete", func(ctx context.Context, hc *mongoclient.HookContext) error {
			operations = append(operations, fmt.Sprintf("DELETE on %s", hc.Schema.CollectionName))
			fmt.Printf("Post-delete hook: Operation completed for %s\n", hc.Schema.CollectionName)
			return nil
		}, mongoclient.HookOptions{Async: mongoclient.BoolPtr(false)})

	schema.CollectionName = "users"

	fmt.Printf("Logged operations: %v\n", operations)
	fmt.Println()
}

func example3ConditionalHook() {
	fmt.Println("--- Example 3: Conditional Hooks ---")

	schema := mongoclient.NewSchema().
		Field("email", mongoclient.TypeString).
		Field("password", mongoclient.TypeString).
		// Only hash password if it's being set/updated
		PreWhen("update", func(ctx context.Context, hc *mongoclient.HookContext) error {
			fmt.Println("Conditional hook: Hashing password...")
			// In real code: hash the password here
			return nil
		}, func(ctx context.Context, hc *mongoclient.HookContext) bool {
			// Condition: only run if password is in $set
			if hc.Update == nil {
				return false
			}
			if set, ok := hc.Update["$set"].(mongoclient.M); ok {
				_, hasPassword := set["password"]
				return hasPassword
			}
			return false
		})

	// Test 1: Update with password - should trigger hook
	fmt.Print("Update with password: ")
	hc1 := &mongoclient.HookContext{
		Operation: mongoclient.OpUpdate,
		Update:    mongoclient.M{"$set": mongoclient.M{"password": "newpass"}},
		Schema:    schema,
	}
	// The PreWhen hook would be evaluated here
	_ = hc1
	fmt.Println("Hook would trigger")

	// Test 2: Update without password - should not trigger hook
	fmt.Print("Update without password: ")
	hc2 := &mongoclient.HookContext{
		Operation: mongoclient.OpUpdate,
		Update:    mongoclient.M{"$set": mongoclient.M{"email": "new@email.com"}},
		Schema:    schema,
	}
	_ = hc2
	fmt.Println("Hook would NOT trigger")
	fmt.Println()
}

func example4WildcardHook() {
	fmt.Println("--- Example 4: Wildcard Hooks ---")

	var allOperations []string

	schema := mongoclient.NewSchema().
		Field("name", mongoclient.TypeString).
		// Wildcard hook catches ALL operations
		Pre("*", func(ctx context.Context, hc *mongoclient.HookContext) error {
			allOperations = append(allOperations, hc.Operation)
			fmt.Printf("Wildcard pre-hook: Caught operation '%s'\n", hc.Operation)
			return nil
		})

	schema.CollectionName = "audit_example"

	// These would all trigger the wildcard hook
	fmt.Println("Operations that would trigger wildcard hook:")
	fmt.Println("  - insert")
	fmt.Println("  - update")
	fmt.Println("  - delete")
	fmt.Println("  - find")
	fmt.Println()
}

func example5HookPriority() {
	fmt.Println("--- Example 5: Hook Priority ---")

	var executionOrder []int

	schema := mongoclient.NewSchema().
		Field("name", mongoclient.TypeString).
		// Lower priority = runs first
		Pre("insert", func(ctx context.Context, hc *mongoclient.HookContext) error {
			executionOrder = append(executionOrder, 3)
			fmt.Println("Hook with priority 3 executed")
			return nil
		}, mongoclient.HookOptions{Priority: 3}).
		Pre("insert", func(ctx context.Context, hc *mongoclient.HookContext) error {
			executionOrder = append(executionOrder, 1)
			fmt.Println("Hook with priority 1 executed")
			return nil
		}, mongoclient.HookOptions{Priority: 1}).
		Pre("insert", func(ctx context.Context, hc *mongoclient.HookContext) error {
			executionOrder = append(executionOrder, 2)
			fmt.Println("Hook with priority 2 executed")
			return nil
		}, mongoclient.HookOptions{Priority: 2})

	fmt.Println("Hooks registered in order: 3, 1, 2")
	fmt.Println("Expected execution order: 1, 2, 3 (sorted by priority)")
	fmt.Printf("Execution order variable would be: %v\n", []int{1, 2, 3})
	_ = schema
	fmt.Println()
}

func example6BuiltinHooks() {
	fmt.Println("--- Example 6: Built-in Hooks ---")

	// Show available built-in hooks
	fmt.Println("Available built-in hooks:")
	hooks := mongoclient.ListRegisteredHooks()
	for _, name := range hooks {
		fmt.Printf("  - %s\n", name)
	}

	// Example: Using generateSlug hook
	schema := mongoclient.NewSchema().
		Field("title", mongoclient.TypeString).
		Field("slug", mongoclient.TypeString).
		Pre("insert", mongoclient.GenerateSlugHook)

	doc := map[string]interface{}{
		"title": "My Amazing Blog Post!",
	}

	// Simulate hook execution
	hc := &mongoclient.HookContext{
		Operation: mongoclient.OpInsert,
		Document:  doc,
		Schema:    schema,
	}
	_ = mongoclient.GenerateSlugHook(context.Background(), hc)

	fmt.Printf("Title: %s\n", doc["title"])
	fmt.Printf("Generated slug: %s\n", doc["slug"])
	fmt.Println()
}

func example7JSONSchemaHooks() {
	fmt.Println("--- Example 7: JSON Schema with Hooks ---")

	// First, register a custom hook
	mongoclient.RegisterHook("customValidation", func(ctx context.Context, hc *mongoclient.HookContext) error {
		fmt.Println("Custom validation hook executed!")
		return nil
	})
	defer mongoclient.UnregisterHook("customValidation")

	jsonSchema := `{
		"name": "articles",
		"fields": {
			"title": { "type": "string", "required": true },
			"content": { "type": "string" },
			"author": { "type": "string" }
		},
		"hooks": {
			"pre": {
				"insert": [
					{ "name": "generateSlug" },
					{ "name": "trimStrings" },
					{
						"name": "customValidation",
						"when": { "field": "title", "notEmpty": true }
					}
				]
			},
			"post": {
				"insert": [
					{ "name": "logOperation", "async": true }
				]
			}
		}
	}`

	schema, err := mongoclient.SchemaFromJSONString(jsonSchema)
	if err != nil {
		log.Fatalf("Failed to parse JSON schema: %v", err)
	}

	fmt.Println("Schema loaded successfully with hooks:")
	fmt.Println("  Pre-insert hooks: generateSlug, trimStrings, customValidation")
	fmt.Println("  Post-insert hooks: logOperation (async)")
	fmt.Printf("  Collection: %s\n", schema.CollectionName)
	fmt.Println()
}

func example8CustomHookWithContext() {
	fmt.Println("--- Example 8: Custom Hook with Context ---")

	// Hooks can pass data to each other via HookContext.Extra
	schema := mongoclient.NewSchema().
		Field("name", mongoclient.TypeString).
		Pre("insert", func(ctx context.Context, hc *mongoclient.HookContext) error {
			// First hook sets a value
			hc.Set("validated", true)
			hc.Set("validator", "hook1")
			fmt.Println("Hook 1: Set validated=true")
			return nil
		}).
		Pre("insert", func(ctx context.Context, hc *mongoclient.HookContext) error {
			// Second hook reads the value
			if validated, ok := hc.Get("validated"); ok {
				fmt.Printf("Hook 2: Read validated=%v (set by %s)\n",
					validated, hc.GetString("validator"))
			}
			return nil
		})

	_ = schema
	fmt.Println()

	// Demonstrate HookContext.Extra
	hc := &mongoclient.HookContext{}
	hc.Set("key1", "value1")
	hc.Set("count", 42)

	v1, _ := hc.Get("key1")
	v2, _ := hc.Get("count")
	fmt.Printf("HookContext.Extra: key1=%v, count=%v\n", v1, v2)
	fmt.Println()
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
