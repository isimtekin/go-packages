package natsclient

import (
	"testing"
	"time"
)

func TestWithURL(t *testing.T) {
	config := DefaultConfig()
	opt := WithURL("nats://custom:4222")
	opt(config)

	if config.URL != "nats://custom:4222" {
		t.Errorf("WithURL() = %s, want nats://custom:4222", config.URL)
	}
}

func TestWithName(t *testing.T) {
	config := DefaultConfig()
	opt := WithName("my-client")
	opt(config)

	if config.Name != "my-client" {
		t.Errorf("WithName() = %s, want my-client", config.Name)
	}
}

func TestWithUsername(t *testing.T) {
	config := DefaultConfig()
	opt := WithUsername("testuser")
	opt(config)

	if config.Username != "testuser" {
		t.Errorf("WithUsername() = %s, want testuser", config.Username)
	}
}

func TestWithPassword(t *testing.T) {
	config := DefaultConfig()
	opt := WithPassword("secret123")
	opt(config)

	if config.Password != "secret123" {
		t.Errorf("WithPassword() = %s, want secret123", config.Password)
	}
}

func TestWithToken(t *testing.T) {
	config := DefaultConfig()
	opt := WithToken("my-token")
	opt(config)

	if config.Token != "my-token" {
		t.Errorf("WithToken() = %s, want my-token", config.Token)
	}
}

func TestWithMaxReconnects(t *testing.T) {
	config := DefaultConfig()
	opt := WithMaxReconnects(100)
	opt(config)

	if config.MaxReconnects != 100 {
		t.Errorf("WithMaxReconnects() = %d, want 100", config.MaxReconnects)
	}
}

func TestWithReconnectWait(t *testing.T) {
	config := DefaultConfig()
	opt := WithReconnectWait(5 * time.Second)
	opt(config)

	if config.ReconnectWait != 5*time.Second {
		t.Errorf("WithReconnectWait() = %v, want 5s", config.ReconnectWait)
	}
}

func TestWithTimeout(t *testing.T) {
	config := DefaultConfig()
	opt := WithTimeout(10 * time.Second)
	opt(config)

	if config.Timeout != 10*time.Second {
		t.Errorf("WithTimeout() = %v, want 10s", config.Timeout)
	}
}

func TestWithPingInterval(t *testing.T) {
	config := DefaultConfig()
	opt := WithPingInterval(30 * time.Second)
	opt(config)

	if config.PingInterval != 30*time.Second {
		t.Errorf("WithPingInterval() = %v, want 30s", config.PingInterval)
	}
}

func TestWithAllowReconnect(t *testing.T) {
	config := DefaultConfig()

	// Test disable
	opt := WithAllowReconnect(false)
	opt(config)
	if config.AllowReconnect != false {
		t.Errorf("WithAllowReconnect(false) = %v, want false", config.AllowReconnect)
	}

	// Test enable
	opt = WithAllowReconnect(true)
	opt(config)
	if config.AllowReconnect != true {
		t.Errorf("WithAllowReconnect(true) = %v, want true", config.AllowReconnect)
	}
}

func TestWithNoEcho(t *testing.T) {
	config := DefaultConfig()

	opt := WithNoEcho(true)
	opt(config)

	if config.NoEcho != true {
		t.Errorf("WithNoEcho() = %v, want true", config.NoEcho)
	}
}

func TestWithTLS(t *testing.T) {
	config := DefaultConfig()
	opt := WithTLS("cert.pem", "key.pem", "ca.pem")
	opt(config)

	if !config.TLSEnabled {
		t.Error("WithTLS() should enable TLS")
	}
	if config.TLSCertFile != "cert.pem" {
		t.Errorf("WithTLS() CertFile = %s, want cert.pem", config.TLSCertFile)
	}
	if config.TLSKeyFile != "key.pem" {
		t.Errorf("WithTLS() KeyFile = %s, want key.pem", config.TLSKeyFile)
	}
	if config.TLSCAFile != "ca.pem" {
		t.Errorf("WithTLS() CAFile = %s, want ca.pem", config.TLSCAFile)
	}
}

func TestWithJetStream(t *testing.T) {
	config := DefaultConfig()

	// Enable
	opt := WithJetStream(true)
	opt(config)
	if config.EnableJetStream != true {
		t.Errorf("WithJetStream(true) = %v, want true", config.EnableJetStream)
	}

	// Disable
	opt = WithJetStream(false)
	opt(config)
	if config.EnableJetStream != false {
		t.Errorf("WithJetStream(false) = %v, want false", config.EnableJetStream)
	}
}

func TestMultipleOptions(t *testing.T) {
	config := DefaultConfig()

	opts := []Option{
		WithURL("nats://myserver:4222"),
		WithName("test-client"),
		WithUsername("user"),
		WithPassword("pass"),
		WithTimeout(5 * time.Second),
		WithMaxReconnects(10),
		WithJetStream(true),
	}

	for _, opt := range opts {
		opt(config)
	}

	if config.URL != "nats://myserver:4222" {
		t.Errorf("URL = %s, want nats://myserver:4222", config.URL)
	}
	if config.Name != "test-client" {
		t.Errorf("Name = %s, want test-client", config.Name)
	}
	if config.Username != "user" {
		t.Errorf("Username = %s, want user", config.Username)
	}
	if config.Password != "pass" {
		t.Errorf("Password = %s, want pass", config.Password)
	}
	if config.Timeout != 5*time.Second {
		t.Errorf("Timeout = %v, want 5s", config.Timeout)
	}
	if config.MaxReconnects != 10 {
		t.Errorf("MaxReconnects = %d, want 10", config.MaxReconnects)
	}
	if !config.EnableJetStream {
		t.Error("EnableJetStream should be true")
	}
}
