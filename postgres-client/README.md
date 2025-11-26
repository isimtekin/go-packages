# postgres-client

A PostgreSQL client wrapper for Go with connection pooling, transaction support, and convenient utilities.

## Installation

```bash
go get github.com/isimtekin/go-packages/postgres-client@v0.0.1
```

## Features

- Connection pooling with pgxpool
- Transaction support with automatic commit/rollback
- Functional options pattern for configuration
- Environment variable configuration
- Error helpers for common PostgreSQL errors
- Health checks and connection statistics
- Batch operations support
- COPY protocol support for bulk inserts

## Quick Start

### Basic Usage

```go
package main

import (
    "context"
    "log"

    postgresclient "github.com/isimtekin/go-packages/postgres-client"
)

func main() {
    ctx := context.Background()

    // Create client with configuration
    config := &postgresclient.Config{
        Host:     "localhost",
        Port:     5432,
        Database: "mydb",
        Username: "postgres",
        Password: "secret",
        SSLMode:  "disable",
    }

    client, err := postgresclient.New(ctx, config)
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // Execute a query
    _, err = client.Exec(ctx, "INSERT INTO users (name, email) VALUES ($1, $2)", "John", "john@example.com")
    if err != nil {
        log.Fatal(err)
    }

    // Query single row
    var name string
    err = client.QueryRow(ctx, "SELECT name FROM users WHERE id = $1", 1).Scan(&name)
    if err != nil {
        if postgresclient.IsNoRows(err) {
            log.Println("User not found")
        } else {
            log.Fatal(err)
        }
    }
}
```

### Using Functional Options

```go
client, err := postgresclient.NewWithOptions(ctx,
    postgresclient.WithHost("localhost"),
    postgresclient.WithPort(5432),
    postgresclient.WithDatabase("mydb"),
    postgresclient.WithUsername("postgres"),
    postgresclient.WithPassword("secret"),
    postgresclient.WithSSLMode("disable"),
    postgresclient.WithMaxConns(50),
    postgresclient.WithMinConns(10),
    postgresclient.WithQueryTimeout(30*time.Second),
)
```

### Using Environment Variables

```go
// Using default POSTGRES_ prefix
client, err := postgresclient.NewFromEnvWithDefaults(ctx)

// Using custom prefix
client, err := postgresclient.NewFromEnv(ctx, "DB_")
```

Environment variables (with POSTGRES_ prefix):
- `POSTGRES_HOST` - PostgreSQL host (default: localhost)
- `POSTGRES_PORT` - PostgreSQL port (default: 5432)
- `POSTGRES_DATABASE` or `POSTGRES_DB` - Database name
- `POSTGRES_USERNAME` or `POSTGRES_USER` - Username (default: postgres)
- `POSTGRES_PASSWORD` or `POSTGRES_PASS` - Password
- `POSTGRES_SSL_MODE` - SSL mode (default: disable)
- `POSTGRES_MAX_CONNS` - Maximum connections (default: 25)
- `POSTGRES_MIN_CONNS` - Minimum connections (default: 5)
- `POSTGRES_MAX_CONN_LIFETIME` - Max connection lifetime (default: 1h)
- `POSTGRES_MAX_CONN_IDLE_TIME` - Max idle time (default: 30m)
- `POSTGRES_HEALTH_CHECK_PERIOD` - Health check period (default: 1m)
- `POSTGRES_CONNECT_TIMEOUT` - Connection timeout (default: 10s)
- `POSTGRES_QUERY_TIMEOUT` - Query timeout (default: 30s)

## Transactions

### Simple Transaction

```go
err := client.WithTransaction(ctx, func(tx pgx.Tx) error {
    _, err := tx.Exec(ctx, "INSERT INTO users (name) VALUES ($1)", "John")
    if err != nil {
        return err
    }

    _, err = tx.Exec(ctx, "INSERT INTO logs (action) VALUES ($1)", "user_created")
    if err != nil {
        return err
    }

    return nil
})
if err != nil {
    log.Fatal(err)
}
```

### Transaction with Options

```go
err := client.WithTransactionOptions(ctx, pgx.TxOptions{
    IsoLevel: pgx.Serializable,
}, func(tx pgx.Tx) error {
    // Your transaction logic here
    return nil
})
```

### Manual Transaction

```go
tx, err := client.Begin(ctx)
if err != nil {
    log.Fatal(err)
}

_, err = tx.Exec(ctx, "INSERT INTO users (name) VALUES ($1)", "John")
if err != nil {
    tx.Rollback(ctx)
    log.Fatal(err)
}

err = tx.Commit(ctx)
if err != nil {
    log.Fatal(err)
}
```

## Error Handling

```go
_, err := client.Exec(ctx, "INSERT INTO users (email) VALUES ($1)", "duplicate@example.com")
if err != nil {
    switch {
    case postgresclient.IsUniqueViolation(err):
        log.Println("Email already exists")
    case postgresclient.IsForeignKeyViolation(err):
        log.Println("Referenced record does not exist")
    case postgresclient.IsNotNullViolation(err):
        log.Println("Required field is missing")
    case postgresclient.IsConnectionError(err):
        log.Println("Database connection error")
    default:
        log.Printf("Database error: %s (code: %s)",
            postgresclient.GetPgErrorMessage(err),
            postgresclient.GetPgErrorCode(err))
    }
}
```

## Batch Operations

```go
batch := &pgx.Batch{}
batch.Queue("INSERT INTO users (name) VALUES ($1)", "User 1")
batch.Queue("INSERT INTO users (name) VALUES ($1)", "User 2")
batch.Queue("INSERT INTO users (name) VALUES ($1)", "User 3")

results := client.SendBatch(ctx, batch)
defer results.Close()

for i := 0; i < batch.Len(); i++ {
    _, err := results.Exec()
    if err != nil {
        log.Printf("Batch item %d failed: %v", i, err)
    }
}
```

## Bulk Insert with COPY

```go
rows := [][]any{
    {"John", "john@example.com"},
    {"Jane", "jane@example.com"},
    {"Bob", "bob@example.com"},
}

copyCount, err := client.CopyFrom(
    ctx,
    pgx.Identifier{"users"},
    []string{"name", "email"},
    pgx.CopyFromRows(rows),
)
if err != nil {
    log.Fatal(err)
}
log.Printf("Inserted %d rows", copyCount)
```

## Health Checks and Statistics

```go
// Check connection health
err := client.Health(ctx)
if err != nil {
    log.Printf("Database unhealthy: %v", err)
}

// Get pool statistics
stats := client.Stats()
log.Printf("Total connections: %d", stats.TotalConns())
log.Printf("Idle connections: %d", stats.IdleConns())
log.Printf("Acquired connections: %d", stats.AcquiredConns())
```

## Configuration Options

| Option | Description | Default |
|--------|-------------|---------|
| Host | PostgreSQL host | localhost |
| Port | PostgreSQL port | 5432 |
| Database | Database name | postgres |
| Username | Username | postgres |
| Password | Password | (empty) |
| SSLMode | SSL mode | disable |
| MaxConns | Maximum connections | 25 |
| MinConns | Minimum connections | 5 |
| MaxConnLifetime | Max connection lifetime | 1h |
| MaxConnIdleTime | Max connection idle time | 30m |
| HealthCheckPeriod | Health check period | 1m |
| ConnectTimeout | Connection timeout | 10s |
| QueryTimeout | Default query timeout | 30s |

## SSL Modes

- `disable` - No SSL
- `allow` - Try non-SSL first, then SSL
- `prefer` - Try SSL first, then non-SSL
- `require` - Only SSL (no certificate verification)
- `verify-ca` - SSL with CA verification
- `verify-full` - SSL with full verification

## License

MIT License
