# redis-client

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/license-MIT-blue)](LICENSE)

A high-level Redis client wrapper for Go with multi-database support, connection pooling, and a clean API built on top of [go-redis/v9](https://github.com/redis/go-redis).

## ( Features

- = **Simple API** - Clean, intuitive methods for all Redis operations
- =� **Multi-Database Support** - Manage multiple Redis databases with singleton pattern and configurable names
- = **Connection Pooling** - Automatic connection pool management per database
- = **TLS/SSL Support** - Optional encryption with certificate management
- <� **Functional Options** - Clean configuration with `WithAddr()`, `WithPassword()`, etc.
- � **High Performance** - Built on go-redis/v9 with connection reuse
- =� **Thread-Safe** - All operations protected with proper locking
- < **Environment Config** - Load configuration from environment variables
-  **Comprehensive Testing** - 96 tests with extensive coverage
- =� **All Redis Types** - String, Hash, List, Set, Sorted Set operations
- =� **Advanced Features** - Pipelines, Transactions, Pub/Sub

## =� Installation

```bash
go get github.com/isimtekin/go-packages/redis-client
```

## =� Quick Start

### Single Database

```go
package main

import (
    "context"
    "time"

    redisclient "github.com/isimtekin/go-packages/redis-client"
)

func main() {
    ctx := context.Background()

    // Create client
    client, err := redisclient.NewWithOptions(
        redisclient.WithAddr("localhost:6379"),
        redisclient.WithPassword("your-password"),
        redisclient.WithDB(0),
    )
    if err != nil {
        panic(err)
    }
    defer client.Close()

    // Set a key
    client.Set(ctx, "user:123", "John Doe", 1*time.Hour)

    // Get a key
    value, _ := client.Get(ctx, "user:123")
    println(value) // "John Doe"
}
```

### Multiple Databases (Recommended)

```go
package main

import (
    "context"
    "time"

    redisclient "github.com/isimtekin/go-packages/redis-client"
)

func main() {
    ctx := context.Background()

    // Create database manager with named databases
    manager, _ := redisclient.NewDBManagerWithOptions(
        redisclient.WithAddr("localhost:6379"),
        redisclient.WithPoolSize(100),
        redisclient.WithDatabaseNames(map[string]int{
            "cache":   0,
            "session": 1,
            "user":    2,
        }),
    )
    defer manager.Close()

    // Get clients for different databases by name
    sessions := manager.MustWithDB("session")
    cache := manager.MustWithDB("cache")
    users := manager.MustWithDB("user")

    // Use them independently
    sessions.Set(ctx, "session:abc", "active", 30*time.Minute)
    cache.Set(ctx, "page:home", "<html>...", 5*time.Minute)
    users.HSet(ctx, "user:123", "name", "John", "email", "john@example.com")
}
```

### From Environment Variables

```go
// Set environment variables
// REDIS_ADDR=localhost:6379
// REDIS_PASSWORD=secret
// REDIS_DB=0

client, err := redisclient.NewFromEnvWithDefaults(ctx)
if err != nil {
    panic(err)
}
defer client.Close()
```

## =� Documentation

### Creating Clients

#### Option 1: With Functional Options

```go
client, err := redisclient.NewWithOptions(
    redisclient.WithAddr("localhost:6379"),
    redisclient.WithPassword("secret"),
    redisclient.WithDB(0),
    redisclient.WithPoolSize(100),
    redisclient.WithDialTimeout(5*time.Second),
)
```

#### Option 2: With Config Struct

```go
config := &redisclient.Config{
    Addr:        "localhost:6379",
    Password:    "secret",
    DB:          0,
    PoolSize:    100,
    DialTimeout: 5 * time.Second,
}
client, err := redisclient.New(config)
```

#### Option 3: From Environment

```go
// Uses REDIS_ prefix by default
client, err := redisclient.NewFromEnvWithDefaults(ctx)

// Or with custom prefix
client, err := redisclient.NewFromEnv(ctx, "CACHE_")
```

### Multi-Database Management

#### Creating a Database Manager

```go
manager, err := redisclient.NewDBManagerWithOptions(
    redisclient.WithAddr("localhost:6379"),
    redisclient.WithPassword("secret"),
)
defer manager.Close()
```

#### Getting Database Clients

```go
// By number
db0, err := manager.DB(0)

// By name (if configured with WithDatabaseNames)
cache, err := manager.DB("cache")     // Maps to DB number from config
sessions, _ := manager.DB("session")
queue, _ := manager.DB("queue")

// Using DBClient wrapper
sessions := manager.MustWithDB("session")
cache := manager.MustWithDB("cache")

// Both ways return the same client instance (singleton)
client1, _ := manager.DB(0)
client2, _ := manager.DB("cache")
// client1 == client2 (if "cache" is configured as DB 0) 
```

#### Global Singleton Manager

```go
// Initialize once at app startup
redisclient.InitGlobalManagerWithOptions(
    redisclient.WithAddr("localhost:6379"),
)

// Use anywhere in your app
manager, _ := redisclient.GetGlobalManager()
db := manager.MustWithDB("cache") // If configured
```

#### Configuring Database Names

You can configure friendly names for your databases:

```go
// Option 1: Configure all at once
manager, _ := redisclient.NewDBManagerWithOptions(
    redisclient.WithAddr("localhost:6379"),
    redisclient.WithDatabaseNames(map[string]int{
        "cache":    0,
        "session":  1,
        "queue":    2,
        "user":     3,
        "analytic": 4,
    }),
)

// Option 2: Add names one by one
manager, _ := redisclient.NewDBManagerWithOptions(
    redisclient.WithAddr("localhost:6379"),
    redisclient.WithDatabaseName("cache", 0),
    redisclient.WithDatabaseName("session", 1),
    redisclient.WithDatabaseName("queue", 2),
)

// Now use friendly names
cache := manager.MustDB("cache")      // DB 0
session := manager.MustDB("session")  // DB 1
queue := manager.MustDB("queue")      // DB 2
```

### String Operations

```go
// Set with TTL
client.Set(ctx, "key", "value", 1*time.Hour)

// Set without TTL
client.Set(ctx, "key", "value", 0)

// Set only if not exists
ok, _ := client.SetNX(ctx, "key", "value", 1*time.Hour)

// Get
value, err := client.Get(ctx, "key")
if redisclient.IsNil(err) {
    // Key doesn't exist
}

// Multiple get/set
client.MSet(ctx, "key1", "val1", "key2", "val2")
values, _ := client.MGet(ctx, "key1", "key2")

// Increment/Decrement
count, _ := client.Incr(ctx, "counter")
count, _ := client.IncrBy(ctx, "counter", 5)
count, _ := client.Decr(ctx, "counter")
```

### Hash Operations

```go
// Set hash fields
client.HSet(ctx, "user:123", "name", "John", "age", "30")

// Get single field
name, _ := client.HGet(ctx, "user:123", "name")

// Get all fields
data, _ := client.HGetAll(ctx, "user:123")

// Delete fields
client.HDel(ctx, "user:123", "age")

// Check if field exists
exists, _ := client.HExists(ctx, "user:123", "name")

// Increment field
client.HIncrBy(ctx, "user:123", "loginCount", 1)

// Get all field names
keys, _ := client.HKeys(ctx, "user:123")
```

### List Operations

```go
// Push to head/tail
client.LPush(ctx, "queue", "item1", "item2")
client.RPush(ctx, "queue", "item3")

// Pop from head/tail
item, _ := client.LPop(ctx, "queue")
item, _ := client.RPop(ctx, "queue")

// Get range
items, _ := client.LRange(ctx, "queue", 0, -1) // All items

// Get length
length, _ := client.LLen(ctx, "queue")
```

### Set Operations

```go
// Add members
client.SAdd(ctx, "tags", "go", "redis", "database")

// Get all members
members, _ := client.SMembers(ctx, "tags")

// Check membership
exists, _ := client.SIsMember(ctx, "tags", "go")

// Remove members
client.SRem(ctx, "tags", "redis")

// Get count
count, _ := client.SCard(ctx, "tags")
```

### Sorted Set Operations

```go
// Add members with scores
client.ZAdd(ctx, "leaderboard",
    redis.Z{Score: 100, Member: "player1"},
    redis.Z{Score: 95, Member: "player2"},
)

// Get range by rank
players, _ := client.ZRange(ctx, "leaderboard", 0, 9) // Top 10

// Get range with scores
scores, _ := client.ZRangeWithScores(ctx, "leaderboard", 0, 9)

// Get score
score, _ := client.ZScore(ctx, "leaderboard", "player1")

// Remove members
client.ZRem(ctx, "leaderboard", "player1")
```

### Key Operations

```go
// Delete keys
count, _ := client.Del(ctx, "key1", "key2", "key3")

// Check existence
exists, _ := client.Exists(ctx, "key1", "key2")

// Set expiration
client.Expire(ctx, "key", 1*time.Hour)

// Get TTL
ttl, _ := client.TTL(ctx, "key")

// Remove expiration
client.Persist(ctx, "key")

// Get all keys matching pattern
keys, _ := client.Keys(ctx, "user:*")
```

### Pipeline Operations

```go
pipe := client.Pipeline()

// Queue commands
pipe.Set(ctx, "key1", "value1", 0)
pipe.Set(ctx, "key2", "value2", 0)
pipe.Incr(ctx, "counter")

// Execute all at once
cmds, err := pipe.Exec(ctx)
```

### Pub/Sub

```go
// Publish
client.Publish(ctx, "notifications", "Hello!")

// Subscribe
pubsub := client.Subscribe(ctx, "notifications")
defer pubsub.Close()

for msg := range pubsub.Channel() {
    fmt.Println(msg.Payload)
}
```

### Transactions

```go
err := client.Watch(ctx, func(tx *redis.Tx) error {
    // Get current value
    val, _ := tx.Get(ctx, "counter").Result()

    // Modify
    newVal := parseInt(val) + 1

    // Set in transaction
    _, err := tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
        pipe.Set(ctx, "counter", newVal, 0)
        return nil
    })

    return err
}, "counter")
```

## < Environment Variables

The package supports loading configuration from environment variables:

### Default Prefix: `REDIS_`

```bash
# Required
REDIS_ADDR=localhost:6379

# Optional
REDIS_PASSWORD=secret
REDIS_DB=0
REDIS_MAX_RETRIES=3
REDIS_MIN_IDLE_CONNS=5
REDIS_MAX_IDLE_CONNS=10
REDIS_POOL_SIZE=100
REDIS_POOL_TIMEOUT=4s
REDIS_CONN_MAX_IDLE_TIME=5m
REDIS_CONN_MAX_LIFETIME=0
REDIS_DIAL_TIMEOUT=5s
REDIS_READ_TIMEOUT=3s
REDIS_WRITE_TIMEOUT=3s

# TLS/SSL
REDIS_TLS_ENABLED=false
REDIS_TLS_CERT_FILE=/path/to/cert.pem
REDIS_TLS_KEY_FILE=/path/to/key.pem
REDIS_TLS_CA_FILE=/path/to/ca.pem
```

### Custom Prefix

```go
// Use CACHE_ prefix instead
// CACHE_ADDR, CACHE_PASSWORD, etc.
client, err := redisclient.NewFromEnv(ctx, "CACHE_")
```

### Using .env Files

```go
import envutil "github.com/isimtekin/go-packages/env-util"

// Load .env file
envutil.LoadEnvFile(".env")

// Then create client
client, err := redisclient.NewFromEnvWithDefaults(ctx)
```

## � Configuration Options

All available configuration options:

```go
config := &redisclient.Config{
    // Connection
    Addr:     "localhost:6379", // host:port
    Password: "",               // optional password
    DB:       0,                // database number

    // Connection pool
    MaxRetries:      3,
    MinIdleConns:    5,
    MaxIdleConns:    10,
    PoolSize:        100,
    PoolTimeout:     4 * time.Second,
    ConnMaxIdleTime: 5 * time.Minute,
    ConnMaxLifetime: 0, // 0 = no limit

    // Timeouts
    DialTimeout:  5 * time.Second,
    ReadTimeout:  3 * time.Second,
    WriteTimeout: 3 * time.Second,

    // TLS
    TLSEnabled:  false,
    TLSCertFile: "",
    TLSKeyFile:  "",
    TLSCAFile:   "",
}
```

### Functional Options

```go
// All available options
client, _ := redisclient.NewWithOptions(
    redisclient.WithAddr("localhost:6379"),
    redisclient.WithPassword("secret"),
    redisclient.WithDB(0),
    redisclient.WithMaxRetries(5),
    redisclient.WithPoolSize(200),
    redisclient.WithMinIdleConns(10),
    redisclient.WithMaxIdleConns(20),
    redisclient.WithPoolTimeout(5*time.Second),
    redisclient.WithConnMaxIdleTime(10*time.Minute),
    redisclient.WithConnMaxLifetime(1*time.Hour),
    redisclient.WithDialTimeout(10*time.Second),
    redisclient.WithReadTimeout(5*time.Second),
    redisclient.WithWriteTimeout(5*time.Second),
    redisclient.WithTLS("/cert.pem", "/key.pem", "/ca.pem"),
)
```

## =' Advanced Usage

### TLS/SSL Configuration

```go
client, err := redisclient.NewWithOptions(
    redisclient.WithAddr("secure-redis.example.com:6380"),
    redisclient.WithTLS(
        "/path/to/client-cert.pem",
        "/path/to/client-key.pem",
        "/path/to/ca-cert.pem",
    ),
)
```

### Accessing Underlying go-redis Client

```go
// Get underlying client for advanced operations
underlyingClient := client.Client()

// Use go-redis API directly
result := underlyingClient.Do(ctx, "CUSTOM", "COMMAND")
```

### Error Handling

```go
value, err := client.Get(ctx, "key")
if err != nil {
    if redisclient.IsNil(err) {
        // Key doesn't exist
        fmt.Println("Key not found")
    } else if redisclient.IsConnectionError(err) {
        // Connection issue
        fmt.Println("Connection failed")
    } else {
        // Other error
        fmt.Println("Error:", err)
    }
}
```

## <� Real-World Examples

### E-Commerce Application

```go
manager, _ := redisclient.NewDBManagerWithOptions(
    redisclient.WithAddr("localhost:6379"),
)
defer manager.Close()

// DB0: User sessions
sessions := manager.MustWithDB(0)
sessions.Set(ctx, "session:user123", "logged_in", 24*time.Hour)

// DB1: Product cache
products := manager.MustWithDB(1)
products.HSet(ctx, "product:999",
    "name", "Laptop",
    "price", "1299.99",
    "stock", "50",
)

// DB2: Shopping carts
carts := manager.MustWithDB(2)
carts.LPush(ctx, "cart:user123", "product:999", "product:888")

// DB3: Rate limiting
rateLimit := manager.MustWithDB(3)
count, _ := rateLimit.Incr(ctx, "ratelimit:user123")
if count == 1 {
    rateLimit.Client().Expire(ctx, "ratelimit:user123", 1*time.Minute)
}
```

### Caching with TTL

```go
// Cache with different TTLs
cache := manager.MustWithDB(1)

// Short-lived cache (5 minutes)
cache.Set(ctx, "homepage", "<html>...</html>", 5*time.Minute)

// Medium-lived cache (1 hour)
cache.Set(ctx, "user:profile:123", `{"name":"John"}`, 1*time.Hour)

// Long-lived cache (24 hours)
cache.Set(ctx, "settings", `{"theme":"dark"}`, 24*time.Hour)
```

### Distributed Locking

```go
lockKey := "lock:resource:123"
lockValue := "unique-token"

// Acquire lock
ok, _ := client.SetNX(ctx, lockKey, lockValue, 10*time.Second)
if !ok {
    // Lock already held
    return errors.New("resource locked")
}

// Do work...

// Release lock
client.Del(ctx, lockKey)
```

## =� Performance

- **Connection Pooling**: Reuses connections for optimal performance
- **Pipelining**: Batch multiple commands for reduced latency
- **Lazy Initialization**: Database connections created on first use
- **Singleton Pattern**: Each database connection is reused across calls

##  Testing

```bash
# Run all tests
go test -v ./...

# Run with coverage
go test -v -cover ./...

# Run specific test
go test -v -run TestDBManager

# Run with race detection
go test -race ./...
```

**Test Coverage**: 68 tests covering:
- Configuration validation
- Connection management
- All Redis operations
- Multi-database functionality
- Error handling
- Thread safety

## =� Examples

See the [examples/](./examples/) directory for complete examples:

- **[multi-db/](./examples/multi-db/)** - Multi-database usage patterns
- More examples coming soon!

## > Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## =� License

MIT License - See LICENSE file for details.

## = Links

- [GitHub Repository](https://github.com/isimtekin/go-packages)
- [go-redis Documentation](https://redis.uptrace.dev/)
- [Redis Documentation](https://redis.io/docs/)

## =O Acknowledgments

Built on top of [go-redis/v9](https://github.com/redis/go-redis) - the official Redis client for Go.

## =� Related Packages

- [env-util](../env-util) - Environment variable utilities
- [mongo-client](../mongo-client) - MongoDB client wrapper
