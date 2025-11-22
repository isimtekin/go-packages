# mail-sender

A simple and flexible email sending library for Go with support for multiple email service providers. Currently supports SendGrid with an extensible architecture for adding more providers.

## Features

- **Multi-Provider Support**: Abstract interface for different email providers (SendGrid supported, more coming soon)
- **SendGrid Integration**: Full support for SendGrid email service
- **Async/Event-Based Sending**: Non-blocking email sending with worker pools and event handlers
- **Template Support**: Built-in HTML and plain text template rendering using Go templates
- **Flexible Configuration**: Configure via code, functional options, or environment variables
- **Rich Email Features**:
  - Plain text and HTML emails
  - Multiple recipients (To, Cc, Bcc)
  - Custom reply-to addresses
  - Sender name customization
- **Advanced Async Features**:
  - Worker pool for concurrent email sending
  - Event handlers (OnSuccess, OnFailure, OnRetry)
  - Automatic retry logic with configurable attempts and delays
  - Graceful shutdown with timeout support
  - Real-time statistics (sent, failed, pending, retried)
  - Queue-based architecture with configurable buffer size
- **Simple API**: Easy-to-use interface with sensible defaults
- **Well Tested**: Comprehensive unit tests with >90% coverage

## Installation

```bash
go get github.com/isimtekin/go-packages/mail-sender@v0.1.1
```

## Quick Start

### Basic Usage with SendGrid

```go
package main

import (
    "context"
    "log"

    mailsender "github.com/isimtekin/go-packages/mail-sender"
)

func main() {
    // Create a SendGrid sender
    sender, err := mailsender.NewSendGridWithOptions(
        mailsender.WithAPIKey("your-sendgrid-api-key"),
        mailsender.WithDefaultFrom("sender@example.com"),
        mailsender.WithDefaultFromName("My App"),
    )
    if err != nil {
        log.Fatal(err)
    }
    defer sender.Close()

    // Send an email
    err = sender.Send(context.Background(), &mailsender.EmailMessage{
        To:        []string{"recipient@example.com"},
        Subject:   "Hello from mail-sender!",
        PlainText: "This is a test email.",
    })
    if err != nil {
        log.Fatal(err)
    }
}
```

### Using Environment Variables

```go
// Set environment variables:
// SENDGRID_API_KEY=your-api-key
// SENDGRID_DEFAULT_FROM=sender@example.com
// SENDGRID_DEFAULT_FROM_NAME=My App

sender, err := mailsender.NewSendGridFromEnv()
if err != nil {
    log.Fatal(err)
}
defer sender.Close()

err = sender.Send(context.Background(), &mailsender.EmailMessage{
    To:        []string{"recipient@example.com"},
    Subject:   "Hello!",
    PlainText: "Configured from environment variables.",
})
```

### Async/Non-Blocking Email Sending

For high-throughput scenarios, use the async sender with worker pools:

```go
// Create a SendGrid sender
sender, err := mailsender.NewSendGridWithOptions(
    mailsender.WithAPIKey("your-api-key"),
    mailsender.WithDefaultFrom("sender@example.com"),
)
if err != nil {
    log.Fatal(err)
}

// Create async sender with 5 workers and event handlers
asyncSender := mailsender.NewAsyncSender(sender,
    mailsender.WithWorkers(5),
    mailsender.WithQueueSize(100),
    mailsender.WithRetry(3, time.Second),
    mailsender.WithOnSuccess(func(msg *mailsender.EmailMessage) {
        log.Printf("Email sent to %v", msg.To)
    }),
    mailsender.WithOnFailure(func(msg *mailsender.EmailMessage, err error) {
        log.Printf("Failed to send email: %v", err)
    }),
)
defer asyncSender.Close() // Gracefully waits for all queued emails

// Send emails asynchronously (non-blocking)
for i := 0; i < 100; i++ {
    err := asyncSender.SendAsync(context.Background(), &mailsender.EmailMessage{
        To:        []string{fmt.Sprintf("user%d@example.com", i)},
        Subject:   "Bulk Email",
        PlainText: "This is sent asynchronously!",
    })
    if err != nil {
        log.Printf("Failed to queue email: %v", err)
    }
}

// Check statistics
stats := asyncSender.Stats()
fmt.Printf("Sent: %d, Failed: %d, Pending: %d\n",
    stats.Sent, stats.Failed, stats.Pending)
```

## Usage Examples

### Sending HTML Email

```go
err := sender.Send(context.Background(), &mailsender.EmailMessage{
    To:      []string{"recipient@example.com"},
    Subject: "HTML Email",
    HTML:    "<h1>Hello World</h1><p>This is an <strong>HTML</strong> email.</p>",
})
```

### Sending Email with Both Plain Text and HTML

```go
err := sender.Send(context.Background(), &mailsender.EmailMessage{
    To:        []string{"recipient@example.com"},
    Subject:   "Multi-part Email",
    PlainText: "This is the plain text version.",
    HTML:      "<h1>This is the HTML version</h1>",
})
```

### Multiple Recipients

```go
err := sender.Send(context.Background(), &mailsender.EmailMessage{
    To:        []string{"recipient1@example.com", "recipient2@example.com"},
    Cc:        []string{"cc@example.com"},
    Bcc:       []string{"bcc@example.com"},
    Subject:   "Team Update",
    PlainText: "This goes to the whole team.",
})
```

### Custom Reply-To Address

```go
err := sender.Send(context.Background(), &mailsender.EmailMessage{
    To:        []string{"customer@example.com"},
    Subject:   "Customer Support",
    PlainText: "Please reply to our support address.",
    ReplyTo:   "support@example.com",
})
```

### Using Templates

#### HTML Template

```go
htmlTemplate := `
<html>
    <body>
        <h1>Hello {{.Name}}!</h1>
        <p>Your order #{{.OrderID}} has been confirmed.</p>
        <p>Total: ${{.Total}}</p>
    </body>
</html>
`

data := map[string]interface{}{
    "Name":    "John Doe",
    "OrderID": "12345",
    "Total":   "99.99",
}

htmlContent, err := mailsender.RenderHTMLTemplate(htmlTemplate, data)
if err != nil {
    log.Fatal(err)
}

err = sender.Send(context.Background(), &mailsender.EmailMessage{
    To:      []string{"customer@example.com"},
    Subject: "Order Confirmation",
    HTML:    htmlContent,
})
```

#### Plain Text Template

```go
textTemplate := `
Hello {{.Name}},

Your order has been confirmed!

Order Details:
{{range .Items}}
- {{.}}
{{end}}

Total: ${{.Total}}

Thanks for your purchase!
`

data := map[string]interface{}{
    "Name":  "Jane Smith",
    "Items": []string{"Product A", "Product B"},
    "Total": "149.99",
}

textContent, err := mailsender.RenderTextTemplate(textTemplate, data)
if err != nil {
    log.Fatal(err)
}

err = sender.Send(context.Background(), &mailsender.EmailMessage{
    To:        []string{"customer@example.com"},
    Subject:   "Order Confirmation",
    PlainText: textContent,
})
```

#### Template with Conditionals

```go
template := `
Hello {{.Name}},

{{if .Premium}}
Thank you for being a premium member!
{{else}}
Consider upgrading to premium for exclusive benefits.
{{end}}

{{if gt .MessageCount 0}}
You have {{.MessageCount}} new messages.
{{end}}
`

data := map[string]interface{}{
    "Name":         "Alice",
    "Premium":      true,
    "MessageCount": 5,
}

content, err := mailsender.RenderTextTemplate(template, data)
```

## Configuration

### Configuration Options

```go
type Config struct {
    Provider        Provider  // Email provider (e.g., "sendgrid")
    APIKey          string    // Provider API key
    DefaultFrom     string    // Default sender email (optional)
    DefaultFromName string    // Default sender name (optional)
}
```

### Functional Options

```go
sender, err := mailsender.NewSendGridWithOptions(
    mailsender.WithAPIKey("your-api-key"),
    mailsender.WithDefaultFrom("sender@example.com"),
    mailsender.WithDefaultFromName("My Application"),
)
```

Available options:
- `WithProvider(provider)` - Set the email provider
- `WithAPIKey(key)` - Set the API key
- `WithDefaultFrom(email)` - Set default sender email
- `WithDefaultFromName(name)` - Set default sender name

### Environment Variables

#### SendGrid-specific (recommended)

```bash
SENDGRID_API_KEY=your-sendgrid-api-key
SENDGRID_DEFAULT_FROM=sender@example.com
SENDGRID_DEFAULT_FROM_NAME=Your App Name
```

#### Generic Email Configuration

```bash
EMAIL_PROVIDER=sendgrid
EMAIL_API_KEY=your-api-key
EMAIL_DEFAULT_FROM=sender@example.com
EMAIL_DEFAULT_FROM_NAME=Your App Name
```

## Email Message Structure

```go
type EmailMessage struct {
    From      string   // Sender email address
    FromName  string   // Sender name (optional)
    To        []string // Recipient email addresses (required)
    Cc        []string // Carbon copy recipients (optional)
    Bcc       []string // Blind carbon copy recipients (optional)
    Subject   string   // Email subject (required)
    PlainText string   // Plain text body
    HTML      string   // HTML body
    ReplyTo   string   // Reply-to address (optional)
}
```

**Validation Rules:**
- `From` is required (can be set via defaults)
- `To` must have at least one recipient
- `Subject` is required
- Either `PlainText` or `HTML` (or both) must be provided

## Async Email Sending

The async sender provides non-blocking email sending with worker pools, event handlers, and automatic retry logic.

### Basic Async Usage

```go
// Create async sender
asyncSender := mailsender.NewAsyncSender(sender,
    mailsender.WithWorkers(5),
    mailsender.WithQueueSize(100),
)
defer asyncSender.Close()

// Send emails asynchronously (returns immediately)
err := asyncSender.SendAsync(ctx, message)
```

### Async Configuration Options

```go
// WithWorkers sets the number of concurrent workers (default: 3)
mailsender.WithWorkers(10)

// WithQueueSize sets the queue buffer size (default: 100)
mailsender.WithQueueSize(1000)

// WithRetry sets retry attempts and delay (default: 0 retries)
mailsender.WithRetry(3, 500*time.Millisecond)

// WithEventHandlers sets all event handlers at once
mailsender.WithEventHandlers(&mailsender.EventHandlers{
    OnSuccess: func(msg *mailsender.EmailMessage) { /* ... */ },
    OnFailure: func(msg *mailsender.EmailMessage, err error) { /* ... */ },
    OnRetry:   func(msg *mailsender.EmailMessage, attempt int, err error) { /* ... */ },
})

// Or set individual handlers
mailsender.WithOnSuccess(func(msg *mailsender.EmailMessage) {
    log.Printf("Email sent to %v", msg.To)
})

mailsender.WithOnFailure(func(msg *mailsender.EmailMessage, err error) {
    log.Printf("Failed to send: %v", err)
})

mailsender.WithOnRetry(func(msg *mailsender.EmailMessage, attempt int, err error) {
    log.Printf("Retrying (attempt %d): %v", attempt, err)
})
```

### Event Handlers

Event handlers allow you to react to email sending events:

```go
asyncSender := mailsender.NewAsyncSender(sender,
    mailsender.WithOnSuccess(func(msg *mailsender.EmailMessage) {
        // Called when email is sent successfully
        metrics.IncrementEmailsSent()
    }),
    mailsender.WithOnFailure(func(msg *mailsender.EmailMessage, err error) {
        // Called when email fails after all retries
        alerting.NotifyFailure(msg, err)
    }),
    mailsender.WithOnRetry(func(msg *mailsender.EmailMessage, attempt int, err error) {
        // Called before each retry attempt
        log.Printf("Retry %d for %v: %v", attempt, msg.To, err)
    }),
)
```

### Retry Logic

Configure automatic retries for failed sends:

```go
asyncSender := mailsender.NewAsyncSender(sender,
    mailsender.WithRetry(3, 2*time.Second), // 3 retries, 2s delay between retries
)
```

The sender will automatically retry failed sends up to the specified number of attempts with the configured delay between each retry.

### Statistics

Track email sending metrics in real-time:

```go
stats := asyncSender.Stats()
fmt.Printf("Sent: %d\n", stats.Sent)
fmt.Printf("Failed: %d\n", stats.Failed)
fmt.Printf("Pending: %d\n", stats.Pending)
fmt.Printf("Retried: %d\n", stats.Retried)
```

### Graceful Shutdown

The async sender supports graceful shutdown:

```go
// Close() waits for all queued emails to be sent
err := asyncSender.Close()

// CloseWithTimeout() forces close after timeout
err := asyncSender.CloseWithTimeout(5 * time.Second)
```

### SendAsync vs SendAsyncBlocking

```go
// SendAsync returns immediately if queue is full (non-blocking)
err := asyncSender.SendAsync(ctx, message)
if err != nil {
    // Queue is full or sender is closed
}

// SendAsyncBlocking waits for queue space (blocking)
err := asyncSender.SendAsyncBlocking(ctx, message)
// Blocks until message is queued or context is cancelled
```

### Complete Async Example

```go
// Create async sender with all features
asyncSender := mailsender.NewAsyncSender(sender,
    mailsender.WithWorkers(5),
    mailsender.WithQueueSize(100),
    mailsender.WithRetry(3, time.Second),
    mailsender.WithOnSuccess(func(msg *mailsender.EmailMessage) {
        log.Printf("✓ Sent to %v", msg.To)
    }),
    mailsender.WithOnFailure(func(msg *mailsender.EmailMessage, err error) {
        log.Printf("✗ Failed to send to %v: %v", msg.To, err)
    }),
    mailsender.WithOnRetry(func(msg *mailsender.EmailMessage, attempt int, err error) {
        log.Printf("⟳ Retry %d for %v", attempt, msg.To)
    }),
)
defer asyncSender.Close()

// Send emails asynchronously
for i := 0; i < 1000; i++ {
    err := asyncSender.SendAsync(context.Background(), &mailsender.EmailMessage{
        To:        []string{fmt.Sprintf("user%d@example.com", i)},
        Subject:   "Notification",
        PlainText: "Your notification message",
    })
    if err != nil {
        log.Printf("Failed to queue: %v", err)
    }
}

// Check statistics
stats := asyncSender.Stats()
fmt.Printf("Stats: Sent=%d, Failed=%d, Pending=%d, Retried=%d\n",
    stats.Sent, stats.Failed, stats.Pending, stats.Retried)
```

## Supported Providers

### SendGrid

SendGrid is a cloud-based email delivery service.

**Setup:**
1. Sign up for [SendGrid](https://sendgrid.com/)
2. Create an API key in your SendGrid dashboard
3. Configure your application with the API key

**Example:**
```go
sender, err := mailsender.NewSendGridWithOptions(
    mailsender.WithAPIKey("SG.xxxxxxxxxxxxxxxxxxxxx"),
    mailsender.WithDefaultFrom("noreply@example.com"),
)
```

### Future Providers

The library is designed to support multiple providers. Future providers may include:
- Amazon SES
- Mailgun
- Postmark
- SMTP (generic)

## Error Handling

The library provides specific error types for common scenarios:

```go
var (
    ErrMissingFrom       error // Missing sender email
    ErrMissingRecipients error // No recipients specified
    ErrMissingSubject    error // Missing email subject
    ErrMissingContent    error // No email body provided
    ErrInvalidProvider   error // Unsupported provider
    ErrMissingAPIKey     error // API key not provided
    ErrSendFailed        error // Email sending failed
)
```

**Example:**
```go
err := sender.Send(ctx, message)
if err != nil {
    if errors.Is(err, mailsender.ErrMissingRecipients) {
        log.Println("No recipients specified")
    } else if errors.Is(err, mailsender.ErrSendFailed) {
        log.Printf("Failed to send email: %v", err)
    }
}
```

## Template Features

The library uses Go's standard `html/template` and `text/template` packages.

**Supported Features:**
- Variable substitution: `{{.FieldName}}`
- Conditionals: `{{if .Condition}}...{{else}}...{{end}}`
- Loops: `{{range .Items}}...{{end}}`
- Comparisons: `{{if eq .Status "active"}}...{{end}}`
- Arithmetic: `{{if gt .Count 0}}...{{end}}`

**Template Functions:**
- `RenderHTMLTemplate(templateStr, data)` - Render HTML template
- `RenderTextTemplate(templateStr, data)` - Render plain text template

## Testing

Run all tests:
```bash
go test -v ./...
```

Run tests with coverage:
```bash
go test -v -cover ./...
```

Run tests with race detection:
```bash
go test -v -race ./...
```

## Examples

The `examples/` directory contains complete working examples:

- `examples/sendgrid/` - SendGrid usage examples
- `examples/template/` - Template rendering examples
- `examples/async/` - Async/event-based sending examples

To run an example:
```bash
cd examples/sendgrid
go run main.go
```

Make sure to set your API key before running examples.

## Best Practices

1. **Always close senders**: Use `defer sender.Close()` or `defer asyncSender.Close()` to ensure cleanup
2. **Use context**: Pass context for cancellation and timeout control
3. **Choose the right sending mode**:
   - Use **sync sending** for immediate, critical emails (password resets, OTPs)
   - Use **async sending** for bulk emails, notifications, newsletters
4. **Configure worker pools appropriately**:
   - More workers = higher throughput but more resource usage
   - Start with 3-5 workers and scale based on your needs
5. **Implement event handlers**: Use `OnFailure` handler to log or alert on failures
6. **Use retry logic wisely**: Enable retries for transient failures, but not for validation errors
7. **Monitor statistics**: Track `Stats()` to understand email sending patterns and failures
8. **Validate before sending**: The library validates messages, but pre-validation can improve UX
9. **Use templates for complex emails**: Keep your email HTML/text separate from code
10. **Set default sender**: Configure `DefaultFrom` to avoid repeating it in every message
11. **Handle errors appropriately**: Check for specific error types when needed
12. **Use environment variables**: Keep API keys out of your code
13. **Graceful shutdown**: Always call `Close()` to ensure all queued emails are sent

## Security Considerations

- **Never commit API keys**: Use environment variables or secret management
- **Validate user input**: Sanitize data before using in templates
- **Use HTTPS**: All provider communications use HTTPS
- **Rate limiting**: Respect provider rate limits (implement in your application)
- **Email validation**: Validate email addresses before sending

## Contributing

Contributions are welcome! Please follow these guidelines:

1. Write tests for new features
2. Follow existing code style
3. Update documentation
4. Add examples for new functionality

## License

This package is part of the [go-packages](https://github.com/isimtekin/go-packages) monorepo.

## Support

For issues, questions, or contributions, please visit the [GitHub repository](https://github.com/isimtekin/go-packages).

## Changelog

### v0.1.0 (Initial Release)
- SendGrid provider support
- HTML and plain text email sending
- Template rendering (HTML and text)
- Environment variable configuration
- Functional options pattern
- Multiple recipients (To, Cc, Bcc)
- Custom reply-to addresses
- Async/event-based email sending
- Worker pool for concurrent sending
- Event handlers (OnSuccess, OnFailure, OnRetry)
- Automatic retry logic with configurable attempts
- Real-time statistics tracking
- Graceful shutdown support
- Comprehensive tests and examples
