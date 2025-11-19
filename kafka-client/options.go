package kafkaclient

import "time"

// Option is a functional option for configuring the client
type Option func(*Config)

// WithBrokers sets the Kafka broker addresses
func WithBrokers(brokers []string) Option {
	return func(c *Config) {
		c.Brokers = brokers
	}
}

// WithVersion sets the Kafka version
func WithVersion(version string) Option {
	return func(c *Config) {
		c.Version = version
	}
}

// WithClientID sets the client ID
func WithClientID(clientID string) Option {
	return func(c *Config) {
		c.ClientID = clientID
	}
}

// WithWorkspace sets the workspace prefix for all topics
// When set, all topics will be prefixed with "{workspace}."
// Example: workspace="production" => topic="orders" becomes "production.orders"
func WithWorkspace(workspace string) Option {
	return func(c *Config) {
		c.Workspace = workspace
	}
}

// WithConsumerGroup sets the consumer group ID
func WithConsumerGroup(groupID string) Option {
	return func(c *Config) {
		c.Consumer.GroupID = groupID
	}
}

// WithTopics sets the topics to consume from
func WithTopics(topics []string) Option {
	return func(c *Config) {
		c.Consumer.Topics = topics
	}
}

// WithOffsetInitial sets the initial offset for consuming
// Use sarama.OffsetNewest or sarama.OffsetOldest
func WithOffsetInitial(offset int64) Option {
	return func(c *Config) {
		c.Consumer.OffsetInitial = offset
	}
}

// WithAutoCommit enables or disables auto-commit of offsets
func WithAutoCommit(enable bool) Option {
	return func(c *Config) {
		c.Consumer.AutoCommit = enable
	}
}

// WithCompression sets the compression codec for producer
// Options: "none", "gzip", "snappy", "lz4", "zstd"
func WithCompression(compression string) Option {
	return func(c *Config) {
		c.Producer.Compression = compression
	}
}

// WithRequiredAcks sets the required acknowledgments for producer
// 0 = NoResponse, 1 = WaitForLocal, -1 = WaitForAll
func WithRequiredAcks(acks int16) Option {
	return func(c *Config) {
		c.Producer.RequiredAcks = acks
	}
}

// WithIdempotentWrites enables or disables idempotent writes
func WithIdempotentWrites(enable bool) Option {
	return func(c *Config) {
		c.Producer.IdempotentWrites = enable
	}
}

// WithPartitioner sets the partitioner strategy
// Options: "hash", "random", "roundrobin"
func WithPartitioner(partitioner string) Option {
	return func(c *Config) {
		c.Producer.Partitioner = partitioner
	}
}

// WithMaxMessageBytes sets the maximum message size
func WithMaxMessageBytes(bytes int) Option {
	return func(c *Config) {
		c.Producer.MaxMessageBytes = bytes
	}
}

// WithRetryMax sets the maximum number of retry attempts
func WithRetryMax(max int) Option {
	return func(c *Config) {
		c.Producer.RetryMax = max
	}
}

// WithTimeout sets the network timeout
func WithTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		c.Timeout = timeout
	}
}

// WithSASL enables SASL authentication
func WithSASL(mechanism, username, password string) Option {
	return func(c *Config) {
		c.Security.Enabled = true
		c.Security.Mechanism = mechanism
		c.Security.Username = username
		c.Security.Password = password
	}
}

// WithTLS enables TLS encryption
func WithTLS(enable bool) Option {
	return func(c *Config) {
		c.Security.EnableTLS = enable
		if enable {
			c.Security.Enabled = true
		}
	}
}

// WithDebug enables or disables debug mode
func WithDebug(enable bool) Option {
	return func(c *Config) {
		c.EnableDebug = enable
	}
}
