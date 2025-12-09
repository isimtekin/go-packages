## mongo-client

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/license-MIT-blue)](LICENSE)

A minimalist, high-level MongoDB client wrapper for Go with convenient methods, transaction support, and Mongoose-like schema patterns.

## ✨ Features

- 🚀 **High-level CRUD operations** - Simple methods for common database operations
- 🔄 **Transaction support** - Easy-to-use transaction helpers with automatic rollback
- ⚙️ **Connection pooling** - Built-in connection pool management and health checks
- ⏱️ **Context management** - Automatic timeout handling for all operations
- 📊 **Aggregation helpers** - Simplified aggregation pipeline building
- 🎯 **Auto-timestamps** - Automatic `createdAt`/`updatedAt` management
- 🔐 **Schema Validation** - Mongoose-like schema with required fields, type checking, and custom validators
- 📝 **Model Pattern** - Schema-bound models that enforce validation on all operations
- 📋 **JSON Schema Support** - Define schemas in JSON format and load from files
- 🔄 **Transform Support** - Convert documents for API responses (rename _id to id, omit fields, snake_case)
- 🔧 **Functional options** - Clean configuration with functional options pattern
- 🛠️ **Query builders** - Helper functions for building MongoDB queries
- 📋 **Pagination support** - Built-in pagination helpers
- ✅ **Type-safe** - Full TypeScript-like experience with Go
- 🧪 **Well-tested** - Comprehensive test coverage

## 🚀 Quick Start

```bash
# Installation
go get github.com/isimtekin/go-packages/mongo-client@v0.1.0
```

## 📋 Basic Usage

### Creating a Client

**Option 1: From Environment Variables (Recommended for Production)**

```go
package main

import (
    "context"
    "log"

    mongoclient "github.com/isimtekin/go-packages/mongo-client"
    envutil "github.com/isimtekin/go-packages/env-util"
)

func main() {
    ctx := context.Background()

    // Load .env file (optional, for local development)
    envutil.LoadEnvFile(".env")

    // Create client from environment variables
    // Reads MONGO_URI, MONGO_DATABASE, etc.
    client, err := mongoclient.NewFromEnvWithDefaults(ctx)
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close(ctx)
}
```

**Environment Variables:**
```bash
# Required
MONGO_URI=mongodb://localhost:27017
MONGO_DATABASE=myapp

# Optional
MONGO_MAX_POOL_SIZE=100
MONGO_CONNECT_TIMEOUT=10s
```

See [.env.example](./.env.example) for complete configuration options.

**Option 2: Using Functional Options**

```go
client, err := mongoclient.NewWithOptions(ctx,
    mongoclient.WithURI("mongodb://localhost:27017"),
    mongoclient.WithDatabase("myapp"),
    mongoclient.WithMaxPoolSize(100),
    mongoclient.WithConnectTimeout(10*time.Second),
)
```

**Option 3: Using Config Struct**

```go
config := &mongoclient.Config{
    URI:      "mongodb://localhost:27017",
    Database: "myapp",
}
client, err := mongoclient.New(ctx, config)
```

---

## 🔐 Schema & Model (Mongoose-like)

mongo-client provides a Mongoose-like schema validation system. **All operations require a schema** - this ensures data integrity at the application level.

### Defining a Schema

```go
// Define a user schema
userSchema := mongoclient.NewSchema().
    Field("email", mongoclient.TypeString,
        mongoclient.Required(),
        mongoclient.MatchRegex(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)).
    Field("name", mongoclient.TypeString,
        mongoclient.Required(),
        mongoclient.MinLen(2),
        mongoclient.MaxLen(100)).
    Field("age", mongoclient.TypeInt,
        mongoclient.MinValue(0),
        mongoclient.MaxValue(150)).
    Field("role", mongoclient.TypeString,
        mongoclient.Enum("admin", "user", "guest"),
        mongoclient.Default("user")).
    Field("tags", mongoclient.TypeArray,
        mongoclient.ArrayOf(mongoclient.TypeString)).
    Field("isActive", mongoclient.TypeBool,
        mongoclient.Default(true)).
    AddIndex(mongoclient.M{"email": 1}, true, "email_unique").
    WithTimestamps(true) // Auto createdAt/updatedAt
```

### Creating a Model

```go
// Create model (schema is REQUIRED)
userModel, err := mongoclient.NewModel(client, "users", userSchema)
if err != nil {
    log.Fatal(err) // Returns ErrSchemaRequired if schema is nil
}

// Create indexes defined in schema
err = userModel.EnsureIndexes(ctx)
```

### CRUD with Schema Validation

```go
// Define your struct
type User struct {
    mongoclient.BaseModel `bson:",inline"` // Includes _id, createdAt, updatedAt
    Email    string `bson:"email" json:"email"`
    Name     string `bson:"name" json:"name"`
    Age      int    `bson:"age" json:"age"`
    Role     string `bson:"role" json:"role"`
    Tags     []string `bson:"tags" json:"tags"`
    IsActive bool   `bson:"isActive" json:"isActive"`
}

// INSERT - validates before insert, auto-generates _id and timestamps
user := &User{
    Email: "john@example.com",
    Name:  "John Doe",
    Age:   30,
}
result, err := userModel.InsertOne(ctx, user)
// Error if validation fails: "validation failed: field 'email' is required"

// After insert, user.ID is automatically set to a new ObjectID
fmt.Println(user.GetID()) // "507f1f77bcf86cd799439011"

// UPDATE - validates update data against schema
_, err = userModel.UpdateOneByID(ctx, user.ID, mongoclient.M{
    "$set": mongoclient.M{"age": 31},
})
// Error if field not in schema: "validation failed: field 'unknown' is not defined in schema"

// DELETE - protected against empty filter
_, err = userModel.DeleteOne(ctx, mongoclient.M{})
// Error: "filter cannot be empty"

// FIND - no validation needed
var foundUser User
err = userModel.FindOneByID(ctx, user.ID).Decode(&foundUser)
```

### Schema Field Types

| Type | Go Types | Description |
|------|----------|-------------|
| `TypeString` | `string` | Text values |
| `TypeInt` | `int`, `int32`, `int64` | Integer values |
| `TypeInt64` | `int64` | 64-bit integers |
| `TypeFloat64` | `float32`, `float64` | Floating point numbers |
| `TypeBool` | `bool` | Boolean values |
| `TypeTime` | `time.Time`, `primitive.DateTime` | Date/time values |
| `TypeObjectID` | `primitive.ObjectID`, `string` | MongoDB ObjectIDs |
| `TypeArray` | `[]T` | Array/slice values |
| `TypeObject` | `struct`, `map` | Nested objects |
| `TypeAny` | `interface{}` | Any type |

### Schema Field Options

```go
// Required field
mongoclient.Required()

// Default value
mongoclient.Default("user")

// String length constraints
mongoclient.MinLen(2)
mongoclient.MaxLen(100)

// Numeric range constraints
mongoclient.MinValue(0)
mongoclient.MaxValue(150)

// Enum (allowed values)
mongoclient.Enum("admin", "user", "guest")

// Regex pattern
mongoclient.MatchRegex(`^[a-z]+@[a-z]+\.[a-z]+$`)

// Reference to another collection
mongoclient.Ref("users")

// Array element type
mongoclient.ArrayOf(mongoclient.TypeString)

// Nested schema
mongoclient.NestedSchema(addressSchema)

// Custom validator
mongoclient.Validator(func(value interface{}) error {
    email := value.(string)
    if strings.Contains(email, "spam") {
        return errors.New("spam emails not allowed")
    }
    return nil
})
```

### Nested Objects

```go
// Define nested address schema
addressSchema := mongoclient.NewSchema().
    Field("street", mongoclient.TypeString, mongoclient.Required()).
    Field("city", mongoclient.TypeString, mongoclient.Required()).
    Field("zipCode", mongoclient.TypeString, mongoclient.MinLen(5))

// Use in parent schema
userSchema := mongoclient.NewSchema().
    Field("name", mongoclient.TypeString, mongoclient.Required()).
    Field("address", mongoclient.TypeObject,
        mongoclient.Required(),
        mongoclient.NestedSchema(addressSchema))
```

### Strict Mode

By default, schemas are in **strict mode** - fields not defined in schema are rejected:

```go
// Strict mode (default: true)
schema := mongoclient.NewSchema().
    Field("name", mongoclient.TypeString)

// This will fail - "unknown" field not in schema
doc := map[string]interface{}{
    "name": "test",
    "unknown": "field", // ERROR!
}
err := schema.Validate(doc) // "field 'unknown' is not defined in schema"

// Disable strict mode to allow extra fields
schema.WithStrict(false)
```

---

## 📋 JSON Schema Support

Define schemas in JSON format for easy configuration and sharing:

### JSON Schema Format

```json
{
  "name": "users",
  "timestamps": true,
  "strict": true,
  "fields": {
    "email": {
      "type": "string",
      "required": true,
      "pattern": "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"
    },
    "name": {
      "type": "string",
      "required": true,
      "minLength": 2,
      "maxLength": 100
    },
    "age": {
      "type": "int",
      "min": 0,
      "max": 150
    },
    "role": {
      "type": "string",
      "enum": ["admin", "user", "guest"],
      "default": "user"
    },
    "tags": {
      "type": "array",
      "items": "string"
    },
    "address": {
      "type": "object",
      "fields": {
        "street": { "type": "string", "required": true },
        "city": { "type": "string", "required": true },
        "zipCode": { "type": "string", "minLength": 5 }
      }
    },
    "managerId": {
      "type": "objectid",
      "ref": "users"
    }
  },
  "indexes": [
    { "keys": { "email": 1 }, "unique": true, "name": "email_unique" },
    { "keys": { "createdAt": -1 } }
  ]
}
```

### Loading Schema from JSON

```go
// From JSON string
jsonStr := `{
    "name": "users",
    "fields": {
        "email": { "type": "string", "required": true },
        "name": { "type": "string", "required": true }
    }
}`
schema, err := mongoclient.SchemaFromJSONString(jsonStr)

// From JSON file
schema, err := mongoclient.SchemaFromJSONFile("schemas/user.json")

// From JSON bytes
schema, err := mongoclient.SchemaFromJSONBytes(jsonData)

// Create model with loaded schema
userModel, err := mongoclient.NewModel(client, "users", schema)
```

### Converting Schema to JSON

```go
// Define schema programmatically
schema := mongoclient.NewSchema().
    Field("email", mongoclient.TypeString, mongoclient.Required()).
    Field("name", mongoclient.TypeString)
schema.CollectionName = "users"

// Convert to JSON string
jsonStr, err := mongoclient.SchemaToJSONString(schema)

// Convert to JSON bytes
jsonBytes, err := mongoclient.SchemaToJSONBytes(schema)
```

### JSON Field Reference

| JSON Field | Type | Description |
|------------|------|-------------|
| `type` | string | `string`, `int`, `int64`, `float`, `bool`, `time`, `objectid`, `array`, `object`, `any` |
| `required` | bool | Field is required |
| `default` | any | Default value |
| `minLength` | int | Min string length |
| `maxLength` | int | Max string length |
| `min` | number | Min numeric value |
| `max` | number | Max numeric value |
| `enum` | array | Allowed values |
| `pattern` | string | Regex pattern |
| `ref` | string | Reference collection (for ObjectID) |
| `items` | string | Array element type |
| `fields` | object | Nested object fields |

---

## 🔄 Transform (toJSON / toObject)

Transform documents for API responses - rename `_id` to `id`, omit sensitive fields, convert to snake_case, and more. Similar to Mongoose's `toJSON` and `toObject` options.

### Basic Usage

```go
// Define schema with transform options
schema := mongoclient.NewSchema().
    Field("email", mongoclient.TypeString).
    Field("name", mongoclient.TypeString).
    Field("password", mongoclient.TypeString).
    WithTransform(mongoclient.TransformOptions{
        RenameID: true,                           // _id → id
        Omit:     []string{"password", "apiKey"}, // Remove sensitive fields
    })

// Transform a document
user := &User{ID: someObjectID, Email: "john@example.com", Password: "secret"}
result := schema.Transform(user)
// Result: {"id": "507f1f77bcf86cd799439011", "email": "john@example.com", "name": "John"}
// Note: password is omitted, _id is renamed to id

// Get as JSON
jsonBytes, _ := schema.ToJSON(user)
jsonString, _ := schema.ToJSONString(user)

// Transform multiple documents
users := []User{user1, user2, user3}
results := schema.TransformMany(users)
jsonArray, _ := schema.ToJSONMany(users)
```

### Transform Options

```go
type TransformOptions struct {
    RenameID  bool                    // Rename _id to id
    Rename    map[string]string       // Rename fields (e.g., createdAt → created_at)
    Omit      []string                // Fields to exclude from output
    Pick      []string                // Only include these fields (whitelist)
    OmitEmpty bool                    // Remove fields with zero/empty values
    Custom    func(map[string]interface{}) map[string]interface{} // Custom transform function
}
```

### Examples

**Rename Fields (camelCase → snake_case):**

```go
schema := mongoclient.NewSchema().
    Field("email", mongoclient.TypeString).
    Field("firstName", mongoclient.TypeString).
    WithTransform(mongoclient.TransformOptions{
        RenameID: true,
        Rename: map[string]string{
            "firstName": "first_name",
            "lastName":  "last_name",
            "createdAt": "created_at",
            "updatedAt": "updated_at",
        },
    })
```

**Pick Specific Fields:**

```go
schema := mongoclient.NewSchema().
    Field("email", mongoclient.TypeString).
    Field("name", mongoclient.TypeString).
    WithTransform(mongoclient.TransformOptions{
        RenameID: true,
        Pick:     []string{"id", "email", "name"}, // Only these fields
    })
```

**Custom Transform Function:**

```go
schema := mongoclient.NewSchema().
    Field("email", mongoclient.TypeString).
    Field("name", mongoclient.TypeString).
    WithTransform(mongoclient.TransformOptions{
        RenameID: true,
        Custom: func(doc map[string]interface{}) map[string]interface{} {
            // Add computed fields
            if name, ok := doc["name"].(string); ok {
                doc["displayName"] = "@" + name
            }
            // Add metadata
            doc["_version"] = "v1"
            return doc
        },
    })
```

### Preset Transforms

Use built-in presets for common scenarios:

```go
// Default API Transform - renames _id to id, removes empty fields
schema.WithTransform(mongoclient.DefaultAPITransform())

// Secure Transform - omits sensitive fields
schema.WithTransform(mongoclient.SecureTransform("password", "apiKey", "token"))

// Snake Case Transform - converts common fields to snake_case
schema.WithTransform(mongoclient.SnakeCaseTransform())
```

### JSON Schema with Transform

Define transforms in JSON schema files:

```json
{
  "name": "users",
  "fields": {
    "email": { "type": "string", "required": true },
    "name": { "type": "string", "required": true },
    "password": { "type": "string" }
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
}
```

```go
schema, _ := mongoclient.SchemaFromJSONFile("schemas/user.json")
result := schema.Transform(user)
```

### Transform Methods Reference

| Method | Description |
|--------|-------------|
| `schema.Transform(doc)` | Transform single document to `map[string]interface{}` |
| `schema.TransformMany(docs)` | Transform slice of documents |
| `schema.ToJSON(doc)` | Transform and marshal to JSON bytes |
| `schema.ToJSONString(doc)` | Transform and marshal to JSON string |
| `schema.ToJSONMany(docs)` | Transform slice and marshal to JSON bytes |
| `schema.ToObject(doc)` | Alias for Transform |

---

## 📝 Models Without Schema (Direct Collection Access)

For simple cases or when you want to manage validation yourself, you can still use collections directly:

```go
// Direct collection access (no schema validation)
users := client.Collection("users")

// Insert - no validation
result, err := users.InsertOne(ctx, user)

// But you lose:
// - Automatic _id generation guarantee
// - Schema validation
// - Strict mode protection
```

**Recommendation**: Always use Models with Schemas for production code.

---

## 🎯 Auto-Timestamps

### How It Works

When `schema.Timestamps = true` (default):

**On Insert:**
- `_id` is auto-generated if not set (MongoDB ObjectID)
- `createdAt` is set to current time
- `updatedAt` is set to current time

**On Update:**
- `updatedAt` is automatically added to `$set` operations

### Using BaseModel

```go
type User struct {
    mongoclient.BaseModel `bson:",inline"` // Includes ID, CreatedAt, UpdatedAt
    Email string `bson:"email"`
    Name  string `bson:"name"`
}

user := &User{Email: "john@example.com", Name: "John"}

// Insert - _id, createdAt, updatedAt auto-set
result, _ := userModel.InsertOne(ctx, user)

fmt.Println(user.ID)        // Auto-generated ObjectID
fmt.Println(user.CreatedAt) // Current time
fmt.Println(user.UpdatedAt) // Current time

// Update - updatedAt auto-updated
userModel.UpdateOneByID(ctx, user.ID, mongoclient.M{
    "$set": mongoclient.M{"name": "John Updated"},
})
// updatedAt is automatically added to $set
```

### Using SimpleModel (No Timestamps)

```go
type Product struct {
    mongoclient.SimpleModel `bson:",inline"` // Only ID
    Name  string  `bson:"name"`
    Price float64 `bson:"price"`
}

// Create schema without timestamps
productSchema := mongoclient.NewSchema().
    Field("name", mongoclient.TypeString, mongoclient.Required()).
    Field("price", mongoclient.TypeFloat64).
    WithTimestamps(false)
```

---

## 🔍 Advanced Queries

### Query Operators

```go
// Comparison operators
userModel.FindAll(ctx, mongoclient.M{
    "age": mongoclient.Gt(25),  // age > 25
}, &users)

userModel.FindAll(ctx, mongoclient.M{
    "status": mongoclient.In("active", "pending"),
}, &users)

userModel.FindAll(ctx, mongoclient.M{
    "name": mongoclient.Regex("^John", "i"),
}, &users)

// Logical operators
userModel.FindAll(ctx, mongoclient.Or(
    mongoclient.M{"age": mongoclient.Gt(30)},
    mongoclient.M{"role": "admin"},
), &users)
```

### Update Operators

```go
// $set - Set field values
userModel.UpdateOne(ctx, filter, mongoclient.Set(mongoclient.M{
    "name": "New Name",
}))

// $inc - Increment
userModel.UpdateOne(ctx, filter, mongoclient.Inc(mongoclient.M{
    "loginCount": 1,
}))

// $push - Add to array
userModel.UpdateOne(ctx, filter, mongoclient.Push("tags", "newTag"))

// $pull - Remove from array
userModel.UpdateOne(ctx, filter, mongoclient.Pull("tags", "oldTag"))
```

---

## 📊 Aggregation Pipelines

```go
pipeline := mongoclient.A{
    mongoclient.Match(mongoclient.M{"isActive": true}),
    mongoclient.Group("$role", mongoclient.M{
        "count":  mongoclient.M{"$sum": 1},
        "avgAge": mongoclient.M{"$avg": "$age"},
    }),
    mongoclient.Sort(mongoclient.M{"count": -1}),
    mongoclient.Limit(10),
}

var results []AggResult
err := userModel.AggregateAll(ctx, pipeline, &results)
```

---

## 🔄 Transactions

```go
err := client.WithTransaction(ctx, func(sessCtx mongo.SessionContext) error {
    _, err := userModel.InsertOne(sessCtx, &User{
        Email: "tx@example.com",
        Name:  "Transaction User",
    })
    if err != nil {
        return err // Rollback
    }

    _, err = userModel.UpdateOne(sessCtx,
        mongoclient.M{"email": "other@example.com"},
        mongoclient.Set(mongoclient.M{"role": "admin"}),
    )
    return err // Commit if nil, rollback if error
})
```

---

## 📋 Pagination

```go
pagination := &mongoclient.PaginationOptions{
    Page:     1,
    PageSize: 20,
}
pagination.Validate()

opts := options.Find().
    SetSkip(pagination.GetSkip()).
    SetLimit(pagination.GetLimit())

cursor, err := userModel.Find(ctx, filter, opts)
```

---

## 🏥 Health Checks

```go
// Ping
err := client.Ping(ctx)

// Health check
err := client.Health(ctx)

// Stats
stats := client.Stats()
```

---

## 🧪 Testing

```bash
# Run all tests
go test -v ./...

# Run with coverage
go test -v -cover -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run specific tests
go test -v -run "Schema" ./...
go test -v -run "JSON" ./...
go test -v -run "Model" ./...
```

---

## 📚 Complete Examples

See [examples/](./examples/) directory:

- **[Basic Example](./examples/basic/main.go)** - CRUD, queries, aggregations
- **[Environment Config](./examples/env-config/main.go)** - Configuration patterns
- **[Transform Example](./examples/transform/main.go)** - Transform documents for API responses

---

## 🔗 Links

- [GitHub Repository](https://github.com/isimtekin/go-packages)
- [Package Directory](https://github.com/isimtekin/go-packages/tree/main/mongo-client)
- [MongoDB Go Driver Docs](https://pkg.go.dev/go.mongodb.org/mongo-driver/mongo)
- [Report Issues](https://github.com/isimtekin/go-packages/issues)

---

## 📄 License

MIT License - See LICENSE file for details.
