# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Context

This is the `mongo-client` package within a Go monorepo (`go-packages`). The monorepo uses **independent versioning** per package with the format `package-name/vX.Y.Z`. This package is a minimalist MongoDB client wrapper providing high-level CRUD operations, transaction support, and optional auto-timestamps.

## Development Commands

### Testing
```bash
# Run all tests (recommended - uses race detection)
make test

# Run tests with coverage report
make coverage

# Run integration tests (requires MongoDB)
make test-integration
# Or manually:
make docker-up && go test -v ./... && make docker-down
```

### Running Examples
```bash
# Basic example (requires MongoDB running)
make run-example
# Or: cd examples/basic && go run main.go

# Environment configuration example (no MongoDB needed)
make run-env-example
# Or: cd examples/env-config && go run main.go

# Run all examples
make run-all-examples
```

### Code Quality
```bash
# Format code
make fmt

# Run linter (requires golangci-lint)
make lint

# Clean artifacts
make clean
```

### MongoDB Docker
```bash
# Start MongoDB container
make docker-up

# Stop MongoDB container
make docker-down
```

### Monorepo Release (from parent directory)
```bash
cd ..  # Go to monorepo root

# Release new version
./scripts/release.sh mongo-client 1.0.0

# Auto-increment version
./scripts/bump-version.sh mongo-client patch  # 1.0.0 → 1.0.1
./scripts/bump-version.sh mongo-client minor  # 1.0.1 → 1.1.0
./scripts/bump-version.sh mongo-client major  # 1.1.0 → 2.0.0
```

## Architecture Overview

### Core Components

**Client (`client.go`)**: Main MongoDB client wrapper
- Manages connection lifecycle and health checks
- Provides transaction support via `WithTransaction()`
- Handles connection pooling and automatic reconnection
- All operations use context with configurable timeouts

**Collection (`collection.go`)**: Collection wrapper with CRUD operations
- Wraps `mongo.Collection` with high-level methods
- **Auto-timestamp injection**: Automatically adds timestamps on insert/update
- Context management: Creates operation contexts with timeouts
- ID conversion: Handles both string and `primitive.ObjectID` types

**Configuration (`config.go`)**: Configuration management
- Supports functional options pattern via `options.go`
- Provides `DefaultConfig()` with sensible defaults
- Validates configuration before client creation

**Schema (`schema.go`)**: Optional auto-timestamp system
- **Key Design**: Interface-based detection for optional timestamps
- `Timestamped` interface: Implement to enable auto-timestamps
- `BaseModel`: Embeddable struct WITH timestamps (implements `Timestamped`)
- `SimpleModel`: Embeddable struct WITHOUT timestamps (doesn't implement `Timestamped`)

**Environment Integration (`env.go`)**: Load config from environment variables
- Uses `env-util` package (local dependency in monorepo)
- Supports custom prefixes (default: `MONGO_`)
- Builds URI from components if `MONGO_URI` not provided
- Functions: `NewFromEnvWithDefaults()`, `NewFromEnv()`, `LoadConfigFromEnv()`

**Helpers (`helpers.go`)**: Query builders and utilities
- Type aliases: `M` (bson.M), `D` (bson.D), `A` (bson.A)
- Update operators: `Set()`, `Inc()`, `Push()`, `Pull()`, `AddToSet()`, etc.
- Aggregation helpers: `Match()`, `Group()`, `Sort()`, `Limit()`, `Lookup()`, etc.
- Query operators: `Gt()`, `Gte()`, `Lt()`, `In()`, `Regex()`, `Or()`, `And()`

### Auto-Timestamp System (Critical Architecture)

This is the **most important architectural pattern** in the codebase:

1. **Timestamped Interface**: Documents implementing this interface get automatic timestamps
2. **Detection at Runtime**: `applyTimestamps()` checks if document implements `Timestamped`
3. **Injection Points**:
   - `InsertOne()`/`InsertMany()`: Calls `applyTimestamps(doc, true)` - sets both `createdAt` and `updatedAt`
   - `UpdateOne()`/`UpdateMany()`: Calls `addUpdatedAtToUpdate()` - adds `updatedAt` to `$set` operator
4. **Truly Optional**: Documents with `SimpleModel` don't implement `Timestamped`, so no timestamps are applied

**Example Usage**:
```go
// WITH auto-timestamps
type User struct {
    mongoclient.BaseModel `bson:",inline"`  // Implements Timestamped
    Email string `bson:"email"`
}
// createdAt and updatedAt are automatic!

// WITHOUT auto-timestamps
type Product struct {
    mongoclient.SimpleModel `bson:",inline"`  // Does NOT implement Timestamped
    Name string `bson:"name"`
}
// No timestamps are added
```

### Local Dependencies

This package depends on `env-util` from the same monorepo:
- `go.mod` uses `replace` directive: `replace github.com/isimtekin/go-packages/env-util => ../env-util`
- When adding imports, use: `import envutil "github.com/isimtekin/go-packages/env-util"`
- Both packages must be present in the monorepo structure

## Testing Strategy

### Test Files
- `client_test.go`: Client creation, configuration validation, timeout handling
- `collection_test.go`: Auto-timestamp injection in updates, ObjectID conversion, context creation
- `schema_test.go`: `BaseModel`/`SimpleModel` behavior, `Timestamped` interface verification, `applyTimestamps()` logic
- `helpers_test.go`: Query builders, aggregation helpers, utility functions

### Critical Test Patterns

**Testing Interface Implementation**:
```go
// Verify BaseModel implements Timestamped
var _ Timestamped = &BaseModel{}

// Verify SimpleModel does NOT implement Timestamped
_, ok := interface{}(&SimpleModel{}).(Timestamped)
// ok should be false
```

**Testing Auto-Timestamp Injection**:
```go
// Test that updatedAt is added to $set operations
update := bson.M{"$set": bson.M{"name": "New Name"}}
result := collection.addUpdatedAtToUpdate(update)
// result should have updatedAt in $set
```

### Running Single Tests
```bash
# Run specific test
go test -v -run TestBaseModel_Timestamped

# Run tests in specific file
go test -v -run Test.*Timestamp

# Run with race detection
go test -v -race -run TestName
```

## Common Patterns

### Creating a Client
```go
// From environment (production)
client, err := mongoclient.NewFromEnvWithDefaults(ctx)

// With functional options (programmatic)
client, err := mongoclient.NewWithOptions(ctx,
    mongoclient.WithURI("mongodb://localhost:27017"),
    mongoclient.WithDatabase("mydb"),
)
```

### CRUD with Auto-Timestamps
```go
type User struct {
    mongoclient.BaseModel `bson:",inline"`  // Auto-timestamps enabled
    Email string `bson:"email"`
}

users := client.Collection("users")

// Insert - createdAt and updatedAt set automatically
user := &User{Email: "test@example.com"}
users.InsertOne(ctx, user)

// Update - updatedAt added to $set automatically
users.UpdateOne(ctx,
    mongoclient.M{"email": "test@example.com"},
    mongoclient.Set(mongoclient.M{"name": "New Name"}),
)
```

### Transactions
```go
err := client.WithTransaction(ctx, func(sessCtx mongo.SessionContext) error {
    _, err := collection.InsertOne(sessCtx, doc)
    if err != nil {
        return err  // Automatic rollback
    }
    return nil  // Automatic commit
})
```

## Environment Variables

Default prefix: `MONGO_`

**Required**:
- `MONGO_URI` OR (`MONGO_HOST` + `MONGO_PORT`)
- `MONGO_DATABASE`

**Optional**:
- `MONGO_USERNAME`, `MONGO_PASSWORD`, `MONGO_AUTH_SOURCE`
- `MONGO_MAX_POOL_SIZE`, `MONGO_MIN_POOL_SIZE`
- `MONGO_CONNECT_TIMEOUT`, `MONGO_SOCKET_TIMEOUT`, `MONGO_OPERATION_TIMEOUT`
- `MONGO_RETRY_WRITES`, `MONGO_RETRY_READS`

See `.env.example` for complete list and examples.

## Code Style

- Use functional options pattern for configuration
- All operations accept `context.Context` as first parameter
- Return wrapped result types (`*InsertOneResult`, `*UpdateResult`, etc.)
- Use `bson.M` for filters and updates (aliased as `mongoclient.M`)
- Error wrapping: Use `fmt.Errorf()` with `%w` for error chains
- Auto-timestamps: Always call via `applyTimestamps()` - never manually set timestamps

## Important Notes

1. **Examples Structure**: Examples are in separate subdirectories (`examples/basic/`, `examples/env-config/`) to avoid "main redeclared" errors
2. **Monorepo Context**: This is part of a multi-package repo - use relative paths for local dependencies
3. **Interface-Based Design**: The `Timestamped` interface pattern is critical - don't break it by adding timestamps manually
4. **Context Timeouts**: All operations automatically create contexts with `OperationTimeout` from config
5. **Local Module Replacement**: `go.mod` uses `replace` for `env-util` - this is intentional for monorepo development

## Examples Location

All examples are in `examples/` directory:
- `examples/basic/main.go`: Comprehensive feature demonstration
- `examples/env-config/main.go`: Environment variable configuration patterns
- `examples/README.md`: Documentation for running examples
