package redisclient

import (
	"context"
	"testing"
	"time"
)

// TestConfig_Validate tests the configuration validation
func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name:    "valid config",
			config:  DefaultConfig(),
			wantErr: false,
		},
		{
			name: "empty addr",
			config: &Config{
				Addr:        "",
				DB:          0,
				PoolSize:    100,
				DialTimeout: 5 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "negative DB",
			config: &Config{
				Addr:        "localhost:6379",
				DB:          -1,
				PoolSize:    100,
				DialTimeout: 5 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "negative MaxRetries",
			config: &Config{
				Addr:        "localhost:6379",
				DB:          0,
				MaxRetries:  -1,
				PoolSize:    100,
				DialTimeout: 5 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "zero pool size",
			config: &Config{
				Addr:        "localhost:6379",
				DB:          0,
				PoolSize:    0,
				DialTimeout: 5 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "zero dial timeout",
			config: &Config{
				Addr:        "localhost:6379",
				DB:          0,
				PoolSize:    100,
				DialTimeout: 0,
			},
			wantErr: true,
		},
		{
			name: "TLS with only cert file",
			config: &Config{
				Addr:        "localhost:6379",
				DB:          0,
				PoolSize:    100,
				DialTimeout: 5 * time.Second,
				TLSEnabled:  true,
				TLSCertFile: "/path/to/cert.pem",
				TLSKeyFile:  "",
			},
			wantErr: true,
		},
		{
			name: "TLS with only key file",
			config: &Config{
				Addr:        "localhost:6379",
				DB:          0,
				PoolSize:    100,
				DialTimeout: 5 * time.Second,
				TLSEnabled:  true,
				TLSCertFile: "",
				TLSKeyFile:  "/path/to/key.pem",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestDefaultConfig tests the default configuration
func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.Addr != "localhost:6379" {
		t.Errorf("DefaultConfig().Addr = %v, want localhost:6379", config.Addr)
	}

	if config.DB != 0 {
		t.Errorf("DefaultConfig().DB = %v, want 0", config.DB)
	}

	if config.PoolSize != 100 {
		t.Errorf("DefaultConfig().PoolSize = %v, want 100", config.PoolSize)
	}

	if config.DialTimeout != 5*time.Second {
		t.Errorf("DefaultConfig().DialTimeout = %v, want 5s", config.DialTimeout)
	}

	if err := config.Validate(); err != nil {
		t.Errorf("DefaultConfig().Validate() error = %v, want nil", err)
	}
}

// TestNewWithOptions tests creating client with functional options
func TestNewWithOptions(t *testing.T) {
	// Note: This test doesn't actually connect to Redis
	// We're just testing the config building
	config := DefaultConfig()

	// Apply some options
	opt1 := WithAddr("redis:6379")
	opt2 := WithPassword("secret")
	opt3 := WithDB(1)
	opt4 := WithPoolSize(50)

	opt1(config)
	opt2(config)
	opt3(config)
	opt4(config)

	if config.Addr != "redis:6379" {
		t.Errorf("WithAddr() failed, got %v", config.Addr)
	}

	if config.Password != "secret" {
		t.Errorf("WithPassword() failed, got %v", config.Password)
	}

	if config.DB != 1 {
		t.Errorf("WithDB() failed, got %v", config.DB)
	}

	if config.PoolSize != 50 {
		t.Errorf("WithPoolSize() failed, got %v", config.PoolSize)
	}
}

// TestOptions tests all functional options
func TestOptions(t *testing.T) {
	tests := []struct {
		name     string
		option   Option
		validate func(*Config) error
	}{
		{
			name:   "WithAddr",
			option: WithAddr("redis:6379"),
			validate: func(c *Config) error {
				if c.Addr != "redis:6379" {
					t.Errorf("WithAddr() = %v, want redis:6379", c.Addr)
				}
				return nil
			},
		},
		{
			name:   "WithPassword",
			option: WithPassword("test123"),
			validate: func(c *Config) error {
				if c.Password != "test123" {
					t.Errorf("WithPassword() = %v, want test123", c.Password)
				}
				return nil
			},
		},
		{
			name:   "WithDB",
			option: WithDB(5),
			validate: func(c *Config) error {
				if c.DB != 5 {
					t.Errorf("WithDB() = %v, want 5", c.DB)
				}
				return nil
			},
		},
		{
			name:   "WithMaxRetries",
			option: WithMaxRetries(10),
			validate: func(c *Config) error {
				if c.MaxRetries != 10 {
					t.Errorf("WithMaxRetries() = %v, want 10", c.MaxRetries)
				}
				return nil
			},
		},
		{
			name:   "WithPoolSize",
			option: WithPoolSize(200),
			validate: func(c *Config) error {
				if c.PoolSize != 200 {
					t.Errorf("WithPoolSize() = %v, want 200", c.PoolSize)
				}
				return nil
			},
		},
		{
			name:   "WithMinIdleConns",
			option: WithMinIdleConns(10),
			validate: func(c *Config) error {
				if c.MinIdleConns != 10 {
					t.Errorf("WithMinIdleConns() = %v, want 10", c.MinIdleConns)
				}
				return nil
			},
		},
		{
			name:   "WithDialTimeout",
			option: WithDialTimeout(10 * time.Second),
			validate: func(c *Config) error {
				if c.DialTimeout != 10*time.Second {
					t.Errorf("WithDialTimeout() = %v, want 10s", c.DialTimeout)
				}
				return nil
			},
		},
		{
			name:   "WithReadTimeout",
			option: WithReadTimeout(5 * time.Second),
			validate: func(c *Config) error {
				if c.ReadTimeout != 5*time.Second {
					t.Errorf("WithReadTimeout() = %v, want 5s", c.ReadTimeout)
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultConfig()
			tt.option(config)
			tt.validate(config)
		})
	}
}

// TestErrorHelpers tests error helper functions
func TestErrorHelpers(t *testing.T) {
	t.Run("IsNil with ErrNil", func(t *testing.T) {
		if !IsNil(ErrNil) {
			t.Error("IsNil(ErrNil) = false, want true")
		}
	})

	t.Run("IsNil with other error", func(t *testing.T) {
		if IsNil(ErrClientClosed) {
			t.Error("IsNil(ErrClientClosed) = true, want false")
		}
	})

	t.Run("IsConnectionError with ErrClientClosed", func(t *testing.T) {
		if !IsConnectionError(ErrClientClosed) {
			t.Error("IsConnectionError(ErrClientClosed) = false, want true")
		}
	})

	t.Run("IsConnectionError with other error", func(t *testing.T) {
		if IsConnectionError(ErrInvalidKey) {
			t.Error("IsConnectionError(ErrInvalidKey) = true, want false")
		}
	})
}

// TestClient_ClosedOperations tests operations on a closed client
func TestClient_ClosedOperations(t *testing.T) {
	// Create a mock closed client
	client := &Client{
		config: DefaultConfig(),
		closed: true,
	}

	ctx := context.Background()

	t.Run("Ping on closed client", func(t *testing.T) {
		err := client.Ping(ctx)
		if err != ErrClientClosed {
			t.Errorf("Ping() error = %v, want ErrClientClosed", err)
		}
	})

	t.Run("Get on closed client", func(t *testing.T) {
		_, err := client.Get(ctx, "key")
		if err != ErrClientClosed {
			t.Errorf("Get() error = %v, want ErrClientClosed", err)
		}
	})

	t.Run("Set on closed client", func(t *testing.T) {
		err := client.Set(ctx, "key", "value", 0)
		if err != ErrClientClosed {
			t.Errorf("Set() error = %v, want ErrClientClosed", err)
		}
	})

	t.Run("Del on closed client", func(t *testing.T) {
		_, err := client.Del(ctx, "key")
		if err != ErrClientClosed {
			t.Errorf("Del() error = %v, want ErrClientClosed", err)
		}
	})

	t.Run("Close already closed client", func(t *testing.T) {
		err := client.Close()
		if err != ErrAlreadyClosed {
			t.Errorf("Close() error = %v, want ErrAlreadyClosed", err)
		}
	})
}

// TestClient_InvalidKeyOperations tests operations with invalid keys
func TestClient_InvalidKeyOperations(t *testing.T) {
	// Create a client without actual Redis connection
	client := &Client{
		config: DefaultConfig(),
		closed: false,
	}

	ctx := context.Background()

	t.Run("Get with empty key", func(t *testing.T) {
		_, err := client.Get(ctx, "")
		if err != ErrInvalidKey {
			t.Errorf("Get(\"\") error = %v, want ErrInvalidKey", err)
		}
	})

	t.Run("Set with empty key", func(t *testing.T) {
		err := client.Set(ctx, "", "value", 0)
		if err != ErrInvalidKey {
			t.Errorf("Set(\"\") error = %v, want ErrInvalidKey", err)
		}
	})

	t.Run("Incr with empty key", func(t *testing.T) {
		_, err := client.Incr(ctx, "")
		if err != ErrInvalidKey {
			t.Errorf("Incr(\"\") error = %v, want ErrInvalidKey", err)
		}
	})

	t.Run("HSet with empty key", func(t *testing.T) {
		_, err := client.HSet(ctx, "", "field", "value")
		if err != ErrInvalidKey {
			t.Errorf("HSet(\"\") error = %v, want ErrInvalidKey", err)
		}
	})

	t.Run("LPush with empty key", func(t *testing.T) {
		_, err := client.LPush(ctx, "", "value")
		if err != ErrInvalidKey {
			t.Errorf("LPush(\"\") error = %v, want ErrInvalidKey", err)
		}
	})
}

// TestSetEX_InvalidTTL tests SetEX with invalid TTL
func TestSetEX_InvalidTTL(t *testing.T) {
	client := &Client{
		config: DefaultConfig(),
		closed: false,
	}

	ctx := context.Background()

	t.Run("SetEX with zero TTL", func(t *testing.T) {
		err := client.SetEX(ctx, "key", "value", 0)
		if err != ErrInvalidTTL {
			t.Errorf("SetEX() with TTL=0 error = %v, want ErrInvalidTTL", err)
		}
	})

	t.Run("SetEX with negative TTL", func(t *testing.T) {
		err := client.SetEX(ctx, "key", "value", -1*time.Second)
		if err != ErrInvalidTTL {
			t.Errorf("SetEX() with negative TTL error = %v, want ErrInvalidTTL", err)
		}
	})
}
