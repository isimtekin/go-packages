package logger

import (
	"testing"
)

func TestLoadConfigFromEnv(t *testing.T) {
	t.Run("loads config with custom prefix", func(t *testing.T) {
		t.Setenv("APP_LOG_ENABLED", "false")
		t.Setenv("APP_LOG_LEVEL", "debug")
		t.Setenv("APP_LOG_OUTPUT", "console")
		t.Setenv("APP_LOG_FORMAT", "json")
		t.Setenv("APP_LOG_SERVICE_NAME", "test-service")
		t.Setenv("APP_LOG_ENVIRONMENT", "production")
		t.Setenv("APP_LOG_VERSION", "1.0.0")

		config, err := LoadConfigFromEnv("APP_LOG_")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if config.Enabled {
			t.Error("expected Enabled to be false")
		}
		if config.Level != "debug" {
			t.Errorf("expected Level 'debug', got %s", config.Level)
		}
		if config.Format != "json" {
			t.Errorf("expected Format 'json', got %s", config.Format)
		}
		if config.ServiceName != "test-service" {
			t.Errorf("expected ServiceName 'test-service', got %s", config.ServiceName)
		}
		if config.Environment != "production" {
			t.Errorf("expected Environment 'production', got %s", config.Environment)
		}
		if config.Version != "1.0.0" {
			t.Errorf("expected Version '1.0.0', got %s", config.Version)
		}
	})

	t.Run("loads config with default prefix", func(t *testing.T) {
		t.Setenv("LOG_ENABLED", "true")
		t.Setenv("LOG_LEVEL", "warn")
		t.Setenv("LOG_SERVICE_NAME", "default-service")

		config, err := LoadConfigFromEnvWithDefaults()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if !config.Enabled {
			t.Error("expected Enabled to be true")
		}
		if config.Level != "warn" {
			t.Errorf("expected Level 'warn', got %s", config.Level)
		}
		if config.ServiceName != "default-service" {
			t.Errorf("expected ServiceName 'default-service', got %s", config.ServiceName)
		}
	})

	t.Run("loads logstash config from env", func(t *testing.T) {
		t.Setenv("LOG_LOGSTASH_HOST", "logstash.example.com")
		t.Setenv("LOG_LOGSTASH_PORT", "5044")
		t.Setenv("LOG_LOGSTASH_PROTOCOL", "tcp")
		t.Setenv("LOG_LOGSTASH_TIMEOUT", "10s")
		t.Setenv("LOG_LOGSTASH_RETRIES", "5")

		config, err := LoadConfigFromEnvWithDefaults()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if config.LogstashHost != "logstash.example.com" {
			t.Errorf("expected LogstashHost 'logstash.example.com', got %s", config.LogstashHost)
		}
		if config.LogstashPort != 5044 {
			t.Errorf("expected LogstashPort 5044, got %d", config.LogstashPort)
		}
		if config.LogstashProtocol != "tcp" {
			t.Errorf("expected LogstashProtocol 'tcp', got %s", config.LogstashProtocol)
		}
		if config.LogstashRetries != 5 {
			t.Errorf("expected LogstashRetries 5, got %d", config.LogstashRetries)
		}
	})

	t.Run("loads file config from env", func(t *testing.T) {
		t.Setenv("LOG_FILE_PATH", "/var/log/app.log")
		t.Setenv("LOG_FILE_MAX_SIZE", "200")
		t.Setenv("LOG_FILE_MAX_AGE", "60")
		t.Setenv("LOG_FILE_MAX_BACKUPS", "20")

		config, err := LoadConfigFromEnvWithDefaults()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if config.FilePath != "/var/log/app.log" {
			t.Errorf("expected FilePath '/var/log/app.log', got %s", config.FilePath)
		}
		if config.FileMaxSize != 200 {
			t.Errorf("expected FileMaxSize 200, got %d", config.FileMaxSize)
		}
		if config.FileMaxAge != 60 {
			t.Errorf("expected FileMaxAge 60, got %d", config.FileMaxAge)
		}
		if config.FileMaxBackups != 20 {
			t.Errorf("expected FileMaxBackups 20, got %d", config.FileMaxBackups)
		}
	})

	t.Run("uses defaults when env vars not set", func(t *testing.T) {
		config, err := LoadConfigFromEnv("NONEXISTENT_")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		defaults := DefaultConfig()
		if config.Enabled != defaults.Enabled {
			t.Error("expected default Enabled value")
		}
		if config.Level != defaults.Level {
			t.Error("expected default Level value")
		}
		if config.Output != defaults.Output {
			t.Error("expected default Output value")
		}
	})
}

func TestNewFromEnv(t *testing.T) {
	t.Run("creates logger from env with custom prefix", func(t *testing.T) {
		t.Setenv("CUSTOM_ENABLED", "true")
		t.Setenv("CUSTOM_LEVEL", "debug")
		t.Setenv("CUSTOM_OUTPUT", "console")

		logger, err := NewFromEnv("CUSTOM_")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if !logger.IsEnabled() {
			t.Error("expected logger to be enabled")
		}
	})

	t.Run("creates logger from env with defaults", func(t *testing.T) {
		t.Setenv("LOG_ENABLED", "true")
		t.Setenv("LOG_LEVEL", "info")

		logger, err := NewFromEnvWithDefaults()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if !logger.IsEnabled() {
			t.Error("expected logger to be enabled")
		}
	})

	t.Run("fails with invalid config from env", func(t *testing.T) {
		t.Setenv("LOG_OUTPUT", "invalid")

		_, err := NewFromEnvWithDefaults()
		if err == nil {
			t.Error("expected error with invalid config")
		}
	})
}

func TestSetGlobalFromEnv(t *testing.T) {
	t.Run("sets global logger from env", func(t *testing.T) {
		t.Setenv("APP_SERVICE_NAME", "global-service")
		t.Setenv("APP_LEVEL", "debug")

		err := SetGlobalFromEnv("APP_")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		global := GetGlobal()
		if global.config.ServiceName != "global-service" {
			t.Error("global logger not set from env")
		}
	})

	t.Run("sets global logger with defaults", func(t *testing.T) {
		t.Setenv("LOG_SERVICE_NAME", "default-global")

		err := SetGlobalFromEnvWithDefaults()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		global := GetGlobal()
		if global.config.ServiceName != "default-global" {
			t.Error("global logger not set from env with defaults")
		}
	})
}
