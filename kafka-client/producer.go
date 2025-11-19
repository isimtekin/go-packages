package kafkaclient

import (
	"context"
	"fmt"
	"sync"

	"github.com/IBM/sarama"
)

// Producer wraps a Sarama sync producer
type Producer struct {
	producer sarama.SyncProducer
	config   *Config
	mu       sync.RWMutex
	closed   bool
}

// Message represents a Kafka message
type Message struct {
	Topic     string
	Key       []byte
	Value     []byte
	Headers   map[string]string
	Partition int32 // -1 for automatic partitioning
}

// NewProducer creates a new Kafka producer
func NewProducer(config *Config) (*Producer, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	saramaConfig, err := config.ToSaramaConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to create sarama config: %w", err)
	}

	producer, err := sarama.NewSyncProducer(config.Brokers, saramaConfig)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrConnectionFailed, err)
	}

	return &Producer{
		producer: producer,
		config:   config,
		closed:   false,
	}, nil
}

// SendMessage sends a single message to Kafka
func (p *Producer) SendMessage(ctx context.Context, msg *Message) (partition int32, offset int64, err error) {
	p.mu.RLock()
	if p.closed {
		p.mu.RUnlock()
		return 0, 0, ErrProducerClosed
	}
	p.mu.RUnlock()

	if msg.Topic == "" {
		return 0, 0, ErrInvalidTopic
	}

	// Apply workspace prefix to topic if configured
	topic := p.config.ApplyWorkspacePrefix(msg.Topic)

	// Build Sarama producer message
	producerMsg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(msg.Value),
	}

	if len(msg.Key) > 0 {
		producerMsg.Key = sarama.ByteEncoder(msg.Key)
	}

	if msg.Partition >= 0 {
		producerMsg.Partition = msg.Partition
	}

	// Add headers
	if len(msg.Headers) > 0 {
		headers := make([]sarama.RecordHeader, 0, len(msg.Headers))
		for k, v := range msg.Headers {
			headers = append(headers, sarama.RecordHeader{
				Key:   []byte(k),
				Value: []byte(v),
			})
		}
		producerMsg.Headers = headers
	}

	// Send with context support
	done := make(chan struct{})
	var sendErr error

	go func() {
		partition, offset, sendErr = p.producer.SendMessage(producerMsg)
		close(done)
	}()

	select {
	case <-ctx.Done():
		return 0, 0, fmt.Errorf("%w: %v", ErrTimeout, ctx.Err())
	case <-done:
		if sendErr != nil {
			return 0, 0, fmt.Errorf("failed to send message: %w", sendErr)
		}
		return partition, offset, nil
	}
}

// SendMessages sends multiple messages to Kafka
func (p *Producer) SendMessages(ctx context.Context, messages []*Message) error {
	p.mu.RLock()
	if p.closed {
		p.mu.RUnlock()
		return ErrProducerClosed
	}
	p.mu.RUnlock()

	if len(messages) == 0 {
		return nil
	}

	// Convert to Sarama messages
	producerMsgs := make([]*sarama.ProducerMessage, 0, len(messages))
	for _, msg := range messages {
		if msg.Topic == "" {
			return ErrInvalidTopic
		}

		// Apply workspace prefix to topic if configured
		topic := p.config.ApplyWorkspacePrefix(msg.Topic)

		producerMsg := &sarama.ProducerMessage{
			Topic: topic,
			Value: sarama.ByteEncoder(msg.Value),
		}

		if len(msg.Key) > 0 {
			producerMsg.Key = sarama.ByteEncoder(msg.Key)
		}

		if msg.Partition >= 0 {
			producerMsg.Partition = msg.Partition
		}

		if len(msg.Headers) > 0 {
			headers := make([]sarama.RecordHeader, 0, len(msg.Headers))
			for k, v := range msg.Headers {
				headers = append(headers, sarama.RecordHeader{
					Key:   []byte(k),
					Value: []byte(v),
				})
			}
			producerMsg.Headers = headers
		}

		producerMsgs = append(producerMsgs, producerMsg)
	}

	// Send with context support
	done := make(chan error)

	go func() {
		err := p.producer.SendMessages(producerMsgs)
		done <- err
	}()

	select {
	case <-ctx.Done():
		return fmt.Errorf("%w: %v", ErrTimeout, ctx.Err())
	case err := <-done:
		if err != nil {
			return fmt.Errorf("failed to send messages: %w", err)
		}
		return nil
	}
}

// Close closes the producer
func (p *Producer) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return ErrAlreadyClosed
	}

	p.closed = true
	return p.producer.Close()
}
