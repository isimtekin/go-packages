# nats-client

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/license-MIT-blue)](LICENSE)

A high-level NATS client wrapper for Go with clean API, automatic reconnection, and JetStream support built on top of [nats.go](https://github.com/nats-io/nats.go).

## Features

- **Simple API** - Clean, intuitive methods for pub/sub and request/reply patterns
- **Automatic Reconnection** - Configurable reconnection with exponential backoff
- **JetStream Support** - First-class support for NATS JetStream
- **TLS/SSL Support** - Optional encryption with certificate management
- **Functional Options** - Clean configuration with `WithURL()`, `WithUsername()`, etc.
- **Authentication** - Support for username/password and token authentication
- **Environment Config** - Load configuration from environment variables
- **Thread-Safe** - All operations are safe for concurrent use
- **Connection Monitoring** - Built-in connection state callbacks and monitoring

## Installation

```bash
go get github.com/isimtekin/go-packages/nats-client@v0.0.1
```

## Quick Start

### Basic Publish/Subscribe

```go
package main

import (
    "context"
    "log"

    natsclient "github.com/isimtekin/go-packages/nats-client"
)

func main() {
    ctx := context.Background()

    // Create client
    client, err := natsclient.NewWithOptions(
        natsclient.WithURL("nats://localhost:4222"),
        natsclient.WithName("my-service"),
    )
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // Subscribe to a subject
    sub, err := client.Subscribe("orders.created", func(msg *nats.Msg) {
        log.Printf("Received: %s", string(msg.Data))
    })
    if err != nil {
        log.Fatal(err)
    }
    defer sub.Unsubscribe()

    // Publish a message
    err = client.Publish("orders.created", []byte("order-123"))
    if err != nil {
        log.Fatal(err)
    }
}
```

### Request/Reply Pattern

```go
// Responder
sub, err := client.Subscribe("get.user", func(msg *nats.Msg) {
    user := `{"id": 123, "name": "John"}`
    msg.Respond([]byte(user))
})
defer sub.Unsubscribe()

// Requester
response, err := client.Request("get.user", []byte("123"), 2*time.Second)
if err != nil {
    log.Fatal(err)
}
log.Printf("User: %s", string(response.Data))
```

### Queue Groups

```go
// Multiple subscribers in the same queue group share the work
for i := 0; i < 3; i++ {
    client.QueueSubscribe("tasks", "workers", func(msg *nats.Msg) {
        log.Printf("Worker processing: %s", string(msg.Data))
        // Process task
    })
}
```

## Configuration

### Using Functional Options

```go
client, err := natsclient.NewWithOptions(
    natsclient.WithURL("nats://localhost:4222"),
    natsclient.WithName("my-service"),
    natsclient.WithUsername("user"),
    natsclient.WithPassword("password"),
    natsclient.WithMaxReconnects(10),
    natsclient.WithReconnectWait(2*time.Second),
    natsclient.WithTimeout(5*time.Second),
    natsclient.WithJetStream(true),
)
```

### Using Config Struct

```go
config := &natsclient.Config{
    URL:            "nats://localhost:4222",
    Name:           "my-service",
    Username:       "user",
    Password:       "password",
    MaxReconnects:  10,
    ReconnectWait:  2 * time.Second,
    Timeout:        5 * time.Second,
    AllowReconnect: true,
}

client, err := natsclient.New(config)
```

### From Environment Variables

```bash
# .env file
NATS_URL=nats://localhost:4222
NATS_NAME=my-service
NATS_USERNAME=user
NATS_PASSWORD=secret
NATS_MAX_RECONNECTS=10
NATS_TIMEOUT=5s
NATS_ENABLE_JETSTREAM=true
```

```go
client, err := natsclient.NewFromEnvWithDefaults(ctx)
if err != nil {
    log.Fatal(err)
}
defer client.Close()
```

## Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| URL | string | nats://localhost:4222 | NATS server URL |
| Name | string | nats-client | Client name |
| Username | string | - | Username for authentication |
| Password | string | - | Password for authentication |
| Token | string | - | Token for authentication |
| MaxReconnects | int | 60 | Maximum reconnection attempts (-1 for unlimited) |
| ReconnectWait | duration | 2s | Wait time between reconnects |
| ReconnectJitter | duration | 100ms | Random jitter for reconnects |
| Timeout | duration | 2s | Connection timeout |
| PingInterval | duration | 2m | Interval between pings |
| MaxPingsOut | int | 2 | Max outstanding pings |
| AllowReconnect | bool | true | Enable automatic reconnection |
| NoRandomize | bool | false | Disable server randomization |
| NoEcho | bool | false | Disable echo of published messages |
| RetryOnFailedConn | bool | true | Retry on failed connection |
| TLSEnabled | bool | false | Enable TLS/SSL |
| TLSCertFile | string | - | Path to TLS certificate |
| TLSKeyFile | string | - | Path to TLS key |
| TLSCAFile | string | - | Path to TLS CA certificate |
| EnableJetStream | bool | false | Enable JetStream support |

## API Reference

### Client Methods

#### Connection Management

```go
// New creates a new client with the given configuration
func New(config *Config) (*Client, error)

// NewWithOptions creates a new client using functional options
func NewWithOptions(opts ...Option) (*Client, error)

// NewFromEnvWithDefaults creates a client from environment variables with NATS_ prefix
func NewFromEnvWithDefaults(ctx context.Context) (*Client, error)

// Close closes the connection
func (c *Client) Close() error

// IsClosed returns true if the connection is closed
func (c *Client) IsClosed() bool

// IsConnected returns true if currently connected
func (c *Client) IsConnected() bool

// IsReconnecting returns true if currently reconnecting
func (c *Client) IsReconnecting() bool
```

#### Publishing

```go
// Publish publishes a message to a subject
func (c *Client) Publish(subject string, data []byte) error

// PublishMsg publishes a NATS message
func (c *Client) PublishMsg(msg *nats.Msg) error

// PublishRequest publishes a request with a reply subject
func (c *Client) PublishRequest(subject, reply string, data []byte) error

// Flush flushes any buffered messages
func (c *Client) Flush() error
```

#### Subscribing

```go
// Subscribe creates a subscription
func (c *Client) Subscribe(subject string, handler nats.MsgHandler) (*nats.Subscription, error)

// QueueSubscribe creates a queue group subscription
func (c *Client) QueueSubscribe(subject, queue string, handler nats.MsgHandler) (*nats.Subscription, error)

// ChanSubscribe creates a channel subscription
func (c *Client) ChanSubscribe(subject string, ch chan *nats.Msg) (*nats.Subscription, error)

// QueueChanSubscribe creates a queue channel subscription
func (c *Client) QueueChanSubscribe(subject, queue string, ch chan *nats.Msg) (*nats.Subscription, error)
```

#### Request/Reply

```go
// Request sends a request and waits for a reply
func (c *Client) Request(subject string, data []byte, timeout time.Duration) (*nats.Msg, error)

// RequestMsg sends a request message and waits for a reply
func (c *Client) RequestMsg(msg *nats.Msg, timeout time.Duration) (*nats.Msg, error)
```

#### JetStream

```go
// JetStream returns a JetStream context (if enabled)
func (c *Client) JetStream() (nats.JetStreamContext, error)
```

## Environment Variables

Default prefix: `NATS_`

**Connection Settings:**
- `NATS_URL` - Server URL (default: nats://localhost:4222)
- `NATS_NAME` - Client name
- `NATS_USERNAME` - Username for auth
- `NATS_PASSWORD` - Password for auth
- `NATS_TOKEN` - Token for auth

**Reconnection Settings:**
- `NATS_MAX_RECONNECTS` - Max reconnection attempts
- `NATS_RECONNECT_WAIT` - Wait between reconnects (e.g., "2s")
- `NATS_RECONNECT_JITTER` - Jitter for reconnects (e.g., "100ms")
- `NATS_ALLOW_RECONNECT` - Enable auto-reconnect (true/false)
- `NATS_RETRY_ON_FAILED_CONN` - Retry on failed connection (true/false)

**Timeout Settings:**
- `NATS_TIMEOUT` - Connection timeout (e.g., "5s")
- `NATS_PING_INTERVAL` - Ping interval (e.g., "2m")
- `NATS_MAX_PINGS_OUT` - Max outstanding pings

**TLS Settings:**
- `NATS_TLS_ENABLED` - Enable TLS (true/false)
- `NATS_TLS_CERT_FILE` - Path to certificate
- `NATS_TLS_KEY_FILE` - Path to key
- `NATS_TLS_CA_FILE` - Path to CA certificate

**JetStream:**
- `NATS_ENABLE_JETSTREAM` - Enable JetStream (true/false)

See `.env.example` for complete list.

## Use Cases

### Microservices Communication

```go
// Service A - Publisher
client.Publish("user.created", []byte(`{"id": 123, "name": "John"}`))

// Service B - Subscriber
client.Subscribe("user.created", func(msg *nats.Msg) {
    var user User
    json.Unmarshal(msg.Data, &user)
    // Process user creation
})
```

### Load Balancing with Queue Groups

```go
// Multiple instances of the same service
// Only one instance receives each message
client.QueueSubscribe("tasks", "task-workers", func(msg *nats.Msg) {
    processTask(msg.Data)
})
```

### Request/Reply RPC Pattern

```go
// Service - Responder
client.Subscribe("math.add", func(msg *nats.Msg) {
    var req struct{ A, B int }
    json.Unmarshal(msg.Data, &req)

    result := req.A + req.B
    response, _ := json.Marshal(result)
    msg.Respond(response)
})

// Client - Requester
req, _ := json.Marshal(map[string]int{"a": 5, "b": 3})
resp, err := client.Request("math.add", req, 1*time.Second)
// resp.Data contains the result
```

### Event Sourcing with JetStream

```go
client, _ := natsclient.NewWithOptions(
    natsclient.WithURL("nats://localhost:4222"),
    natsclient.WithJetStream(true),
)

js, _ := client.JetStream()

// Create stream
js.AddStream(&nats.StreamConfig{
    Name:     "EVENTS",
    Subjects: []string{"events.>"},
})

// Publish event
js.Publish("events.order.created", []byte("order-123"))

// Subscribe to events
js.Subscribe("events.>", func(msg *nats.Msg) {
    msg.Ack()
    // Process event
})
```

## Error Handling

```go
_, err := client.Publish("subject", data)
if err != nil {
    if natsclient.IsConnectionError(err) {
        // Handle connection errors
        log.Println("Connection issue:", err)
    } else if natsclient.IsTimeoutError(err) {
        // Handle timeout errors
        log.Println("Timeout:", err)
    } else {
        // Handle other errors
        log.Println("Error:", err)
    }
}
```

**Common Errors:**
- `ErrClientClosed` - Client has been closed
- `ErrConnectionFailed` - Failed to establish connection
- `ErrTimeout` - Operation timed out
- `ErrNoResponders` - No responders available for request
- `ErrInvalidSubject` - Invalid subject name
- `ErrSlowConsumer` - Consumer too slow, messages dropped

## Connection Callbacks

```go
config := natsclient.DefaultConfig()

// Set callbacks for connection events
opts := []nats.Option{
    nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
        log.Printf("Disconnected: %v", err)
    }),
    nats.ReconnectHandler(func(nc *nats.Conn) {
        log.Println("Reconnected")
    }),
    nats.ClosedHandler(func(nc *nats.Conn) {
        log.Println("Connection closed")
    }),
}

client, _ := natsclient.New(config)
```

## Testing

```bash
# Run all tests
go test -v ./...

# Run with race detection
go test -v -race ./...

# Run with coverage
go test -v -cover -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Examples

See the `examples/` directory for complete working examples:
- Basic publish/subscribe
- Request/reply patterns
- Queue groups
- JetStream usage
- Environment configuration

## License

MIT License - see [LICENSE](../LICENSE) file for details.

## Related Packages

- [env-util](../env-util) - Environment variable utilities
- [mongo-client](../mongo-client) - MongoDB client wrapper
- [redis-client](../redis-client) - Redis client wrapper

## Resources

- [NATS Documentation](https://docs.nats.io/)
- [nats.go Client](https://github.com/nats-io/nats.go)
- [JetStream Guide](https://docs.nats.io/nats-concepts/jetstream)
