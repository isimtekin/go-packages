package postgresclient

import (
	"fmt"
	"time"
)

// Config holds PostgreSQL client configuration
type Config struct {
	// Connection settings
	Host     string // PostgreSQL host
	Port     int    // PostgreSQL port
	Database string // Database name
	Username string // Username
	Password string // Password
	SSLMode  string // SSL mode (disable, require, verify-ca, verify-full)

	// Connection pool settings
	MaxConns          int32         // Maximum number of connections in the pool
	MinConns          int32         // Minimum number of connections in the pool
	MaxConnLifetime   time.Duration // Maximum lifetime of a connection
	MaxConnIdleTime   time.Duration // Maximum idle time of a connection
	HealthCheckPeriod time.Duration // Period between health checks

	// Timeout settings
	ConnectTimeout time.Duration // Timeout for establishing new connections
	QueryTimeout   time.Duration // Default timeout for queries
}

// DefaultConfig returns a config with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		Host:              "localhost",
		Port:              5432,
		Database:          "postgres",
		Username:          "postgres",
		Password:          "",
		SSLMode:           "disable",
		MaxConns:          25,
		MinConns:          5,
		MaxConnLifetime:   time.Hour,
		MaxConnIdleTime:   30 * time.Minute,
		HealthCheckPeriod: time.Minute,
		ConnectTimeout:    10 * time.Second,
		QueryTimeout:      30 * time.Second,
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Host == "" {
		return fmt.Errorf("host is required")
	}

	if c.Port <= 0 || c.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}

	if c.Database == "" {
		return fmt.Errorf("database is required")
	}

	if c.Username == "" {
		return fmt.Errorf("username is required")
	}

	if c.MaxConns > 0 && c.MinConns > c.MaxConns {
		return fmt.Errorf("MinConns (%d) cannot be greater than MaxConns (%d)",
			c.MinConns, c.MaxConns)
	}

	validSSLModes := map[string]bool{
		"disable":     true,
		"require":     true,
		"verify-ca":   true,
		"verify-full": true,
		"prefer":      true,
		"allow":       true,
	}
	if c.SSLMode != "" && !validSSLModes[c.SSLMode] {
		return fmt.Errorf("invalid SSL mode: %s", c.SSLMode)
	}

	return nil
}

// DSN returns the PostgreSQL connection string (Data Source Name)
func (c *Config) DSN() string {
	dsn := fmt.Sprintf("host=%s port=%d dbname=%s user=%s",
		c.Host, c.Port, c.Database, c.Username)

	if c.Password != "" {
		dsn += fmt.Sprintf(" password=%s", c.Password)
	}

	if c.SSLMode != "" {
		dsn += fmt.Sprintf(" sslmode=%s", c.SSLMode)
	}

	if c.ConnectTimeout > 0 {
		dsn += fmt.Sprintf(" connect_timeout=%d", int(c.ConnectTimeout.Seconds()))
	}

	return dsn
}

// ConnectionURL returns the PostgreSQL connection URL
func (c *Config) ConnectionURL() string {
	var auth string
	if c.Username != "" {
		auth = c.Username
		if c.Password != "" {
			auth += ":" + c.Password
		}
		auth += "@"
	}

	url := fmt.Sprintf("postgres://%s%s:%d/%s", auth, c.Host, c.Port, c.Database)

	if c.SSLMode != "" {
		url += "?sslmode=" + c.SSLMode
	}

	return url
}
