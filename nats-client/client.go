package natsclient

import (
	"context"
	"fmt"
	"sync"

	"github.com/nats-io/nats.go"
)

// Client represents the NATS client wrapper
type Client struct {
	config *Config
	conn   *nats.Conn
	js     nats.JetStreamContext
	mu     sync.RWMutex
	closed bool
}

// New creates a new client with the given configuration
func New(config *Config) (*Client, error) {
	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	client := &Client{
		config: config,
		closed: false,
	}

	// Initialize your client here
	if err := client.connect(); err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

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

// connect establishes the connection to NATS server
func (c *Client) connect() error {
	opts := []nats.Option{
		nats.Name(c.config.Name),
		nats.MaxReconnects(c.config.MaxReconnects),
		nats.ReconnectWait(c.config.ReconnectWait),
		nats.ReconnectJitter(c.config.ReconnectJitter, c.config.ReconnectJitter),
		nats.Timeout(c.config.Timeout),
		nats.PingInterval(c.config.PingInterval),
		nats.MaxPingsOutstanding(c.config.MaxPingsOut),
	}

	// Authentication
	if c.config.Username != "" && c.config.Password != "" {
		opts = append(opts, nats.UserInfo(c.config.Username, c.config.Password))
	} else if c.config.Token != "" {
		opts = append(opts, nats.Token(c.config.Token))
	}

	// Connection options
	if !c.config.AllowReconnect {
		opts = append(opts, nats.NoReconnect())
	}
	if c.config.NoRandomize {
		opts = append(opts, nats.DontRandomize())
	}
	if c.config.NoEcho {
		opts = append(opts, nats.NoEcho())
	}
	if c.config.RetryOnFailedConn {
		opts = append(opts, nats.RetryOnFailedConnect(true))
	}

	// TLS configuration
	if c.config.TLSEnabled {
		// Basic TLS without client certificates
		opts = append(opts, nats.Secure())

		// If client certificates are provided
		if c.config.TLSCertFile != "" && c.config.TLSKeyFile != "" {
			opts = append(opts, nats.ClientCert(c.config.TLSCertFile, c.config.TLSKeyFile))
		}

		// If CA certificate is provided
		if c.config.TLSCAFile != "" {
			opts = append(opts, nats.RootCAs(c.config.TLSCAFile))
		}
	}

	// Connect to NATS
	conn, err := nats.Connect(c.config.URL, opts...)
	if err != nil {
		return fmt.Errorf("failed to connect to NATS: %w", err)
	}

	c.conn = conn

	// Initialize JetStream if enabled
	if c.config.EnableJetStream {
		js, err := conn.JetStream()
		if err != nil {
			conn.Close()
			return fmt.Errorf("failed to initialize JetStream: %w", err)
		}
		c.js = js
	}

	return nil
}

// Close closes the client and releases resources
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return ErrAlreadyClosed
	}

	if c.conn != nil {
		c.conn.Close()
	}

	c.closed = true
	return nil
}

// IsClosed returns true if the connection is closed
func (c *Client) IsClosed() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.closed
}

// IsConnected returns true if connected to NATS server
func (c *Client) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed || c.conn == nil {
		return false
	}

	return c.conn.IsConnected()
}

// IsReconnecting returns true if currently reconnecting
func (c *Client) IsReconnecting() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed || c.conn == nil {
		return false
	}

	return c.conn.IsReconnecting()
}

// Stats returns connection statistics
func (c *Client) Stats() nats.Statistics {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.conn == nil {
		return nats.Statistics{}
	}

	return c.conn.Stats()
}

// Publish publishes a message to a subject
func (c *Client) Publish(subject string, data []byte) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return ErrClientClosed
	}

	if c.conn == nil {
		return ErrConnectionFailed
	}

	return c.conn.Publish(subject, data)
}

// Subscribe creates a subscription to a subject
func (c *Client) Subscribe(subject string, handler nats.MsgHandler) (*nats.Subscription, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return nil, ErrClientClosed
	}

	if c.conn == nil {
		return nil, ErrConnectionFailed
	}

	return c.conn.Subscribe(subject, handler)
}

// QueueSubscribe creates a queue group subscription
func (c *Client) QueueSubscribe(subject, queue string, handler nats.MsgHandler) (*nats.Subscription, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return nil, ErrClientClosed
	}

	if c.conn == nil {
		return nil, ErrConnectionFailed
	}

	return c.conn.QueueSubscribe(subject, queue, handler)
}

// Request sends a request and waits for a response
func (c *Client) Request(subject string, data []byte, timeout context.Context) (*nats.Msg, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return nil, ErrClientClosed
	}

	if c.conn == nil {
		return nil, ErrConnectionFailed
	}

	// Get timeout duration from context
	_, ok := timeout.Deadline()
	if !ok {
		// No deadline, use config timeout
		return c.conn.Request(subject, data, c.config.Timeout)
	}

	return c.conn.RequestWithContext(timeout, subject, data)
}

// Flush flushes any buffered messages
func (c *Client) Flush() error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return ErrClientClosed
	}

	if c.conn == nil {
		return ErrConnectionFailed
	}

	return c.conn.Flush()
}

// JetStream returns the JetStream context
func (c *Client) JetStream() (nats.JetStreamContext, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return nil, ErrClientClosed
	}

	if !c.config.EnableJetStream {
		return nil, fmt.Errorf("JetStream not enabled in configuration")
	}

	if c.js == nil {
		return nil, fmt.Errorf("JetStream context not initialized")
	}

	return c.js, nil
}

// Conn returns the underlying NATS connection
func (c *Client) Conn() *nats.Conn {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.conn
}
