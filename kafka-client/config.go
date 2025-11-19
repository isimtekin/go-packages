package kafkaclient

import (
	"fmt"
	"time"

	"github.com/IBM/sarama"
)

// Config holds the configuration for Kafka client
type Config struct {
	// Brokers is the list of Kafka broker addresses
	Brokers []string

	// Version is the Kafka cluster version
	Version string

	// ClientID is the client identifier sent to Kafka
	ClientID string

	// Workspace is an optional prefix for all topics
	// If set, all topics will be prefixed with "{workspace}."
	// Example: workspace="production" => topic="orders" becomes "production.orders"
	// This is useful for multi-tenancy or environment separation
	Workspace string

	// Consumer configuration
	Consumer ConsumerConfig

	// Producer configuration
	Producer ProducerConfig

	// Security configuration
	Security SecurityConfig

	// Timeout for network operations
	Timeout time.Duration

	// EnableDebug enables debug logging
	EnableDebug bool
}

// ConsumerConfig holds consumer-specific configuration
type ConsumerConfig struct {
	// GroupID is the consumer group ID
	GroupID string

	// Topics to consume from
	Topics []string

	// SessionTimeout for consumer group session
	SessionTimeout time.Duration

	// RebalanceTimeout for consumer group rebalance
	RebalanceTimeout time.Duration

	// OffsetInitial determines where to start consuming
	// Can be sarama.OffsetNewest or sarama.OffsetOldest
	OffsetInitial int64

	// AutoCommit enables auto-commit of offsets
	AutoCommit bool

	// AutoCommitInterval is the interval for auto-committing offsets
	AutoCommitInterval time.Duration

	// MaxProcessingTime is the maximum time to process a message
	MaxProcessingTime time.Duration
}

// ProducerConfig holds producer-specific configuration
type ProducerConfig struct {
	// RequiredAcks determines acknowledgment level
	// 0 = NoResponse, 1 = WaitForLocal, -1 = WaitForAll
	RequiredAcks int16

	// Compression codec (None, Gzip, Snappy, LZ4, Zstd)
	Compression string

	// MaxMessageBytes is the max message size
	MaxMessageBytes int

	// IdempotentWrites ensures exactly-once semantics
	IdempotentWrites bool

	// RetryMax is the maximum number of retry attempts
	RetryMax int

	// RetryBackoff is the backoff duration between retries
	RetryBackoff time.Duration

	// Timeout for produce requests
	Timeout time.Duration

	// Partitioner strategy (hash, random, manual, roundrobin)
	Partitioner string
}

// SecurityConfig holds security-related configuration
type SecurityConfig struct {
	// Enabled determines if security is enabled
	Enabled bool

	// Mechanism is the SASL mechanism (PLAIN, SCRAM-SHA-256, SCRAM-SHA-512)
	Mechanism string

	// Username for SASL authentication
	Username string

	// Password for SASL authentication
	Password string

	// EnableTLS enables TLS encryption
	EnableTLS bool

	// TLSSkipVerify skips TLS certificate verification (insecure)
	TLSSkipVerify bool
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		Brokers:  []string{"localhost:9092"},
		Version:  "3.6.0",
		ClientID: "kafka-client-go",
		Consumer: ConsumerConfig{
			GroupID:            "default-consumer-group",
			Topics:             []string{},
			SessionTimeout:     10 * time.Second,
			RebalanceTimeout:   60 * time.Second,
			OffsetInitial:      sarama.OffsetNewest,
			AutoCommit:         true,
			AutoCommitInterval: 1 * time.Second,
			MaxProcessingTime:  5 * time.Minute,
		},
		Producer: ProducerConfig{
			RequiredAcks:     -1, // WaitForAll
			Compression:      "snappy",
			MaxMessageBytes:  1000000, // 1MB
			IdempotentWrites: true,
			RetryMax:         3,
			RetryBackoff:     100 * time.Millisecond,
			Timeout:          10 * time.Second,
			Partitioner:      "hash",
		},
		Security: SecurityConfig{
			Enabled: false,
		},
		Timeout:     30 * time.Second,
		EnableDebug: false,
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if len(c.Brokers) == 0 {
		return fmt.Errorf("at least one broker must be specified")
	}

	if c.ClientID == "" {
		return fmt.Errorf("client ID cannot be empty")
	}

	if c.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}

	// Validate consumer config if topics are specified
	if len(c.Consumer.Topics) > 0 {
		if c.Consumer.GroupID == "" {
			return fmt.Errorf("consumer group ID cannot be empty when topics are specified")
		}
		if c.Consumer.SessionTimeout <= 0 {
			return fmt.Errorf("consumer session timeout must be positive")
		}
	}

	// Validate producer config
	if c.Producer.MaxMessageBytes <= 0 {
		return fmt.Errorf("max message bytes must be positive")
	}

	if c.Producer.RetryMax < 0 {
		return fmt.Errorf("retry max cannot be negative")
	}

	return nil
}

// ToSaramaConfig converts our config to Sarama config
func (c *Config) ToSaramaConfig() (*sarama.Config, error) {
	config := sarama.NewConfig()

	// Parse Kafka version
	version, err := sarama.ParseKafkaVersion(c.Version)
	if err != nil {
		return nil, fmt.Errorf("invalid Kafka version: %w", err)
	}
	config.Version = version
	config.ClientID = c.ClientID

	// Consumer configuration
	config.Consumer.Group.Session.Timeout = c.Consumer.SessionTimeout
	config.Consumer.Group.Rebalance.Timeout = c.Consumer.RebalanceTimeout
	config.Consumer.Offsets.Initial = c.Consumer.OffsetInitial
	config.Consumer.Offsets.AutoCommit.Enable = c.Consumer.AutoCommit
	config.Consumer.Offsets.AutoCommit.Interval = c.Consumer.AutoCommitInterval
	config.Consumer.MaxProcessingTime = c.Consumer.MaxProcessingTime
	config.Consumer.Return.Errors = true

	// Producer configuration
	config.Producer.RequiredAcks = sarama.RequiredAcks(c.Producer.RequiredAcks)
	config.Producer.MaxMessageBytes = c.Producer.MaxMessageBytes
	config.Producer.Idempotent = c.Producer.IdempotentWrites
	config.Producer.Retry.Max = c.Producer.RetryMax
	config.Producer.Retry.Backoff = c.Producer.RetryBackoff
	config.Producer.Timeout = c.Producer.Timeout
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true

	// Set compression
	switch c.Producer.Compression {
	case "none":
		config.Producer.Compression = sarama.CompressionNone
	case "gzip":
		config.Producer.Compression = sarama.CompressionGZIP
	case "snappy":
		config.Producer.Compression = sarama.CompressionSnappy
	case "lz4":
		config.Producer.Compression = sarama.CompressionLZ4
	case "zstd":
		config.Producer.Compression = sarama.CompressionZSTD
	default:
		config.Producer.Compression = sarama.CompressionSnappy
	}

	// Set partitioner
	switch c.Producer.Partitioner {
	case "hash":
		config.Producer.Partitioner = sarama.NewHashPartitioner
	case "random":
		config.Producer.Partitioner = sarama.NewRandomPartitioner
	case "roundrobin":
		config.Producer.Partitioner = sarama.NewRoundRobinPartitioner
	default:
		config.Producer.Partitioner = sarama.NewHashPartitioner
	}

	// Security configuration
	if c.Security.Enabled {
		config.Net.SASL.Enable = true
		config.Net.SASL.User = c.Security.Username
		config.Net.SASL.Password = c.Security.Password

		switch c.Security.Mechanism {
		case "PLAIN":
			config.Net.SASL.Mechanism = sarama.SASLTypePlaintext
		case "SCRAM-SHA-256":
			config.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA256
		case "SCRAM-SHA-512":
			config.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA512
		default:
			config.Net.SASL.Mechanism = sarama.SASLTypePlaintext
		}

		if c.Security.EnableTLS {
			config.Net.TLS.Enable = true
		}
	}

	// Network timeout
	config.Net.DialTimeout = c.Timeout
	config.Net.ReadTimeout = c.Timeout
	config.Net.WriteTimeout = c.Timeout

	return config, nil
}

// ApplyWorkspacePrefix applies the workspace prefix to a topic name if workspace is configured
func (c *Config) ApplyWorkspacePrefix(topic string) string {
	if c.Workspace == "" || topic == "" {
		return topic
	}
	return fmt.Sprintf("%s.%s", c.Workspace, topic)
}

// ApplyWorkspacePrefixToTopics applies the workspace prefix to multiple topic names
func (c *Config) ApplyWorkspacePrefixToTopics(topics []string) []string {
	if c.Workspace == "" {
		return topics
	}

	prefixedTopics := make([]string, len(topics))
	for i, topic := range topics {
		prefixedTopics[i] = c.ApplyWorkspacePrefix(topic)
	}
	return prefixedTopics
}
