package postgresclient

import (
	"context"
	"fmt"
	"time"

	envutil "github.com/isimtekin/go-packages/env-util"
)

// LoadConfigFromEnv loads PostgreSQL configuration from environment variables
// with optional prefix (e.g., "POSTGRES_" or "DB_")
//
// Environment variables:
//   - {PREFIX}HOST - PostgreSQL host (default: localhost)
//   - {PREFIX}PORT - PostgreSQL port (default: 5432)
//   - {PREFIX}DATABASE or {PREFIX}DB - Database name (required)
//   - {PREFIX}USERNAME or {PREFIX}USER - Username (default: postgres)
//   - {PREFIX}PASSWORD or {PREFIX}PASS - Password
//   - {PREFIX}SSL_MODE - SSL mode (disable, require, verify-ca, verify-full)
//   - {PREFIX}MAX_CONNS - Maximum number of connections
//   - {PREFIX}MIN_CONNS - Minimum number of connections
//   - {PREFIX}MAX_CONN_LIFETIME - Maximum connection lifetime (duration)
//   - {PREFIX}MAX_CONN_IDLE_TIME - Maximum connection idle time (duration)
//   - {PREFIX}HEALTH_CHECK_PERIOD - Health check period (duration)
//   - {PREFIX}CONNECT_TIMEOUT - Connection timeout (duration)
//   - {PREFIX}QUERY_TIMEOUT - Default query timeout (duration)
func LoadConfigFromEnv(prefix string) (*Config, error) {
	env := envutil.NewWithOptions(
		envutil.WithPrefix(prefix),
		envutil.WithSilent(true),
	)

	config := &Config{
		// Connection settings
		Host:     env.GetString("HOST", "localhost"),
		Port:     env.GetInt("PORT", 5432),
		Database: env.GetString("DATABASE", env.GetString("DB", "")),
		Username: env.GetString("USERNAME", env.GetString("USER", "postgres")),
		Password: env.GetString("PASSWORD", env.GetString("PASS", "")),
		SSLMode:  env.GetString("SSL_MODE", "disable"),

		// Pool settings
		MaxConns:          int32(env.GetInt("MAX_CONNS", 25)),
		MinConns:          int32(env.GetInt("MIN_CONNS", 5)),
		MaxConnLifetime:   env.GetDuration("MAX_CONN_LIFETIME", time.Hour),
		MaxConnIdleTime:   env.GetDuration("MAX_CONN_IDLE_TIME", 30*time.Minute),
		HealthCheckPeriod: env.GetDuration("HEALTH_CHECK_PERIOD", time.Minute),

		// Timeout settings
		ConnectTimeout: env.GetDuration("CONNECT_TIMEOUT", 10*time.Second),
		QueryTimeout:   env.GetDuration("QUERY_TIMEOUT", 30*time.Second),
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration from environment: %w", err)
	}

	return config, nil
}

// LoadConfigFromEnvWithDefaults loads config from environment with default prefix
// Uses "POSTGRES_" as the default prefix (e.g., POSTGRES_HOST, POSTGRES_DATABASE)
func LoadConfigFromEnvWithDefaults() (*Config, error) {
	return LoadConfigFromEnv("POSTGRES_")
}

// NewFromEnv creates a new PostgreSQL client from environment variables
// with optional prefix
func NewFromEnv(ctx context.Context, prefix string) (*Client, error) {
	config, err := LoadConfigFromEnv(prefix)
	if err != nil {
		return nil, err
	}
	return New(ctx, config)
}

// NewFromEnvWithDefaults creates a new PostgreSQL client from environment
// using default "POSTGRES_" prefix
func NewFromEnvWithDefaults(ctx context.Context) (*Client, error) {
	return NewFromEnv(ctx, "POSTGRES_")
}

// MustLoadConfigFromEnv loads config from environment or panics
func MustLoadConfigFromEnv(prefix string) *Config {
	config, err := LoadConfigFromEnv(prefix)
	if err != nil {
		panic(fmt.Sprintf("failed to load PostgreSQL config from environment: %v", err))
	}
	return config
}
