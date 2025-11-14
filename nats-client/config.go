package natsclient

import (
	"fmt"
	"time"
)

// Config holds the configuration for NATS client
type Config struct {
	// Connection settings
	URL      string `json:"url" yaml:"url"`           // NATS server URL (nats://localhost:4222)
	Name     string `json:"name" yaml:"name"`         // Client name
	Username string `json:"username" yaml:"username"` // Username for authentication
	Password string `json:"password" yaml:"password"` // Password for authentication
	Token    string `json:"token" yaml:"token"`       // Token for authentication

	// Connection pool settings
	MaxReconnects     int           `json:"max_reconnects" yaml:"max_reconnects"`
	ReconnectWait     time.Duration `json:"reconnect_wait" yaml:"reconnect_wait"`
	ReconnectJitter   time.Duration `json:"reconnect_jitter" yaml:"reconnect_jitter"`
	Timeout           time.Duration `json:"timeout" yaml:"timeout"`
	PingInterval      time.Duration `json:"ping_interval" yaml:"ping_interval"`
	MaxPingsOut       int           `json:"max_pings_out" yaml:"max_pings_out"`
	AllowReconnect    bool          `json:"allow_reconnect" yaml:"allow_reconnect"`
	NoRandomize       bool          `json:"no_randomize" yaml:"no_randomize"`
	NoEcho            bool          `json:"no_echo" yaml:"no_echo"`
	RetryOnFailedConn bool          `json:"retry_on_failed_conn" yaml:"retry_on_failed_conn"`

	// TLS settings
	TLSEnabled  bool   `json:"tls_enabled" yaml:"tls_enabled"`
	TLSCertFile string `json:"tls_cert_file" yaml:"tls_cert_file"`
	TLSKeyFile  string `json:"tls_key_file" yaml:"tls_key_file"`
	TLSCAFile   string `json:"tls_ca_file" yaml:"tls_ca_file"`

	// JetStream settings
	EnableJetStream bool `json:"enable_jetstream" yaml:"enable_jetstream"`
}

// DefaultConfig returns the default configuration for NATS
func DefaultConfig() *Config {
	return &Config{
		URL:               "nats://localhost:4222",
		Name:              "nats-client",
		MaxReconnects:     60,
		ReconnectWait:     2 * time.Second,
		ReconnectJitter:   100 * time.Millisecond,
		Timeout:           2 * time.Second,
		PingInterval:      2 * time.Minute,
		MaxPingsOut:       2,
		AllowReconnect:    true,
		NoRandomize:       false,
		NoEcho:            false,
		RetryOnFailedConn: true,
		TLSEnabled:        false,
		EnableJetStream:   false,
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.URL == "" {
		return fmt.Errorf("url cannot be empty")
	}

	if c.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}

	if c.MaxReconnects < -1 {
		return fmt.Errorf("max_reconnects must be >= -1")
	}

	if c.TLSEnabled {
		if c.TLSCertFile == "" && c.TLSKeyFile != "" {
			return fmt.Errorf("tls_cert_file required when tls_key_file is set")
		}
		if c.TLSKeyFile == "" && c.TLSCertFile != "" {
			return fmt.Errorf("tls_key_file required when tls_cert_file is set")
		}
	}

	return nil
}
