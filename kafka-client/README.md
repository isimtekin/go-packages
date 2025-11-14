# Kafka Client

A robust, idiomatic Go client library for Apache Kafka with producer and consumer capabilities.

## Features

- **Producer**: Send single or batch messages with configurable compression and partitioning
- **Consumer**: Consumer group support with automatic offset management
- **Admin Operations**: Create/delete topics, list topics, get metadata
- **Flexible Configuration**: Support for SASL authentication and TLS encryption
- **Idempotent Writes**: Exactly-once semantics for producers
- **Multiple Compression Codecs**: Snappy, GZIP, LZ4, Zstd
- **Context Support**: All operations support context for timeouts and cancellation
- **Thread-Safe**: Safe for concurrent use

## Installation

\`\`\`bash
go get github.com/isimtekin/go-packages/kafka-client
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
