package mongoclient

import (
	"fmt"
	"time"
)

// Config holds MongoDB client configuration
type Config struct {
	// Connection settings
	URI      string // MongoDB connection URI
	Database string // Database name

	// Connection pool settings
	MaxPoolSize     uint64        // Maximum number of connections in the pool
	MinPoolSize     uint64        // Minimum number of connections in the pool
	MaxConnIdleTime time.Duration // Maximum time a connection can be idle

	// Timeout settings
	ConnectTimeout         time.Duration // Timeout for initial connection
	SocketTimeout          time.Duration // Timeout for socket operations
	ServerSelectionTimeout time.Duration // Timeout for server selection
	OperationTimeout       time.Duration // Default timeout for operations

	// Retry settings
	RetryWrites bool // Enable retry writes
	RetryReads  bool // Enable retry reads
}

// DefaultConfig returns a config with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		URI:                    "mongodb://localhost:27017",
		Database:               "test",
		MaxPoolSize:            100,
		MinPoolSize:            10,
		MaxConnIdleTime:        5 * time.Minute,
		ConnectTimeout:         10 * time.Second,
		SocketTimeout:          30 * time.Second,
		ServerSelectionTimeout: 10 * time.Second,
		OperationTimeout:       30 * time.Second,
		RetryWrites:            true,
		RetryReads:             true,
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.URI == "" {
		return fmt.Errorf("URI is required")
	}

	if c.Database == "" {
		return fmt.Errorf("Database is required")
	}

	if c.MaxPoolSize > 0 && c.MinPoolSize > c.MaxPoolSize {
		return fmt.Errorf("MinPoolSize (%d) cannot be greater than MaxPoolSize (%d)",
			c.MinPoolSize, c.MaxPoolSize)
	}

	return nil
}
