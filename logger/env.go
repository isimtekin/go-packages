package logger

import (
	"fmt"
	"os"
	"time"

	envutil "github.com/isimtekin/go-packages/env-util"
)

// LoadConfigFromEnv loads logger configuration from environment variables
// with a custom prefix (e.g., "APP_LOG_")
func LoadConfigFromEnv(prefix string) (*Config, error) {
	config := DefaultConfig()
	env := envutil.NewWithOptions(
		envutil.WithPrefix(prefix),
	)

	// Basic settings
	config.Enabled = env.GetBool("ENABLED", config.Enabled)
	config.Level = env.GetString("LEVEL", config.Level)
	config.Output = env.GetString("OUTPUT", config.Output)
	config.Format = env.GetString("FORMAT", config.Format)

	// Console settings
	config.ConsoleColors = env.GetBool("CONSOLE_COLORS", config.ConsoleColors)

	// Logstash settings
	config.LogstashHost = env.GetString("LOGSTASH_HOST", config.LogstashHost)
	config.LogstashPort = env.GetInt("LOGSTASH_PORT", config.LogstashPort)
	config.LogstashProtocol = env.GetString("LOGSTASH_PROTOCOL", config.LogstashProtocol)

	if timeout := env.GetString("LOGSTASH_TIMEOUT", ""); timeout != "" {
		if d, err := time.ParseDuration(timeout); err == nil {
			config.LogstashTimeout = d
		}
	}
	config.LogstashRetries = env.GetInt("LOGSTASH_RETRIES", config.LogstashRetries)

	// File settings
	config.FilePath = env.GetString("FILE_PATH", config.FilePath)
	config.FileMaxSize = env.GetInt("FILE_MAX_SIZE", config.FileMaxSize)
	config.FileMaxAge = env.GetInt("FILE_MAX_AGE", config.FileMaxAge)
	config.FileMaxBackups = env.GetInt("FILE_MAX_BACKUPS", config.FileMaxBackups)

	// Application metadata
	config.ServiceName = env.GetString("SERVICE_NAME", config.ServiceName)
	config.Environment = env.GetString("ENVIRONMENT", config.Environment)
	config.Version = env.GetString("VERSION", config.Version)
	config.Host = env.GetString("HOST", config.Host)

	// Performance settings
	config.ReportCaller = env.GetBool("REPORT_CALLER", config.ReportCaller)

	return config, nil
}

// LoadConfigFromEnvWithDefaults loads configuration using default "LOG_" prefix
func LoadConfigFromEnvWithDefaults() (*Config, error) {
	return LoadConfigFromEnv("LOG_")
}

// NewFromEnv creates a new logger from environment variables with custom prefix
func NewFromEnv(prefix string) (*Logger, error) {
	config, err := LoadConfigFromEnv(prefix)
	if err != nil {
		return nil, fmt.Errorf("failed to load config from env: %w", err)
	}

	return New(config)
}

// NewFromEnvWithDefaults creates a new logger from environment variables with default prefix
func NewFromEnvWithDefaults() (*Logger, error) {
	return NewFromEnv("LOG_")
}

// SetGlobalFromEnv sets the global logger from environment variables
func SetGlobalFromEnv(prefix string) error {
	logger, err := NewFromEnv(prefix)
	if err != nil {
		return err
	}
	SetGlobal(logger)
	return nil
}

// SetGlobalFromEnvWithDefaults sets the global logger from environment with default prefix
func SetGlobalFromEnvWithDefaults() error {
	return SetGlobalFromEnv("LOG_")
}

// getEnv is a helper function to get environment variable with default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
