package natsclient

import (
	"context"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
)

// StreamConfig represents configuration for creating a stream
type StreamConfig struct {
	Name         string          // Stream name
	Description  string          // Description of the stream
	Subjects     []string        // Subjects the stream listens on
	Retention    RetentionPolicy // Retention policy
	MaxConsumers int             // Maximum number of consumers
	MaxMsgs      int64           // Maximum number of messages
	MaxBytes     int64           // Maximum bytes in the stream
	MaxAge       time.Duration   // Maximum age of messages
	MaxMsgSize   int32           // Maximum message size
	Storage      StorageType     // Storage type (file or memory)
	Replicas     int             // Number of replicas
	NoAck        bool            // Disable acknowledgements
	Duplicates   time.Duration   // Duplicate detection window
}

// RetentionPolicy defines how messages are retained
type RetentionPolicy int

const (
	// LimitsPolicy retains messages until limits are reached
	LimitsPolicy RetentionPolicy = iota
	// InterestPolicy retains messages while there are consumers
	InterestPolicy
	// WorkQueuePolicy retains messages until acknowledged
	WorkQueuePolicy
)

// StorageType defines storage backend
type StorageType int

const (
	// FileStorage stores messages on disk
	FileStorage StorageType = iota
	// MemoryStorage stores messages in memory
	MemoryStorage
)

// ConsumerConfig represents configuration for creating a consumer
type ConsumerConfig struct {
	Durable         string        // Durable name (empty for ephemeral)
	Description     string        // Description
	DeliverSubject  string        // Delivery subject for push consumers
	DeliverGroup    string        // Queue group for push consumers
	DeliverPolicy   DeliverPolicy // Where to start delivering
	OptStartSeq     uint64        // Start sequence (for DeliverByStartSequence)
	OptStartTime    *time.Time    // Start time (for DeliverByStartTime)
	AckPolicy       AckPolicy     // Acknowledgement policy
	AckWait         time.Duration // How long to wait for ack
	MaxDeliver      int           // Maximum delivery attempts
	FilterSubject   string        // Filter by subject
	ReplayPolicy    ReplayPolicy  // Replay policy
	MaxAckPending   int           // Maximum pending acks
	MaxWaiting      int           // Maximum waiting pull requests
	InactiveThreshold time.Duration // Inactive threshold for ephemeral consumers
}

// DeliverPolicy defines where to start delivering messages
type DeliverPolicy int

const (
	// DeliverAll starts from the beginning
	DeliverAll DeliverPolicy = iota
	// DeliverLast starts from the last message
	DeliverLast
	// DeliverNew starts from new messages only
	DeliverNew
	// DeliverByStartSequence starts from a specific sequence
	DeliverByStartSequence
	// DeliverByStartTime starts from a specific time
	DeliverByStartTime
	// DeliverLastPerSubject starts from last message per subject
	DeliverLastPerSubject
)

// AckPolicy defines acknowledgement requirements
type AckPolicy int

const (
	// AckNone requires no acknowledgement
	AckNone AckPolicy = iota
	// AckAll acknowledges all messages up to sequence
	AckAll
	// AckExplicit requires explicit acknowledgement
	AckExplicit
)

// ReplayPolicy defines how messages are replayed
type ReplayPolicy int

const (
	// ReplayInstant replays messages as fast as possible
	ReplayInstant ReplayPolicy = iota
	// ReplayOriginal replays at original speed
	ReplayOriginal
)

// toNatsStreamConfig converts StreamConfig to nats.StreamConfig
func (sc *StreamConfig) toNatsStreamConfig() *nats.StreamConfig {
	cfg := &nats.StreamConfig{
		Name:         sc.Name,
		Description:  sc.Description,
		Subjects:     sc.Subjects,
		MaxConsumers: sc.MaxConsumers,
		MaxMsgs:      sc.MaxMsgs,
		MaxBytes:     sc.MaxBytes,
		MaxAge:       sc.MaxAge,
		MaxMsgSize:   sc.MaxMsgSize,
		Replicas:     sc.Replicas,
		NoAck:        sc.NoAck,
		Duplicates:   sc.Duplicates,
	}

	// Set retention policy
	switch sc.Retention {
	case LimitsPolicy:
		cfg.Retention = nats.LimitsPolicy
	case InterestPolicy:
		cfg.Retention = nats.InterestPolicy
	case WorkQueuePolicy:
		cfg.Retention = nats.WorkQueuePolicy
	}

	// Set storage type
	switch sc.Storage {
	case FileStorage:
		cfg.Storage = nats.FileStorage
	case MemoryStorage:
		cfg.Storage = nats.MemoryStorage
	}

	return cfg
}

// toNatsConsumerConfig converts ConsumerConfig to nats.ConsumerConfig
func (cc *ConsumerConfig) toNatsConsumerConfig() *nats.ConsumerConfig {
	cfg := &nats.ConsumerConfig{
		Durable:           cc.Durable,
		Description:       cc.Description,
		DeliverSubject:    cc.DeliverSubject,
		DeliverGroup:      cc.DeliverGroup,
		OptStartSeq:       cc.OptStartSeq,
		OptStartTime:      cc.OptStartTime,
		AckWait:           cc.AckWait,
		MaxDeliver:        cc.MaxDeliver,
		FilterSubject:     cc.FilterSubject,
		MaxAckPending:     cc.MaxAckPending,
		MaxWaiting:        cc.MaxWaiting,
		InactiveThreshold: cc.InactiveThreshold,
	}

	// Set deliver policy
	switch cc.DeliverPolicy {
	case DeliverAll:
		cfg.DeliverPolicy = nats.DeliverAllPolicy
	case DeliverLast:
		cfg.DeliverPolicy = nats.DeliverLastPolicy
	case DeliverNew:
		cfg.DeliverPolicy = nats.DeliverNewPolicy
	case DeliverByStartSequence:
		cfg.DeliverPolicy = nats.DeliverByStartSequencePolicy
	case DeliverByStartTime:
		cfg.DeliverPolicy = nats.DeliverByStartTimePolicy
	case DeliverLastPerSubject:
		cfg.DeliverPolicy = nats.DeliverLastPerSubjectPolicy
	}

	// Set ack policy
	switch cc.AckPolicy {
	case AckNone:
		cfg.AckPolicy = nats.AckNonePolicy
	case AckAll:
		cfg.AckPolicy = nats.AckAllPolicy
	case AckExplicit:
		cfg.AckPolicy = nats.AckExplicitPolicy
	}

	// Set replay policy
	switch cc.ReplayPolicy {
	case ReplayInstant:
		cfg.ReplayPolicy = nats.ReplayInstantPolicy
	case ReplayOriginal:
		cfg.ReplayPolicy = nats.ReplayOriginalPolicy
	}

	return cfg
}

// ============================================================================
// Stream Management
// ============================================================================

// CreateStream creates a new stream with the given configuration
func (c *Client) CreateStream(ctx context.Context, config *StreamConfig) (*nats.StreamInfo, error) {
	js, err := c.getJetStream()
	if err != nil {
		return nil, err
	}

	info, err := js.AddStream(config.toNatsStreamConfig(), nats.Context(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to create stream: %w", err)
	}

	return info, nil
}

// UpdateStream updates an existing stream
func (c *Client) UpdateStream(ctx context.Context, config *StreamConfig) (*nats.StreamInfo, error) {
	js, err := c.getJetStream()
	if err != nil {
		return nil, err
	}

	info, err := js.UpdateStream(config.toNatsStreamConfig(), nats.Context(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to update stream: %w", err)
	}

	return info, nil
}

// DeleteStream deletes a stream by name
func (c *Client) DeleteStream(ctx context.Context, name string) error {
	js, err := c.getJetStream()
	if err != nil {
		return err
	}

	if err := js.DeleteStream(name, nats.Context(ctx)); err != nil {
		return fmt.Errorf("failed to delete stream: %w", err)
	}

	return nil
}

// GetStream returns information about a stream
func (c *Client) GetStream(ctx context.Context, name string) (*nats.StreamInfo, error) {
	js, err := c.getJetStream()
	if err != nil {
		return nil, err
	}

	info, err := js.StreamInfo(name, nats.Context(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to get stream info: %w", err)
	}

	return info, nil
}

// ListStreams returns a list of all streams
func (c *Client) ListStreams(ctx context.Context) ([]*nats.StreamInfo, error) {
	js, err := c.getJetStream()
	if err != nil {
		return nil, err
	}

	var streams []*nats.StreamInfo
	for info := range js.Streams(nats.Context(ctx)) {
		streams = append(streams, info)
	}

	return streams, nil
}

// StreamNames returns a list of stream names
func (c *Client) StreamNames(ctx context.Context) ([]string, error) {
	js, err := c.getJetStream()
	if err != nil {
		return nil, err
	}

	var names []string
	for name := range js.StreamNames(nats.Context(ctx)) {
		names = append(names, name)
	}

	return names, nil
}

// PurgeStream removes all messages from a stream
func (c *Client) PurgeStream(ctx context.Context, name string) error {
	js, err := c.getJetStream()
	if err != nil {
		return err
	}

	if err := js.PurgeStream(name, nats.Context(ctx)); err != nil {
		return fmt.Errorf("failed to purge stream: %w", err)
	}

	return nil
}

// ============================================================================
// Consumer Management
// ============================================================================

// CreateConsumer creates a new consumer on a stream
func (c *Client) CreateConsumer(ctx context.Context, stream string, config *ConsumerConfig) (*nats.ConsumerInfo, error) {
	js, err := c.getJetStream()
	if err != nil {
		return nil, err
	}

	info, err := js.AddConsumer(stream, config.toNatsConsumerConfig(), nats.Context(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer: %w", err)
	}

	return info, nil
}

// UpdateConsumer updates an existing consumer
func (c *Client) UpdateConsumer(ctx context.Context, stream string, config *ConsumerConfig) (*nats.ConsumerInfo, error) {
	js, err := c.getJetStream()
	if err != nil {
		return nil, err
	}

	info, err := js.UpdateConsumer(stream, config.toNatsConsumerConfig(), nats.Context(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to update consumer: %w", err)
	}

	return info, nil
}

// DeleteConsumer deletes a consumer from a stream
func (c *Client) DeleteConsumer(ctx context.Context, stream, consumer string) error {
	js, err := c.getJetStream()
	if err != nil {
		return err
	}

	if err := js.DeleteConsumer(stream, consumer, nats.Context(ctx)); err != nil {
		return fmt.Errorf("failed to delete consumer: %w", err)
	}

	return nil
}

// GetConsumer returns information about a consumer
func (c *Client) GetConsumer(ctx context.Context, stream, consumer string) (*nats.ConsumerInfo, error) {
	js, err := c.getJetStream()
	if err != nil {
		return nil, err
	}

	info, err := js.ConsumerInfo(stream, consumer, nats.Context(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to get consumer info: %w", err)
	}

	return info, nil
}

// ListConsumers returns a list of all consumers for a stream
func (c *Client) ListConsumers(ctx context.Context, stream string) ([]*nats.ConsumerInfo, error) {
	js, err := c.getJetStream()
	if err != nil {
		return nil, err
	}

	var consumers []*nats.ConsumerInfo
	for info := range js.Consumers(stream, nats.Context(ctx)) {
		consumers = append(consumers, info)
	}

	return consumers, nil
}

// ConsumerNames returns a list of consumer names for a stream
func (c *Client) ConsumerNames(ctx context.Context, stream string) ([]string, error) {
	js, err := c.getJetStream()
	if err != nil {
		return nil, err
	}

	var names []string
	for name := range js.ConsumerNames(stream, nats.Context(ctx)) {
		names = append(names, name)
	}

	return names, nil
}

// ============================================================================
// JetStream Publish
// ============================================================================

// PublishToStream publishes a message to a subject and waits for acknowledgement
func (c *Client) PublishToStream(ctx context.Context, subject string, data []byte) (*nats.PubAck, error) {
	js, err := c.getJetStream()
	if err != nil {
		return nil, err
	}

	ack, err := js.Publish(subject, data, nats.Context(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to publish to stream: %w", err)
	}

	return ack, nil
}

// PublishToStreamAsync publishes a message asynchronously
func (c *Client) PublishToStreamAsync(subject string, data []byte) (nats.PubAckFuture, error) {
	js, err := c.getJetStream()
	if err != nil {
		return nil, err
	}

	future, err := js.PublishAsync(subject, data)
	if err != nil {
		return nil, fmt.Errorf("failed to publish async: %w", err)
	}

	return future, nil
}

// PublishMsgToStream publishes a NATS message to a stream
func (c *Client) PublishMsgToStream(ctx context.Context, msg *nats.Msg) (*nats.PubAck, error) {
	js, err := c.getJetStream()
	if err != nil {
		return nil, err
	}

	ack, err := js.PublishMsg(msg, nats.Context(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to publish message to stream: %w", err)
	}

	return ack, nil
}

// ============================================================================
// JetStream Subscribe
// ============================================================================

// PullSubscribe creates a pull-based subscription
func (c *Client) PullSubscribe(subject, durable string, opts ...nats.SubOpt) (*nats.Subscription, error) {
	js, err := c.getJetStream()
	if err != nil {
		return nil, err
	}

	sub, err := js.PullSubscribe(subject, durable, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create pull subscription: %w", err)
	}

	return sub, nil
}

// SubscribeToStream creates a push-based subscription to a stream
func (c *Client) SubscribeToStream(subject string, handler nats.MsgHandler, opts ...nats.SubOpt) (*nats.Subscription, error) {
	js, err := c.getJetStream()
	if err != nil {
		return nil, err
	}

	sub, err := js.Subscribe(subject, handler, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to stream: %w", err)
	}

	return sub, nil
}

// QueueSubscribeToStream creates a queue group subscription to a stream
func (c *Client) QueueSubscribeToStream(subject, queue string, handler nats.MsgHandler, opts ...nats.SubOpt) (*nats.Subscription, error) {
	js, err := c.getJetStream()
	if err != nil {
		return nil, err
	}

	sub, err := js.QueueSubscribe(subject, queue, handler, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to queue subscribe to stream: %w", err)
	}

	return sub, nil
}

// ChanSubscribeToStream creates a channel-based subscription to a stream
func (c *Client) ChanSubscribeToStream(subject string, ch chan *nats.Msg, opts ...nats.SubOpt) (*nats.Subscription, error) {
	js, err := c.getJetStream()
	if err != nil {
		return nil, err
	}

	sub, err := js.ChanSubscribe(subject, ch, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to chan subscribe to stream: %w", err)
	}

	return sub, nil
}

// ============================================================================
// Message Operations
// ============================================================================

// GetMsg retrieves a message from a stream by sequence number
func (c *Client) GetMsg(ctx context.Context, stream string, seq uint64) (*nats.RawStreamMsg, error) {
	js, err := c.getJetStream()
	if err != nil {
		return nil, err
	}

	msg, err := js.GetMsg(stream, seq, nats.Context(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to get message: %w", err)
	}

	return msg, nil
}

// GetLastMsg retrieves the last message for a subject
func (c *Client) GetLastMsg(ctx context.Context, stream, subject string) (*nats.RawStreamMsg, error) {
	js, err := c.getJetStream()
	if err != nil {
		return nil, err
	}

	msg, err := js.GetLastMsg(stream, subject, nats.Context(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to get last message: %w", err)
	}

	return msg, nil
}

// DeleteMsg deletes a message from a stream by sequence number
func (c *Client) DeleteMsg(ctx context.Context, stream string, seq uint64) error {
	js, err := c.getJetStream()
	if err != nil {
		return err
	}

	if err := js.DeleteMsg(stream, seq, nats.Context(ctx)); err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}

	return nil
}

// SecureDeleteMsg securely deletes a message (overwrites with random data)
func (c *Client) SecureDeleteMsg(ctx context.Context, stream string, seq uint64) error {
	js, err := c.getJetStream()
	if err != nil {
		return err
	}

	if err := js.SecureDeleteMsg(stream, seq, nats.Context(ctx)); err != nil {
		return fmt.Errorf("failed to secure delete message: %w", err)
	}

	return nil
}

// ============================================================================
// Helper Methods
// ============================================================================

// getJetStream returns the JetStream context with proper error handling
func (c *Client) getJetStream() (nats.JetStreamContext, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return nil, ErrClientClosed
	}

	if !c.config.EnableJetStream {
		return nil, ErrJetStreamNotEnabled
	}

	if c.js == nil {
		return nil, ErrJetStreamNotInitialized
	}

	return c.js, nil
}

// PublishAckFutureWait waits for all pending async publishes
func (c *Client) PublishAckFutureWait() <-chan struct{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.js == nil {
		ch := make(chan struct{})
		close(ch)
		return ch
	}

	return c.js.PublishAsyncComplete()
}
