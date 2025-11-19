package kafkaclient

import (
	"testing"
	"time"

	"github.com/IBM/sarama"
)

func TestOptions_AllOptions(t *testing.T) {
	config := DefaultConfig()

	// Apply all options
	WithBrokers([]string{"broker1:9092", "broker2:9092"})(config)
	WithVersion("3.5.0")(config)
	WithClientID("test-client-id")(config)
	WithWorkspace("production")(config)
	WithConsumerGroup("test-consumer-group")(config)
	WithTopics([]string{"topic1", "topic2"})(config)
	WithOffsetInitial(sarama.OffsetOldest)(config)
	WithAutoCommit(false)(config)
	WithCompression("gzip")(config)
	WithRequiredAcks(1)(config)
	WithIdempotentWrites(false)(config)
	WithPartitioner("random")(config)
	WithMaxMessageBytes(5 * 1024 * 1024)(config)
	WithRetryMax(5)(config)
	WithTimeout(60 * time.Second)(config)
	WithSASL("PLAIN", "user", "pass")(config)
	WithTLS(true)(config)
	WithDebug(true)(config)

	// Verify all options were applied
	if len(config.Brokers) != 2 {
		t.Errorf("Brokers count = %d, want 2", len(config.Brokers))
	}
	if config.Version != "3.5.0" {
		t.Errorf("Version = %s, want 3.5.0", config.Version)
	}
	if config.ClientID != "test-client-id" {
		t.Errorf("ClientID = %s, want test-client-id", config.ClientID)
	}
	if config.Workspace != "production" {
		t.Errorf("Workspace = %s, want production", config.Workspace)
	}
	if config.Consumer.GroupID != "test-consumer-group" {
		t.Errorf("GroupID = %s, want test-consumer-group", config.Consumer.GroupID)
	}
	if len(config.Consumer.Topics) != 2 {
		t.Errorf("Topics count = %d, want 2", len(config.Consumer.Topics))
	}
	if config.Consumer.OffsetInitial != sarama.OffsetOldest {
		t.Errorf("OffsetInitial = %d, want %d", config.Consumer.OffsetInitial, sarama.OffsetOldest)
	}
	if config.Consumer.AutoCommit != false {
		t.Error("AutoCommit should be false")
	}
	if config.Producer.Compression != "gzip" {
		t.Errorf("Compression = %s, want gzip", config.Producer.Compression)
	}
	if config.Producer.RequiredAcks != 1 {
		t.Errorf("RequiredAcks = %d, want 1", config.Producer.RequiredAcks)
	}
	if config.Producer.IdempotentWrites != false {
		t.Error("IdempotentWrites should be false")
	}
	if config.Producer.Partitioner != "random" {
		t.Errorf("Partitioner = %s, want random", config.Producer.Partitioner)
	}
	if config.Producer.MaxMessageBytes != 5*1024*1024 {
		t.Errorf("MaxMessageBytes = %d, want %d", config.Producer.MaxMessageBytes, 5*1024*1024)
	}
	if config.Producer.RetryMax != 5 {
		t.Errorf("RetryMax = %d, want 5", config.Producer.RetryMax)
	}
	if config.Timeout != 60*time.Second {
		t.Errorf("Timeout = %v, want 60s", config.Timeout)
	}
	if !config.Security.Enabled {
		t.Error("Security should be enabled")
	}
	if config.Security.Mechanism != "PLAIN" {
		t.Errorf("SASL Mechanism = %s, want PLAIN", config.Security.Mechanism)
	}
	if !config.Security.EnableTLS {
		t.Error("TLS should be enabled")
	}
	if !config.EnableDebug {
		t.Error("Debug should be enabled")
	}
}

func TestOptions_ConsumerOptions(t *testing.T) {
	tests := []struct {
		name     string
		option   Option
		validate func(*Config) error
	}{
		{
			name:   "WithConsumerGroup",
			option: WithConsumerGroup("my-group"),
			validate: func(c *Config) error {
				if c.Consumer.GroupID != "my-group" {
					return ErrInvalidConfig
				}
				return nil
			},
		},
		{
			name:   "WithTopics single",
			option: WithTopics([]string{"orders"}),
			validate: func(c *Config) error {
				if len(c.Consumer.Topics) != 1 || c.Consumer.Topics[0] != "orders" {
					return ErrInvalidConfig
				}
				return nil
			},
		},
		{
			name:   "WithTopics multiple",
			option: WithTopics([]string{"orders", "users", "events"}),
			validate: func(c *Config) error {
				if len(c.Consumer.Topics) != 3 {
					return ErrInvalidConfig
				}
				return nil
			},
		},
		{
			name:   "WithAutoCommit enabled",
			option: WithAutoCommit(true),
			validate: func(c *Config) error {
				if !c.Consumer.AutoCommit {
					return ErrInvalidConfig
				}
				return nil
			},
		},
		{
			name:   "WithAutoCommit disabled",
			option: WithAutoCommit(false),
			validate: func(c *Config) error {
				if c.Consumer.AutoCommit {
					return ErrInvalidConfig
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultConfig()
			tt.option(config)
			if err := tt.validate(config); err != nil {
				t.Errorf("Option validation failed: %v", err)
			}
		})
	}
}

func TestOptions_ProducerOptions(t *testing.T) {
	tests := []struct {
		name     string
		option   Option
		validate func(*Config) error
	}{
		{
			name:   "WithIdempotentWrites true",
			option: WithIdempotentWrites(true),
			validate: func(c *Config) error {
				if !c.Producer.IdempotentWrites {
					return ErrInvalidConfig
				}
				return nil
			},
		},
		{
			name:   "WithIdempotentWrites false",
			option: WithIdempotentWrites(false),
			validate: func(c *Config) error {
				if c.Producer.IdempotentWrites {
					return ErrInvalidConfig
				}
				return nil
			},
		},
		{
			name:   "WithPartitioner hash",
			option: WithPartitioner("hash"),
			validate: func(c *Config) error {
				if c.Producer.Partitioner != "hash" {
					return ErrInvalidConfig
				}
				return nil
			},
		},
		{
			name:   "WithPartitioner roundrobin",
			option: WithPartitioner("roundrobin"),
			validate: func(c *Config) error {
				if c.Producer.Partitioner != "roundrobin" {
					return ErrInvalidConfig
				}
				return nil
			},
		},
		{
			name:   "WithMaxMessageBytes 1MB",
			option: WithMaxMessageBytes(1024 * 1024),
			validate: func(c *Config) error {
				if c.Producer.MaxMessageBytes != 1024*1024 {
					return ErrInvalidConfig
				}
				return nil
			},
		},
		{
			name:   "WithMaxMessageBytes 10MB",
			option: WithMaxMessageBytes(10 * 1024 * 1024),
			validate: func(c *Config) error {
				if c.Producer.MaxMessageBytes != 10*1024*1024 {
					return ErrInvalidConfig
				}
				return nil
			},
		},
		{
			name:   "WithRetryMax 0",
			option: WithRetryMax(0),
			validate: func(c *Config) error {
				if c.Producer.RetryMax != 0 {
					return ErrInvalidConfig
				}
				return nil
			},
		},
		{
			name:   "WithRetryMax 10",
			option: WithRetryMax(10),
			validate: func(c *Config) error {
				if c.Producer.RetryMax != 10 {
					return ErrInvalidConfig
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultConfig()
			tt.option(config)
			if err := tt.validate(config); err != nil {
				t.Errorf("Option validation failed: %v", err)
			}
		})
	}
}

func TestOptions_SecurityOptions(t *testing.T) {
	tests := []struct {
		name     string
		option   Option
		validate func(*Config) error
	}{
		{
			name:   "WithSASL PLAIN",
			option: WithSASL("PLAIN", "testuser", "testpass"),
			validate: func(c *Config) error {
				if !c.Security.Enabled {
					return ErrInvalidConfig
				}
				if c.Security.Mechanism != "PLAIN" {
					return ErrInvalidConfig
				}
				if c.Security.Username != "testuser" {
					return ErrInvalidConfig
				}
				if c.Security.Password != "testpass" {
					return ErrInvalidConfig
				}
				return nil
			},
		},
		{
			name:   "WithSASL SCRAM-SHA-256",
			option: WithSASL("SCRAM-SHA-256", "user", "pass"),
			validate: func(c *Config) error {
				if c.Security.Mechanism != "SCRAM-SHA-256" {
					return ErrInvalidConfig
				}
				return nil
			},
		},
		{
			name:   "WithTLS enabled",
			option: WithTLS(true),
			validate: func(c *Config) error {
				if !c.Security.EnableTLS {
					return ErrInvalidConfig
				}
				if !c.Security.Enabled {
					return ErrInvalidConfig
				}
				return nil
			},
		},
		{
			name:   "WithTLS disabled",
			option: WithTLS(false),
			validate: func(c *Config) error {
				if c.Security.EnableTLS {
					return ErrInvalidConfig
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultConfig()
			tt.option(config)
			if err := tt.validate(config); err != nil {
				t.Errorf("Option validation failed: %v", err)
			}
		})
	}
}

func TestOptions_DebugOption(t *testing.T) {
	tests := []struct {
		name  string
		debug bool
	}{
		{"debug enabled", true},
		{"debug disabled", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultConfig()
			WithDebug(tt.debug)(config)

			if config.EnableDebug != tt.debug {
				t.Errorf("EnableDebug = %v, want %v", config.EnableDebug, tt.debug)
			}
		})
	}
}

func TestOptions_ChainingMultipleOptions(t *testing.T) {
	// Test applying multiple options in sequence
	config := DefaultConfig()

	// Chain 1: Basic setup
	WithBrokers([]string{"localhost:9092"})(config)
	WithClientID("chain-test")(config)
	WithWorkspace("test-env")(config)

	if len(config.Brokers) != 1 {
		t.Error("First chain failed")
	}

	// Chain 2: Producer settings
	WithCompression("lz4")(config)
	WithRequiredAcks(-1)(config)
	WithIdempotentWrites(true)(config)

	if config.Producer.Compression != "lz4" {
		t.Error("Second chain failed")
	}

	// Chain 3: Consumer settings
	WithConsumerGroup("chain-group")(config)
	WithTopics([]string{"topic1", "topic2", "topic3"})(config)
	WithAutoCommit(true)(config)

	if len(config.Consumer.Topics) != 3 {
		t.Error("Third chain failed")
	}

	// Verify all settings are still applied
	if config.ClientID != "chain-test" {
		t.Error("Earlier setting was overwritten")
	}
	if config.Producer.RequiredAcks != -1 {
		t.Error("Earlier setting was overwritten")
	}
}

func TestOptions_OverwritingOptions(t *testing.T) {
	config := DefaultConfig()

	// First set
	WithClientID("first-id")(config)
	if config.ClientID != "first-id" {
		t.Error("First set failed")
	}

	// Overwrite
	WithClientID("second-id")(config)
	if config.ClientID != "second-id" {
		t.Error("Overwrite failed")
	}

	// Test with compression
	WithCompression("snappy")(config)
	if config.Producer.Compression != "snappy" {
		t.Error("First compression failed")
	}

	WithCompression("gzip")(config)
	if config.Producer.Compression != "gzip" {
		t.Error("Compression overwrite failed")
	}
}

func TestOptions_EmptyValues(t *testing.T) {
	tests := []struct {
		name   string
		option Option
	}{
		{"empty workspace", WithWorkspace("")},
		{"empty consumer group", WithConsumerGroup("")},
		{"empty topics", WithTopics([]string{})},
		{"empty client ID", WithClientID("")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultConfig()
			tt.option(config)
			// Should not panic, validation happens separately
		})
	}
}

func TestOptions_BoundaryValues(t *testing.T) {
	tests := []struct {
		name   string
		option Option
	}{
		{"zero retry max", WithRetryMax(0)},
		{"large retry max", WithRetryMax(100)},
		{"very small timeout", WithTimeout(1 * time.Millisecond)},
		{"very large timeout", WithTimeout(1 * time.Hour)},
		{"small message size", WithMaxMessageBytes(1024)},
		{"large message size", WithMaxMessageBytes(100 * 1024 * 1024)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultConfig()
			tt.option(config)
			// Should not panic
		})
	}
}
