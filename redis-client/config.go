package redisclient

import (
	"fmt"
	"time"
)

// Config holds the configuration for Redis client
type Config struct {
	// Connection settings
	Addr     string `json:"addr" yaml:"addr"`         // host:port address
	Password string `json:"password" yaml:"password"` // password (optional)
	DB       int    `json:"db" yaml:"db"`             // database number

	// Connection pool settings
	MaxRetries      int           `json:"max_retries" yaml:"max_retries"`
	MinIdleConns    int           `json:"min_idle_conns" yaml:"min_idle_conns"`
	MaxIdleConns    int           `json:"max_idle_conns" yaml:"max_idle_conns"`
	PoolSize        int           `json:"pool_size" yaml:"pool_size"`
	PoolTimeout     time.Duration `json:"pool_timeout" yaml:"pool_timeout"`
	ConnMaxIdleTime time.Duration `json:"conn_max_idle_time" yaml:"conn_max_idle_time"`
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime" yaml:"conn_max_lifetime"`

	// Timeout settings
	DialTimeout  time.Duration `json:"dial_timeout" yaml:"dial_timeout"`
	ReadTimeout  time.Duration `json:"read_timeout" yaml:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout" yaml:"write_timeout"`

	// TLS/SSL settings
	TLSEnabled  bool   `json:"tls_enabled" yaml:"tls_enabled"`
	TLSCertFile string `json:"tls_cert_file" yaml:"tls_cert_file"`
	TLSKeyFile  string `json:"tls_key_file" yaml:"tls_key_file"`
	TLSCAFile   string `json:"tls_ca_file" yaml:"tls_ca_file"`

	// Database name mappings (for multi-database manager)
	DatabaseNames map[string]int `json:"database_names" yaml:"database_names"`
}

// DefaultConfig returns the default configuration for Redis
func DefaultConfig() *Config {
	return &Config{
		Addr:            "localhost:6379",
		Password:        "",
		DB:              0,
		MaxRetries:      3,
		MinIdleConns:    5,
		MaxIdleConns:    10,
		PoolSize:        100,
		PoolTimeout:     4 * time.Second,
		ConnMaxIdleTime: 5 * time.Minute,
		ConnMaxLifetime: 0, // 0 means connections are not closed due to age
		DialTimeout:     5 * time.Second,
		ReadTimeout:     3 * time.Second,
		WriteTimeout:    3 * time.Second,
		TLSEnabled:      false,
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Addr == "" {
		return fmt.Errorf("addr cannot be empty")
	}

	if c.DB < 0 {
		return fmt.Errorf("db must be non-negative")
	}

	if c.MaxRetries < 0 {
		return fmt.Errorf("max_retries must be non-negative")
	}

	if c.PoolSize <= 0 {
		return fmt.Errorf("pool_size must be positive")
	}

	if c.DialTimeout <= 0 {
		return fmt.Errorf("dial_timeout must be positive")
	}

	if c.TLSEnabled {
		if c.TLSCertFile == "" && c.TLSKeyFile != "" {
			return fmt.Errorf("tls_cert_file required when tls_key_file is set")
		}
		if c.TLSKeyFile == "" && c.TLSCertFile != "" {
			return fmt.Errorf("tls_key_file required when tls_cert_file is set")
		}
	}

	return nil
}
