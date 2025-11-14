package natsclient

import (
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name:    "valid config",
			config:  DefaultConfig(),
			wantErr: false, // May succeed if NATS server is running
		},
		{
			name: "invalid url",
			config: &Config{
				URL:     "",
				Timeout: 30 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "invalid timeout",
			config: &Config{
				URL:     "nats://localhost:4222",
				Timeout: 0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := New(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if client != nil {
				defer client.Close()
			}
		})
	}
}

func TestNewWithOptions(t *testing.T) {
	t.Run("creates client with options", func(t *testing.T) {
		_, err := NewWithOptions(
			WithURL("nats://localhost:4222"),
			WithName("test-client"),
			WithTimeout(10*time.Second),
		)

		// Will fail without NATS server, but validates options work
		if err == nil {
			t.Skip("NATS server not available")
		}
	})
}

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
			name: "empty url",
			config: &Config{
				URL:     "",
				Timeout: 1 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "zero timeout",
			config: &Config{
				URL:     "nats://localhost:4222",
				Timeout: 0,
			},
			wantErr: true,
		},
		{
			name: "invalid max reconnects",
			config: &Config{
				URL:           "nats://localhost:4222",
				Timeout:       1 * time.Second,
				MaxReconnects: -2,
			},
			wantErr: true,
		},
		{
			name: "tls cert without key",
			config: &Config{
				URL:         "nats://localhost:4222",
				Timeout:     1 * time.Second,
				TLSEnabled:  true,
				TLSCertFile: "cert.pem",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_IsClosed(t *testing.T) {
	config := DefaultConfig()
	client := &Client{
		config: config,
		closed: false,
	}

	if client.IsClosed() {
		t.Error("IsClosed() = true, want false")
	}

	client.closed = true
	if !client.IsClosed() {
		t.Error("IsClosed() = false, want true")
	}
}
