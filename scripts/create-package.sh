#!/bin/bash
# scripts/create-package.sh

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Get package name from argument
PACKAGE_NAME=$1

if [ -z "$PACKAGE_NAME" ]; then
    echo -e "${RED}Error: Package name is required${NC}"
    echo "Usage: ./scripts/create-package.sh <package-name>"
    echo "Example: ./scripts/create-package.sh redis-client"
    exit 1
fi

# Check if package already exists
if [ -d "$PACKAGE_NAME" ]; then
    echo -e "${RED}Error: Package '$PACKAGE_NAME' already exists${NC}"
    exit 1
fi

# Convert package-name to packagename (remove hyphens for Go package name)
GO_PACKAGE_NAME=$(echo "$PACKAGE_NAME" | tr '-' '')

echo -e "${YELLOW}Creating package: $PACKAGE_NAME${NC}"

# Create package directory
mkdir -p "$PACKAGE_NAME"
cd "$PACKAGE_NAME"

# Create go.mod
echo -e "${GREEN}✓${NC} Creating go.mod..."
cat > go.mod << EOF
module github.com/isimtekin/go-packages/${PACKAGE_NAME}

go 1.21
EOF

# Create config.go
echo -e "${GREEN}✓${NC} Creating config.go..."
cat > config.go << EOF
package ${GO_PACKAGE_NAME}

import (
	"fmt"
	"time"
)

// Config holds the configuration for ${PACKAGE_NAME}
type Config struct {
	// Add your configuration fields here
	Host           string        \`json:"host" yaml:"host"\`
	Port           int           \`json:"port" yaml:"port"\`
	Timeout        time.Duration \`json:"timeout" yaml:"timeout"\`
	MaxRetries     int           \`json:"max_retries" yaml:"max_retries"\`
	EnableDebug    bool          \`json:"enable_debug" yaml:"enable_debug"\`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Host:       "localhost",
		Port:       8080,
		Timeout:    30 * time.Second,
		MaxRetries: 3,
		EnableDebug: false,
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Host == "" {
		return fmt.Errorf("host cannot be empty")
	}

	if c.Port <= 0 || c.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}

	if c.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}

	return nil
}
EOF

# Create client.go
echo -e "${GREEN}✓${NC} Creating client.go..."
cat > client.go << EOF
package ${GO_PACKAGE_NAME}

import (
	"context"
	"fmt"
	"sync"
)

// Client represents the ${PACKAGE_NAME} client
type Client struct {
	config *Config

	// Add your client fields here
	// conn   *Connection
	// pool   *ConnectionPool

	mu     sync.RWMutex
	closed bool
}

// New creates a new client with the given configuration
func New(config *Config) (*Client, error) {
	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	client := &Client{
		config: config,
		closed: false,
	}

	// Initialize your client here
	if err := client.connect(); err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	return client, nil
}

// NewWithOptions creates a new client with functional options
func NewWithOptions(opts ...Option) (*Client, error) {
	config := DefaultConfig()

	for _, opt := range opts {
		opt(config)
	}

	return New(config)
}

// connect establishes the connection
func (c *Client) connect() error {
	// TODO: Implement connection logic
	return nil
}

// Close closes the client and releases resources
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return ErrAlreadyClosed
	}

	// TODO: Clean up resources

	c.closed = true
	return nil
}

// Ping checks if the connection is alive
func (c *Client) Ping(ctx context.Context) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return ErrClientClosed
	}

	// TODO: Implement ping logic
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

// Example method - replace with your actual methods
func (c *Client) DoSomething(ctx context.Context, param string) (string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return "", ErrClientClosed
	}

	// TODO: Implement your logic here
	return "result: " + param, nil
}
EOF

# Create options.go
echo -e "${GREEN}✓${NC} Creating options.go..."
cat > options.go << EOF
package ${GO_PACKAGE_NAME}

import "time"

// Option is a functional option for configuring the client
type Option func(*Config)

// WithHost sets the host
func WithHost(host string) Option {
	return func(c *Config) {
		c.Host = host
	}
}

// WithPort sets the port
func WithPort(port int) Option {
	return func(c *Config) {
		c.Port = port
	}
}

// WithTimeout sets the timeout
func WithTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		c.Timeout = timeout
	}
}

// WithMaxRetries sets the maximum number of retries
func WithMaxRetries(retries int) Option {
	return func(c *Config) {
		c.MaxRetries = retries
	}
}

// WithDebug enables or disables debug mode
func WithDebug(enable bool) Option {
	return func(c *Config) {
		c.EnableDebug = enable
	}
}
EOF

# Create errors.go
echo -e "${GREEN}✓${NC} Creating errors.go..."
cat > errors.go << EOF
package ${GO_PACKAGE_NAME}

import "errors"

var (
	// ErrClientClosed is returned when operating on a closed client
	ErrClientClosed = errors.New("client is closed")

	// ErrAlreadyClosed is returned when closing an already closed client
	ErrAlreadyClosed = errors.New("client is already closed")

	// ErrConnectionFailed is returned when connection fails
	ErrConnectionFailed = errors.New("connection failed")

	// ErrTimeout is returned when an operation times out
	ErrTimeout = errors.New("operation timeout")

	// ErrInvalidResponse is returned when the response is invalid
	ErrInvalidResponse = errors.New("invalid response")

	// Add more errors as needed
)

// IsConnectionError returns true if the error is connection related
func IsConnectionError(err error) bool {
	return errors.Is(err, ErrConnectionFailed) ||
		errors.Is(err, ErrClientClosed)
}

// IsTimeoutError returns true if the error is timeout related
func IsTimeoutError(err error) bool {
	return errors.Is(err, ErrTimeout)
}
EOF

# Create client_test.go
echo -e "${GREEN}✓${NC} Creating client_test.go..."
cat > client_test.go << EOF
package ${GO_PACKAGE_NAME}_test

import (
	"context"
	"testing"
	"time"

	${GO_PACKAGE_NAME} "github.com/isimtekin/go-packages/${PACKAGE_NAME}"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		config  *${GO_PACKAGE_NAME}.Config
		wantErr bool
	}{
		{
			name:    "default config",
			config:  ${GO_PACKAGE_NAME}.DefaultConfig(),
			wantErr: false,
		},
		{
			name: "invalid host",
			config: &${GO_PACKAGE_NAME}.Config{
				Host:    "",
				Port:    8080,
				Timeout: 30 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "invalid port",
			config: &${GO_PACKAGE_NAME}.Config{
				Host:    "localhost",
				Port:    0,
				Timeout: 30 * time.Second,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := ${GO_PACKAGE_NAME}.New(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if client != nil {
				defer client.Close()
			}
		})
	}
}

func TestNewWithOptions(t *testing.T) {
	client, err := ${GO_PACKAGE_NAME}.NewWithOptions(
		${GO_PACKAGE_NAME}.WithHost("example.com"),
		${GO_PACKAGE_NAME}.WithPort(9090),
		${GO_PACKAGE_NAME}.WithTimeout(10*time.Second),
		${GO_PACKAGE_NAME}.WithDebug(true),
	)

	if err != nil {
		t.Fatalf("NewWithOptions() failed: %v", err)
	}
	defer client.Close()
}

func TestClient_Ping(t *testing.T) {
	client, err := ${GO_PACKAGE_NAME}.New(${GO_PACKAGE_NAME}.DefaultConfig())
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	if err := client.Ping(ctx); err != nil {
		t.Errorf("Ping() failed: %v", err)
	}
}

func TestClient_Close(t *testing.T) {
	client, err := ${GO_PACKAGE_NAME}.New(${GO_PACKAGE_NAME}.DefaultConfig())
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// First close should succeed
	if err := client.Close(); err != nil {
		t.Errorf("First Close() failed: %v", err)
	}

	// Second close should return ErrAlreadyClosed
	if err := client.Close(); err != ${GO_PACKAGE_NAME}.ErrAlreadyClosed {
		t.Errorf("Second Close() error = %v, want %v", err, ${GO_PACKAGE_NAME}.ErrAlreadyClosed)
	}
}
EOF

# Create README.md
echo -e "${GREEN}✓${NC} Creating README.md..."
cat > README.md << EOF
# ${PACKAGE_NAME^}

A Go client library for ${PACKAGE_NAME//-/ } operations.

## Installation

\`\`\`bash
go get github.com/isimtekin/go-packages/${PACKAGE_NAME}
\`\`\`

## Quick Start

\`\`\`go
package main

import (
    "context"
    "log"

    ${GO_PACKAGE_NAME} "github.com/isimtekin/go-packages/${PACKAGE_NAME}"
)

func main() {
    // Create client with default config
    client, err := ${GO_PACKAGE_NAME}.New(${GO_PACKAGE_NAME}.DefaultConfig())
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // Ping to check connection
    ctx := context.Background()
    if err := client.Ping(ctx); err != nil {
        log.Fatal("ping failed:", err)
    }

    // Use the client
    result, err := client.DoSomething(ctx, "example")
    if err != nil {
        log.Fatal(err)
    }
    log.Println("Result:", result)
}
\`\`\`

## Using with Options

\`\`\`go
client, err := ${GO_PACKAGE_NAME}.NewWithOptions(
    ${GO_PACKAGE_NAME}.WithHost("example.com"),
    ${GO_PACKAGE_NAME}.WithPort(9090),
    ${GO_PACKAGE_NAME}.WithTimeout(10*time.Second),
    ${GO_PACKAGE_NAME}.WithDebug(true),
)
if err != nil {
    log.Fatal(err)
}
defer client.Close()
\`\`\`

## Configuration

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| Host | string | localhost | Server host |
| Port | int | 8080 | Server port |
| Timeout | time.Duration | 30s | Operation timeout |
| MaxRetries | int | 3 | Maximum retry attempts |
| EnableDebug | bool | false | Enable debug logging |

## API Reference

### Client Methods

#### New(config *Config) (*Client, error)
Creates a new client with the given configuration.

#### NewWithOptions(opts ...Option) (*Client, error)
Creates a new client using functional options.

#### (c *Client) Ping(ctx context.Context) error
Checks if the connection is alive.

#### (c *Client) Close() error
Closes the client and releases all resources.

#### (c *Client) DoSomething(ctx context.Context, param string) (string, error)
Example method - replace with actual implementation.

## Error Handling

Common errors:

- \`ErrClientClosed\`: Client has been closed
- \`ErrAlreadyClosed\`: Client is already closed
- \`ErrConnectionFailed\`: Failed to establish connection
- \`ErrTimeout\`: Operation timed out
- \`ErrInvalidResponse\`: Received invalid response

Example error handling:

\`\`\`go
result, err := client.DoSomething(ctx, "param")
if err != nil {
    if ${GO_PACKAGE_NAME}.IsTimeoutError(err) {
        // Handle timeout
        return fmt.Errorf("operation timed out: %w", err)
    }
    if ${GO_PACKAGE_NAME}.IsConnectionError(err) {
        // Handle connection error
        return fmt.Errorf("connection issue: %w", err)
    }
    return err
}
\`\`\`

## Testing

Run tests:

\`\`\`bash
go test -v -cover ./...
\`\`\`

Run benchmarks:

\`\`\`bash
go test -bench=. -benchmem
\`\`\`

## License

MIT License - see [LICENSE](../LICENSE) file for details.

## Contributing

Please see [CONTRIBUTING.md](../CONTRIBUTING.md) for details on submitting patches and the contribution workflow.

## TODO

- [ ] Implement actual connection logic
- [ ] Add connection pooling
- [ ] Add retry mechanism
- [ ] Add metrics/monitoring support
- [ ] Add more comprehensive tests
- [ ] Add benchmarks
EOF

# Create .gitignore
echo -e "${GREEN}✓${NC} Creating .gitignore..."
cat > .gitignore << EOF
# Binaries for programs and plugins
*.exe
*.exe~
*.dll
*.so
*.dylib

# Test binary, built with go test -c
*.test

# Output of the go coverage tool
*.out

# Go workspace file
go.work

# IDE
.idea/
.vscode/
*.swp
*.swo
*~

# OS
.DS_Store
Thumbs.db
EOF

cd ..

echo -e "${GREEN}✅ Package '$PACKAGE_NAME' created successfully!${NC}"
echo ""
echo "Next steps:"
echo "  1. cd $PACKAGE_NAME"
echo "  2. Implement your logic in client.go"
echo "  3. Update the README.md with actual documentation"
echo "  4. Run tests: go test -v ./..."
echo "  5. Release: ./scripts/bump-version.sh $PACKAGE_NAME minor"
echo ""
echo -e "${YELLOW}Don't forget to:${NC}"
echo "  - Replace 'DoSomething' method with actual functionality"
echo "  - Update config fields for your specific needs"
echo "  - Add proper error types in errors.go"
echo "  - Implement connection logic in connect() method"