package postgresclient

import (
	"context"
	"testing"
	"time"
)

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
		Host: "localhost",
		Port: 5432,
		// Missing Database and Username
	}

	client, err := New(ctx, config)

	if err == nil {
		t.Error("Expected error for invalid config")
	}

	if client != nil {
		t.Error("Expected client to be nil")
	}
}

func TestNewWithOptions(t *testing.T) {
	// This test verifies that options are applied correctly
	// It will fail to connect (no PostgreSQL server) but the config should be set
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := NewWithOptions(ctx,
		WithHost("invalid-host"),
		WithPort(5432),
		WithDatabase("testdb"),
		WithUsername("testuser"),
		WithConnectTimeout(100*time.Millisecond),
	)

	// We expect an error because there's no PostgreSQL server at invalid-host
	if err == nil {
		t.Error("Expected connection error")
	}
}

func TestOptions(t *testing.T) {
	config := DefaultConfig()

	WithHost("testhost")(config)
	if config.Host != "testhost" {
		t.Errorf("Expected Host to be testhost, got %s", config.Host)
	}

	WithPort(5433)(config)
	if config.Port != 5433 {
		t.Errorf("Expected Port to be 5433, got %d", config.Port)
	}

	WithDatabase("testdb")(config)
	if config.Database != "testdb" {
		t.Errorf("Expected Database to be testdb, got %s", config.Database)
	}

	WithUsername("testuser")(config)
	if config.Username != "testuser" {
		t.Errorf("Expected Username to be testuser, got %s", config.Username)
	}

	WithPassword("testpass")(config)
	if config.Password != "testpass" {
		t.Errorf("Expected Password to be testpass, got %s", config.Password)
	}

	WithSSLMode("require")(config)
	if config.SSLMode != "require" {
		t.Errorf("Expected SSLMode to be require, got %s", config.SSLMode)
	}

	WithMaxConns(50)(config)
	if config.MaxConns != 50 {
		t.Errorf("Expected MaxConns to be 50, got %d", config.MaxConns)
	}

	WithMinConns(10)(config)
	if config.MinConns != 10 {
		t.Errorf("Expected MinConns to be 10, got %d", config.MinConns)
	}

	WithMaxConnLifetime(2 * time.Hour)(config)
	if config.MaxConnLifetime != 2*time.Hour {
		t.Errorf("Expected MaxConnLifetime to be 2h, got %v", config.MaxConnLifetime)
	}

	WithMaxConnIdleTime(15 * time.Minute)(config)
	if config.MaxConnIdleTime != 15*time.Minute {
		t.Errorf("Expected MaxConnIdleTime to be 15m, got %v", config.MaxConnIdleTime)
	}

	WithHealthCheckPeriod(30 * time.Second)(config)
	if config.HealthCheckPeriod != 30*time.Second {
		t.Errorf("Expected HealthCheckPeriod to be 30s, got %v", config.HealthCheckPeriod)
	}

	WithConnectTimeout(5 * time.Second)(config)
	if config.ConnectTimeout != 5*time.Second {
		t.Errorf("Expected ConnectTimeout to be 5s, got %v", config.ConnectTimeout)
	}

	WithQueryTimeout(10 * time.Second)(config)
	if config.QueryTimeout != 10*time.Second {
		t.Errorf("Expected QueryTimeout to be 10s, got %v", config.QueryTimeout)
	}
}

func TestGetTimeout(t *testing.T) {
	client := &Client{
		config: &Config{
			QueryTimeout: 15 * time.Second,
		},
	}

	timeout := client.GetTimeout()
	if timeout != 15*time.Second {
		t.Errorf("Expected timeout to be 15s, got %v", timeout)
	}

	// Test default timeout
	client.config.QueryTimeout = 0
	timeout = client.GetTimeout()
	if timeout != 30*time.Second {
		t.Errorf("Expected default timeout to be 30s, got %v", timeout)
	}
}

func TestClient_Close_Nil(t *testing.T) {
	client := &Client{
		pool:   nil,
		config: DefaultConfig(),
	}

	// Should not panic when pool is nil
	client.Close()
}

func TestClient_Ping_NotConnected(t *testing.T) {
	client := &Client{
		pool:   nil,
		config: DefaultConfig(),
	}

	err := client.Ping(context.Background())
	if err != ErrClientNotConnected {
		t.Errorf("Expected ErrClientNotConnected, got %v", err)
	}
}

func TestClient_Exec_NotConnected(t *testing.T) {
	client := &Client{
		pool:   nil,
		config: DefaultConfig(),
	}

	_, err := client.Exec(context.Background(), "SELECT 1")
	if err != ErrClientNotConnected {
		t.Errorf("Expected ErrClientNotConnected, got %v", err)
	}
}

func TestClient_Query_NotConnected(t *testing.T) {
	client := &Client{
		pool:   nil,
		config: DefaultConfig(),
	}

	_, err := client.Query(context.Background(), "SELECT 1")
	if err != ErrClientNotConnected {
		t.Errorf("Expected ErrClientNotConnected, got %v", err)
	}
}

func TestClient_BeginTx_NotConnected(t *testing.T) {
	client := &Client{
		pool:   nil,
		config: DefaultConfig(),
	}

	_, err := client.Begin(context.Background())
	if err != ErrClientNotConnected {
		t.Errorf("Expected ErrClientNotConnected, got %v", err)
	}
}

func TestClient_Health_NotConnected(t *testing.T) {
	client := &Client{
		pool:   nil,
		config: DefaultConfig(),
	}

	err := client.Health(context.Background())
	if err != ErrClientNotConnected {
		t.Errorf("Expected ErrClientNotConnected, got %v", err)
	}
}

func TestClient_Stats_NotConnected(t *testing.T) {
	client := &Client{
		pool:   nil,
		config: DefaultConfig(),
	}

	stats := client.Stats()
	if stats != nil {
		t.Error("Expected nil stats when not connected")
	}
}

func TestClient_Pool_NotConnected(t *testing.T) {
	client := &Client{
		pool:   nil,
		config: DefaultConfig(),
	}

	pool := client.Pool()
	if pool != nil {
		t.Error("Expected nil pool when not connected")
	}
}

func TestClient_Config(t *testing.T) {
	config := DefaultConfig()
	config.Database = "testdb"

	client := &Client{
		pool:   nil,
		config: config,
	}

	returnedConfig := client.Config()
	if returnedConfig.Database != "testdb" {
		t.Errorf("Expected Database to be testdb, got %s", returnedConfig.Database)
	}
}

func TestClient_AcquireConn_NotConnected(t *testing.T) {
	client := &Client{
		pool:   nil,
		config: DefaultConfig(),
	}

	_, err := client.AcquireConn(context.Background())
	if err != ErrClientNotConnected {
		t.Errorf("Expected ErrClientNotConnected, got %v", err)
	}
}

func TestClient_CopyFrom_NotConnected(t *testing.T) {
	client := &Client{
		pool:   nil,
		config: DefaultConfig(),
	}

	_, err := client.CopyFrom(context.Background(), nil, nil, nil)
	if err != ErrClientNotConnected {
		t.Errorf("Expected ErrClientNotConnected, got %v", err)
	}
}
