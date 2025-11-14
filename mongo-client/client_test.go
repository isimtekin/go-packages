package mongoclient

import (
	"context"
	"testing"
	"time"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				URI:      "mongodb://localhost:27017",
				Database: "testdb",
			},
			wantErr: false,
		},
		{
			name: "missing URI",
			config: &Config{
				Database: "testdb",
			},
			wantErr: true,
		},
		{
			name: "missing database",
			config: &Config{
				URI: "mongodb://localhost:27017",
			},
			wantErr: true,
		},
		{
			name: "invalid pool sizes",
			config: &Config{
				URI:         "mongodb://localhost:27017",
				Database:    "testdb",
				MinPoolSize: 100,
				MaxPoolSize: 10,
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

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.URI == "" {
		t.Error("Expected default URI to be set")
	}

	if config.Database == "" {
		t.Error("Expected default Database to be set")
	}

	if config.MaxPoolSize == 0 {
		t.Error("Expected default MaxPoolSize to be set")
	}

	if config.ConnectTimeout == 0 {
		t.Error("Expected default ConnectTimeout to be set")
	}
}

func TestOptions(t *testing.T) {
	config := DefaultConfig()

	WithURI("mongodb://test:27017")(config)
	if config.URI != "mongodb://test:27017" {
		t.Errorf("Expected URI to be mongodb://test:27017, got %s", config.URI)
	}

	WithDatabase("testdb")(config)
	if config.Database != "testdb" {
		t.Errorf("Expected Database to be testdb, got %s", config.Database)
	}

	WithMaxPoolSize(50)(config)
	if config.MaxPoolSize != 50 {
		t.Errorf("Expected MaxPoolSize to be 50, got %d", config.MaxPoolSize)
	}

	WithConnectTimeout(5 * time.Second)(config)
	if config.ConnectTimeout != 5*time.Second {
		t.Errorf("Expected ConnectTimeout to be 5s, got %v", config.ConnectTimeout)
	}
}

func TestNew_NilConfig(t *testing.T) {
	ctx := context.Background()
	client, err := New(ctx, nil)

	if err != ErrNilConfig {
		t.Errorf("Expected ErrNilConfig, got %v", err)
	}

	if client != nil {
		t.Error("Expected client to be nil")
	}
}

func TestNew_InvalidConfig(t *testing.T) {
	ctx := context.Background()
	config := &Config{
		URI: "mongodb://localhost:27017",
		// Missing Database
	}

	client, err := New(ctx, config)

	if err == nil {
		t.Error("Expected error for invalid config")
	}

	if client != nil {
		t.Error("Expected client to be nil")
	}
}

// Note: Integration tests require a running MongoDB instance
// These are unit tests that don't require MongoDB

func TestGetTimeout(t *testing.T) {
	client := &Client{
		config: &Config{
			OperationTimeout: 15 * time.Second,
		},
	}

	timeout := client.GetTimeout()
	if timeout != 15*time.Second {
		t.Errorf("Expected timeout to be 15s, got %v", timeout)
	}

	// Test default timeout
	client.config.OperationTimeout = 0
	timeout = client.GetTimeout()
	if timeout != 30*time.Second {
		t.Errorf("Expected default timeout to be 30s, got %v", timeout)
	}
}
