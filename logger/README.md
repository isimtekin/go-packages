# Logger

A flexible, production-ready logging package for Go applications with support for multiple outputs including console, Logstash, and file logging. Built on top of [logrus](https://github.com/sirupsen/logrus) with enhanced features for structured logging and ELK stack integration.

## Features

- **Multiple Output Destinations**
  - Console (with colored output)
  - Logstash (TCP/UDP with automatic reconnection)
  - File (with rotation support)
  - Multi-output (combine multiple destinations)

- **Enable/Disable Control**
  - Globally enable or disable logging
  - Check debug mode status at runtime
  - Zero overhead when disabled

- **Structured Logging**
  - JSON or text format
  - Automatic field management
  - Context-aware logging
  - Service metadata (name, environment, version, host)

- **ELK Stack Integration**
  - Logstash-compatible JSON format
  - Automatic @timestamp and @version fields
  - Buffering and automatic reconnection
  - Configurable retry logic

- **Flexible Configuration**
  - Environment variable support
  - Functional options pattern
  - Sensible defaults
  - Comprehensive validation

- **Production Ready**
  - Connection pooling for Logstash
  - Automatic error recovery
  - Thread-safe operations
  - Minimal performance impact

## Installation

```bash
go get github.com/isimtekin/go-packages/logger@v0.1.0
```

## Quick Start

### Simple Console Logging

```go
package main

import (
    "github.com/isimtekin/go-packages/logger"
)

func main() {
    // Use default configuration (console output, info level)
    logger.Info("Application started")
    logger.Debug("This won't be logged (level is info)")

    logger.Info("User logged in",
        logger.String("user_id", "123"),
        logger.String("ip", "192.168.1.1"),
    )

    logger.Error("Failed to connect to database",
        logger.Err(err),
        logger.String("host", "localhost"),
    )
}
```

### Custom Logger Instance

```go
package main

import (
    "github.com/isimtekin/go-packages/logger"
)

func main() {
    config := logger.DefaultConfig()
    config.Level = "debug"
    config.Format = "json"
    config.ServiceName = "my-service"
    config.Environment = "production"

    log, err := logger.New(config)
    if err != nil {
        panic(err)
    }

    log.Info("Application started")
    log.Debug("Debug information",
        logger.String("module", "main"),
    )
}
```

### Using Functional Options

```go
package main

import (
    "github.com/isimtekin/go-packages/logger"
)

func main() {
    log, err := logger.NewWithOptions(
        logger.WithEnabled(true),
        logger.WithLevel("debug"),
        logger.WithFormat("json"),
        logger.WithServiceName("my-service"),
        logger.WithEnvironment("production"),
        logger.WithVersion("1.0.0"),
    )
    if err != nil {
        panic(err)
    }

    log.Info("Application started")
}
```

## Configuration

### Config Structure

```go
type Config struct {
    // Enable/disable logging
    Enabled bool

    // Log level: "trace", "debug", "info", "warn", "error", "fatal", "panic"
    Level string

    // Output: "console", "logstash", "file", "multi"
    Output string

    // Format: "json", "text"
    Format string

    // Console settings
    ConsoleColors bool

    // Logstash settings
    LogstashHost     string
    LogstashPort     int
    LogstashProtocol string        // "tcp" or "udp"
    LogstashTimeout  time.Duration
    LogstashRetries  int

    // File settings
    FilePath       string
    FileMaxSize    int  // MB
    FileMaxAge     int  // days
    FileMaxBackups int  // number of files

    // Service metadata
    ServiceName string
    Environment string
    Version     string
    Host        string

    // Performance
    ReportCaller bool  // Include file:line in logs
}
```

### Default Configuration

```go
config := logger.DefaultConfig()
// Enabled:          true
// Level:            "info"
// Output:           "console"
// Format:           "text"
// ConsoleColors:    true
// LogstashPort:     5000
// LogstashProtocol: "tcp"
// LogstashTimeout:  5s
// LogstashRetries:  3
// FileMaxSize:      100 MB
// FileMaxAge:       30 days
// FileMaxBackups:   10 files
```

## Environment Variables

Load configuration from environment variables:

```go
// Using default prefix "LOG_"
logger, err := logger.NewFromEnvWithDefaults()

// Using custom prefix
logger, err := logger.NewFromEnv("APP_LOG_")
```

### Available Environment Variables

```bash
# Basic settings
LOG_ENABLED=true
LOG_LEVEL=debug
LOG_OUTPUT=console
LOG_FORMAT=json

# Console settings
LOG_CONSOLE_COLORS=true

# Logstash settings
LOG_LOGSTASH_HOST=logstash.example.com
LOG_LOGSTASH_PORT=5044
LOG_LOGSTASH_PROTOCOL=tcp
LOG_LOGSTASH_TIMEOUT=10s
LOG_LOGSTASH_RETRIES=5

# File settings
LOG_FILE_PATH=/var/log/app.log
LOG_FILE_MAX_SIZE=200
LOG_FILE_MAX_AGE=60
LOG_FILE_MAX_BACKUPS=20

# Service metadata
LOG_SERVICE_NAME=my-service
LOG_ENVIRONMENT=production
LOG_VERSION=1.0.0
LOG_HOST=server-01

# Performance
LOG_REPORT_CALLER=true
```

### .env.example

```bash
# Enable/disable logging
LOG_ENABLED=true

# Log level (trace, debug, info, warn, error, fatal, panic)
LOG_LEVEL=info

# Output destination (console, logstash, file, multi)
LOG_OUTPUT=console

# Log format (json, text)
LOG_FORMAT=text

# Console colors (true, false)
LOG_CONSOLE_COLORS=true

# Logstash configuration
LOG_LOGSTASH_HOST=localhost
LOG_LOGSTASH_PORT=5000
LOG_LOGSTASH_PROTOCOL=tcp
LOG_LOGSTASH_TIMEOUT=5s
LOG_LOGSTASH_RETRIES=3

# File configuration
LOG_FILE_PATH=/var/log/app.log
LOG_FILE_MAX_SIZE=100
LOG_FILE_MAX_AGE=30
LOG_FILE_MAX_BACKUPS=10

# Service metadata
LOG_SERVICE_NAME=my-application
LOG_ENVIRONMENT=development
LOG_VERSION=1.0.0

# Performance options
LOG_REPORT_CALLER=false
```

## Usage Examples

### Enable/Disable Logging

```go
// Disable logging completely (zero overhead)
config := logger.DefaultConfig()
config.Enabled = false
log, _ := logger.New(config)

// All log calls become no-ops
log.Info("This won't be logged")
log.Debug("This won't be logged")

// Check if logging is enabled
if log.IsEnabled() {
    // Perform expensive operations only if logging is enabled
    data := gatherDebugData()
    log.Debug("Debug data", logger.Any("data", data))
}
```

### Debug Mode

```go
config := logger.DefaultConfig()
config.Level = "debug"
log, _ := logger.New(config)

// Check if debug is enabled
if log.IsDebugEnabled() {
    // Only execute expensive debug operations when needed
    complexData := performExpensiveOperation()
    log.Debug("Complex data", logger.Any("data", complexData))
}
```

### Structured Logging

```go
log.Info("User action",
    logger.String("action", "login"),
    logger.String("user_id", "123"),
    logger.String("ip", "192.168.1.1"),
    logger.Int("attempts", 3),
    logger.Bool("success", true),
    logger.Float64("duration", 0.523),
)

// JSON output:
// {
//   "@timestamp": "2025-11-26T10:00:00.000Z",
//   "level": "info",
//   "message": "User action",
//   "action": "login",
//   "user_id": "123",
//   "ip": "192.168.1.1",
//   "attempts": 3,
//   "success": true,
//   "duration": 0.523
// }
```

### Context-Aware Logging

```go
import "context"

// Add fields to logger
logger := log.WithFields(
    logger.String("request_id", "abc-123"),
    logger.String("user_id", "456"),
)

// Use throughout request
logger.Info("Processing request")
logger.Debug("Validation passed")
logger.Error("Database error", logger.Err(err))

// Extract from context
ctx := context.WithValue(context.Background(), "request_id", "abc-123")
logger = log.WithContext(ctx)
logger.Info("Request received")
```

### Logstash Output

```go
log, err := logger.NewWithOptions(
    logger.WithOutput("logstash"),
    logger.WithFormat("json"),
    logger.WithLogstash("logstash.example.com", 5044, "tcp"),
    logger.WithLogstashTimeout(10 * time.Second),
    logger.WithLogstashRetries(5),
    logger.WithServiceName("my-service"),
    logger.WithEnvironment("production"),
)

log.Info("Application started")

// Sends to Logstash:
// {
//   "@timestamp": "2025-11-26T10:00:00.000Z",
//   "@version": "1",
//   "level": "info",
//   "message": "Application started",
//   "service": "my-service",
//   "environment": "production",
//   "host": "server-01"
// }
```

### File Output

```go
log, err := logger.NewWithOptions(
    logger.WithOutput("file"),
    logger.WithFile("/var/log/myapp.log"),
    logger.WithFileRotation(100, 30, 10), // 100MB, 30 days, 10 backups
)

log.Info("Writing to file")
```

### Multi-Output

```go
config := logger.DefaultConfig()
config.Output = "multi"
config.MultiOutputs = []logger.OutputConfig{
    {
        Type:    "console",
        Level:   "info",
        Format:  "text",
        Enabled: true,
    },
    {
        Type:     "logstash",
        Level:    "warn",
        Format:   "json",
        Enabled:  true,
        Host:     "logstash.example.com",
        Port:     5044,
    },
    {
        Type:     "file",
        Level:    "error",
        Format:   "json",
        Enabled:  true,
        FilePath: "/var/log/errors.log",
    },
}

log, err := logger.New(config)
// Info logs go to console only
// Warn logs go to console and Logstash
// Error logs go to all three outputs
```

### Global Logger

```go
// Set global logger
logger, _ := logger.NewWithOptions(
    logger.WithLevel("debug"),
    logger.WithServiceName("my-service"),
)
logger.SetGlobal(logger)

// Use global logger anywhere
logger.Info("Using global logger")
logger.Debug("Debug message")

// Or from environment
logger.SetGlobalFromEnvWithDefaults()
logger.Info("Global logger from env")
```

### Error Logging

```go
err := someOperation()
if err != nil {
    log.Error("Operation failed",
        logger.Err(err),
        logger.String("operation", "database_query"),
        logger.String("table", "users"),
    )
}
```

### Field Helpers

```go
// Available field types
logger.String("key", "value")
logger.Int("count", 42)
logger.Int64("count", int64(42))
logger.Float64("price", 19.99)
logger.Bool("active", true)
logger.Err(err)
logger.Any("data", complexStruct)
```

## Logstash Integration

### Logstash Configuration

```ruby
input {
  tcp {
    port => 5044
    codec => json_lines
  }
}

filter {
  # Your filters here
  if [service] == "my-service" {
    mutate {
      add_field => { "[@metadata][index]" => "my-service-%{+YYYY.MM.dd}" }
    }
  }
}

output {
  elasticsearch {
    hosts => ["elasticsearch:9200"]
    index => "%{[@metadata][index]}"
  }
}
```

### Docker Compose Example

```yaml
version: '3.8'

services:
  app:
    build: .
    environment:
      LOG_ENABLED: "true"
      LOG_LEVEL: "info"
      LOG_OUTPUT: "logstash"
      LOG_FORMAT: "json"
      LOG_LOGSTASH_HOST: "logstash"
      LOG_LOGSTASH_PORT: "5044"
      LOG_SERVICE_NAME: "my-service"
      LOG_ENVIRONMENT: "production"
    depends_on:
      - logstash

  logstash:
    image: docker.elastic.co/logstash/logstash:8.11.0
    ports:
      - "5044:5044"
    volumes:
      - ./logstash.conf:/usr/share/logstash/pipeline/logstash.conf

  elasticsearch:
    image: docker.elastic.co/elasticsearch/elasticsearch:8.11.0
    environment:
      - discovery.type=single-node
    ports:
      - "9200:9200"

  kibana:
    image: docker.elastic.co/kibana/kibana:8.11.0
    ports:
      - "5601:5601"
    depends_on:
      - elasticsearch
```

## Advanced Features

### Custom Logger with All Options

```go
log, err := logger.NewWithOptions(
    // Enable/disable
    logger.WithEnabled(true),

    // Level and format
    logger.WithLevel("debug"),
    logger.WithFormat("json"),

    // Output
    logger.WithOutput("logstash"),
    logger.WithLogstash("logstash.example.com", 5044, "tcp"),
    logger.WithLogstashTimeout(10 * time.Second),
    logger.WithLogstashRetries(5),

    // Metadata
    logger.WithServiceName("my-service"),
    logger.WithEnvironment("production"),
    logger.WithVersion("1.2.3"),
    logger.WithHost("server-01"),

    // Performance
    logger.WithReportCaller(true),
)
```

### Performance Considerations

```go
// Disable logging in production for sensitive paths
config := logger.DefaultConfig()
config.Enabled = false
sensitiveLogger, _ := logger.New(config)

// Use IsDebugEnabled to avoid expensive operations
if log.IsDebugEnabled() {
    // This expensive operation only runs when debug is enabled
    data := gatherComplexDebugInfo()
    log.Debug("Debug info", logger.Any("data", data))
}
```

## Testing

The package includes 56 tests with 53.7% code coverage:

- Configuration validation
- Enable/disable functionality
- All log levels (trace, debug, info, warn, error, fatal, panic)
- Field helpers
- Context support
- Output formats (JSON, text)
- Environment variable loading
- Functional options
- Global logger

Run tests:

```bash
cd logger
go test -v -cover ./...
```

## Best Practices

1. **Use Structured Logging**: Always use field helpers for better searchability in ELK
   ```go
   // Good
   log.Info("User logged in", logger.String("user_id", "123"))

   // Avoid
   log.Info("User 123 logged in")
   ```

2. **Check Debug Mode**: Prevent expensive operations when debug is disabled
   ```go
   if log.IsDebugEnabled() {
       complexData := expensiveOperation()
       log.Debug("Data", logger.Any("data", complexData))
   }
   ```

3. **Use Service Metadata**: Always set service name, environment, and version
   ```go
   logger.WithServiceName("my-service"),
   logger.WithEnvironment("production"),
   logger.WithVersion("1.0.0"),
   ```

4. **Enable Buffering for Logstash**: Use appropriate retry and timeout settings
   ```go
   logger.WithLogstashRetries(5),
   logger.WithLogstashTimeout(10 * time.Second),
   ```

5. **Use Context**: Extract request-specific fields from context
   ```go
   logger := log.WithContext(ctx)
   logger.Info("Processing request")
   ```

## Error Handling

The logger handles errors gracefully:

- **Validation errors**: Returned during logger creation
- **Connection errors**: Automatic reconnection for Logstash
- **Buffering**: Messages buffered during Logstash downtime (max 1000)
- **Disabled state**: All operations become no-ops when disabled

## Dependencies

- [github.com/sirupsen/logrus](https://github.com/sirupsen/logrus) - Core logging library
- [github.com/isimtekin/go-packages/env-util](https://github.com/isimtekin/go-packages/env-util) - Environment variable utilities

## License

MIT License

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Support

For issues and questions, please open an issue on GitHub.

## Changelog

### v0.1.0 (Initial Release)
- Console, Logstash, and file output support
- Enable/disable functionality
- Debug mode detection
- Structured logging with field helpers
- Environment variable configuration
- Functional options pattern
- Multi-output support
- Logstash reconnection and buffering
- Context-aware logging
- Comprehensive test coverage
