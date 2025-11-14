package redisclient

import (
	"context"
	"fmt"
	"time"

	envutil "github.com/isimtekin/go-packages/env-util"
)

// LoadConfigFromEnv loads Redis configuration from environment variables
// with a custom prefix (e.g., "REDIS_", "CACHE_", etc.)
func LoadConfigFromEnv(prefix string) (*Config, error) {
	env := envutil.NewWithOptions(envutil.WithPrefix(prefix))

	config := &Config{
		Addr:            env.GetString("ADDR", "localhost:6379"),
		Password:        env.GetString("PASSWORD", ""),
		DB:              env.GetInt("DB", 0),
		MaxRetries:      env.GetInt("MAX_RETRIES", 3),
		MinIdleConns:    env.GetInt("MIN_IDLE_CONNS", 5),
		MaxIdleConns:    env.GetInt("MAX_IDLE_CONNS", 10),
		PoolSize:        env.GetInt("POOL_SIZE", 100),
		PoolTimeout:     env.GetDuration("POOL_TIMEOUT", 4*time.Second),
		ConnMaxIdleTime: env.GetDuration("CONN_MAX_IDLE_TIME", 5*time.Minute),
		ConnMaxLifetime: env.GetDuration("CONN_MAX_LIFETIME", 0),
		DialTimeout:     env.GetDuration("DIAL_TIMEOUT", 5*time.Second),
		ReadTimeout:     env.GetDuration("READ_TIMEOUT", 3*time.Second),
		WriteTimeout:    env.GetDuration("WRITE_TIMEOUT", 3*time.Second),
		TLSEnabled:      env.GetBool("TLS_ENABLED", false),
		TLSCertFile:     env.GetString("TLS_CERT_FILE", ""),
		TLSKeyFile:      env.GetString("TLS_KEY_FILE", ""),
		TLSCAFile:       env.GetString("TLS_CA_FILE", ""),
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration from environment: %w", err)
	}

	return config, nil
}

// LoadConfigFromEnvWithDefaults loads Redis configuration from environment
// variables using the default "REDIS_" prefix
func LoadConfigFromEnvWithDefaults() (*Config, error) {
	return LoadConfigFromEnv("REDIS_")
}

// NewFromEnv creates a new Redis client from environment variables
// with a custom prefix
func NewFromEnv(ctx context.Context, prefix string) (*Client, error) {
	config, err := LoadConfigFromEnv(prefix)
	if err != nil {
		return nil, err
	}

	client, err := New(config)
	if err != nil {
		return nil, err
	}

	// Verify connection
	if err := client.Ping(ctx); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to ping Redis: %w", err)
	}

	return client, nil
}

// NewFromEnvWithDefaults creates a new Redis client from environment variables
// using the default "REDIS_" prefix
func NewFromEnvWithDefaults(ctx context.Context) (*Client, error) {
	return NewFromEnv(ctx, "REDIS_")
}
