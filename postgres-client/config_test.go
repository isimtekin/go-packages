package postgresclient

import (
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
				Host:     "localhost",
				Port:     5432,
				Database: "testdb",
				Username: "testuser",
				SSLMode:  "disable",
			},
			wantErr: false,
		},
		{
			name: "missing host",
			config: &Config{
				Port:     5432,
				Database: "testdb",
				Username: "testuser",
			},
			wantErr: true,
		},
		{
			name: "invalid port - zero",
			config: &Config{
				Host:     "localhost",
				Port:     0,
				Database: "testdb",
				Username: "testuser",
			},
			wantErr: true,
		},
		{
			name: "invalid port - too high",
			config: &Config{
				Host:     "localhost",
				Port:     70000,
				Database: "testdb",
				Username: "testuser",
			},
			wantErr: true,
		},
		{
			name: "missing database",
			config: &Config{
				Host:     "localhost",
				Port:     5432,
				Username: "testuser",
			},
			wantErr: true,
		},
		{
			name: "missing username",
			config: &Config{
				Host:     "localhost",
				Port:     5432,
				Database: "testdb",
			},
			wantErr: true,
		},
		{
			name: "invalid pool sizes",
			config: &Config{
				Host:     "localhost",
				Port:     5432,
				Database: "testdb",
				Username: "testuser",
				MinConns: 100,
				MaxConns: 10,
			},
			wantErr: true,
		},
		{
			name: "invalid SSL mode",
			config: &Config{
				Host:     "localhost",
				Port:     5432,
				Database: "testdb",
				Username: "testuser",
				SSLMode:  "invalid",
			},
			wantErr: true,
		},
		{
			name: "valid SSL mode - require",
			config: &Config{
				Host:     "localhost",
				Port:     5432,
				Database: "testdb",
				Username: "testuser",
				SSLMode:  "require",
			},
			wantErr: false,
		},
		{
			name: "valid SSL mode - verify-full",
			config: &Config{
				Host:     "localhost",
				Port:     5432,
				Database: "testdb",
				Username: "testuser",
				SSLMode:  "verify-full",
			},
			wantErr: false,
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

	if config.Host == "" {
		t.Error("Expected default Host to be set")
	}

	if config.Port == 0 {
		t.Error("Expected default Port to be set")
	}

	if config.Database == "" {
		t.Error("Expected default Database to be set")
	}

	if config.Username == "" {
		t.Error("Expected default Username to be set")
	}

	if config.MaxConns == 0 {
		t.Error("Expected default MaxConns to be set")
	}

	if config.ConnectTimeout == 0 {
		t.Error("Expected default ConnectTimeout to be set")
	}

	// Verify defaults are sensible
	if config.Port != 5432 {
		t.Errorf("Expected default Port to be 5432, got %d", config.Port)
	}

	if config.SSLMode != "disable" {
		t.Errorf("Expected default SSLMode to be 'disable', got %s", config.SSLMode)
	}
}

func TestConfig_DSN(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		expected string
	}{
		{
			name: "basic DSN",
			config: &Config{
				Host:     "localhost",
				Port:     5432,
				Database: "testdb",
				Username: "testuser",
			},
			expected: "host=localhost port=5432 dbname=testdb user=testuser",
		},
		{
			name: "DSN with password",
			config: &Config{
				Host:     "localhost",
				Port:     5432,
				Database: "testdb",
				Username: "testuser",
				Password: "secret",
			},
			expected: "host=localhost port=5432 dbname=testdb user=testuser password=secret",
		},
		{
			name: "DSN with SSL mode",
			config: &Config{
				Host:     "localhost",
				Port:     5432,
				Database: "testdb",
				Username: "testuser",
				SSLMode:  "require",
			},
			expected: "host=localhost port=5432 dbname=testdb user=testuser sslmode=require",
		},
		{
			name: "DSN with connect timeout",
			config: &Config{
				Host:           "localhost",
				Port:           5432,
				Database:       "testdb",
				Username:       "testuser",
				ConnectTimeout: 10 * time.Second,
			},
			expected: "host=localhost port=5432 dbname=testdb user=testuser connect_timeout=10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dsn := tt.config.DSN()
			if dsn != tt.expected {
				t.Errorf("Config.DSN() = %q, expected %q", dsn, tt.expected)
			}
		})
	}
}

func TestConfig_ConnectionURL(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		expected string
	}{
		{
			name: "basic URL",
			config: &Config{
				Host:     "localhost",
				Port:     5432,
				Database: "testdb",
				Username: "testuser",
			},
			expected: "postgres://testuser@localhost:5432/testdb",
		},
		{
			name: "URL with password",
			config: &Config{
				Host:     "localhost",
				Port:     5432,
				Database: "testdb",
				Username: "testuser",
				Password: "secret",
			},
			expected: "postgres://testuser:secret@localhost:5432/testdb",
		},
		{
			name: "URL with SSL mode",
			config: &Config{
				Host:     "localhost",
				Port:     5432,
				Database: "testdb",
				Username: "testuser",
				SSLMode:  "require",
			},
			expected: "postgres://testuser@localhost:5432/testdb?sslmode=require",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := tt.config.ConnectionURL()
			if url != tt.expected {
				t.Errorf("Config.ConnectionURL() = %q, expected %q", url, tt.expected)
			}
		})
	}
}
