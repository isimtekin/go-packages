package kafkaclient

import (
	"context"
	"fmt"
	"sync"

	"github.com/IBM/sarama"
)

// Consumer wraps a Sarama consumer group
type Consumer struct {
	consumerGroup sarama.ConsumerGroup
	config        *Config
	handler       *consumerGroupHandler
	mu            sync.RWMutex
	closed        bool
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
}

// ConsumedMessage represents a consumed Kafka message
type ConsumedMessage struct {
	Topic     string
	Partition int32
	Offset    int64
	Key       []byte
	Value     []byte
	Headers   map[string]string
	Timestamp int64
}

// MessageHandler is a function type for handling consumed messages
type MessageHandler func(ctx context.Context, msg *ConsumedMessage) error

// consumerGroupHandler implements sarama.ConsumerGroupHandler
type consumerGroupHandler struct {
	handler MessageHandler
	ready   chan bool
}

func (h *consumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	close(h.ready)
	return nil
}

func (h *consumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message := <-claim.Messages():
			if message == nil {
				return nil
			}

			// Convert Sarama message to our format
			headers := make(map[string]string)
			for _, header := range message.Headers {
				headers[string(header.Key)] = string(header.Value)
			}

			msg := &ConsumedMessage{
				Topic:     message.Topic,
				Partition: message.Partition,
				Offset:    message.Offset,
				Key:       message.Key,
				Value:     message.Value,
				Headers:   headers,
				Timestamp: message.Timestamp.Unix(),
			}

			// Call user's handler
			if err := h.handler(session.Context(), msg); err != nil {
				// Log error but continue processing
				// In production, you might want to implement error handling strategy
				continue
			}

			// Mark message as processed
			session.MarkMessage(message, "")

		case <-session.Context().Done():
			return nil
		}
	}
}

// NewConsumer creates a new Kafka consumer
func NewConsumer(config *Config, handler MessageHandler) (*Consumer, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	if len(config.Consumer.Topics) == 0 {
		return nil, fmt.Errorf("no topics specified in consumer config")
	}

	if handler == nil {
		return nil, fmt.Errorf("message handler cannot be nil")
	}

	saramaConfig, err := config.ToSaramaConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to create sarama config: %w", err)
	}

	consumerGroup, err := sarama.NewConsumerGroup(config.Brokers, config.Consumer.GroupID, saramaConfig)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrConnectionFailed, err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	consumer := &Consumer{
		consumerGroup: consumerGroup,
		config:        config,
		handler: &consumerGroupHandler{
			handler: handler,
			ready:   make(chan bool),
		},
		closed: false,
		ctx:    ctx,
		cancel: cancel,
	}

	// Start consuming in background
	consumer.wg.Add(1)
	go consumer.consume()

	// Wait for consumer to be ready
	<-consumer.handler.ready

	return consumer, nil
}

// consume runs the consumer group loop
func (c *Consumer) consume() {
	defer c.wg.Done()

	for {
		// Check if context is cancelled
		select {
		case <-c.ctx.Done():
			return
		default:
		}

		// Consume messages
		if err := c.consumerGroup.Consume(c.ctx, c.config.Consumer.Topics, c.handler); err != nil {
			// Log error or handle it appropriately
			// For now, we'll just continue
			select {
			case <-c.ctx.Done():
				return
			default:
				continue
			}
		}

		// Check if context was cancelled
		if c.ctx.Err() != nil {
			return
		}

		// Reset ready channel for next session
		c.handler.ready = make(chan bool)
	}
}

// Close closes the consumer
func (c *Consumer) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return ErrAlreadyClosed
	}

	c.closed = true
	c.cancel()
	c.wg.Wait()

	return c.consumerGroup.Close()
}

// Errors returns a channel of consumer errors
func (c *Consumer) Errors() <-chan error {
	return c.consumerGroup.Errors()
}
