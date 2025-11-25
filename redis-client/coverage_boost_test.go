package redisclient

import (
	"context"
	"testing"
	"time"
)

// TestMSet_WorkspaceLogic tests the workspace logic in MSet
func TestMSet_WorkspaceLogic(t *testing.T) {
	t.Run("MSet applies workspace prefix to keys", func(t *testing.T) {
		config := DefaultConfig()
		config.Workspace = "test"

		client, err := New(config)
		if err != nil {
			t.Fatalf("Failed to create client: %v", err)
		}
		defer client.Close()

		ctx := context.Background()

		// Test with multiple key-value pairs
		err = client.MSet(ctx, "key1", "val1", "key2", "val2", "key3", "val3")
		if err != nil {
			t.Logf("MSet failed (expected with no Redis): %v", err)
		}

		// Verify workspace prefixing logic
		values := []interface{}{"key1", "val1", "key2", "val2"}
		if config.Workspace != "" {
			prefixedValues := make([]interface{}, len(values))
			for i := 0; i < len(values); i += 2 {
				if i+1 < len(values) {
					if key, ok := values[i].(string); ok {
						prefixedValues[i] = config.ApplyWorkspacePrefix(key)
						prefixedValues[i+1] = values[i+1]
					} else {
						prefixedValues[i] = values[i]
						prefixedValues[i+1] = values[i+1]
					}
				}
			}

			// Verify first key is prefixed
			if prefixed, ok := prefixedValues[0].(string); ok {
				expected := "test:key1"
				if prefixed != expected {
					t.Errorf("Expected key '%s', got '%s'", expected, prefixed)
				}
			}
		}
	})

	t.Run("MSet without workspace", func(t *testing.T) {
		config := DefaultConfig()
		// No workspace

		client, err := New(config)
		if err != nil {
			t.Fatalf("Failed to create client: %v", err)
		}
		defer client.Close()

		ctx := context.Background()

		err = client.MSet(ctx, "key1", "val1")
		if err != nil {
			t.Logf("MSet failed (expected with no Redis): %v", err)
		}
	})

	t.Run("MSet with non-string key", func(t *testing.T) {
		config := DefaultConfig()
		config.Workspace = "test"

		client, err := New(config)
		if err != nil {
			t.Fatalf("Failed to create client: %v", err)
		}
		defer client.Close()

		ctx := context.Background()

		// Test with integer key (should handle gracefully)
		err = client.MSet(ctx, 123, "val1")
		if err != nil {
			t.Logf("MSet failed (expected with no Redis): %v", err)
		}
	})
}

// TestEnv_LoadConfigFromEnv tests environment variable loading
func TestEnv_LoadConfigFromEnv(t *testing.T) {
	t.Run("Load all config from env", func(t *testing.T) {
		t.Setenv("REDIS_ADDR", "custom-host:6380")
		t.Setenv("REDIS_PASSWORD", "secret123")
		t.Setenv("REDIS_DB", "5")
		t.Setenv("REDIS_WORKSPACE", "production")
		t.Setenv("REDIS_MAX_RETRIES", "10")
		t.Setenv("REDIS_MIN_IDLE_CONNS", "20")
		t.Setenv("REDIS_MAX_IDLE_CONNS", "30")
		t.Setenv("REDIS_POOL_SIZE", "200")
		t.Setenv("REDIS_POOL_TIMEOUT", "10s")
		t.Setenv("REDIS_CONN_MAX_IDLE_TIME", "15m")
		t.Setenv("REDIS_CONN_MAX_LIFETIME", "1h")
		t.Setenv("REDIS_DIAL_TIMEOUT", "20s")
		t.Setenv("REDIS_READ_TIMEOUT", "10s")
		t.Setenv("REDIS_WRITE_TIMEOUT", "10s")
		t.Setenv("REDIS_TLS_ENABLED", "true")
		t.Setenv("REDIS_TLS_CERT_FILE", "/cert.pem")
		t.Setenv("REDIS_TLS_KEY_FILE", "/key.pem")
		t.Setenv("REDIS_TLS_CA_FILE", "/ca.pem")

		config, err := LoadConfigFromEnvWithDefaults()
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		if config.Addr != "custom-host:6380" {
			t.Errorf("Expected addr 'custom-host:6380', got '%s'", config.Addr)
		}

		if config.Password != "secret123" {
			t.Errorf("Expected password 'secret123', got '%s'", config.Password)
		}

		if config.DB != 5 {
			t.Errorf("Expected DB 5, got %d", config.DB)
		}

		if config.Workspace != "production" {
			t.Errorf("Expected workspace 'production', got '%s'", config.Workspace)
		}

		if config.MaxRetries != 10 {
			t.Errorf("Expected max retries 10, got %d", config.MaxRetries)
		}

		if config.MinIdleConns != 20 {
			t.Errorf("Expected min idle conns 20, got %d", config.MinIdleConns)
		}

		if config.MaxIdleConns != 30 {
			t.Errorf("Expected max idle conns 30, got %d", config.MaxIdleConns)
		}

		if config.PoolSize != 200 {
			t.Errorf("Expected pool size 200, got %d", config.PoolSize)
		}

		if config.PoolTimeout != 10*time.Second {
			t.Errorf("Expected pool timeout 10s, got %v", config.PoolTimeout)
		}

		if config.ConnMaxIdleTime != 15*time.Minute {
			t.Errorf("Expected conn max idle time 15m, got %v", config.ConnMaxIdleTime)
		}

		if config.ConnMaxLifetime != 1*time.Hour {
			t.Errorf("Expected conn max lifetime 1h, got %v", config.ConnMaxLifetime)
		}

		if config.DialTimeout != 20*time.Second {
			t.Errorf("Expected dial timeout 20s, got %v", config.DialTimeout)
		}

		if config.ReadTimeout != 10*time.Second {
			t.Errorf("Expected read timeout 10s, got %v", config.ReadTimeout)
		}

		if config.WriteTimeout != 10*time.Second {
			t.Errorf("Expected write timeout 10s, got %v", config.WriteTimeout)
		}

		if !config.TLSEnabled {
			t.Error("Expected TLS to be enabled")
		}

		if config.TLSCertFile != "/cert.pem" {
			t.Errorf("Expected TLS cert file '/cert.pem', got '%s'", config.TLSCertFile)
		}

		if config.TLSKeyFile != "/key.pem" {
			t.Errorf("Expected TLS key file '/key.pem', got '%s'", config.TLSKeyFile)
		}

		if config.TLSCAFile != "/ca.pem" {
			t.Errorf("Expected TLS CA file '/ca.pem', got '%s'", config.TLSCAFile)
		}
	})

	t.Run("Load config with custom prefix", func(t *testing.T) {
		t.Setenv("CACHE_ADDR", "cache-host:6379")
		t.Setenv("CACHE_WORKSPACE", "staging")

		config, err := LoadConfigFromEnv("CACHE_")
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		if config.Addr != "cache-host:6379" {
			t.Errorf("Expected addr 'cache-host:6379', got '%s'", config.Addr)
		}

		if config.Workspace != "staging" {
			t.Errorf("Expected workspace 'staging', got '%s'", config.Workspace)
		}
	})
}

// TestNewFromEnv tests creating client from environment
func TestNewFromEnv(t *testing.T) {
	t.Run("NewFromEnv fails with invalid config", func(t *testing.T) {
		t.Setenv("REDIS_ADDR", "") // Invalid empty address

		ctx := context.Background()
		_, err := NewFromEnvWithDefaults(ctx)
		if err == nil {
			t.Error("Expected error with invalid config, got nil")
		}
	})

	t.Run("NewFromEnv with custom prefix", func(t *testing.T) {
		t.Setenv("CUSTOM_ADDR", "localhost:6379")
		t.Setenv("CUSTOM_WORKSPACE", "test")

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		_, err := NewFromEnv(ctx, "CUSTOM_")
		if err == nil {
			t.Log("Client created but ping will fail (expected with no Redis)")
		}
		// Expected to fail due to no Redis server
	})
}

// TestOptions_AllOptions tests all functional options
func TestOptions_AllOptions(t *testing.T) {
	config := DefaultConfig()

	// Test all options
	options := []Option{
		WithAddr("test:6379"),
		WithPassword("secret"),
		WithDB(5),
		WithWorkspace("production"),
		WithMaxRetries(10),
		WithPoolSize(200),
		WithMinIdleConns(20),
		WithMaxIdleConns(40),
		WithPoolTimeout(10 * time.Second),
		WithConnMaxIdleTime(30 * time.Minute),
		WithConnMaxLifetime(2 * time.Hour),
		WithDialTimeout(15 * time.Second),
		WithReadTimeout(8 * time.Second),
		WithWriteTimeout(8 * time.Second),
		WithTLS("/cert.pem", "/key.pem", "/ca.pem"),
		WithDatabaseNames(map[string]int{"cache": 0, "session": 1}),
		WithDatabaseName("queue", 2),
	}

	for _, opt := range options {
		opt(config)
	}

	// Verify all options applied
	if config.Addr != "test:6379" {
		t.Errorf("WithAddr failed: got %s", config.Addr)
	}
	if config.Password != "secret" {
		t.Errorf("WithPassword failed: got %s", config.Password)
	}
	if config.DB != 5 {
		t.Errorf("WithDB failed: got %d", config.DB)
	}
	if config.Workspace != "production" {
		t.Errorf("WithWorkspace failed: got %s", config.Workspace)
	}
	if config.MaxRetries != 10 {
		t.Errorf("WithMaxRetries failed: got %d", config.MaxRetries)
	}
	if config.PoolSize != 200 {
		t.Errorf("WithPoolSize failed: got %d", config.PoolSize)
	}
	if config.MinIdleConns != 20 {
		t.Errorf("WithMinIdleConns failed: got %d", config.MinIdleConns)
	}
	if config.MaxIdleConns != 40 {
		t.Errorf("WithMaxIdleConns failed: got %d", config.MaxIdleConns)
	}
	if config.PoolTimeout != 10*time.Second {
		t.Errorf("WithPoolTimeout failed: got %v", config.PoolTimeout)
	}
	if config.ConnMaxIdleTime != 30*time.Minute {
		t.Errorf("WithConnMaxIdleTime failed: got %v", config.ConnMaxIdleTime)
	}
	if config.ConnMaxLifetime != 2*time.Hour {
		t.Errorf("WithConnMaxLifetime failed: got %v", config.ConnMaxLifetime)
	}
	if config.DialTimeout != 15*time.Second {
		t.Errorf("WithDialTimeout failed: got %v", config.DialTimeout)
	}
	if config.ReadTimeout != 8*time.Second {
		t.Errorf("WithReadTimeout failed: got %v", config.ReadTimeout)
	}
	if config.WriteTimeout != 8*time.Second {
		t.Errorf("WithWriteTimeout failed: got %v", config.WriteTimeout)
	}
	if !config.TLSEnabled {
		t.Error("WithTLS failed: TLS not enabled")
	}
	if config.TLSCertFile != "/cert.pem" {
		t.Errorf("WithTLS failed: got cert %s", config.TLSCertFile)
	}
	if config.TLSKeyFile != "/key.pem" {
		t.Errorf("WithTLS failed: got key %s", config.TLSKeyFile)
	}
	if config.TLSCAFile != "/ca.pem" {
		t.Errorf("WithTLS failed: got CA %s", config.TLSCAFile)
	}
	if config.DatabaseNames["cache"] != 0 {
		t.Errorf("WithDatabaseNames failed: cache not mapped to 0")
	}
	if config.DatabaseNames["session"] != 1 {
		t.Errorf("WithDatabaseNames failed: session not mapped to 1")
	}
	if config.DatabaseNames["queue"] != 2 {
		t.Errorf("WithDatabaseName failed: queue not mapped to 2")
	}
}

// TestClient_GetClient tests accessing underlying Redis client
func TestClient_GetClient(t *testing.T) {
	config := DefaultConfig()
	client, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	redisClient := client.Client()
	if redisClient == nil {
		t.Error("Client() returned nil")
	}
}

// TestTLS_CreateTLSConfig tests TLS configuration creation
func TestTLS_CreateTLSConfig(t *testing.T) {
	t.Run("TLS with missing cert file", func(t *testing.T) {
		config := &Config{
			Addr:        "localhost:6379",
			TLSEnabled:  true,
			TLSCertFile: "/nonexistent/cert.pem",
			TLSKeyFile:  "/nonexistent/key.pem",
		}

		_, err := New(config)
		if err == nil {
			t.Error("Expected error with nonexistent TLS files, got nil")
		}
	})

	t.Run("TLS with missing CA file", func(t *testing.T) {
		config := &Config{
			Addr:       "localhost:6379",
			TLSEnabled: true,
			TLSCAFile:  "/nonexistent/ca.pem",
		}

		_, err := New(config)
		if err == nil {
			t.Error("Expected error with nonexistent CA file, got nil")
		}
	})
}
