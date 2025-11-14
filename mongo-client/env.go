package mongoclient

import (
	"context"
	"fmt"
	"time"

	envutil "github.com/isimtekin/go-packages/env-util"
)

// LoadConfigFromEnv loads MongoDB configuration from environment variables
// with optional prefix (e.g., "MONGO_" or "DB_")
//
// Environment variables:
//   - {PREFIX}URI or {PREFIX}URL - MongoDB connection URI (required)
//   - {PREFIX}HOST - MongoDB host (alternative to URI)
//   - {PREFIX}PORT - MongoDB port (default: 27017)
//   - {PREFIX}USERNAME - MongoDB username
//   - {PREFIX}PASSWORD - MongoDB password
//   - {PREFIX}DATABASE or {PREFIX}DB - Database name (required)
//   - {PREFIX}AUTH_SOURCE - Authentication database (default: admin)
//   - {PREFIX}MAX_POOL_SIZE - Maximum connection pool size
//   - {PREFIX}MIN_POOL_SIZE - Minimum connection pool size
//   - {PREFIX}MAX_CONN_IDLE_TIME - Maximum connection idle time (duration)
//   - {PREFIX}CONNECT_TIMEOUT - Connection timeout (duration)
//   - {PREFIX}SOCKET_TIMEOUT - Socket timeout (duration)
//   - {PREFIX}SERVER_SELECTION_TIMEOUT - Server selection timeout (duration)
//   - {PREFIX}OPERATION_TIMEOUT - Default operation timeout (duration)
//   - {PREFIX}RETRY_WRITES - Enable retry writes (bool, default: true)
//   - {PREFIX}RETRY_READS - Enable retry reads (bool, default: true)
func LoadConfigFromEnv(prefix string) (*Config, error) {
	// Create client with prefix
	env := envutil.NewWithOptions(
		envutil.WithPrefix(prefix),
		envutil.WithSilent(true),
	)

	config := &Config{
		// Connection settings
		URI:      env.GetString("URI", ""),
		Database: env.GetString("DATABASE", env.GetString("DB", "")),

		// Pool settings
		MaxPoolSize:     uint64(env.GetInt("MAX_POOL_SIZE", 100)),
		MinPoolSize:     uint64(env.GetInt("MIN_POOL_SIZE", 10)),
		MaxConnIdleTime: env.GetDuration("MAX_CONN_IDLE_TIME", 5*time.Minute),

		// Timeout settings
		ConnectTimeout:         env.GetDuration("CONNECT_TIMEOUT", 10*time.Second),
		SocketTimeout:          env.GetDuration("SOCKET_TIMEOUT", 30*time.Second),
		ServerSelectionTimeout: env.GetDuration("SERVER_SELECTION_TIMEOUT", 10*time.Second),
		OperationTimeout:       env.GetDuration("OPERATION_TIMEOUT", 30*time.Second),

		// Retry settings
		RetryWrites: env.GetBool("RETRY_WRITES", true),
		RetryReads:  env.GetBool("RETRY_READS", true),
	}

	// Build URI if not provided directly
	if config.URI == "" {
		config.URI = buildURIFromEnv(env)
	}

	// Validate required fields
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration from environment: %w", err)
	}

	return config, nil
}

// LoadConfigFromEnvWithDefaults loads config from environment with a default prefix
// Uses "MONGO_" as the default prefix (e.g., MONGO_URI, MONGO_DATABASE)
func LoadConfigFromEnvWithDefaults() (*Config, error) {
	return LoadConfigFromEnv("MONGO_")
}

// buildURIFromEnv constructs MongoDB URI from individual environment variables
func buildURIFromEnv(env *envutil.Client) string {
	// Check for URL (alias for URI)
	if uri := env.GetString("URL", ""); uri != "" {
		return uri
	}

	// Build from components
	host := env.GetString("HOST", "localhost")
	port := env.GetInt("PORT", 27017)
	username := env.GetString("USERNAME", "")
	password := env.GetString("PASSWORD", "")
	authSource := env.GetString("AUTH_SOURCE", "admin")

	// Build connection string
	var uri string
	if username != "" && password != "" {
		// With authentication
		uri = fmt.Sprintf("mongodb://%s:%s@%s:%d/?authSource=%s",
			username, password, host, port, authSource)
	} else {
		// Without authentication
		uri = fmt.Sprintf("mongodb://%s:%d", host, port)
	}

	return uri
}

// NewFromEnv creates a new MongoDB client from environment variables
// with optional prefix
func NewFromEnv(ctx context.Context, prefix string) (*Client, error) {
	config, err := LoadConfigFromEnv(prefix)
	if err != nil {
		return nil, err
	}

	return New(ctx, config)
}

// NewFromEnvWithDefaults creates a new MongoDB client from environment
// using default "MONGO_" prefix
func NewFromEnvWithDefaults(ctx context.Context) (*Client, error) {
	return NewFromEnv(ctx, "MONGO_")
}

// MustLoadConfigFromEnv loads config from environment or panics
func MustLoadConfigFromEnv(prefix string) *Config {
	config, err := LoadConfigFromEnv(prefix)
	if err != nil {
		panic(fmt.Sprintf("failed to load MongoDB config from environment: %v", err))
	}
	return config
}
