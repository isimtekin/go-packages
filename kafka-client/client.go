package kafkaclient

import (
	"context"
	"fmt"
	"sync"

	"github.com/IBM/sarama"
)

// Client represents a unified Kafka client with producer and consumer capabilities
type Client struct {
	config   *Config
	producer *Producer
	consumer *Consumer
	admin    sarama.ClusterAdmin

	mu     sync.RWMutex
	closed bool
}

// New creates a new Kafka client with the given configuration
func New(config *Config) (*Client, error) {
	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidConfig, err)
	}

	client := &Client{
		config: config,
		closed: false,
	}

	// Initialize admin client for metadata operations
	if err := client.initAdmin(); err != nil {
		return nil, fmt.Errorf("failed to initialize admin client: %w", err)
	}

	// Initialize producer (always available)
	producer, err := NewProducer(config)
	if err != nil {
		client.admin.Close()
		return nil, fmt.Errorf("failed to create producer: %w", err)
	}
	client.producer = producer

	return client, nil
}

// NewWithOptions creates a new client with functional options
func NewWithOptions(opts ...Option) (*Client, error) {
	config := DefaultConfig()

	for _, opt := range opts {
		opt(config)
	}

	return New(config)
}

// NewWithConsumer creates a new client with both producer and consumer
func NewWithConsumer(config *Config, handler MessageHandler) (*Client, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidConfig, err)
	}

	client := &Client{
		config: config,
		closed: false,
	}

	// Initialize admin client
	if err := client.initAdmin(); err != nil {
		return nil, fmt.Errorf("failed to initialize admin client: %w", err)
	}

	// Initialize producer
	producer, err := NewProducer(config)
	if err != nil {
		client.admin.Close()
		return nil, fmt.Errorf("failed to create producer: %w", err)
	}
	client.producer = producer

	// Initialize consumer
	consumer, err := NewConsumer(config, handler)
	if err != nil {
		client.producer.Close()
		client.admin.Close()
		return nil, fmt.Errorf("failed to create consumer: %w", err)
	}
	client.consumer = consumer

	return client, nil
}

// initAdmin initializes the admin client
func (c *Client) initAdmin() error {
	saramaConfig, err := c.config.ToSaramaConfig()
	if err != nil {
		return fmt.Errorf("failed to create sarama config: %w", err)
	}

	admin, err := sarama.NewClusterAdmin(c.config.Brokers, saramaConfig)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrConnectionFailed, err)
	}

	c.admin = admin
	return nil
}

// SendMessage sends a single message to Kafka
func (c *Client) SendMessage(ctx context.Context, msg *Message) (partition int32, offset int64, err error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return 0, 0, ErrClientClosed
	}

	if c.producer == nil {
		return 0, 0, ErrNoProducer
	}

	return c.producer.SendMessage(ctx, msg)
}

// SendMessages sends multiple messages to Kafka
func (c *Client) SendMessages(ctx context.Context, messages []*Message) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return ErrClientClosed
	}

	if c.producer == nil {
		return ErrNoProducer
	}

	return c.producer.SendMessages(ctx, messages)
}

// ListTopics returns a list of all topics
func (c *Client) ListTopics(ctx context.Context) ([]string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return nil, ErrClientClosed
	}

	topics, err := c.admin.ListTopics()
	if err != nil {
		return nil, fmt.Errorf("failed to list topics: %w", err)
	}

	topicNames := make([]string, 0, len(topics))
	for topic := range topics {
		topicNames = append(topicNames, topic)
	}

	return topicNames, nil
}

// CreateTopic creates a new topic
func (c *Client) CreateTopic(ctx context.Context, topic string, numPartitions int32, replicationFactor int16) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return ErrClientClosed
	}

	topicDetail := &sarama.TopicDetail{
		NumPartitions:     numPartitions,
		ReplicationFactor: replicationFactor,
	}

	err := c.admin.CreateTopic(topic, topicDetail, false)
	if err != nil {
		return fmt.Errorf("failed to create topic: %w", err)
	}

	return nil
}

// DeleteTopic deletes a topic
func (c *Client) DeleteTopic(ctx context.Context, topic string) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return ErrClientClosed
	}

	err := c.admin.DeleteTopic(topic)
	if err != nil {
		return fmt.Errorf("failed to delete topic: %w", err)
	}

	return nil
}

// GetTopicMetadata returns metadata for a specific topic
func (c *Client) GetTopicMetadata(ctx context.Context, topic string) (*sarama.TopicMetadata, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return nil, ErrClientClosed
	}

	topics, err := c.admin.DescribeTopics([]string{topic})
	if err != nil {
		return nil, fmt.Errorf("failed to describe topic: %w", err)
	}

	if len(topics) == 0 {
		return nil, fmt.Errorf("topic %s not found", topic)
	}

	return topics[0], nil
}

// ConsumerErrors returns the consumer error channel
func (c *Client) ConsumerErrors() <-chan error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.consumer == nil {
		ch := make(chan error)
		close(ch)
		return ch
	}

	return c.consumer.Errors()
}

// Close closes the client and releases all resources
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return ErrAlreadyClosed
	}

	var errs []error

	// Close consumer first if it exists
	if c.consumer != nil {
		if err := c.consumer.Close(); err != nil {
			errs = append(errs, fmt.Errorf("consumer close error: %w", err))
		}
	}

	// Close producer
	if c.producer != nil {
		if err := c.producer.Close(); err != nil {
			errs = append(errs, fmt.Errorf("producer close error: %w", err))
		}
	}

	// Close admin client
	if c.admin != nil {
		if err := c.admin.Close(); err != nil {
			errs = append(errs, fmt.Errorf("admin close error: %w", err))
		}
	}

	c.closed = true

	if len(errs) > 0 {
		return fmt.Errorf("errors during close: %v", errs)
	}

	return nil
}

// Ping checks if the Kafka cluster is reachable
func (c *Client) Ping(ctx context.Context) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return ErrClientClosed
	}

	// Try to list topics as a ping operation
	_, err := c.admin.ListTopics()
	if err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}

	return nil
}

// IsClosed returns true if the client is closed
func (c *Client) IsClosed() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.closed
}
