# Kafka Client

A robust, idiomatic Go client library for Apache Kafka with producer and consumer capabilities.

## Features

- **Producer**: Send single or batch messages with configurable compression and partitioning
- **Consumer**: Consumer group support with automatic offset management
- **Admin Operations**: Create/delete topics, list topics, get metadata
- **Workspace Support**: Multi-tenancy and environment separation with topic prefixing
- **Flexible Configuration**: Support for SASL authentication and TLS encryption
- **Idempotent Writes**: Exactly-once semantics for producers
- **Multiple Compression Codecs**: Snappy, GZIP, LZ4, Zstd
- **Context Support**: All operations support context for timeouts and cancellation
- **Thread-Safe**: Safe for concurrent use

## Installation

\`\`\`bash
go get github.com/isimtekin/go-packages/kafka-client@v0.1.0
\`\`\`

## Quick Start

### Producer Example

\`\`\`go
package main

import (
    "context"
    "log"

    kafkaclient "github.com/isimtekin/go-packages/kafka-client"
)

func main() {
    // Create client with default config
    client, err := kafkaclient.New(kafkaclient.DefaultConfig())
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // Send a message
    ctx := context.Background()
    msg := &kafkaclient.Message{
        Topic: "my-topic",
        Key:   []byte("key1"),
        Value: []byte("Hello, Kafka!"),
        Headers: map[string]string{
            "content-type": "text/plain",
        },
    }

    partition, offset, err := client.SendMessage(ctx, msg)
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Message sent to partition %d at offset %d\n", partition, offset)
}
\`\`\`

### Consumer Example

\`\`\`go
package main

import (
    "context"
    "log"

    kafkaclient "github.com/isimtekin/go-packages/kafka-client"
)

func main() {
    // Configure consumer
    config := kafkaclient.DefaultConfig()
    config.Consumer.Topics = []string{"my-topic"}
    config.Consumer.GroupID = "my-consumer-group"

    // Message handler function
    handler := func(ctx context.Context, msg *kafkaclient.ConsumedMessage) error {
        log.Printf("Received message: %s\n", string(msg.Value))
        return nil
    }

    // Create client with consumer
    client, err := kafkaclient.NewWithConsumer(config, handler)
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // Consumer runs in background
    // Monitor for errors
    for err := range client.ConsumerErrors() {
        log.Printf("Consumer error: %v\n", err)
    }
}
\`\`\`

## License

MIT License - see [LICENSE](../LICENSE) file for details.

### Workspace Support

Use workspaces to separate topics by environment, tenant, or team. When a workspace is configured, all topic names are automatically prefixed.

#### Producer with Workspace

```go
package main

import (
    "context"
    "log"

    kafkaclient "github.com/isimtekin/go-packages/kafka-client"
)

func main() {
    // Create client with workspace for production environment
    client, err := kafkaclient.NewWithOptions(
        kafkaclient.WithBrokers([]string{"localhost:9092"}),
        kafkaclient.WithWorkspace("production"),  // All topics will be prefixed with "production."
    )
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    ctx := context.Background()

    // Send to "orders" topic -> actually sends to "production.orders"
    msg := &kafkaclient.Message{
        Topic: "orders",
        Value: []byte("Order data"),
    }

    partition, offset, err := client.SendMessage(ctx, msg)
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Message sent to production.orders at partition %d, offset %d\n", partition, offset)
}
```

#### Multi-Environment Setup

```go
// Development environment
devClient, _ := kafkaclient.NewWithOptions(
    kafkaclient.WithWorkspace("dev"),
    // sends to: dev.orders, dev.users, dev.events
)

// Staging environment
stagingClient, _ := kafkaclient.NewWithOptions(
    kafkaclient.WithWorkspace("staging"),
    // sends to: staging.orders, staging.users, staging.events
)

// Production environment
prodClient, _ := kafkaclient.NewWithOptions(
    kafkaclient.WithWorkspace("production"),
    // sends to: production.orders, production.users, production.events
)
```

#### Multi-Tenancy

```go
// Tenant-specific topics
tenant1Client, _ := kafkaclient.NewWithOptions(
    kafkaclient.WithWorkspace("tenant-123"),
    // sends to: tenant-123.orders, tenant-123.events
)

tenant2Client, _ := kafkaclient.NewWithOptions(
    kafkaclient.WithWorkspace("tenant-456"),
    // sends to: tenant-456.orders, tenant-456.events
)
```

#### Consumer with Workspace

```go
// Consumer will subscribe to workspace-prefixed topics
config := kafkaclient.DefaultConfig()
config.Workspace = "production"
config.Consumer.Topics = []string{"orders", "events"}  // Actually subscribes to: production.orders, production.events
config.Consumer.GroupID = "order-processor"

handler := func(ctx context.Context, msg *kafkaclient.ConsumedMessage) error {
    log.Printf("Received from %s: %s\n", string(msg.Value))
    return nil
}

client, err := kafkaclient.NewWithConsumer(config, handler)
if err != nil {
    log.Fatal(err)
}
defer client.Close()

// Consume messages from production.orders and production.events
for err := range client.ConsumerErrors() {
    log.Printf("Consumer error: %v\n", err)
}
```

**Benefits of Workspace:**
- **Environment Isolation**: Separate dev, staging, and production topics on the same cluster
- **Multi-Tenancy**: Isolate data for different tenants/customers
- **Team Separation**: Different teams can use the same topic names without conflicts
- **Easy Migration**: Move between environments by changing workspace configuration
- **Cost Efficiency**: Share Kafka infrastructure across environments
