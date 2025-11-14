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
- 🎯 **Auto-timestamps** - Optional automatic `createdAt`/`updatedAt` (like Mongoose)
- 📝 **Flexible schemas** - BaseModel (with timestamps) or SimpleModel (without)
- 🔧 **Functional options** - Clean configuration with functional options pattern
- 🛠️ **Query builders** - Helper functions for building MongoDB queries
- 📋 **Pagination support** - Built-in pagination helpers
- ✅ **Type-safe** - Full TypeScript-like experience with Go
- 🧪 **Well-tested** - Comprehensive test coverage

## 🚀 Quick Start

```bash
# Installation (after first release)
go get github.com/isimtekin/go-packages/mongo-client
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

### Defining Models

**Option 1: With Auto-Timestamps (Recommended for most use cases)**

Use `BaseModel` for automatic `createdAt` and `updatedAt` management:

```go
// User model WITH auto-timestamps
type User struct {
    mongoclient.BaseModel `bson:",inline"` // Includes ID, CreatedAt, UpdatedAt
    Email                 string   `bson:"email" json:"email"`
    Name                  string   `bson:"name" json:"name"`
    Age                   int      `bson:"age" json:"age"`
    Active                bool     `bson:"active" json:"active"`
}

// No hooks needed! Timestamps are automatic:
// - createdAt: set automatically on insert
// - updatedAt: set automatically on insert and update
```

**Option 2: Without Auto-Timestamps**

Use `SimpleModel` when you don't need automatic timestamps:

```go
// Product model WITHOUT auto-timestamps
type Product struct {
    mongoclient.SimpleModel `bson:",inline"` // Only includes ID
    Name                    string  `bson:"name" json:"name"`
    Price                   float64 `bson:"price" json:"price"`
    SKU                     string  `bson:"sku" json:"sku"`
}

// No timestamps will be added automatically
```

**How Auto-Timestamps Work:**

1. **On Insert** (`InsertOne`, `InsertMany`):
   - `createdAt` is set to current time (if zero)
   - `updatedAt` is set to current time (if zero)

2. **On Update** (`UpdateOne`, `UpdateMany`, `UpdateOneByID`):
   - `updatedAt` is automatically added to `$set` operations

```go
// Insert - both timestamps set automatically
user := &User{Email: "john@example.com", Name: "John"}
result, _ := users.InsertOne(ctx, user)
// user.CreatedAt and user.UpdatedAt are now set!

// Update - updatedAt set automatically
users.UpdateOne(ctx,
    mongoclient.M{"email": "john@example.com"},
    mongoclient.Set(mongoclient.M{"age": 31}),
    // updatedAt is automatically added to $set!
)
```

### CRUD Operations

```go
// Get collection
users := client.Collection("users")

// Insert one document
user := &User{
    Email:  "john@example.com",
    Name:   "John Doe",
    Age:    30,
    Active: true,
}
user.BeforeInsert()

result, err := users.InsertOne(ctx, user)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Inserted ID: %v\n", result.InsertedID)

// Insert multiple documents
users := []interface{}{
    &User{Email: "jane@example.com", Name: "Jane Smith", Age: 28},
    &User{Email: "bob@example.com", Name: "Bob Johnson", Age: 35},
}
manyResult, err := users.InsertMany(ctx, users)

// Find one document
var foundUser User
err = users.FindOne(ctx, mongoclient.M{"email": "john@example.com"}).Decode(&foundUser)
if mongoclient.IsNoDocuments(err) {
    fmt.Println("User not found")
}

// Find by ID
err = users.FindOneByID(ctx, userID).Decode(&foundUser)

// Find all documents
var allUsers []User
filter := mongoclient.M{"active": true}
err = users.FindAll(ctx, filter, &allUsers)

// Update one document
updateResult, err := users.UpdateOne(ctx,
    mongoclient.M{"email": "john@example.com"},
    mongoclient.Set(mongoclient.M{"age": 31}),
)

// Update by ID
updateResult, err := users.UpdateOneByID(ctx, userID,
    mongoclient.Set(mongoclient.M{"name": "John Updated"}),
)

// Update many documents
updateResult, err := users.UpdateMany(ctx,
    mongoclient.M{"active": false},
    mongoclient.Set(mongoclient.M{"status": "inactive"}),
)

// Delete one document
deleteResult, err := users.DeleteOne(ctx, mongoclient.M{"email": "john@example.com"})

// Delete by ID
deleteResult, err := users.DeleteOneByID(ctx, userID)

// Delete many documents
deleteResult, err := users.DeleteMany(ctx, mongoclient.M{"active": false})
```

## 🔍 Advanced Queries

### Query Operators

```go
// Comparison operators
users.FindAll(ctx, mongoclient.M{
    "age": mongoclient.Gt(25),  // age > 25
})

users.FindAll(ctx, mongoclient.M{
    "age": mongoclient.Gte(25), // age >= 25
})

users.FindAll(ctx, mongoclient.M{
    "age": mongoclient.Lt(50),  // age < 50
})

users.FindAll(ctx, mongoclient.M{
    "status": mongoclient.In("active", "pending"), // status IN (...)
})

users.FindAll(ctx, mongoclient.M{
    "name": mongoclient.Regex("^John", "i"), // regex match
})

// Logical operators
users.FindAll(ctx, mongoclient.Or(
    mongoclient.M{"age": mongoclient.Gt(30)},
    mongoclient.M{"status": "premium"},
))

users.FindAll(ctx, mongoclient.And(
    mongoclient.M{"active": true},
    mongoclient.M{"age": mongoclient.Gte(18)},
))
```

### Update Operators

```go
// $set - Set field values
users.UpdateOne(ctx, filter, mongoclient.Set(mongoclient.M{
    "name": "New Name",
    "age":  30,
}))

// $inc - Increment field
users.UpdateOne(ctx, filter, mongoclient.Inc(mongoclient.M{
    "loginCount": 1,
}))

// $push - Add to array
users.UpdateOne(ctx, filter, mongoclient.Push("tags", "newTag"))

// $pull - Remove from array
users.UpdateOne(ctx, filter, mongoclient.Pull("tags", "oldTag"))

// $addToSet - Add unique value to array
users.UpdateOne(ctx, filter, mongoclient.AddToSet("tags", "uniqueTag"))

// $unset - Remove fields
users.UpdateOne(ctx, filter, mongoclient.Unset("temporaryField"))

// $currentDate - Set to current date
users.UpdateOne(ctx, filter, mongoclient.CurrentDate("lastModified"))
```

## 📊 Aggregation Pipelines

```go
pipeline := mongoclient.A{
    // Match active users
    mongoclient.Match(mongoclient.M{"active": true}),

    // Group by category
    mongoclient.Group("$category", mongoclient.M{
        "count":  mongoclient.M{"$sum": 1},
        "avgAge": mongoclient.M{"$avg": "$age"},
    }),

    // Sort by count
    mongoclient.Sort(mongoclient.M{"count": -1}),

    // Limit results
    mongoclient.Limit(10),
}

type AggResult struct {
    ID     string  `bson:"_id"`
    Count  int     `bson:"count"`
    AvgAge float64 `bson:"avgAge"`
}

var results []AggResult
err := users.AggregateAll(ctx, pipeline, &results)

// Or get a single result
var result AggResult
err := users.AggregateOne(ctx, pipeline, &result)
```

### Joins with $lookup

```go
pipeline := mongoclient.A{
    // Lookup related documents
    mongoclient.Lookup("orders", "userId", "_id", "userOrders"),

    // Unwind array
    mongoclient.Unwind("$userOrders"),

    // Project specific fields
    mongoclient.Project(mongoclient.M{
        "name":       1,
        "orderTotal": "$userOrders.total",
    }),
}
```

## 🔄 Transactions

```go
err := client.WithTransaction(ctx, func(sessCtx mongo.SessionContext) error {
    // All operations in this function are part of the transaction

    // Insert document
    _, err := users.InsertOne(sessCtx, &User{
        Email: "tx@example.com",
        Name:  "Transaction User",
        Age:   25,
    })
    if err != nil {
        return err // Rolls back transaction
    }

    // Update document
    _, err = users.UpdateOne(sessCtx,
        mongoclient.M{"email": "jane@example.com"},
        mongoclient.Set(mongoclient.M{"age": 29}),
    )
    if err != nil {
        return err // Rolls back transaction
    }

    // If no errors, transaction commits automatically
    return nil
})

if err != nil {
    log.Printf("Transaction failed: %v", err)
}
```

## 📋 Pagination

```go
// Define pagination options
pagination := &mongoclient.PaginationOptions{
    Page:     1,    // Current page
    PageSize: 20,   // Items per page
}
pagination.Validate() // Ensures values are valid

// Use in find operation
import "go.mongodb.org/mongo-driver/mongo/options"

opts := options.Find().
    SetSkip(pagination.GetSkip()).
    SetLimit(pagination.GetLimit())

var users []User
cursor, err := collection.Find(ctx, filter, opts)
```

## 🏥 Health Checks

```go
// Ping database
err := client.Ping(ctx)
if err != nil {
    log.Printf("Database unreachable: %v", err)
}

// Health check with details
err := client.Health(ctx)
if err != nil {
    log.Printf("Health check failed: %v", err)
}

// Get connection stats
stats := client.Stats()
```

## 🌍 Environment Variable Configuration

The package supports loading configuration from environment variables, making it easy to configure for different environments (dev, staging, production).

### Quick Start with Environment Variables

```bash
# Set environment variables
export MONGO_URI=mongodb://localhost:27017
export MONGO_DATABASE=myapp

# Or use .env file
cat > .env << EOF
MONGO_URI=mongodb://localhost:27017
MONGO_DATABASE=myapp
EOF
```

```go
// Load and connect
envutil.LoadEnvFile(".env")
client, err := mongoclient.NewFromEnvWithDefaults(ctx)
```

### Supported Environment Variables

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `MONGO_URI` | Full MongoDB connection URI | - | `mongodb://localhost:27017` |
| `MONGO_DATABASE` | Database name | - | `myapp` |
| `MONGO_HOST` | Host (if not using URI) | `localhost` | `db.example.com` |
| `MONGO_PORT` | Port (if not using URI) | `27017` | `27017` |
| `MONGO_USERNAME` | Username for auth | - | `admin` |
| `MONGO_PASSWORD` | Password for auth | - | `secret` |
| `MONGO_AUTH_SOURCE` | Auth database | `admin` | `admin` |
| `MONGO_MAX_POOL_SIZE` | Max connections | `100` | `50` |
| `MONGO_MIN_POOL_SIZE` | Min connections | `10` | `5` |
| `MONGO_MAX_CONN_IDLE_TIME` | Max idle time | `5m` | `10m` |
| `MONGO_CONNECT_TIMEOUT` | Connect timeout | `10s` | `30s` |
| `MONGO_SOCKET_TIMEOUT` | Socket timeout | `30s` | `60s` |
| `MONGO_SERVER_SELECTION_TIMEOUT` | Server selection timeout | `10s` | `5s` |
| `MONGO_OPERATION_TIMEOUT` | Default operation timeout | `30s` | `60s` |
| `MONGO_RETRY_WRITES` | Enable retry writes | `true` | `true` |
| `MONGO_RETRY_READS` | Enable retry reads | `true` | `true` |

### Custom Prefix

Use a custom prefix for your environment variables:

```go
// Use DB_ prefix instead of MONGO_
// DB_URI, DB_DATABASE, etc.
client, err := mongoclient.NewFromEnv(ctx, "DB_")

// Use MYAPP_MONGO_ prefix
// MYAPP_MONGO_URI, MYAPP_MONGO_DATABASE, etc.
client, err := mongoclient.NewFromEnv(ctx, "MYAPP_MONGO_")
```

### Building URI from Components

If `MONGO_URI` is not set, the URI is built from individual components:

```bash
# With authentication
MONGO_HOST=db.example.com
MONGO_PORT=27017
MONGO_USERNAME=myuser
MONGO_PASSWORD=mypass
MONGO_AUTH_SOURCE=admin

# Results in: mongodb://myuser:mypass@db.example.com:27017/?authSource=admin
```

```bash
# Without authentication
MONGO_HOST=localhost
MONGO_PORT=27017

# Results in: mongodb://localhost:27017
```

### Environment-Specific Configurations

**Development (.env.development):**
```bash
MONGO_URI=mongodb://localhost:27017
MONGO_DATABASE=dev_db
MONGO_MAX_POOL_SIZE=10
```

**Production (.env.production):**
```bash
MONGO_HOST=prod-db.example.com
MONGO_PORT=27017
MONGO_USERNAME=${DB_USER}
MONGO_PASSWORD=${DB_PASSWORD}
MONGO_DATABASE=prod_db
MONGO_MAX_POOL_SIZE=100
MONGO_AUTH_SOURCE=admin
```

**MongoDB Atlas:**
```bash
MONGO_URI=mongodb+srv://username:password@cluster.mongodb.net/?retryWrites=true&w=majority
MONGO_DATABASE=atlas_db
```

**Docker Compose:**
```yaml
environment:
  MONGO_URI: mongodb://mongo:27017
  MONGO_DATABASE: app_db
```

## 🔧 Configuration Options

### Available Options

```go
// Connection
mongoclient.WithURI(uri string)
mongoclient.WithDatabase(name string)

// Connection pool
mongoclient.WithMaxPoolSize(size uint64)
mongoclient.WithMinPoolSize(size uint64)
mongoclient.WithMaxConnIdleTime(duration time.Duration)

// Timeouts
mongoclient.WithConnectTimeout(duration time.Duration)
mongoclient.WithSocketTimeout(duration time.Duration)
mongoclient.WithServerSelectionTimeout(duration time.Duration)
mongoclient.WithOperationTimeout(duration time.Duration)

// Retry
mongoclient.WithRetryWrites(enabled bool)
mongoclient.WithRetryReads(enabled bool)
```

### Default Configuration

```go
{
    URI:                    "mongodb://localhost:27017",
    Database:               "test",
    MaxPoolSize:            100,
    MinPoolSize:            10,
    MaxConnIdleTime:        5 * time.Minute,
    ConnectTimeout:         10 * time.Second,
    SocketTimeout:          30 * time.Second,
    ServerSelectionTimeout: 10 * time.Second,
    OperationTimeout:       30 * time.Second,
    RetryWrites:            true,
    RetryReads:             true,
}
```

## 🔨 Utility Functions

```go
// ObjectID helpers
id := mongoclient.NewObjectID()
objID, err := mongoclient.ObjectIDFromHex("507f1f77bcf86cd799439011")
isValid := mongoclient.IsValidObjectID("507f1f77bcf86cd799439011")

// Error helpers
if mongoclient.IsNoDocuments(err) {
    // Handle not found
}

if mongoclient.IsDuplicateKey(err) {
    // Handle duplicate key error
}

// Aliases for convenience
filter := mongoclient.M{"status": "active"}      // bson.M
doc := mongoclient.D{{"key", "value"}}          // bson.D
arr := mongoclient.A{"val1", "val2"}            // bson.A
```

## 📚 Complete Examples

See [examples/main.go](./examples/main.go) for comprehensive examples including:
- Creating and configuring clients
- CRUD operations
- Advanced queries
- Aggregation pipelines
- Transactions
- Pagination
- Health checks
- Index management

## 🧪 Testing

```bash
# Run tests
go test -v ./...

# Run tests with coverage
go test -v -cover -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Note: Integration tests require MongoDB running
docker run -d -p 27017:27017 mongo:latest
```

## 📊 Performance Considerations

- **Connection pooling**: Reuses connections for better performance
- **Context timeouts**: All operations have configurable timeouts
- **Batch operations**: Use `InsertMany`, `UpdateMany`, `BulkWrite` for bulk operations
- **Indexes**: Create indexes for frequently queried fields
- **Projection**: Use field projection to reduce data transfer

## 🔒 Security Best Practices

1. **Use authentication**: Always use authenticated MongoDB connections
2. **Limit permissions**: Use database users with minimal required permissions
3. **Validate input**: Always validate and sanitize user input
4. **Use TLS**: Enable TLS/SSL for production deployments
5. **Connection strings**: Never hardcode credentials, use environment variables

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📄 License

MIT License - See LICENSE file for details.

## 🔗 Links

- [GitHub Repository](https://github.com/isimtekin/go-packages)
- [Package Directory](https://github.com/isimtekin/go-packages/tree/main/mongo-client)
- [MongoDB Go Driver Docs](https://pkg.go.dev/go.mongodb.org/mongo-driver/mongo)
- [Report Issues](https://github.com/isimtekin/go-packages/issues)
- [Example Code](./examples/main.go)

## 🙏 Acknowledgments

Built on top of the official [MongoDB Go Driver](https://github.com/mongodb/mongo-go-driver).
