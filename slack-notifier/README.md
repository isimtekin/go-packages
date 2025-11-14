# Slack Notifier

A robust and easy-to-use Go client for sending notifications to Slack via incoming webhooks. Features include retry logic, message builders, Block Kit support, and comprehensive error handling.

## Features

- **Simple API**: Easy-to-use client with sensible defaults
- **Webhook Integration**: Send messages via Slack incoming webhooks
- **Message Builder Pattern**: Fluent API for constructing messages
- **Block Kit Support**: Create rich, interactive messages with Slack's Block Kit
- **Color-coded Messages**: Built-in support for success, warning, error, and info messages
- **Retry Logic**: Automatic retry with exponential backoff for failed requests
- **Context Support**: Full context.Context integration for timeouts and cancellation
- **Thread Support**: Send messages in threads for better organization
- **Customization**: Configure username, icon, channel, and more
- **Type-safe**: Strongly-typed message structures
- **Well-tested**: 81.8% test coverage with comprehensive test suite

## Installation

```bash
go get github.com/isimtekin/go-packages/slack-notifier
```

## Quick Start

### Basic Usage

```go
package main

import (
    "context"
    "log"

    notifier "github.com/isimtekin/go-packages/slack-notifier"
)

func main() {
    // Create a new client
    client, err := notifier.NewWithOptions(
        notifier.WithWebhookURL("https://hooks.slack.com/services/YOUR/WEBHOOK/URL"),
        notifier.WithChannel("#general"),
        notifier.WithUsername("My Bot"),
        notifier.WithIconEmoji(":robot_face:"),
    )
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    ctx := context.Background()

    // Send a simple text message
    if err := client.SendText(ctx, "Hello, Slack!"); err != nil {
        log.Printf("Failed to send message: %v", err)
    }
}
```

### Color-coded Messages

```go
// Send a success message (green)
client.SendSuccess(ctx, "Deployment completed successfully!")

// Send a warning message (yellow)
client.SendWarning(ctx, "Disk usage is at 85%")

// Send an error message (red)
client.SendError(ctx, "Database connection failed")

// Send an info message (blue)
client.SendInfo(ctx, "Scheduled maintenance in 1 hour")
```

### Using Message Builder

```go
message := notifier.NewMessage("Deployment Status").
    Channel("#deploys").
    Username("Deploy Bot").
    IconEmoji(":rocket:").
    AddAttachment(notifier.Attachment{
        Fallback: "Deployment completed",
        Text:     "Application v2.0.0 deployed successfully",
        Color:    notifier.ColorGood,
        Fields: []notifier.AttachmentField{
            {Title: "Environment", Value: "Production", Short: true},
            {Title: "Version", Value: "v2.0.0", Short: true},
            {Title: "Duration", Value: "5m 32s", Short: true},
            {Title: "Status", Value: "Success", Short: true},
        },
    }).
    Build()

if err := client.Send(ctx, message); err != nil {
    log.Printf("Failed to send: %v", err)
}
```

### Using Block Kit

```go
blocks := []notifier.Block{
    notifier.NewHeaderBlock("Production Alert"),
    notifier.NewSectionBlock("*High CPU Usage Detected*\nServer: web-01\nCPU: 95%"),
    notifier.NewDividerBlock(),
    notifier.NewSectionBlock("Action required: Please investigate immediately."),
}

if err := client.SendWithBlocks(ctx, blocks); err != nil {
    log.Printf("Failed to send: %v", err)
}
```

### Thread Support

```go
// Send a message and get thread timestamp
// (Note: Getting thread_ts requires Slack API, not webhooks)
// You can set a known thread_ts to reply in a thread

client, _ := notifier.NewWithOptions(
    notifier.WithWebhookURL("https://hooks.slack.com/services/YOUR/WEBHOOK/URL"),
    notifier.WithThreadTS("1234567890.123456"), // Reply in thread
)

client.SendText(ctx, "This is a threaded reply")
```

## Configuration

### Using Config Struct

```go
config := &notifier.Config{
    WebhookURL:       "https://hooks.slack.com/services/YOUR/WEBHOOK/URL",
    DefaultChannel:   "#general",
    DefaultUsername:  "My Bot",
    DefaultIconEmoji: ":robot_face:",
    Timeout:          30 * time.Second,
    MaxRetries:       3,
    RetryDelay:       time.Second,
    EnableDebug:      false,
}

client, err := notifier.New(config)
```

### Using Functional Options

```go
client, err := notifier.NewWithOptions(
    notifier.WithWebhookURL("https://hooks.slack.com/services/YOUR/WEBHOOK/URL"),
    notifier.WithChannel("#general"),
    notifier.WithUsername("My Bot"),
    notifier.WithIconEmoji(":robot_face:"),
    notifier.WithIconURL("https://example.com/bot-icon.png"),
    notifier.WithTimeout(30*time.Second),
    notifier.WithMaxRetries(3),
    notifier.WithRetryDelay(time.Second),
    notifier.WithDebug(true),
    notifier.WithThreadTS("1234567890.123456"),
)
```

## Message Structures

### Simple Message

```go
message := &notifier.Message{
    Text:     "Hello, Slack!",
    Channel:  "#general",
    Username: "My Bot",
}
```

### Message with Attachments

```go
message := &notifier.Message{
    Text: "Deployment Report",
    Attachments: []notifier.Attachment{
        {
            Fallback:   "Deployment completed",
            Color:      notifier.ColorGood,
            Title:      "Production Deployment",
            TitleLink:  "https://example.com/deploys/123",
            Text:       "Application deployed successfully",
            AuthorName: "Deploy Bot",
            Fields: []notifier.AttachmentField{
                {Title: "Environment", Value: "Production", Short: true},
                {Title: "Version", Value: "v2.0.0", Short: true},
            },
            Footer:    "Deployment System",
            Timestamp: time.Now().Unix(),
        },
    },
}
```

### Message with Block Kit

```go
message := &notifier.Message{
    Blocks: []notifier.Block{
        {
            Type: "header",
            Text: &notifier.TextObject{
                Type: "plain_text",
                Text: "New Feature Released",
            },
        },
        {
            Type: "section",
            Text: &notifier.TextObject{
                Type: "mrkdwn",
                Text: "*Feature Name*: Dark Mode\n*Released*: Today",
            },
        },
        {
            Type: "divider",
        },
    },
}
```

## Error Handling

```go
err := client.SendText(ctx, "Test message")
if err != nil {
    if notifier.IsConnectionError(err) {
        log.Println("Connection error:", err)
    } else if notifier.IsTimeoutError(err) {
        log.Println("Timeout error:", err)
    } else {
        log.Println("Other error:", err)
    }
}
```

## Available Errors

- `ErrClientClosed`: Client has been closed
- `ErrAlreadyClosed`: Attempting to close an already closed client
- `ErrConnectionFailed`: Failed to connect to Slack webhook
- `ErrTimeout`: Operation timed out
- `ErrInvalidResponse`: Invalid response from Slack
- `ErrEmptyWebhookURL`: Webhook URL is empty
- `ErrInvalidMessage`: Message validation failed

## Color Constants

```go
notifier.ColorGood    // Green
notifier.ColorWarning // Yellow
notifier.ColorDanger  // Red
notifier.ColorInfo    // Blue-green
```

## Context and Timeouts

```go
// Use context with timeout
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

err := client.SendText(ctx, "Message with timeout")
```

## Testing

The package includes a comprehensive test suite with 81.8% coverage.

```bash
# Run tests
make test

# Run tests with coverage
make test-coverage

# Run tests with race detector
make test-race

# Generate HTML coverage report
make coverage-html
```

## Examples

See the `examples/` directory for more usage examples.

## Requirements

- Go 1.18 or higher
- Slack incoming webhook URL

## Setting Up Slack Webhook

1. Go to your Slack workspace
2. Navigate to Apps ’ Incoming Webhooks
3. Click "Add to Slack"
4. Choose a channel and click "Add Incoming WebHooks Integration"
5. Copy the webhook URL

## Best Practices

1. **Store webhook URL securely**: Use environment variables or secrets management
2. **Use context with timeouts**: Prevent hanging operations
3. **Handle errors appropriately**: Check for connection and timeout errors
4. **Close the client**: Always defer `client.Close()`
5. **Use retry logic**: Configure appropriate retry settings for your use case

## Limitations

- This package uses Slack incoming webhooks, which have some limitations:
  - Cannot retrieve thread timestamps (thread_ts)
  - Cannot read messages
  - Cannot upload files
  - Rate-limited to 1 message per second per webhook
- For more advanced features, consider using the official Slack API

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License.

## Related Packages

- [kafka-client](../kafka-client) - Kafka client with producer, consumer, and admin operations
- [nats-client](../nats-client) - NATS messaging client
- [mongo-client](../mongo-client) - MongoDB client wrapper
- [redis-client](../redis-client) - Redis client with connection pooling
- [crypto-utils](../crypto-utils) - Cryptographic utilities

## Support

For issues, questions, or contributions, please visit the [GitHub repository](https://github.com/isimtekin/go-packages).
